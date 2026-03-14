package crawler

import (
	"math/rand"
	"time"

	"github.com/vistone/fingerprint/modules/profiles"
)

// initProfilePool - Initialize profile pool
func (c *Crawler) initProfilePool() {
	if len(c.config.ProfilePool) > 0 {
		for _, id := range c.config.ProfilePool {
			if p, ok := profiles.Get(id); ok {
				cp := p
				c.profileManager.profiles = append(c.profileManager.profiles, &cp)
			}
		}
	} else {
		// Use all available profiles from built-in configuration
		all := profiles.GetAll()
		for i := range all {
			c.profileManager.profiles = append(c.profileManager.profiles, &all[i])
		}
	}

	if len(c.profileManager.profiles) == 0 {
		// Use default profile
		p := profiles.GetRandom()
		c.profileManager.profiles = append(c.profileManager.profiles, &p)
	}

	c.logger.Info("profile pool initialized",
		"count", len(c.profileManager.profiles),
		"strategy", c.config.ProfileStrategy)
}

// getProfileForTask - Get profile for task
func (c *Crawler) getProfileForTask(task *crawlTask) *profiles.ClientProfile {
	c.profileMu.RLock()
	defer c.profileMu.RUnlock()

	pm := c.profileManager
	if len(pm.profiles) == 0 {
		return nil
	}

	switch pm.strategy {
	case ProfileStrategySticky:
		// Session persistence - same URL uses same profile
		if p, ok := pm.sessionMap[task.URL]; ok {
			return p
		}
		p := pm.profiles[rand.Intn(len(pm.profiles))]
		pm.sessionMap[task.URL] = p
		return p

	case ProfileStrategyRotate:
		pm.mu.Lock()
		p := pm.profiles[pm.current]
		pm.current = (pm.current + 1) % len(pm.profiles)
		pm.mu.Unlock()
		return p

	case ProfileStrategyAdaptive:
		// Adaptive selection - based on historical success rate
		return c.selectAdaptiveProfile()

	default: // ProfileStrategyRandom
		return pm.profiles[rand.Intn(len(pm.profiles))]
	}
}

// selectAdaptiveProfile - Select profile adaptively using ML-powered UCB1
func (c *Crawler) selectAdaptiveProfile() *profiles.ClientProfile {
	pm := c.profileManager
	if c.mlAdapter != nil {
		return c.mlAdapter.SelectBestProfile(pm.profiles)
	}
	// Fallback to random when ML is not available
	return pm.profiles[rand.Intn(len(pm.profiles))]
}

// profileRotator - Profile rotator
func (c *Crawler) profileRotator() {
	ticker := time.NewTicker(c.config.RotateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.rotateProfile()
		}
	}
}

// rotateProfile - Rotate profile
func (c *Crawler) rotateProfile() {
	c.profileMu.Lock()
	defer c.profileMu.Unlock()

	pm := c.profileManager
	if len(pm.profiles) == 0 {
		return
	}

	pm.current = (pm.current + 1) % len(pm.profiles)
	c.currentProfile = pm.profiles[pm.current]

	c.logger.Debug("profile rotated",
		"current", c.currentProfile.ID,
		"browser", c.currentProfile.BrowserType)
}
