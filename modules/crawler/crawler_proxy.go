package crawler

import (
	"math/rand"
	"net/url"
)

// initProxyPool - Initialize proxy pool
func (c *Crawler) initProxyPool() {
	for _, proxyStr := range c.config.ProxyList {
		if u, err := url.Parse(proxyStr); err == nil {
			c.proxyManager.proxies = append(c.proxyManager.proxies, u)
		}
	}
}

// getProxy - Get proxy
func (c *Crawler) getProxy() *url.URL {
	if c.proxyManager == nil || len(c.proxyManager.proxies) == 0 {
		return nil
	}

	pm := c.proxyManager
	pm.mu.Lock()
	defer pm.mu.Unlock()

	switch pm.strategy {
	case ProxyStrategyRotate:
		p := pm.proxies[pm.current]
		pm.current = (pm.current + 1) % len(pm.proxies)
		return p
	default:
		return pm.proxies[rand.Intn(len(pm.proxies))]
	}
}
