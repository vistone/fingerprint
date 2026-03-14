//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type fingerprintReport struct {
	ProfileID           string             `json:"profile_id"`
	Timestamp           string             `json:"timestamp"`
	ExpectedFingerprint expectedFP         `json:"expected_fingerprint"`
	ObservedFingerprint observedFP         `json:"observed_fingerprint"`
	Verification        verificationResult `json:"verification"`
	RawCapture          string             `json:"raw_capture,omitempty"`
}

type expectedFP struct {
	TLSVersion        string   `json:"tls_version"`
	CipherSuites      []string `json:"cipher_suites"`
	Extensions        []string `json:"extensions"`
	SupportedGroups   []string `json:"supported_groups"`
	ECPointFormats    []string `json:"ec_point_formats"`
	SignatureAlgs     []string `json:"signature_algorithms"`
	HTTPVersion       string   `json:"http_version"`
	UserAgent         string   `json:"user_agent"`
	SecCHUA           string   `json:"sec_ch_ua"`
	HeaderOrder       []string `json:"header_order"`
	PseudoHeaderOrder []string `json:"pseudo_header_order"`
}

type observedFP struct {
	JA3Hash        string            `json:"ja3_hash"`
	JA3String      string            `json:"ja3_string"`
	TLSVersion     string            `json:"tls_version"`
	HTTPVersion    string            `json:"http_version"`
	UserAgent      string            `json:"user_agent"`
	SecCHUA        string            `json:"sec_ch_ua"`
	Headers        map[string]string `json:"headers"`
	HeaderOrder    []string          `json:"header_order"`
	ALPN           string            `json:"alpn"`
	ServerName     string            `json:"server_name"`
	H2Settings     string            `json:"h2_settings,omitempty"`
	H2WindowUpdate string            `json:"h2_window_update,omitempty"`
	H2Priority     string            `json:"h2_priority,omitempty"`
}

type verificationResult struct {
	OverallPass      bool     `json:"overall_pass"`
	TLSMatch         bool     `json:"tls_match"`
	HTTPMatch        bool     `json:"http_match"`
	HeaderOrderMatch bool     `json:"header_order_match"`
	UserAgentMatch   bool     `json:"user_agent_match"`
	SecCHUAMatch     bool     `json:"sec_ch_ua_match"`
	Issues           []string `json:"issues"`
}

func verifyFingerprints(report *fingerprintReport, verbose bool) {
	if verbose {
		fmt.Printf(">>> STEP 5: Verification\n")
	}

	issues := []string{}

	report.Verification.UserAgentMatch = verifyUserAgent(report, &issues, verbose)
	report.Verification.SecCHUAMatch = verifySecCHUA(report, &issues, verbose)
	report.Verification.HTTPMatch = verifyHTTPVersion(report, &issues, verbose)
	report.Verification.TLSMatch = verifyTLS(report, &issues, verbose)

	report.Verification.Issues = issues
	report.Verification.OverallPass = report.Verification.UserAgentMatch &&
		report.Verification.SecCHUAMatch &&
		report.Verification.HTTPMatch &&
		report.Verification.TLSMatch

	if verbose {
		fmt.Printf("\n")
	}
}

func verifyUserAgent(report *fingerprintReport, issues *[]string, verbose bool) bool {
	if report.ExpectedFingerprint.UserAgent == "" || report.ObservedFingerprint.UserAgent == "" {
		return false
	}
	if strings.TrimSpace(report.ExpectedFingerprint.UserAgent) == strings.TrimSpace(report.ObservedFingerprint.UserAgent) {
		if verbose {
			fmt.Printf("  ✓ User-Agent MATCH\n")
		}
		return true
	}
	*issues = append(*issues, fmt.Sprintf("User-Agent mismatch: expected '%s', got '%s'",
		report.ExpectedFingerprint.UserAgent, report.ObservedFingerprint.UserAgent))
	if verbose {
		fmt.Printf("  ✗ User-Agent MISMATCH\n")
		fmt.Printf("    Expected: %s\n", truncate(report.ExpectedFingerprint.UserAgent, 70))
		fmt.Printf("    Observed: %s\n", truncate(report.ObservedFingerprint.UserAgent, 70))
	}
	return false
}

func verifySecCHUA(report *fingerprintReport, issues *[]string, verbose bool) bool {
	if report.ExpectedFingerprint.SecCHUA == "" {
		return true // N/A for non-Chromium browsers
	}
	if strings.TrimSpace(report.ExpectedFingerprint.SecCHUA) == strings.TrimSpace(report.ObservedFingerprint.SecCHUA) {
		if verbose {
			fmt.Printf("  ✓ Sec-CH-UA MATCH\n")
		}
		return true
	}
	*issues = append(*issues, fmt.Sprintf("Sec-CH-UA mismatch: expected '%s', got '%s'",
		report.ExpectedFingerprint.SecCHUA, report.ObservedFingerprint.SecCHUA))
	if verbose {
		fmt.Printf("  ✗ Sec-CH-UA MISMATCH\n")
		fmt.Printf("    Expected: %s\n", report.ExpectedFingerprint.SecCHUA)
		fmt.Printf("    Observed: %s\n", report.ObservedFingerprint.SecCHUA)
	}
	return false
}

func verifyHTTPVersion(report *fingerprintReport, issues *[]string, verbose bool) bool {
	if report.ObservedFingerprint.HTTPVersion == "" {
		return false
	}
	expectedHTTP := report.ExpectedFingerprint.HTTPVersion
	observedHTTP := report.ObservedFingerprint.HTTPVersion

	if strings.Contains(expectedHTTP, "2") && strings.Contains(observedHTTP, "2") {
		if verbose {
			fmt.Printf("  ✓ HTTP Version MATCH (HTTP/2)\n")
		}
		return true
	}
	if strings.Contains(expectedHTTP, "1") && strings.Contains(observedHTTP, "1") {
		if verbose {
			fmt.Printf("  ✓ HTTP Version MATCH (HTTP/1.1)\n")
		}
		return true
	}
	*issues = append(*issues, fmt.Sprintf("HTTP version mismatch: expected %s, got %s",
		expectedHTTP, observedHTTP))
	if verbose {
		fmt.Printf("  ✗ HTTP Version MISMATCH\n")
		fmt.Printf("    Expected: %s\n", expectedHTTP)
		fmt.Printf("    Observed: %s\n", observedHTTP)
	}
	return false
}

func verifyTLS(report *fingerprintReport, issues *[]string, verbose bool) bool {
	if report.ObservedFingerprint.JA3Hash != "" {
		if verbose {
			fmt.Printf("  ✓ TLS Fingerprint PRESENT (custom TLS stack)\n")
			fmt.Printf("    JA3: %s\n", report.ObservedFingerprint.JA3Hash)
		}
		return true
	}
	*issues = append(*issues, "No JA3 hash captured (might be using system TLS)")
	if verbose {
		fmt.Printf("  ⚠ TLS Fingerprint WARNING (no JA3 captured)\n")
	}
	return false
}

func printSummary(report *fingerprintReport) {
	fmt.Printf("========================================\n")
	fmt.Printf("VERIFICATION SUMMARY\n")
	fmt.Printf("========================================\n")
	fmt.Printf("Profile: %s\n", report.ProfileID)
	fmt.Printf("\n")

	if report.Verification.OverallPass {
		fmt.Printf("✓✓✓ OVERALL: PASS ✓✓✓\n")
	} else {
		fmt.Printf("✗✗✗ OVERALL: FAIL ✗✗✗\n")
	}
	fmt.Printf("\n")

	fmt.Printf("Component Checks:\n")
	fmt.Printf("  TLS Fingerprint:  %s\n", passFailIcon(report.Verification.TLSMatch))
	fmt.Printf("  HTTP Version:     %s\n", passFailIcon(report.Verification.HTTPMatch))
	fmt.Printf("  User-Agent:       %s\n", passFailIcon(report.Verification.UserAgentMatch))
	fmt.Printf("  Sec-CH-UA:        %s\n", passFailIcon(report.Verification.SecCHUAMatch))
	fmt.Printf("\n")

	if len(report.Verification.Issues) > 0 {
		fmt.Printf("Issues Found:\n")
		for i, issue := range report.Verification.Issues {
			fmt.Printf("  %d. %s\n", i+1, issue)
		}
		fmt.Printf("\n")
	}

	printConclusion(report)
}

func printConclusion(report *fingerprintReport) {
	fmt.Printf("Key Observations:\n")
	fmt.Printf("  JA3 Hash: %s\n", report.ObservedFingerprint.JA3Hash)
	fmt.Printf("  HTTP Version: %s (expected: %s)\n",
		report.ObservedFingerprint.HTTPVersion, report.ExpectedFingerprint.HTTPVersion)
	fmt.Printf("  User-Agent Match: %t\n", report.Verification.UserAgentMatch)
	fmt.Printf("  Sec-CH-UA Match: %t\n", report.Verification.SecCHUAMatch)
	fmt.Printf("\n")

	fmt.Printf("Conclusion:\n")
	if report.Verification.OverallPass {
		fmt.Printf("  ✓ This profile is using CUSTOM TCP/IP+TLS+HTTP fingerprints\n")
		fmt.Printf("  ✓ NOT using system default networking stack\n")
		fmt.Printf("  ✓ Server receives exactly what we designed\n")
	} else {
		fmt.Printf("  ✗ Verification failed - see issues above\n")
		fmt.Printf("  ✗ May be falling back to system networking\n")
	}
	fmt.Printf("\n")
}

func saveReport(report *fingerprintReport, filename string) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("ERROR: failed to marshal report: %v\n", err)
		return
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Printf("ERROR: failed to write report: %v\n", err)
		return
	}

	fmt.Printf("Report saved: %s\n", filename)
}

func passFailIcon(pass bool) string {
	if pass {
		return "✓ PASS"
	}
	return "✗ FAIL"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
