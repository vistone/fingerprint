package crawler

// Running reports whether the crawler worker loop is active.
func (c *Crawler) Running() bool {
	return c.running.Load()
}

// ConfigSnapshot returns a copy of the crawler configuration.
func (c *Crawler) ConfigSnapshot() CrawlerConfig {
	if c == nil || c.config == nil {
		return CrawlerConfig{}
	}
	return *c.config
}
