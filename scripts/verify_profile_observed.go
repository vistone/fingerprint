package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/vistone/fingerprint/modules/client"
	"github.com/vistone/fingerprint/modules/profiles"
	legacyprofiles "github.com/vistone/fingerprint/modules/profiles/legacy"
)

type verifyResult struct {
	ProfileID            string
	BrowserVersion       string
	ExpectedUA           string
	ObservedUAHttpbin    string
	ObservedUAPeet       string
	ObservedCHUA         string
	ObservedHTTPVer      string
	ObservedJA3Hash      string
	MatchUAHttpbin       bool
	MatchUAPeet          bool
	MatchCHUA            bool
	OverallPass          bool
	Error                string
	LegacySpecID         string
	ProfileExtensions    int
	LegacySpecExtensions int
	ProfileCurves        int
	TLSExtensionDiff     int
}

func main() {
	var profileID string
	var timeoutSec int
	var strict bool

	flag.StringVar(&profileID, "profile", "chrome_134", "Profile ID to verify, or 'all'")
	flag.IntVar(&timeoutSec, "timeout", 35, "Request timeout (seconds)")
	flag.BoolVar(&strict, "strict", false, "Enable strict fingerprint mode (disable standard TLS compat fallback)")
	flag.Parse()

	all := profiles.GetAll()
	if len(all) == 0 {
		fmt.Println("no profiles found")
		os.Exit(2)
	}

	targets := make([]profiles.ClientProfile, 0)
	if strings.EqualFold(profileID, "all") {
		targets = all
	} else {
		found := false
		for _, p := range all {
			if p.ID == profileID {
				targets = append(targets, p)
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("profile not found: %s\n", profileID)
			os.Exit(2)
		}
	}

	results := make([]verifyResult, 0, len(targets))
	for _, p := range targets {
		results = append(results, verifyOne(p, time.Duration(timeoutSec)*time.Second, strict))
	}

	sort.Slice(results, func(i, j int) bool { return results[i].ProfileID < results[j].ProfileID })

	pass := 0
	for _, r := range results {
		status := "FAIL"
		if r.OverallPass {
			status = "PASS"
			pass++
		}
		fmt.Printf("[%s] profile=%s version=%s\n", status, r.ProfileID, r.BrowserVersion)
		if r.Error != "" {
			fmt.Printf("  error=%s\n", r.Error)
			continue
		}
		fmt.Printf("  expected_ua=%s\n", r.ExpectedUA)
		fmt.Printf("  observed_httpbin_ua=%s (match=%t)\n", r.ObservedUAHttpbin, r.MatchUAHttpbin)
		fmt.Printf("  observed_peet_ua=%s (match=%t)\n", r.ObservedUAPeet, r.MatchUAPeet)
		fmt.Printf("  observed_sec_ch_ua=%s (match=%t)\n", r.ObservedCHUA, r.MatchCHUA)
		fmt.Printf("  observed_http_version=%s\n", r.ObservedHTTPVer)
		fmt.Printf("  observed_ja3_hash=%s\n", r.ObservedJA3Hash)
		fmt.Printf("  tls_diag_legacy_spec=%s\n", r.LegacySpecID)
		fmt.Printf("  tls_diag_profile_extensions=%d\n", r.ProfileExtensions)
		fmt.Printf("  tls_diag_legacy_extensions=%d\n", r.LegacySpecExtensions)
		fmt.Printf("  tls_diag_extension_diff=%d\n", r.TLSExtensionDiff)
		fmt.Printf("  tls_diag_profile_curves=%d\n", r.ProfileCurves)
	}

	fmt.Printf("\nsummary: pass=%d total=%d pass_rate=%.2f%%\n", pass, len(results), float64(pass)*100/float64(len(results)))
	if pass != len(results) {
		os.Exit(1)
	}
}

func verifyOne(profile profiles.ClientProfile, timeout time.Duration, strict bool) verifyResult {
	fixed := profile
	_ = profiles.ValidateAndRepair(&fixed)

	result := verifyResult{
		ProfileID:         fixed.ID,
		BrowserVersion:    fixed.BrowserVersion,
		ExpectedUA:        fixed.Headers.UserAgent,
		ProfileExtensions: len(fixed.Extensions),
		ProfileCurves:     len(fixed.SupportedCurves),
	}

	result.LegacySpecID, result.LegacySpecExtensions = resolveLegacySpecDiag(fixed)
	if result.LegacySpecExtensions > 0 {
		result.TLSExtensionDiff = result.ProfileExtensions - result.LegacySpecExtensions
	}

	c, err := client.NewBrowserClient(fixed, &client.ClientOptions{
		Timeout:           timeout,
		FollowRedirects:   true,
		StrictFingerprint: strict,
	})
	if err != nil {
		result.Error = fmt.Sprintf("create client failed: %v", err)
		return result
	}
	defer c.Close()

	// 1) httpbin: 服务器看到的请求头（直接回显）
	httpbinBody, err := doGet(c, "https://httpbin.org/anything")
	if err != nil {
		result.Error = fmt.Sprintf("httpbin request failed: %v", err)
		return result
	}
	var httpbin map[string]interface{}
	if err := json.Unmarshal(httpbinBody, &httpbin); err == nil {
		result.ObservedUAHttpbin = asString(dig(httpbin, "headers", "User-Agent"))
		result.ObservedCHUA = asString(dig(httpbin, "headers", "Sec-Ch-Ua"))
	}

	// 2) tls.peet.ws: 服务器视角 TLS/HTTP 指纹
	peetBody, err := doGet(c, "https://tls.peet.ws/api/all")
	if err == nil {
		var peet map[string]interface{}
		if err := json.Unmarshal(peetBody, &peet); err == nil {
			result.ObservedUAPeet = asString(dig(peet, "user_agent"))
			result.ObservedHTTPVer = asString(dig(peet, "http_version"))
			result.ObservedJA3Hash = asString(dig(peet, "tls", "ja3_hash"))
		}
	}

	result.MatchUAHttpbin = strings.TrimSpace(result.ObservedUAHttpbin) == strings.TrimSpace(result.ExpectedUA)
	result.MatchUAPeet = result.ObservedUAPeet == "" || strings.TrimSpace(result.ObservedUAPeet) == strings.TrimSpace(result.ExpectedUA)
	if fixed.Headers.SecCHUA != "" {
		result.MatchCHUA = strings.TrimSpace(result.ObservedCHUA) == strings.TrimSpace(fixed.Headers.SecCHUA)
	} else {
		result.MatchCHUA = true
	}

	result.OverallPass = result.MatchUAHttpbin && result.MatchUAPeet && result.MatchCHUA
	return result
}

func doGet(c *client.BrowserClient, url string) ([]byte, error) {
	req, err := fhttp.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func dig(m map[string]interface{}, keys ...string) interface{} {
	var cur interface{} = m
	for _, k := range keys {
		nextMap, ok := cur.(map[string]interface{})
		if !ok {
			return nil
		}
		cur = nextMap[k]
	}
	return cur
}

func asString(v interface{}) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func resolveLegacySpecDiag(profile profiles.ClientProfile) (string, int) {
	legacyID := resolveLegacyProfileIDForDiag(profile)
	if legacyID == "" {
		return "", 0
	}
	legacyProfile, ok := legacyprofiles.GetClientProfile(legacyID)
	if !ok {
		return legacyID, 0
	}
	spec, err := legacyProfile.GetClientHelloSpec()
	if err != nil {
		return legacyID, 0
	}
	return legacyID, len(spec.Extensions)
}

func resolveLegacyProfileIDForDiag(profile profiles.ClientProfile) string {
	if profile.ID != "" {
		if _, ok := legacyprofiles.GetClientProfile(profile.ID); ok {
			return profile.ID
		}
	}

	browser := strings.ToLower(string(profile.BrowserType))
	if browser == "" {
		return ""
	}
	prefix := browser + "_"

	hasTargetMajor := false
	targetMajor := 0
	if profile.BrowserVersion != "" {
		if major, ok := parseLeadingMajorDiag(profile.BrowserVersion); ok {
			targetMajor = major
			hasTargetMajor = true
		}
	}

	bestUnderOrEqualID := ""
	bestUnderOrEqualMajor := -1
	latestID := ""
	latestMajor := -1

	legacyIDs := legacyprofiles.GetAllProfiles()
	sort.Strings(legacyIDs)
	for _, id := range legacyIDs {
		idLower := strings.ToLower(id)
		if !strings.HasPrefix(idLower, prefix) {
			continue
		}
		if browser == "safari" && (strings.HasPrefix(idLower, "safari_ios_") || strings.HasPrefix(idLower, "safari_ipad_")) {
			continue
		}

		remainder := strings.TrimPrefix(idLower, prefix)
		major, ok := parseLeadingMajorDiag(remainder)
		if !ok {
			continue
		}

		if major > latestMajor {
			latestMajor = major
			latestID = id
		}
		if hasTargetMajor && major <= targetMajor && major > bestUnderOrEqualMajor {
			bestUnderOrEqualMajor = major
			bestUnderOrEqualID = id
		}
	}

	if bestUnderOrEqualID != "" {
		return bestUnderOrEqualID
	}
	return latestID
}

func parseLeadingMajorDiag(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	idx := 0
	for idx < len(value) {
		c := value[idx]
		if c < '0' || c > '9' {
			break
		}
		idx++
	}
	if idx == 0 {
		return 0, false
	}
	major, err := strconv.Atoi(value[:idx])
	if err != nil {
		return 0, false
	}
	return major, true
}
