package web

import (
	"strings"
	"time"

	"github.com/vistone/fingerprint/modules/gateway"
	"github.com/vistone/fingerprint/modules/profiles"
)

func loadProfiles() []profiles.ClientProfile {
	// 从 profiles 模块加载所有已注册的指纹配置
	return profiles.GetAll()
}

func filterProfiles(profiles []profiles.ClientProfile, query, browser, os string) []map[string]interface{} {
	var result []map[string]interface{}

	for _, p := range profiles {
		// Apply filters
		if query != "" && !strings.Contains(strings.ToLower(p.Name), query) {
			continue
		}

		if browser != "" && !strings.Contains(strings.ToLower(string(p.BrowserType)), browser) {
			continue
		}

		if os != "" && !strings.Contains(strings.ToLower(string(p.OS)), os) {
			continue
		}

		profileData := map[string]interface{}{
			"id":             p.ID,
			"name":           p.Name,
			"browserType":    p.BrowserType,
			"browserVersion": p.BrowserVersion,
			"os":             p.OS,
			"osVersion":      p.OSVersion,
			"tlsVersion":     p.TLSVersion,
			"cipherSuites":   len(p.CipherSuites),
			"extensions":     len(p.Extensions),
		}

		// 添加 TCP/IP 指纹简要信息
		if p.TCPIP != nil {
			profileData["tcpip"] = map[string]interface{}{
				"ttl":        p.TCPIP.TTL,
				"windowSize": p.TCPIP.WindowSize,
				"mss":        p.TCPIP.MSS,
				"ja4t":       p.TCPIP.JA4T,
			}
		}

		result = append(result, profileData)
	}

	return result
}

func getBrowserDistribution(profiles []profiles.ClientProfile) map[string]int {
	distribution := make(map[string]int)

	for _, p := range profiles {
		browser := string(p.BrowserType)
		distribution[browser]++
	}

	// If empty, return sample data
	if len(distribution) == 0 {
		return map[string]int{
			"Chrome":  47,
			"Firefox": 30,
			"Safari":  38,
			"Edge":    10,
			"Opera":   7,
		}
	}

	return distribution
}

func getOSDistribution(profiles []profiles.ClientProfile) map[string]int {
	distribution := make(map[string]int)

	for _, p := range profiles {
		osStr := string(p.OS)
		// 简化 OS 名称
		var group string
		switch {
		case strings.Contains(osStr, "Windows"):
			group = "Windows"
		case strings.Contains(osStr, "Mac OS") || strings.Contains(osStr, "Macintosh"):
			group = "macOS"
		case strings.Contains(osStr, "Android"):
			group = "Android"
		case strings.Contains(osStr, "iPhone") || strings.Contains(osStr, "iPad") || strings.Contains(osStr, "iOS"):
			group = "iOS"
		case strings.Contains(osStr, "Linux") || strings.Contains(osStr, "X11"):
			group = "Linux"
		default:
			group = "Other"
		}
		distribution[group]++
	}

	// If empty, return sample data
	if len(distribution) == 0 {
		return map[string]int{
			"Windows": 45,
			"macOS":   35,
			"Linux":   20,
			"iOS":     15,
			"Android": 25,
		}
	}

	return distribution
}

func getTopFingerprints() []map[string]interface{} {
	return []map[string]interface{}{
		{"hash": "7692c8d76c4f0e4a9c9c8a7b6c5d4e3f", "count": 12543, "percentage": 15.2},
		{"hash": "a1b2c3d4e5f678901234567890123456", "count": 9876, "percentage": 12.1},
		{"hash": "9876543210fedcba9876543210fedcb", "count": 7654, "percentage": 9.8},
		{"hash": "fedcba0987654321fedcba098765432", "count": 5432, "percentage": 7.3},
		{"hash": "11223344556677889900aabbccddeeff", "count": 4321, "percentage": 5.9},
	}
}

func getTrafficData() map[string]interface{} {
	return map[string]interface{}{
		"labels": []string{"00:00", "04:00", "08:00", "12:00", "16:00", "20:00"},
		"data":   []int{120, 80, 250, 380, 420, 350},
	}
}

func getTCPIPDistribution(profiles []profiles.ClientProfile) map[string]int {
	distribution := make(map[string]int)

	for _, p := range profiles {
		if p.TCPIP == nil {
			distribution["Unknown"]++
			continue
		}
		// 根据 TTL 和 Window Size 判断 OS 类型
		ttl := p.TCPIP.TTL
		ws := p.TCPIP.WindowSize

		switch {
		case ttl == 128:
			distribution["Windows"]++
		case ttl == 64 && ws == 65535:
			distribution["macOS/iOS"]++
		case ttl == 64 && ws == 64240:
			distribution["Linux"]++
		default:
			distribution["Other"]++
		}
	}

	return distribution
}

// applyConfigUpdate applies a config update from the admin console
func (h *Handler) applyConfigUpdate(newConfig map[string]interface{}) {
	h.gateway.UpdateConfig(func(cfg *gateway.GatewayConfig) {
		if rl, ok := newConfig["rateLimit"].(map[string]interface{}); ok {
			if v, ok := rl["rps"].(float64); ok {
				cfg.RateLimitRequests = int(v)
			}
			if v, ok := rl["burstSize"].(float64); ok {
				cfg.RateLimitBurst = int(v)
			}
		}
		if cache, ok := newConfig["cache"].(map[string]interface{}); ok {
			if v, ok := cache["enabled"].(bool); ok {
				cfg.CacheEnabled = v
			}
			if v, ok := cache["size"].(float64); ok {
				cfg.CacheSize = int(v)
			}
			if v, ok := cache["ttl"].(float64); ok {
				cfg.CacheTTL = time.Duration(v) * time.Minute
			}
		}
		if p3, ok := newConfig["p3"].(map[string]interface{}); ok {
			if v, ok := p3["enabled"].(bool); ok {
				cfg.P3Enabled = v
			}
			if v, ok := p3["profileId"].(string); ok {
				cfg.P3ProfileID = v
			}
			if v, ok := p3["proxyTarget"].(string); ok {
				cfg.P3ProxyTarget = v
			}
			if v, ok := p3["directProxy"].(bool); ok {
				cfg.P3DirectProxy = v
			}
			if v, ok := p3["injectConsist"].(bool); ok {
				cfg.P3InjectConsist = v
			}
		}
		if scanner, ok := newConfig["scanner"].(map[string]interface{}); ok {
			if v, ok := scanner["useBrowser"].(bool); ok {
				cfg.ScannerUseBrowser = v
			}
			if v, ok := scanner["browserWS"].(string); ok {
				cfg.ScannerBrowserWS = v
			}
			if v, ok := scanner["browserTimeout"].(float64); ok {
				cfg.ScannerBrowserTimeout = time.Duration(v) * time.Second
			}
		}
		if ag, ok := newConfig["agent"].(map[string]interface{}); ok {
			if v, ok := ag["enabled"].(bool); ok {
				cfg.AgentEnabled = v
			}
		}
		if ml, ok := newConfig["ml"].(map[string]interface{}); ok {
			if v, ok := ml["riskThreshold"].(float64); ok {
				cfg.RiskThreshold = v
			}
		}
	})
}

// handleLogStream SSE 实时日志推流
