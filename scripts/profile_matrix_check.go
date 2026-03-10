package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vistone/fingerprint/modules/client"
	"github.com/vistone/fingerprint/modules/profiles"
)

type checkResult struct {
	ProfileID     string                 `json:"profileId"`
	ProfileName   string                 `json:"profileName"`
	BrowserType   string                 `json:"browserType"`
	URL           string                 `json:"url"`
	Success       bool                   `json:"success"`
	StatusCode    int                    `json:"statusCode"`
	Protocol      string                 `json:"protocol"`
	Error         string                 `json:"error,omitempty"`
	ErrorType     string                 `json:"errorType,omitempty"`
	ErrorCode     int                    `json:"errorCode,omitempty"`
	ErrorDetails  map[string]interface{} `json:"errorDetails,omitempty"`
	DurationMilli int64                  `json:"durationMs"`
}

type summary struct {
	TotalProfiles      int                `json:"totalProfiles"`
	TotalURLs          int                `json:"totalUrls"`
	TotalChecks        int                `json:"totalChecks"`
	Success200Checks   int                `json:"success200Checks"`
	FailedChecks       int                `json:"failedChecks"`
	OverallPassRate    float64            `json:"overallPassRate"`
	PerURLPassRate     map[string]float64 `json:"perUrlPassRate"`
	PerProfilePassRate map[string]float64 `json:"perProfilePassRate"`
	GeneratedAt        time.Time          `json:"generatedAt"`
	Results            []checkResult      `json:"results"`
}

type job struct {
	profile profiles.ClientProfile
	url     string
	method  string
}

func main() {
	var urlsArg string
	var method string
	var workers int
	var outputJSON string
	var outputCSV string
	var onlyFailures bool
	var retryTransient bool

	flag.StringVar(&urlsArg, "urls", "https://httpbin.org/get", "Comma-separated URLs to test")
	flag.StringVar(&method, "method", "GET", "HTTP method")
	flag.IntVar(&workers, "workers", 12, "Number of concurrent workers")
	flag.StringVar(&outputJSON, "out-json", "", "Optional output JSON file path")
	flag.StringVar(&outputCSV, "out-csv", "", "Optional output CSV file path")
	flag.BoolVar(&onlyFailures, "only-failures", false, "Print only failed checks to stdout")
	flag.BoolVar(&retryTransient, "retry-transient", true, "Retry transient target-side failures once (429/502/503/504)")
	flag.Parse()

	urls := parseURLs(urlsArg)
	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "no valid urls provided")
		os.Exit(2)
	}

	allProfiles := profiles.GetAll()
	if len(allProfiles) == 0 {
		fmt.Fprintln(os.Stderr, "no profiles found")
		os.Exit(2)
	}

	fmt.Printf("Running matrix check: profiles=%d urls=%d total=%d workers=%d\n", len(allProfiles), len(urls), len(allProfiles)*len(urls), workers)

	jobs := make(chan job)
	resultsCh := make(chan checkResult, workers)

	var processed int64
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				start := time.Now()
				res := client.ExecuteProxyRequest(j.profile, j.url, j.method, "", nil)
				dur := time.Since(start)
				cr := buildCheckResult(j, res, dur)

				if retryTransient && !cr.Success && isTransientTargetFailure(cr) {
					time.Sleep(transientRetryDelay(cr.StatusCode))
					retryStart := time.Now()
					retryRes := client.ExecuteProxyRequest(j.profile, j.url, j.method, "", nil)
					retryDur := time.Since(retryStart)
					retryCR := buildCheckResult(j, retryRes, retryDur)
					if retryCR.Success {
						cr = retryCR
					} else if cr.ErrorDetails == nil {
						cr.ErrorDetails = map[string]interface{}{"script_retry_attempted": true}
					} else {
						cr.ErrorDetails["script_retry_attempted"] = true
					}
				}

				resultsCh <- cr
				curr := atomic.AddInt64(&processed, 1)
				if curr%50 == 0 {
					fmt.Printf("progress: %d/%d\n", curr, len(allProfiles)*len(urls))
				}
			}
		}()
	}

	go func() {
		for _, p := range allProfiles {
			for _, u := range urls {
				jobs <- job{profile: p, url: u, method: strings.ToUpper(method)}
			}
		}
		close(jobs)
		wg.Wait()
		close(resultsCh)
	}()

	allResults := make([]checkResult, 0, len(allProfiles)*len(urls))
	for r := range resultsCh {
		allResults = append(allResults, r)
	}

	sort.Slice(allResults, func(i, j int) bool {
		if allResults[i].ProfileID == allResults[j].ProfileID {
			return allResults[i].URL < allResults[j].URL
		}
		return allResults[i].ProfileID < allResults[j].ProfileID
	})

	s := buildSummary(allProfiles, urls, allResults)
	printSummary(s, onlyFailures)

	if outputJSON != "" {
		if err := writeJSON(outputJSON, s); err != nil {
			fmt.Fprintf(os.Stderr, "write json failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("json report: %s\n", outputJSON)
	}

	if outputCSV != "" {
		if err := writeCSV(outputCSV, allResults); err != nil {
			fmt.Fprintf(os.Stderr, "write csv failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("csv report: %s\n", outputCSV)
	}

	if s.FailedChecks > 0 {
		os.Exit(1)
	}
}

func buildCheckResult(j job, res *client.ProxyResult, dur time.Duration) checkResult {
	statusCode := 0
	protocol := ""
	if res.ResponseTrace != nil {
		statusCode = res.ResponseTrace.StatusCode
		protocol = res.ResponseTrace.Protocol
	}
	if protocol == "" && res.RequestTrace != nil && res.RequestTrace.HTTP != nil {
		protocol = res.RequestTrace.HTTP.Protocol
	}

	return checkResult{
		ProfileID:     j.profile.ID,
		ProfileName:   j.profile.Name,
		BrowserType:   string(j.profile.BrowserType),
		URL:           j.url,
		Success:       res.Success && statusCode == 200,
		StatusCode:    statusCode,
		Protocol:      protocol,
		Error:         res.Error,
		ErrorType:     res.ErrorType,
		ErrorCode:     res.ErrorCode,
		ErrorDetails:  res.ErrorDetails,
		DurationMilli: dur.Milliseconds(),
	}
}

func isTransientTargetFailure(r checkResult) bool {
	if r.ErrorType != "target_http_error" {
		return false
	}
	switch r.StatusCode {
	case 429, 502, 503, 504:
		return true
	default:
		return false
	}
}

func transientRetryDelay(statusCode int) time.Duration {
	if statusCode == 429 {
		return 6 * time.Second
	}
	return 1200 * time.Millisecond
}

func parseURLs(s string) []string {
	parts := strings.Split(s, ",")
	urls := make([]string, 0, len(parts))
	for _, p := range parts {
		u := strings.TrimSpace(p)
		if u == "" {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(u), "http://") && !strings.HasPrefix(strings.ToLower(u), "https://") {
			u = "https://" + u
		}
		urls = append(urls, u)
	}
	return urls
}

func buildSummary(ps []profiles.ClientProfile, urls []string, results []checkResult) summary {
	totalChecks := len(ps) * len(urls)
	success200 := 0
	failed := 0

	urlTotal := map[string]int{}
	urlSuccess := map[string]int{}
	profileTotal := map[string]int{}
	profileSuccess := map[string]int{}

	for _, r := range results {
		urlTotal[r.URL]++
		profileTotal[r.ProfileID]++
		if r.Success {
			success200++
			urlSuccess[r.URL]++
			profileSuccess[r.ProfileID]++
		} else {
			failed++
		}
	}

	perURL := make(map[string]float64, len(urls))
	for _, u := range urls {
		if urlTotal[u] == 0 {
			perURL[u] = 0
			continue
		}
		perURL[u] = float64(urlSuccess[u]) * 100 / float64(urlTotal[u])
	}

	perProfile := make(map[string]float64, len(ps))
	for _, p := range ps {
		t := profileTotal[p.ID]
		if t == 0 {
			perProfile[p.ID] = 0
			continue
		}
		perProfile[p.ID] = float64(profileSuccess[p.ID]) * 100 / float64(t)
	}

	overall := 0.0
	if totalChecks > 0 {
		overall = float64(success200) * 100 / float64(totalChecks)
	}

	return summary{
		TotalProfiles:      len(ps),
		TotalURLs:          len(urls),
		TotalChecks:        totalChecks,
		Success200Checks:   success200,
		FailedChecks:       failed,
		OverallPassRate:    overall,
		PerURLPassRate:     perURL,
		PerProfilePassRate: perProfile,
		GeneratedAt:        time.Now(),
		Results:            results,
	}
}

func printSummary(s summary, onlyFailures bool) {
	fmt.Println("================ Matrix Summary ================")
	fmt.Printf("profiles: %d\n", s.TotalProfiles)
	fmt.Printf("urls: %d\n", s.TotalURLs)
	fmt.Printf("checks: %d\n", s.TotalChecks)
	fmt.Printf("success(200): %d\n", s.Success200Checks)
	fmt.Printf("failed: %d\n", s.FailedChecks)
	fmt.Printf("pass rate: %.2f%%\n", s.OverallPassRate)
	fmt.Println("per-url pass rate:")

	keys := make([]string, 0, len(s.PerURLPassRate))
	for k := range s.PerURLPassRate {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  - %s: %.2f%%\n", k, s.PerURLPassRate[k])
	}

	if s.FailedChecks == 0 {
		fmt.Println("all checks are 200")
		return
	}

	fmt.Println("sample failures:")
	shown := 0
	for _, r := range s.Results {
		if r.Success {
			continue
		}
		if !onlyFailures || shown < 40 {
			fmt.Printf("  - profile=%s url=%s status=%d type=%s err=%s\n", r.ProfileID, r.URL, r.StatusCode, r.ErrorType, truncate(r.Error, 120))
			shown++
		}
		if !onlyFailures && shown >= 40 {
			break
		}
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func writeJSON(path string, s summary) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

func writeCSV(path string, results []checkResult) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	header := []string{"profile_id", "profile_name", "browser_type", "url", "success_200", "status_code", "protocol", "error_type", "error_code", "error", "duration_ms"}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, r := range results {
		row := []string{
			r.ProfileID,
			r.ProfileName,
			r.BrowserType,
			r.URL,
			fmt.Sprintf("%t", r.Success),
			fmt.Sprintf("%d", r.StatusCode),
			r.Protocol,
			r.ErrorType,
			fmt.Sprintf("%d", r.ErrorCode),
			r.Error,
			fmt.Sprintf("%d", r.DurationMilli),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}
