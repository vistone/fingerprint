package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type tlsEntry struct {
	Browser    string
	Version    string
	OS         string
	Ciphers    []uint16
	Extensions []uint16
	Curves     []uint16
	ALPN       []string
}

func collectJA3er(ctx context.Context, client *http.Client) ([]CollectedFingerprint, error) {
	// JA3er public API: top user agents with their JA3 hashes
	urls := []string{
		"https://ja3er.com/getAllUasJson",
	}

	var all []CollectedFingerprint
	for _, u := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "fingerprint-collector/1.0")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("  JA3er API unavailable: %v\n", err)
			fmt.Println("  Falling back to builtin JA3 knowledge base...")
			return collectJA3BuiltinKnowledge(), nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("  JA3er returned %d, using builtin knowledge\n", resp.StatusCode)
			return collectJA3BuiltinKnowledge(), nil
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // 10MB limit
		if err != nil {
			return nil, err
		}

		var entries []struct {
			UserAgent string `json:"User-Agent"`
			JA3Hash   string `json:"ja3_hash"`
			Count     int    `json:"Count"`
		}
		if err := json.Unmarshal(body, &entries); err != nil {
			fmt.Printf("  JA3er parse error: %v, using builtin\n", err)
			return collectJA3BuiltinKnowledge(), nil
		}

		for _, e := range entries {
			fp := CollectedFingerprint{
				Source:      "ja3er",
				CollectedAt: time.Now().UTC().Format(time.RFC3339),
				UserAgent:   e.UserAgent,
				JA3Hash:     e.JA3Hash,
				Metadata:    map[string]any{"count": e.Count},
			}
			parseBrowserFromUA(e.UserAgent, &fp)
			all = append(all, fp)
		}
	}
	return all, nil
}

// collectPeetJS fetches TLS fingerprint data from peet.ws.
func collectPeetJS(ctx context.Context, client *http.Client) ([]CollectedFingerprint, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://tls.peet.ws/api/all", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "fingerprint-collector/1.0")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  Peet.ws API unavailable: %v\n", err)
		fmt.Println("  Falling back to builtin TLS knowledge base...")
		return collectTLSBuiltinKnowledge(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("  Peet.ws returned %d, using builtin knowledge\n", resp.StatusCode)
		return collectTLSBuiltinKnowledge(), nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, err
	}

	var entries []struct {
		Browser      string   `json:"browser"`
		Version      string   `json:"version"`
		OS           string   `json:"os"`
		CipherSuites []uint16 `json:"cipher_suites"`
		Extensions   []uint16 `json:"extensions"`
		Curves       []uint16 `json:"supported_groups"`
		JA3          string   `json:"ja3"`
		JA4          string   `json:"ja4"`
	}

	if err := json.Unmarshal(body, &entries); err != nil {
		fmt.Printf("  Peet.ws parse error: %v, using builtin\n", err)
		return collectTLSBuiltinKnowledge(), nil
	}

	var all []CollectedFingerprint
	for _, e := range entries {
		fp := CollectedFingerprint{
			Source:      "tls_peet",
			CollectedAt: time.Now().UTC().Format(time.RFC3339),
			BrowserName: e.Browser,
			BrowserVer:  e.Version,
			OSName:      e.OS,
			JA3Hash:     e.JA3,
			JA4:         e.JA4,
			TLS: &TLSFingerprint{
				Version:         0x0303,
				CipherSuites:    e.CipherSuites,
				Extensions:      e.Extensions,
				SupportedGroups: e.Curves,
			},
		}
		all = append(all, fp)
	}
	return all, nil
}

// parseBrowserFromUA extracts browser info from User-Agent string.
func parseBrowserFromUA(ua string, fp *CollectedFingerprint) {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "edg/"):
		fp.BrowserName = "edge"
	case strings.Contains(ua, "opr/") || strings.Contains(ua, "opera"):
		fp.BrowserName = "opera"
	case strings.Contains(ua, "brave"):
		fp.BrowserName = "brave"
	case strings.Contains(ua, "yabrowser"):
		fp.BrowserName = "yandex"
	case strings.Contains(ua, "ucbrowser"):
		fp.BrowserName = "uc_browser"
	case strings.Contains(ua, "qqbrowser"):
		fp.BrowserName = "qq_browser"
	case strings.Contains(ua, "samsungbrowser"):
		fp.BrowserName = "samsung"
	case strings.Contains(ua, "vivaldi"):
		fp.BrowserName = "vivaldi"
	case strings.Contains(ua, "whale"):
		fp.BrowserName = "whale"
	case strings.Contains(ua, "firefox"):
		fp.BrowserName = "firefox"
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		fp.BrowserName = "safari"
	case strings.Contains(ua, "chrome"):
		fp.BrowserName = "chrome"
	}

	if strings.Contains(ua, "android") {
		fp.Platform = "mobile"
	} else if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		fp.Platform = "mobile"
	} else {
		fp.Platform = "desktop"
	}
}

// =========================================================================
// Built-in knowledge bases — known fingerprints from public research
// =========================================================================

// collectBuiltinKnowledge returns a comprehensive set of browser fingerprints
// based on published research, JA3/JA4 databases, and browser documentation.
func collectBuiltinKnowledge() []CollectedFingerprint {
	var all []CollectedFingerprint
	all = append(all, collectJA3BuiltinKnowledge()...)
	all = append(all, collectTLSBuiltinKnowledge()...)
	all = append(all, collectBotFingerprints()...)
	all = append(all, collectInternationalBrowsers()...)
	return all
}

// collectJA3BuiltinKnowledge returns known JA3 hashes for major browsers.
// Source: public JA3 fingerprint database, Salesforce research, Censys data.
func collectJA3BuiltinKnowledge() []CollectedFingerprint {
	// Top JA3 fingerprints observed globally (from public research)
	type ja3Entry struct {
		Hash    string
		Browser string
		Version string
		OS      string
	}

	known := []ja3Entry{
		// Chrome
		{Hash: "cd08e31494f9531f560d64c695473da9", Browser: "chrome", Version: "120-130", OS: "Windows"},
		{Hash: "b32309a26951912be7dba376398abc3b", Browser: "chrome", Version: "110-119", OS: "Windows"},
		{Hash: "a0e9f5d64349fb13191bc781f81f42e1", Browser: "chrome", Version: "100-109", OS: "macOS"},
		{Hash: "3b5074b1b5d032e5620f69f9f700ff0e", Browser: "chrome", Version: "120+", OS: "Linux"},
		// Firefox
		{Hash: "839bbe3ed07fed922ded5aaf714d6842", Browser: "firefox", Version: "120+", OS: "Windows"},
		{Hash: "e058f2d45805cdf9e8e25af97b4a9a66", Browser: "firefox", Version: "115-119", OS: "macOS"},
		{Hash: "457e3a2e7a3bde9e62c0e48a41a43617", Browser: "firefox", Version: "100-114", OS: "Linux"},
		// Safari
		{Hash: "773906b0efdefa24a7f2b8eb6985bf37", Browser: "safari", Version: "17+", OS: "macOS"},
		{Hash: "7c02dbae662670edcf60231987fba071", Browser: "safari", Version: "16", OS: "macOS"},
		{Hash: "8a8a1c52f8d3e5d2a7fca8f08e52d3dd", Browser: "safari", Version: "17+", OS: "iOS"},
		// Edge
		{Hash: "cd08e31494f9531f560d64c695473da9", Browser: "edge", Version: "120+", OS: "Windows"},
		{Hash: "b32309a26951912be7dba376398abc3b", Browser: "edge", Version: "110-119", OS: "Windows"},
		// Brave
		{Hash: "cd08e31494f9531f560d64c695473da9", Browser: "brave", Version: "1.60+", OS: "Windows"},
		// Opera
		{Hash: "cd08e31494f9531f560d64c695473da9", Browser: "opera", Version: "100+", OS: "Windows"},
	}

	fps := make([]CollectedFingerprint, 0, len(known))
	for _, k := range known {
		fps = append(fps, CollectedFingerprint{
			Source:      "ja3_knowledge",
			CollectedAt: time.Now().UTC().Format(time.RFC3339),
			BrowserName: k.Browser,
			BrowserVer:  k.Version,
			OSName:      k.OS,
			JA3Hash:     k.Hash,
		})
	}
	return fps
}

// collectTLSBuiltinKnowledge returns detailed TLS fingerprints from known browsers.
func collectTLSBuiltinKnowledge() []CollectedFingerprint {
	collectedAt := time.Now().UTC().Format(time.RFC3339)
	fps := make([]CollectedFingerprint, 0, len(tlsBuiltinKnowledge))
	for _, k := range tlsBuiltinKnowledge {
		fps = append(fps, buildTLSKnowledgeFingerprint(k, collectedAt))
	}
	return fps
}

func buildTLSKnowledgeFingerprint(k tlsEntry, collectedAt string) CollectedFingerprint {
	return CollectedFingerprint{
		Source:      "tls_knowledge",
		CollectedAt: collectedAt,
		BrowserName: k.Browser,
		BrowserVer:  k.Version,
		OSName:      k.OS,
		TLS: &TLSFingerprint{
			Version:         0x0303,
			CipherSuites:    k.Ciphers,
			Extensions:      k.Extensions,
			SupportedGroups: k.Curves,
			ALPNProtocols:   k.ALPN,
			SNI:             true,
		},
	}
}

var tlsBuiltinKnowledge = []tlsEntry{
	// Chrome 120+ (Windows/macOS/Linux)
	{
		Browser: "chrome", Version: "130", OS: "Windows",
		Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc00a, 0xc009, 0xc013, 0xc014, 0x002f, 0x0035, 0x000a},
		Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0010, 0x0005, 0x000d, 0x0012, 0x002b, 0x002d, 0x001c, 0x001b, 0x0033, 0x0015},
		Curves:     []uint16{0x001d, 0x0017, 0x0018, 0x0019},
		ALPN:       []string{"h2", "http/1.1"},
	},
	// Chrome 130+ (macOS)
	{
		Browser: "chrome", Version: "130", OS: "macOS",
		Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc00a, 0xc009, 0xc013, 0xc014, 0x002f, 0x0035},
		Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0010, 0x0005, 0x000d, 0x002b, 0x002d, 0x0033, 0x001b, 0x0015},
		Curves:     []uint16{0x001d, 0x0017, 0x0018},
		ALPN:       []string{"h2", "http/1.1"},
	},
	// Firefox 130+ (Windows)
	{
		Browser: "firefox", Version: "130", OS: "Windows",
		Ciphers:    []uint16{0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc00a, 0xc009, 0xc013, 0xc014},
		Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0010, 0x0005, 0x000d, 0x002b, 0x002d, 0x0033, 0x001c},
		Curves:     []uint16{0x001d, 0x0017, 0x0018, 0x0100},
		ALPN:       []string{"h2", "http/1.1"},
	},
	// Firefox 130+ (macOS)
	{
		Browser: "firefox", Version: "130", OS: "macOS",
		Ciphers:    []uint16{0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030},
		Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0010, 0x0005, 0x000d, 0x002b, 0x002d, 0x0033},
		Curves:     []uint16{0x001d, 0x0017, 0x0018, 0x0100},
		ALPN:       []string{"h2", "http/1.1"},
	},
	// Safari 17+ (macOS)
	{
		Browser: "safari", Version: "17", OS: "macOS",
		Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02c, 0xc02b, 0xc030, 0xc02f, 0xcca9, 0xcca8, 0xc00a, 0xc009, 0xc014, 0xc013, 0x002f, 0x0035},
		Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0010, 0x0005, 0x000d, 0x002b, 0x002d, 0x0033},
		Curves:     []uint16{0x001d, 0x0017, 0x0018},
		ALPN:       []string{"h2", "http/1.1"},
	},
	// Safari 17+ (iOS)
	{
		Browser: "safari", Version: "17", OS: "iOS",
		Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02c, 0xc02b, 0xc030, 0xc02f, 0xcca9, 0xcca8, 0xc00a, 0xc009, 0xc014, 0xc013},
		Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0010, 0x0005, 0x000d, 0x002b, 0x002d, 0x0033},
		Curves:     []uint16{0x001d, 0x0017, 0x0018},
		ALPN:       []string{"h2", "http/1.1"},
	},
}

// collectBotFingerprints returns known fingerprints of headless browsers, bots, and tools.
// These are critical for training the forgery detector.
