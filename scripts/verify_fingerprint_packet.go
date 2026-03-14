//go:build ignore

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
