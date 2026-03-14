package gateway

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// fetchHTMLWithRedirects fetches URL content and returns the final URL and redirect chain
func fetchHTMLWithRedirects(ctx context.Context, rawURL string, followRedirect bool, maxRedirects int, requestTimeout time.Duration) (string, string, []string, error) {
	redirectChain := []string{}
	trimmedURL := strings.TrimSpace(rawURL)
	if trimmedURL == "" {
		return "", "", redirectChain, fmt.Errorf("empty url")
	}
	requestTimeout = clampTimeout(requestTimeout, 3*time.Second, 20*time.Second, 12*time.Second)

	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt < 2; attempt++ {
		attemptChain := []string{}
		client := buildFetchHTTPClient(requestTimeout, followRedirect, maxRedirects, &attemptChain)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, trimmedURL, nil)
		if err != nil {
			return "", "", redirectChain, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

		resp, err = client.Do(req)
		if err == nil {
			redirectChain = attemptChain
			break
		}

		lastErr = err
		if !strings.Contains(strings.ToLower(err.Error()), "tls handshake timeout") || attempt == 1 {
			return "", "", redirectChain, err
		}
		// Retry once for transient TLS handshake stalls.
		time.Sleep(250 * time.Millisecond)
	}

	if resp == nil {
		if lastErr != nil {
			return "", "", redirectChain, lastErr
		}
		return "", "", redirectChain, fmt.Errorf("fetch failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 6*1024*1024))
	if err != nil {
		return "", "", redirectChain, err
	}

	return string(body), resp.Request.URL.String(), redirectChain, nil
}

func buildFetchHTTPClient(timeout time.Duration, followRedirect bool, maxRedirects int, chain *[]string) *http.Client {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   timeout,
			ResponseHeaderTimeout: 8 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			IdleConnTimeout:       30 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}
	if followRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			*chain = append(*chain, req.URL.String())
			if len(via) >= maxRedirects {
				return fmt.Errorf("too many redirects: %d", maxRedirects)
			}
			return nil
		}
	}
	return client
}

func clampTimeout(t, min, max, def time.Duration) time.Duration {
	if t <= 0 {
		return def
	}
	if t < min {
		return min
	}
	if t > max {
		return max
	}
	return t
}

// fetchHTMLWithClientSideRedirects fetches a page and emulates common client-side redirects.
// Supports meta refresh, window.location, location.href, location.replace redirect patterns.
func fetchHTMLWithClientSideRedirects(ctx context.Context, startURL string, maxHops int, waitMs int) (string, string, []string, error) {
	if maxHops <= 0 {
		maxHops = 5
	}
	if maxHops > 10 {
		maxHops = 10
	}

	currentURL := strings.TrimSpace(startURL)
	if currentURL == "" {
		return "", "", nil, fmt.Errorf("empty start url")
	}

	chain := []string{}
	visited := map[string]bool{}

	for hop := 0; hop < maxHops; hop++ {
		if visited[currentURL] {
			return "", "", chain, fmt.Errorf("redirect loop detected")
		}
		visited[currentURL] = true

		html, finalURL, httpChain, err := fetchHTMLWithRedirects(ctx, currentURL, true, maxHops, 8*time.Second)
		if err != nil {
			return "", "", chain, err
		}
		chain = append(chain, httpChain...)
		chain = append(chain, finalURL)

		// Simple wait, simulating initial page script execution window
		if waitMs > 0 {
			time.Sleep(time.Duration(waitMs) * time.Millisecond)
		}

		target := extractClientSideRedirectTarget(html)
		if target == "" {
			return html, finalURL, chain, nil
		}

		resolved, err := resolveRedirectURL(finalURL, target)
		if err != nil {
			return html, finalURL, chain, nil
		}

		currentURL = resolved
	}

	return "", currentURL, chain, fmt.Errorf("max client-side redirect hops reached")
}

func extractClientSideRedirectTarget(html string) string {
	if strings.TrimSpace(html) == "" {
		return ""
	}

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)<meta[^>]+http-equiv\s*=\s*['"]?refresh['"]?[^>]+content\s*=\s*['"][^;"']*;\s*url=([^"'>\s]+)`),
		regexp.MustCompile(`(?is)window\.location\.href\s*=\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?is)window\.location\s*=\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?is)location\.href\s*=\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?is)location\.replace\(\s*['"]([^'"]+)['"]\s*\)`),
	}

	for _, p := range patterns {
		m := p.FindStringSubmatch(html)
		if len(m) > 1 {
			return strings.TrimSpace(m[1])
		}
	}

	return ""
}

func resolveRedirectURL(baseURL, target string) (string, error) {
	baseParsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", err
	}
	targetParsed, err := url.Parse(strings.TrimSpace(target))
	if err != nil {
		return "", err
	}
	return baseParsed.ResolveReference(targetParsed).String(), nil
}
