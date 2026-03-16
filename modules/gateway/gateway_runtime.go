package gateway

import (
	"github.com/vistone/fingerprint/modules/crawler"
	"github.com/vistone/fingerprint/modules/waf"
)

// GetCrawler returns the integrated crawler runtime, if configured.
func (g *Gateway) GetCrawler() *crawler.Crawler {
	return g.crawler
}

// GetWAF returns the integrated WAF runtime, if configured.
func (g *Gateway) GetWAF() *waf.WAF {
	return g.waf
}
