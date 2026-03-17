package waf

import (
	"net/http"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
)

// HTTPEngine analyzes HTTP layer characteristics
type HTTPEngine struct {
	knownBots []string
}

// HTTPResult contains HTTP analysis results
type HTTPResult struct {
	Score       float64
	Factors     []core.RiskFactor
	BotDetected string
	IsAutomated bool
}

// NewHTTPEngine creates a new HTTP analysis engine
func NewHTTPEngine() *HTTPEngine {
	return &HTTPEngine{
		knownBots: []string{
			"curl", "wget", "python-requests", "scrapy",
			"selenium", "puppeteer", "playwright",
			"httpclient", "java", "go-http-client",
		},
	}
}

// Analyze performs HTTP layer analysis
func (e *HTTPEngine) Analyze(req *http.Request) *HTTPResult {
	result := &HTTPResult{
		Score:   0,
		Factors: make([]core.RiskFactor, 0),
	}

	ua := strings.ToLower(req.UserAgent())
	e.applyUserAgentChecks(result, ua)
	e.applySuspiciousHeaderChecks(result, req)
	e.applyHeaderQualityChecks(result, req)
	e.applyProtocolChecks(result, req)
	return result
}

func (e *HTTPEngine) applyUserAgentChecks(result *HTTPResult, ua string) {
	for _, bot := range e.knownBots {
		if strings.Contains(ua, bot) {
			result.Score += 0.8
			result.IsAutomated = true
			result.BotDetected = bot
			result.Factors = append(result.Factors, core.RiskFactor{
				Name:        "bot_user_agent",
				Weight:      0.8,
				Description: "Known bot User-Agent: " + bot,
			})
			break
		}
	}
	if ua == "" {
		result.Score += 0.5
		result.Factors = append(result.Factors, core.RiskFactor{
			Name:        "missing_user_agent",
			Weight:      0.5,
			Description: "No User-Agent header",
		})
	}
}

func (e *HTTPEngine) applySuspiciousHeaderChecks(result *HTTPResult, req *http.Request) {
	suspiciousHeaders := []string{"X-Scrapy-Project", "X-Selenium", "X-Puppeteer", "X-PhantomJS"}
	for _, header := range suspiciousHeaders {
		if req.Header.Get(header) != "" {
			result.Score += 0.6
			result.IsAutomated = true
			result.Factors = append(result.Factors, core.RiskFactor{
				Name:        "suspicious_header",
				Weight:      0.6,
				Description: "Suspicious automation header: " + header,
			})
		}
	}
}

func (e *HTTPEngine) applyHeaderQualityChecks(result *HTTPResult, req *http.Request) {
	accept := req.Header.Get("Accept")
	if accept == "" || accept == "*/*" {
		result.Score += 0.2
		result.Factors = append(result.Factors, core.RiskFactor{
			Name:        "generic_accept_header",
			Weight:      0.2,
			Description: "Generic or missing Accept header",
		})
	}
	referer := req.Header.Get("Referer")
	if referer == "" && req.Method != "GET" {
		result.Factors = append(result.Factors, core.RiskFactor{
			Name:        "missing_referer",
			Weight:      0.1,
			Description: "Missing Referer header on non-GET request",
		})
	}
}

func (e *HTTPEngine) applyProtocolChecks(result *HTTPResult, req *http.Request) {
	if req.ProtoMajor < 2 {
		result.Factors = append(result.Factors, core.RiskFactor{
			Name:        "http1_request",
			Weight:      0.1,
			Description: "Using HTTP/1.1 instead of HTTP/2",
		})
	}
}
