package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/vistone/fingerprint/modules/crawler"
	"github.com/vistone/fingerprint/modules/gateway"
	"github.com/vistone/fingerprint/modules/waf"
)

type runtimeModules struct {
	crawler *crawler.Crawler
	waf     *waf.WAF
}

func configureRuntimeModules(config *gateway.GatewayConfig) *runtimeModules {
	modules := &runtimeModules{crawler: nil, waf: nil}

	if shouldEnableWAF(config) {
		modules.waf = waf.NewWAF(loadWAFConfig(config))
	}

	if shouldEnableCrawler(config) {
		modules.crawler = crawler.NewCrawler(loadCrawlerConfig())
	}

	return modules
}

func attachRuntimeModules(gatewayService *gateway.Gateway, modules *runtimeModules) {
	if modules == nil {
		return
	}

	if modules.crawler != nil {
		gatewayService.SetCrawler(modules.crawler)

		if shouldAutostartCrawler(modules.crawler) {
			err := modules.crawler.Start()
			if err != nil {
				log.Printf("crawler autostart failed: %v", err)
			}
		}
	}

	if modules.waf != nil {
		gatewayService.SetWAF(modules.waf)
	}
}

func protectWithWAF(wafInstance *waf.WAF, nextHandler http.HandlerFunc) http.HandlerFunc {
	baseHandler := metricsMiddleware(nextHandler)
	if wafInstance == nil {
		return baseHandler
	}

	protectedHandler := wafInstance.Middleware(baseHandler)

	return func(writer http.ResponseWriter, request *http.Request) {
		protectedHandler.ServeHTTP(writer, request)
	}
}

func shouldEnableCrawler(config *gateway.GatewayConfig) bool {
	if config.ClosedLoopEnabled {
		return true
	}

	return parseEnvBool("FP_CRAWLER_ENABLED", false)
}

func shouldEnableWAF(config *gateway.GatewayConfig) bool {
	if config.ClosedLoopEnabled {
		return true
	}

	return parseEnvBool("FP_WAF_ENABLED", false)
}

func shouldAutostartCrawler(crawlerInstance *crawler.Crawler) bool {
	if crawlerInstance == nil {
		return false
	}

	if !parseEnvBool("FP_CRAWLER_AUTOSTART", false) {
		return false
	}

	crawlerConfig := crawlerInstance.ConfigSnapshot()

	return len(crawlerConfig.TargetURLs) > 0
}

func loadCrawlerConfig() *crawler.CrawlerConfig {
	base := *crawler.DefaultCrawlerConfig

	if value, ok := readEnvString("FP_CRAWLER_TARGETS"); ok {
		base.TargetURLs = splitCSV(value)
	}

	if parsed, ok := readEnvInt("FP_CRAWLER_WORKERS"); ok && parsed > 0 {
		base.Workers = parsed
	}

	if parsed, ok := readEnvInt("FP_CRAWLER_MAX_DEPTH"); ok && parsed >= 0 {
		base.MaxDepth = parsed
	}

	if parsed, ok := readEnvDuration("FP_CRAWLER_RATE_LIMIT"); ok && parsed > 0 {
		base.RateLimit = parsed
	}

	if value, ok := readEnvString("FP_CRAWLER_FEEDBACK_URL"); ok {
		base.FeedbackURL = value
	}

	return &base
}

func loadWAFConfig(config *gateway.GatewayConfig) *waf.WAFConfig {
	base := *waf.DefaultWAFConfig
	base.MLClassifierPath = config.MLClassifierPath
	base.MLEnabled = config.MLServiceEnabled
	base.RiskThreshold = config.RiskThreshold
	base.TrustedProxies = append([]string(nil), config.TrustedProxies...)

	if value, ok := readEnvString("FP_WAF_MODE"); ok {
		base.Mode = waf.WAFMode(value)
	}

	if parsed, ok := readEnvFloat("FP_WAF_RISK_THRESHOLD"); ok && parsed > 0 {
		base.RiskThreshold = parsed
	}

	if parsed, ok := readEnvInt("FP_WAF_RATE_LIMIT_RPS"); ok && parsed > 0 {
		base.RateLimitRPS = parsed
	}

	if parsed, ok := readEnvInt("FP_WAF_RATE_LIMIT_BURST"); ok && parsed > 0 {
		base.RateLimitBurst = parsed
	}

	return &base
}

func parseEnvBool(name string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
