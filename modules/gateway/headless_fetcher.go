package gateway

import (
	"context"
	"fmt"
	htmlpkg "html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// ExternalScriptStats contains external script statistics
type ExternalScriptStats struct {
	Count       int            `json:"count"`
	URLs        []string       `json:"urls"`
	DomainCount map[string]int `json:"domainCount"`
}

// headlessFetchOptions groups configuration for headless browser fetching.
type headlessFetchOptions struct {
	TargetURL    string
	BrowserWS    string
	MaxRedirects int
	WaitMs       int
	Timeout      time.Duration
}

type scriptCaptureState struct {
	mu             sync.Mutex
	requestURLs    map[network.RequestID]string
	scripts        []string
	urls           []string
	capturedBytes  int
	maxScripts     int
	maxScriptBytes int
}

// fetchHTMLWithHeadlessBrowser uses remote headless Chrome to execute scripts and return final DOM.
func fetchHTMLWithHeadlessBrowser(ctx context.Context, opts headlessFetchOptions) (string, string, []string, *ExternalScriptStats, error) {
	opts, err := normalizeHeadlessOptions(opts)
	if err != nil {
		return "", "", nil, nil, err
	}

	runCtx, browserCtx, cancelAll := newHeadlessBrowserContext(ctx, opts)
	defer cancelAll()

	capture := newScriptCaptureState(10, 1024*1024)
	attachScriptCaptureListener(browserCtx, capture)

	finalURL, pageHTML, redirectChain, err := runHeadlessDOMFetch(browserCtx, opts)
	if err != nil {
		return "", "", redirectChain, nil, err
	}

	pageHTML, allURLs := mergeHeadlessAndDOMScripts(runCtx, finalURL, pageHTML, capture)
	stats := buildExternalScriptStats(allURLs)
	return pageHTML, finalURL, redirectChain, stats, nil
}

func normalizeHeadlessOptions(opts headlessFetchOptions) (headlessFetchOptions, error) {
	if strings.TrimSpace(opts.TargetURL) == "" {
		return opts, fmt.Errorf("empty target url")
	}
	if strings.TrimSpace(opts.BrowserWS) == "" {
		return opts, fmt.Errorf("empty browser websocket endpoint")
	}
	if opts.WaitMs <= 0 {
		opts.WaitMs = 1200
	}
	if opts.WaitMs > 8000 {
		opts.WaitMs = 8000
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 25 * time.Second
	}
	if opts.MaxRedirects <= 0 {
		opts.MaxRedirects = 10
	}
	return opts, nil
}

func newHeadlessBrowserContext(parent context.Context, opts headlessFetchOptions) (context.Context, context.Context, context.CancelFunc) {
	runCtx, cancelRun := context.WithTimeout(parent, opts.Timeout)
	allocCtx, cancelAlloc := chromedp.NewRemoteAllocator(runCtx, opts.BrowserWS)
	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)

	cancelAll := func() {
		cancelBrowser()
		cancelAlloc()
		cancelRun()
	}
	return runCtx, browserCtx, cancelAll
}

func newScriptCaptureState(maxScripts, maxBytes int) *scriptCaptureState {
	return &scriptCaptureState{
		requestURLs:    map[network.RequestID]string{},
		scripts:        []string{},
		urls:           []string{},
		maxScripts:     maxScripts,
		maxScriptBytes: maxBytes,
	}
}

func attachScriptCaptureListener(browserCtx context.Context, capture *scriptCaptureState) {
	chromedp.ListenTarget(browserCtx, func(ev interface{}) {
		switch e := ev.(type) {
		case *network.EventResponseReceived:
			capture.recordResponse(e)
		case *network.EventLoadingFinished:
			capture.recordLoadedScript(browserCtx, e)
		}
	})
}

func (s *scriptCaptureState) recordResponse(event *network.EventResponseReceived) {
	if event.Type != network.ResourceTypeScript {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requestURLs[event.RequestID] = event.Response.URL
}

func (s *scriptCaptureState) recordLoadedScript(browserCtx context.Context, event *network.EventLoadingFinished) {
	srcURL, ok := s.takeRequestURL(event.RequestID)
	if !ok {
		return
	}

	bodyBytes, err := network.GetResponseBody(event.RequestID).Do(browserCtx)
	if err != nil {
		return
	}
	body := strings.TrimSpace(string(bodyBytes))
	if body == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.scripts) >= s.maxScripts || s.capturedBytes >= s.maxScriptBytes {
		return
	}
	remaining := s.maxScriptBytes - s.capturedBytes
	if len(body) > remaining {
		body = body[:remaining]
	}
	s.scripts = append(s.scripts, body)
	s.urls = append(s.urls, srcURL)
	s.capturedBytes += len(body)
}

func (s *scriptCaptureState) takeRequestURL(id network.RequestID) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	srcURL, ok := s.requestURLs[id]
	if !ok || len(s.scripts) >= s.maxScripts || s.capturedBytes >= s.maxScriptBytes {
		return "", false
	}
	delete(s.requestURLs, id)
	return srcURL, true
}

func runHeadlessDOMFetch(browserCtx context.Context, opts headlessFetchOptions) (string, string, []string, error) {
	finalURL := ""
	pageHTML := ""
	redirectChain := []string{opts.TargetURL}

	err := chromedp.Run(browserCtx,
		network.Enable(),
		chromedp.Navigate(opts.TargetURL),
		chromedp.Sleep(time.Duration(opts.WaitMs)*time.Millisecond),
		chromedp.Location(&finalURL),
		chromedp.OuterHTML("html", &pageHTML, chromedp.ByQuery),
	)
	if err != nil {
		return "", "", redirectChain, fmt.Errorf("headless browser fetch failed: %w", err)
	}

	if strings.TrimSpace(finalURL) == "" {
		finalURL = opts.TargetURL
	}
	if finalURL != opts.TargetURL {
		redirectChain = append(redirectChain, finalURL)
	}
	if strings.TrimSpace(pageHTML) == "" {
		return "", finalURL, redirectChain, fmt.Errorf("empty html from headless browser")
	}

	return finalURL, pageHTML, redirectChain, nil
}

func mergeHeadlessAndDOMScripts(runCtx context.Context, finalURL, pageHTML string, capture *scriptCaptureState) (string, []string) {
	netURLs, netScripts, usedBytes := capture.snapshot()
	remainingScripts := capture.maxScripts - len(netScripts)
	remainingBytes := capture.maxScriptBytes - usedBytes
	domURLs, domScripts := fetchExternalScriptsByDOM(runCtx, finalURL, pageHTML, remainingScripts, remainingBytes)

	allURLs := append(netURLs, domURLs...)
	allScripts := append(netScripts, domScripts...)
	if len(allScripts) > 0 {
		pageHTML = appendCapturedScriptsToHTML(pageHTML, allURLs, allScripts)
	}
	return pageHTML, allURLs
}

func (s *scriptCaptureState) snapshot() ([]string, []string, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	urls := append([]string(nil), s.urls...)
	scripts := append([]string(nil), s.scripts...)
	return urls, scripts, s.capturedBytes
}

func appendCapturedScriptsToHTML(pageHTML string, urls []string, scripts []string) string {
	if len(scripts) == 0 {
		return pageHTML
	}
	var b strings.Builder
	b.WriteString(pageHTML)
	b.WriteString("\n<!-- external-js-captured-by-headless -->\n")
	for i := range scripts {
		b.WriteString("<script data-captured-src=\"")
		if i < len(urls) {
			b.WriteString(htmlpkg.EscapeString(urls[i]))
		}
		b.WriteString("\">\n")
		b.WriteString(scripts[i])
		b.WriteString("\n</script>\n")
	}
	return b.String()
}

func fetchExternalScriptsByDOM(ctx context.Context, baseURL, pageHTML string, maxScripts, maxBytes int) ([]string, []string) {
	if maxScripts <= 0 || maxBytes <= 0 {
		return nil, nil
	}
	if maxScripts > 6 {
		maxScripts = 6
	}

	srcs := extractScriptSrcsFromHTML(pageHTML)
	if len(srcs) == 0 {
		return nil, nil
	}

	resolved := make([]string, 0, len(srcs))
	seen := map[string]bool{}
	for _, src := range srcs {
		u := resolveScriptURL(baseURL, src)
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		resolved = append(resolved, u)
	}

	client := &http.Client{Timeout: 2 * time.Second}
	urls := []string{}
	scripts := []string{}
	used := 0
	deadline := time.Now().Add(3 * time.Second)
	for _, scriptURL := range resolved {
		if time.Now().After(deadline) {
			break
		}
		if len(scripts) >= maxScripts || used >= maxBytes {
			break
		}
		remain := maxBytes - used
		if remain <= 0 {
			break
		}
		if err := ValidateOutboundTarget(scriptURL, false); err != nil {
			continue
		}

		body, ok := fetchScriptBody(ctx, scriptFetchParams{
			Client:     client,
			ScriptURL:  scriptURL,
			PageURL:    baseURL,
			MaxBytes:   remain,
			ReqTimeout: 1500 * time.Millisecond,
		})
		if !ok {
			continue
		}
		urls = append(urls, scriptURL)
		scripts = append(scripts, body)
		used += len(body)
	}
	return urls, scripts
}

// scriptFetchParams groups parameters for fetchScriptBody.
type scriptFetchParams struct {
	Client     *http.Client
	ScriptURL  string
	PageURL    string
	MaxBytes   int
	ReqTimeout time.Duration
}

func fetchScriptBody(ctx context.Context, p scriptFetchParams) (string, bool) {
	if p.ReqTimeout <= 0 {
		p.ReqTimeout = 1500 * time.Millisecond
	}
	reqCtx, cancel := context.WithTimeout(ctx, p.ReqTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, p.ScriptURL, nil)
	if err != nil {
		return "", false
	}

	referer := strings.TrimSpace(p.PageURL)
	origin := ""
	if parsedPage, err := url.Parse(referer); err == nil && parsedPage.Scheme != "" && parsedPage.Host != "" {
		origin = parsedPage.Scheme + "://" + parsedPage.Host
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/javascript,text/javascript,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	// Simulate browser script-tag fetch metadata for cross-site JS endpoints.
	req.Header.Set("Sec-Fetch-Dest", "script")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	resp, err := p.Client.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", false
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, int64(p.MaxBytes)))
	if err != nil {
		return "", false
	}
	body := string(bodyBytes)
	if strings.TrimSpace(body) == "" {
		return "", false
	}
	return body, true
}

func extractScriptSrcsFromHTML(pageHTML string) []string {
	re := regexp.MustCompile(`(?is)<script[^>]*\bsrc\s*=\s*['"]([^'"]+)['"][^>]*>`)
	matches := re.FindAllStringSubmatch(pageHTML, -1)
	srcs := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			s := strings.TrimSpace(m[1])
			if s != "" {
				srcs = append(srcs, s)
			}
		}
	}
	return srcs
}

func resolveScriptURL(baseURL, ref string) string {
	base, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return ""
	}
	r, err := url.Parse(strings.TrimSpace(ref))
	if err != nil {
		return ""
	}
	return base.ResolveReference(r).String()
}

// buildExternalScriptStats builds external script statistics (deduplicated + grouped by domain)
func buildExternalScriptStats(urls []string) *ExternalScriptStats {
	if len(urls) == 0 {
		return &ExternalScriptStats{
			Count:       0,
			URLs:        []string{},
			DomainCount: map[string]int{},
		}
	}

	seen := make(map[string]bool)
	uniqueURLs := []string{}
	for _, u := range urls {
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		uniqueURLs = append(uniqueURLs, u)
	}

	domainCount := make(map[string]int)
	for _, u := range uniqueURLs {
		parsed, err := url.Parse(u)
		if err != nil || parsed.Host == "" {
			continue
		}
		domainCount[parsed.Host]++
	}

	return &ExternalScriptStats{
		Count:       len(uniqueURLs),
		URLs:        uniqueURLs,
		DomainCount: domainCount,
	}
}
