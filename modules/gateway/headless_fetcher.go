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

// fetchHTMLWithHeadlessBrowser uses remote headless Chrome to execute scripts and return final DOM.
func fetchHTMLWithHeadlessBrowser(ctx context.Context, opts headlessFetchOptions) (string, string, []string, *ExternalScriptStats, error) {
	if strings.TrimSpace(opts.TargetURL) == "" {
		return "", "", nil, nil, fmt.Errorf("empty target url")
	}
	if strings.TrimSpace(opts.BrowserWS) == "" {
		return "", "", nil, nil, fmt.Errorf("empty browser websocket endpoint")
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

	runCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	allocCtx, cancelAlloc := chromedp.NewRemoteAllocator(runCtx, opts.BrowserWS)
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	var finalURL string
	var pageHTML string
	redirectChain := []string{opts.TargetURL}

	const maxCapturedScripts = 10
	const maxCapturedScriptBytes = 1024 * 1024

	var mu sync.Mutex
	scriptReqURLs := map[network.RequestID]string{}
	capturedScripts := []string{}
	capturedURLs := []string{}
	capturedBytes := 0

	chromedp.ListenTarget(browserCtx, func(ev interface{}) {
		switch e := ev.(type) {
		case *network.EventResponseReceived:
			if e.Type == network.ResourceTypeScript {
				mu.Lock()
				scriptReqURLs[e.RequestID] = e.Response.URL
				mu.Unlock()
			}
		case *network.EventLoadingFinished:
			mu.Lock()
			srcURL, ok := scriptReqURLs[e.RequestID]
			if !ok || len(capturedScripts) >= maxCapturedScripts || capturedBytes >= maxCapturedScriptBytes {
				mu.Unlock()
				return
			}
			mu.Unlock()

			bodyBytes, err := network.GetResponseBody(e.RequestID).Do(browserCtx)
			if err != nil {
				return
			}
			body := string(bodyBytes)
			if strings.TrimSpace(body) == "" {
				return
			}

			mu.Lock()
			if len(capturedScripts) < maxCapturedScripts && capturedBytes < maxCapturedScriptBytes {
				remaining := maxCapturedScriptBytes - capturedBytes
				if len(body) > remaining {
					body = body[:remaining]
				}
				capturedScripts = append(capturedScripts, body)
				capturedURLs = append(capturedURLs, srcURL)
				capturedBytes += len(body)
			}
			delete(scriptReqURLs, e.RequestID)
			mu.Unlock()
		}
	})

	err := chromedp.Run(browserCtx,
		network.Enable(),
		chromedp.Navigate(opts.TargetURL),
		chromedp.Sleep(time.Duration(opts.WaitMs)*time.Millisecond),
		chromedp.Location(&finalURL),
		chromedp.OuterHTML("html", &pageHTML, chromedp.ByQuery),
	)
	if err != nil {
		return "", "", redirectChain, nil, fmt.Errorf("headless browser fetch failed: %w", err)
	}

	if strings.TrimSpace(finalURL) == "" {
		finalURL = opts.TargetURL
	}
	if finalURL != opts.TargetURL {
		redirectChain = append(redirectChain, finalURL)
	}
	if strings.TrimSpace(pageHTML) == "" {
		return "", finalURL, redirectChain, nil, fmt.Errorf("empty html from headless browser")
	}

	mu.Lock()
	netScripts := append([]string(nil), capturedScripts...)
	netURLs := append([]string(nil), capturedURLs...)
	usedBytes := capturedBytes
	mu.Unlock()

	remainingScripts := maxCapturedScripts - len(netScripts)
	remainingBytes := maxCapturedScriptBytes - usedBytes
	domURLs, domScripts := fetchExternalScriptsByDOM(runCtx, finalURL, pageHTML, remainingScripts, remainingBytes)

	allURLs := append(netURLs, domURLs...)
	allScripts := append(netScripts, domScripts...)

	if len(allScripts) > 0 {
		pageHTML = appendCapturedScriptsToHTML(pageHTML, allURLs, allScripts)
	}

	stats := buildExternalScriptStats(allURLs)
	return pageHTML, finalURL, redirectChain, stats, nil
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
