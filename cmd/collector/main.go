// Command collector — fetches browser fingerprint data from multiple public sources
// and generates Go profile definitions.
//
// Data Sources:
//   - JA3/JA4 fingerprint databases (ja3er.com, ja4db.com via API)
//   - TLS ClientHello data (peetjs.com, tls.browserleaks.com)
//   - HTTP/2 fingerprint databases (Akamai http2fingerprint)
//   - Public browser fingerprint research datasets
//
// Usage:
//
//	go run ./cmd/collector/ [flags]
//
// Flags:
//
//	-output    Output directory for collected data (default: ./training/collected)
//	-format    Output format: json, go (default: json)
//	-sources   Comma-separated sources: ja3er,tls_peet,http2 (default: all)
//	-timeout   HTTP request timeout in seconds (default: 30)
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CollectedFingerprint represents a browser fingerprint from a public source.
type CollectedFingerprint struct {
	Source      string            `json:"source"`
	CollectedAt string            `json:"collected_at"`
	BrowserName string            `json:"browser_name"`
	BrowserVer  string            `json:"browser_version"`
	OSName      string            `json:"os_name"`
	OSVersion   string            `json:"os_version"`
	Platform    string            `json:"platform"` // desktop, mobile, tablet
	TLS         *TLSFingerprint   `json:"tls,omitempty"`
	HTTP2       *HTTP2Fingerprint `json:"http2,omitempty"`
	JA3Hash     string            `json:"ja3_hash,omitempty"`
	JA3Full     string            `json:"ja3_full,omitempty"`
	JA4         string            `json:"ja4,omitempty"`
	UserAgent   string            `json:"user_agent,omitempty"`
	AcceptLang  string            `json:"accept_language,omitempty"`
	AcceptEnc   string            `json:"accept_encoding,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
}

// TLSFingerprint holds TLS ClientHello fingerprint data.
type TLSFingerprint struct {
	Version         uint16   `json:"version"`
	CipherSuites    []uint16 `json:"cipher_suites"`
	Extensions      []uint16 `json:"extensions"`
	SupportedGroups []uint16 `json:"supported_groups"` // curves
	PointFormats    []uint8  `json:"point_formats"`
	SignatureAlgos  []uint16 `json:"signature_algos,omitempty"`
	ALPNProtocols   []string `json:"alpn_protocols,omitempty"`
	SNI             bool     `json:"sni"`
}

// HTTP2Fingerprint holds HTTP/2 SETTINGS fingerprint.
type HTTP2Fingerprint struct {
	HeaderTableSize      uint32   `json:"header_table_size"`
	EnablePush           uint32   `json:"enable_push"`
	MaxConcurrentStreams uint32   `json:"max_concurrent_streams"`
	InitialWindowSize    uint32   `json:"initial_window_size"`
	MaxFrameSize         uint32   `json:"max_frame_size"`
	MaxHeaderListSize    uint32   `json:"max_header_list_size"`
	PseudoHeaderOrder    []string `json:"pseudo_header_order"`
	ConnectionFlow       uint32   `json:"connection_flow"`
}

// CollectionResult aggregates all collected fingerprints.
type CollectionResult struct {
	Version      string                 `json:"version"`
	CollectedAt  string                 `json:"collected_at"`
	TotalCount   int                    `json:"total_count"`
	Sources      []string               `json:"sources"`
	Fingerprints []CollectedFingerprint `json:"fingerprints"`
	Statistics   map[string]int         `json:"statistics"`
}

func main() {
	outputDir := flag.String("output", "./training/collected", "output directory")
	format := flag.String("format", "json", "output format: json, go")
	sources := flag.String("sources", "all", "data sources: ja3er,tls_peet,http2,all")
	timeout := flag.Int("timeout", 30, "HTTP request timeout in seconds")
	flag.Parse()

	ensureOutputDirectory(*outputDir)
	client := newHTTPClient(*timeout)
	result := newCollectionResult()
	collectFromSources(context.Background(), client, result, parseSources(*sources))
	result.TotalCount = len(result.Fingerprints)

	fmt.Printf("\nTotal collected: %d fingerprints from %d sources\n", result.TotalCount, len(result.Sources))
	saveCollectionResult(*format, *outputDir, result)
}

func ensureOutputDirectory(outputDir string) {
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		log.Fatalf("create output dir: %v", err)
	}
}

func newHTTPClient(timeoutSeconds int) *http.Client {
	return &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}
}

func newCollectionResult() *CollectionResult {
	return &CollectionResult{
		Version:     "1.0.0",
		CollectedAt: time.Now().UTC().Format(time.RFC3339),
		Sources:     []string{},
		Statistics:  make(map[string]int),
	}
}

func collectFromSources(ctx context.Context, client *http.Client, result *CollectionResult, sourceList []string) {
	for _, src := range sourceList {
		fmt.Printf("Collecting from %s...\n", src)
		fps, ok := collectFromSource(ctx, client, src)
		if !ok {
			continue
		}

		fmt.Printf("  Collected %d fingerprints\n", len(fps))
		result.Fingerprints = append(result.Fingerprints, fps...)
		result.Sources = append(result.Sources, src)
		result.Statistics[src] = len(fps)
	}
}

func collectFromSource(ctx context.Context, client *http.Client, src string) ([]CollectedFingerprint, bool) {
	switch src {
	case "ja3er":
		return collectOrLogError(func() ([]CollectedFingerprint, error) {
			return collectJA3er(ctx, client)
		})
	case "tls_peet":
		return collectOrLogError(func() ([]CollectedFingerprint, error) {
			return collectPeetJS(ctx, client)
		})
	case "builtin_knowledge":
		return collectBuiltinKnowledge(), true
	default:
		fmt.Printf("  Unknown source: %s, skipping\n", src)
		return nil, false
	}
}

func collectOrLogError(fetch func() ([]CollectedFingerprint, error)) ([]CollectedFingerprint, bool) {
	fps, err := fetch()
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return nil, false
	}
	return fps, true
}

func saveCollectionResult(format string, outputDir string, result *CollectionResult) {
	switch format {
	case "json":
		saveCollectionJSON(outputDir, result)
	case "go":
		saveCollectionGo(outputDir, result)
	}
}

func saveCollectionJSON(outputDir string, result *CollectionResult) {
	outPath := filepath.Join(outputDir, "fingerprints.json")
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("marshal output: %v", err)
	}
	if err := os.WriteFile(outPath, data, 0o600); err != nil {
		log.Fatalf("write output: %v", err)
	}
	fmt.Printf("Saved to %s (%.1f KB)\n", outPath, float64(len(data))/1024)
}

func saveCollectionGo(outputDir string, result *CollectionResult) {
	outPath := filepath.Join(outputDir, "collected_profiles.go")
	if err := generateGoProfiles(result, outPath); err != nil {
		log.Fatalf("generate Go: %v", err)
	}
	fmt.Printf("Generated Go profiles: %s\n", outPath)
}

func parseSources(s string) []string {
	if s == "all" {
		return []string{"builtin_knowledge", "ja3er", "tls_peet"}
	}
	return strings.Split(s, ",")
}

// collectJA3er fetches JA3 fingerprints from ja3er.com API
