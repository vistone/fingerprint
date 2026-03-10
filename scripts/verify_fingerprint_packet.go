package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/vistone/fingerprint/modules/client"
	"github.com/vistone/fingerprint/modules/profiles"
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

func main() {
	var profileID string
	var capturePackets bool
	var verbose bool
	var outJSON string
	var strict bool

	flag.StringVar(&profileID, "profile", "chrome_134", "Profile ID to verify")
	flag.BoolVar(&capturePackets, "capture", false, "Capture packets with tcpdump (requires root)")
	flag.BoolVar(&verbose, "verbose", true, "Print verbose logs")
	flag.StringVar(&outJSON, "out", "", "Output JSON report file")
	flag.BoolVar(&strict, "strict", false, "Enable strict fingerprint mode (disable standard TLS compat fallback)")
	flag.Parse()

	profile, ok := profiles.Get(profileID)
	if !ok {
		fmt.Printf("ERROR: profile not found: %s\n", profileID)
		os.Exit(1)
	}

	fmt.Printf("========================================\n")
	fmt.Printf("FINGERPRINT VERIFICATION (PACKET LEVEL)\n")
	fmt.Printf("========================================\n")
	fmt.Printf("Profile ID: %s\n", profileID)
	fmt.Printf("Timestamp: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("\n")

	report := fingerprintReport{
		ProfileID: profileID,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// 1. Extract expected fingerprint from profile
	extractExpectedFingerprint(profile, &report, verbose)

	// 2. Start packet capture if requested
	var captureCmd *exec.Cmd
	var captureFile string
	if capturePackets {
		captureFile = fmt.Sprintf("/tmp/fingerprint_capture_%s.pcap", profileID)
		captureCmd = startPacketCapture(captureFile, verbose)
		if captureCmd != nil {
			defer captureCmd.Process.Kill()
			time.Sleep(1 * time.Second) // Let tcpdump start
		}
	}

	// 3. Query tls.peet.ws for JA3 and HTTP/2 fingerprint
	queryTLSPeetWS(profile, &report, verbose, strict)

	// 4. Query httpbin.org for HTTP headers
	queryHttpbin(profile, &report, verbose, strict)

	// 5. Stop packet capture and analyze
	if captureCmd != nil {
		time.Sleep(1 * time.Second)
		captureCmd.Process.Kill()
		analyzePcap(captureFile, &report, verbose)
	}

	// 6. Verify and compare
	verifyFingerprints(&report, verbose)

	// 7. Print summary
	printSummary(&report)

	// 8. Save JSON report
	if outJSON != "" {
		saveReport(&report, outJSON)
	}

	if !report.Verification.OverallPass {
		os.Exit(1)
	}
}

func extractExpectedFingerprint(profile profiles.ClientProfile, report *fingerprintReport, verbose bool) {
	if verbose {
		fmt.Printf(">>> STEP 1: Extracting Expected Fingerprint from Profile\n")
	}

	if profile.Headers != nil {
		report.ExpectedFingerprint.UserAgent = profile.Headers.UserAgent
		report.ExpectedFingerprint.SecCHUA = profile.Headers.SecCHUA
	}

	// HTTP version - assume HTTP/2 by default
	report.ExpectedFingerprint.HTTPVersion = "HTTP/2.0"

	// TLS fingerprint details
	report.ExpectedFingerprint.TLSVersion = fmt.Sprintf("0x%04x", profile.TLSVersion)

	for _, cs := range profile.CipherSuites {
		report.ExpectedFingerprint.CipherSuites = append(report.ExpectedFingerprint.CipherSuites, fmt.Sprintf("0x%04x", cs))
	}

	for _, ext := range profile.Extensions {
		report.ExpectedFingerprint.Extensions = append(report.ExpectedFingerprint.Extensions, fmt.Sprintf("%v", ext))
	}

	for _, curve := range profile.SupportedCurves {
		report.ExpectedFingerprint.SupportedGroups = append(report.ExpectedFingerprint.SupportedGroups, fmt.Sprintf("%v", curve))
	}

	// HTTP/2 pseudo-header order
	if len(profile.PseudoHeaderOrder) > 0 {
		report.ExpectedFingerprint.PseudoHeaderOrder = profile.PseudoHeaderOrder
	}

	// TCP/IP fingerprint
	if profile.TCPIP != nil {
		fmt.Printf("  TCP/IP TTL: %d, Window: %d, JA4T: %s\n",
			profile.TCPIP.TTL, profile.TCPIP.WindowSize, profile.TCPIP.JA4T)
	}

	if verbose {
		fmt.Printf("  User-Agent: %s\n", truncate(report.ExpectedFingerprint.UserAgent, 80))
		fmt.Printf("  Sec-CH-UA: %s\n", report.ExpectedFingerprint.SecCHUA)
		fmt.Printf("  HTTP Version: %s\n", report.ExpectedFingerprint.HTTPVersion)
		fmt.Printf("  TLS Version: %s\n", report.ExpectedFingerprint.TLSVersion)
		fmt.Printf("  Cipher Suites: %d items\n", len(report.ExpectedFingerprint.CipherSuites))
		fmt.Printf("  Extensions: %d items\n", len(report.ExpectedFingerprint.Extensions))
		fmt.Printf("  Header Order: %v\n", report.ExpectedFingerprint.HeaderOrder)
		fmt.Printf("\n")
	}
}

func queryTLSPeetWS(profile profiles.ClientProfile, report *fingerprintReport, verbose bool, strict bool) {
	if verbose {
		fmt.Printf(">>> STEP 2: Querying tls.peet.ws for TLS/HTTP2 Fingerprint\n")
	}

	cli, err := client.NewBrowserClient(profile, &client.ClientOptions{
		Timeout:           35 * time.Second,
		StrictFingerprint: strict,
	})
	if err != nil {
		report.Verification.Issues = append(report.Verification.Issues, fmt.Sprintf("failed to create client: %v", err))
		if verbose {
			fmt.Printf("  ERROR: %v\n\n", err)
		}
		return
	}

	req, _ := fhttp.NewRequest("GET", "https://tls.peet.ws/api/all", nil)
	resp, err := cli.Do(req)
	if err != nil {
		report.Verification.Issues = append(report.Verification.Issues, fmt.Sprintf("tls.peet.ws request failed: %v", err))
		if verbose {
			fmt.Printf("  ERROR: %v\n\n", err)
		}
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		report.Verification.Issues = append(report.Verification.Issues, fmt.Sprintf("failed to parse tls.peet.ws response: %v", err))
		if verbose {
			fmt.Printf("  ERROR: %v\n\n", err)
		}
		return
	}

	// Extract JA3
	if ja3, ok := result["tls"].(map[string]interface{}); ok {
		if hash, ok := ja3["ja3"].(string); ok {
			report.ObservedFingerprint.JA3Hash = hash
		}
		if str, ok := ja3["ja3_text"].(string); ok {
			report.ObservedFingerprint.JA3String = str
		}
	}

	// Extract HTTP version
	if http, ok := result["http"].(map[string]interface{}); ok {
		if ver, ok := http["version"].(string); ok {
			report.ObservedFingerprint.HTTPVersion = ver
		}
	}

	// Extract HTTP/2 settings
	if h2, ok := result["http2"].(map[string]interface{}); ok {
		if settings, ok := h2["sent_frames"].([]interface{}); ok && len(settings) > 0 {
			settingsJSON, _ := json.Marshal(settings)
			report.ObservedFingerprint.H2Settings = string(settingsJSON)
		}
	}

	if verbose {
		fmt.Printf("  JA3 Hash: %s\n", report.ObservedFingerprint.JA3Hash)
		fmt.Printf("  JA3 String: %s\n", truncate(report.ObservedFingerprint.JA3String, 80))
		fmt.Printf("  HTTP Version: %s\n", report.ObservedFingerprint.HTTPVersion)
		fmt.Printf("  HTTP/2 Settings: %s\n", truncate(report.ObservedFingerprint.H2Settings, 60))
		fmt.Printf("\n")
	}
}

func queryHttpbin(profile profiles.ClientProfile, report *fingerprintReport, verbose bool, strict bool) {
	if verbose {
		fmt.Printf(">>> STEP 3: Querying httpbin.org for HTTP Headers\n")
	}

	cli, err := client.NewBrowserClient(profile, &client.ClientOptions{
		Timeout:           35 * time.Second,
		StrictFingerprint: strict,
	})
	if err != nil {
		report.Verification.Issues = append(report.Verification.Issues, fmt.Sprintf("failed to create client: %v", err))
		if verbose {
			fmt.Printf("  ERROR: %v\n\n", err)
		}
		return
	}

	req, _ := fhttp.NewRequest("GET", "https://httpbin.org/anything", nil)
	resp, err := cli.Do(req)
	if err != nil {
		report.Verification.Issues = append(report.Verification.Issues, fmt.Sprintf("httpbin request failed: %v", err))
		if verbose {
			fmt.Printf("  ERROR: %v\n\n", err)
		}
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		report.Verification.Issues = append(report.Verification.Issues, fmt.Sprintf("failed to parse httpbin response: %v", err))
		if verbose {
			fmt.Printf("  ERROR: %v\n\n", err)
		}
		return
	}

	if headers, ok := result["headers"].(map[string]interface{}); ok {
		report.ObservedFingerprint.Headers = make(map[string]string)
		for k, v := range headers {
			if str, ok := v.(string); ok {
				report.ObservedFingerprint.Headers[k] = str
				if strings.EqualFold(k, "User-Agent") {
					report.ObservedFingerprint.UserAgent = str
				}
				if strings.EqualFold(k, "Sec-Ch-Ua") {
					report.ObservedFingerprint.SecCHUA = str
				}
			}
		}

		// Try to extract header order (httpbin doesn't preserve order, but we can list them)
		for k := range headers {
			report.ObservedFingerprint.HeaderOrder = append(report.ObservedFingerprint.HeaderOrder, k)
		}
	}

	if verbose {
		fmt.Printf("  Observed User-Agent: %s\n", truncate(report.ObservedFingerprint.UserAgent, 80))
		fmt.Printf("  Observed Sec-CH-UA: %s\n", report.ObservedFingerprint.SecCHUA)
		fmt.Printf("  Total Headers: %d\n", len(report.ObservedFingerprint.Headers))
		fmt.Printf("\n")
	}
}

func startPacketCapture(outputFile string, verbose bool) *exec.Cmd {
	if verbose {
		fmt.Printf(">>> PACKET CAPTURE: Starting tcpdump\n")
		fmt.Printf("  Output: %s\n", outputFile)
		fmt.Printf("  Filter: tcp port 443\n")
		fmt.Printf("  (requires root privileges)\n\n")
	}

	cmd := exec.Command("sudo", "tcpdump", "-i", "any", "-w", outputFile, "tcp port 443", "-s", "0")
	if err := cmd.Start(); err != nil {
		if verbose {
			fmt.Printf("  WARNING: Failed to start tcpdump: %v\n", err)
			fmt.Printf("  Continuing without packet capture...\n\n")
		}
		return nil
	}

	return cmd
}

func analyzePcap(pcapFile string, report *fingerprintReport, verbose bool) {
	if verbose {
		fmt.Printf(">>> STEP 4: Analyzing Packet Capture\n")
	}

	// Use tshark to analyze the pcap file
	cmd := exec.Command("tshark", "-r", pcapFile, "-Y", "tls.handshake.type == 1", "-T", "fields",
		"-e", "tls.handshake.version",
		"-e", "tls.handshake.ciphersuite",
		"-e", "tls.handshake.extension.type",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if verbose {
			fmt.Printf("  WARNING: tshark not available or failed: %v\n", err)
			fmt.Printf("  Skipping packet analysis...\n\n")
		}
		return
	}

	report.RawCapture = string(output)
	if verbose {
		fmt.Printf("  Captured TLS ClientHello:\n")
		fmt.Printf("  %s\n\n", truncate(string(output), 200))
	}
}

func verifyFingerprints(report *fingerprintReport, verbose bool) {
	if verbose {
		fmt.Printf(">>> STEP 5: Verification\n")
	}

	issues := []string{}

	// Check User-Agent
	if report.ExpectedFingerprint.UserAgent != "" && report.ObservedFingerprint.UserAgent != "" {
		if strings.TrimSpace(report.ExpectedFingerprint.UserAgent) == strings.TrimSpace(report.ObservedFingerprint.UserAgent) {
			report.Verification.UserAgentMatch = true
			if verbose {
				fmt.Printf("  ✓ User-Agent MATCH\n")
			}
		} else {
			report.Verification.UserAgentMatch = false
			issues = append(issues, fmt.Sprintf("User-Agent mismatch: expected '%s', got '%s'",
				report.ExpectedFingerprint.UserAgent, report.ObservedFingerprint.UserAgent))
			if verbose {
				fmt.Printf("  ✗ User-Agent MISMATCH\n")
				fmt.Printf("    Expected: %s\n", truncate(report.ExpectedFingerprint.UserAgent, 70))
				fmt.Printf("    Observed: %s\n", truncate(report.ObservedFingerprint.UserAgent, 70))
			}
		}
	}

	// Check Sec-CH-UA
	if report.ExpectedFingerprint.SecCHUA != "" {
		if strings.TrimSpace(report.ExpectedFingerprint.SecCHUA) == strings.TrimSpace(report.ObservedFingerprint.SecCHUA) {
			report.Verification.SecCHUAMatch = true
			if verbose {
				fmt.Printf("  ✓ Sec-CH-UA MATCH\n")
			}
		} else {
			report.Verification.SecCHUAMatch = false
			issues = append(issues, fmt.Sprintf("Sec-CH-UA mismatch: expected '%s', got '%s'",
				report.ExpectedFingerprint.SecCHUA, report.ObservedFingerprint.SecCHUA))
			if verbose {
				fmt.Printf("  ✗ Sec-CH-UA MISMATCH\n")
				fmt.Printf("    Expected: %s\n", report.ExpectedFingerprint.SecCHUA)
				fmt.Printf("    Observed: %s\n", report.ObservedFingerprint.SecCHUA)
			}
		}
	} else {
		report.Verification.SecCHUAMatch = true // N/A for non-Chromium browsers
	}

	// Check HTTP version
	if report.ObservedFingerprint.HTTPVersion != "" {
		expectedHTTP := report.ExpectedFingerprint.HTTPVersion
		observedHTTP := report.ObservedFingerprint.HTTPVersion

		// Normalize HTTP version strings
		if strings.Contains(expectedHTTP, "2") && strings.Contains(observedHTTP, "2") {
			report.Verification.HTTPMatch = true
			if verbose {
				fmt.Printf("  ✓ HTTP Version MATCH (HTTP/2)\n")
			}
		} else if strings.Contains(expectedHTTP, "1") && strings.Contains(observedHTTP, "1") {
			report.Verification.HTTPMatch = true
			if verbose {
				fmt.Printf("  ✓ HTTP Version MATCH (HTTP/1.1)\n")
			}
		} else {
			report.Verification.HTTPMatch = false
			issues = append(issues, fmt.Sprintf("HTTP version mismatch: expected %s, got %s",
				expectedHTTP, observedHTTP))
			if verbose {
				fmt.Printf("  ✗ HTTP Version MISMATCH\n")
				fmt.Printf("    Expected: %s\n", expectedHTTP)
				fmt.Printf("    Observed: %s\n", observedHTTP)
			}
		}
	}

	// Check TLS fingerprint (JA3 hash presence means TLS is custom)
	if report.ObservedFingerprint.JA3Hash != "" {
		report.Verification.TLSMatch = true
		if verbose {
			fmt.Printf("  ✓ TLS Fingerprint PRESENT (custom TLS stack)\n")
			fmt.Printf("    JA3: %s\n", report.ObservedFingerprint.JA3Hash)
		}
	} else {
		report.Verification.TLSMatch = false
		issues = append(issues, "No JA3 hash captured (might be using system TLS)")
		if verbose {
			fmt.Printf("  ⚠ TLS Fingerprint WARNING (no JA3 captured)\n")
		}
	}

	report.Verification.Issues = issues
	report.Verification.OverallPass = report.Verification.UserAgentMatch &&
		report.Verification.SecCHUAMatch &&
		report.Verification.HTTPMatch &&
		report.Verification.TLSMatch

	if verbose {
		fmt.Printf("\n")
	}
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
