package waf

import (
	"net"
	"net/http"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
)

// NetworkEngine analyzes network layer characteristics
type NetworkEngine struct {
	ipReputation *IPReputationTable
	geoDB        *GeoDatabase
}

// IPReputationTable stores IP reputation data
type IPReputationTable struct {
	blacklist map[string]bool
	greylist  map[string]int
}

// GeoDatabase provides geo-location data
type GeoDatabase struct {
	enabled bool
}

// NetworkResult contains network analysis results
type NetworkResult struct {
	Score   float64
	Factors []core.RiskFactor
	Country string
	IsProxy bool
	IsVPN   bool
	IsTor   bool
}

// NewNetworkEngine creates a new network analysis engine
func NewNetworkEngine() *NetworkEngine {
	return &NetworkEngine{
		ipReputation: &IPReputationTable{
			blacklist: make(map[string]bool),
			greylist:  make(map[string]int),
		},
		geoDB: &GeoDatabase{enabled: false},
	}
}

// Analyze performs network layer analysis
func (e *NetworkEngine) Analyze(req *http.Request) *NetworkResult {
	result := &NetworkResult{
		Score:   0,
		Factors: make([]core.RiskFactor, 0),
	}

	clientIP := getClientIP(req)

	// Check IP reputation
	if e.ipReputation.blacklist[clientIP] {
		result.Score += 1.0
		result.Factors = append(result.Factors, core.RiskFactor{
			Name:        "blacklisted_ip",
			Weight:      1.0,
			Description: "IP is in blacklist",
		})
		return result
	}

	// Check if IP is private/internal
	if isPrivateIP(clientIP) {
		// Internal IPs might be legitimate depending on deployment
		result.Factors = append(result.Factors, core.RiskFactor{
			Name:        "private_ip",
			Weight:      0.1,
			Description: "Private IP address",
		})
	}

	// Check greylist (previous suspicious activity)
	if count, ok := e.ipReputation.greylist[clientIP]; ok && count > 5 {
		result.Score += float64(count) * 0.1
		result.Factors = append(result.Factors, core.RiskFactor{
			Name:        "greylisted_ip",
			Weight:      float64(count) * 0.1,
			Description: "IP has suspicious history",
		})
	}

	return result
}

// BlockIP adds an IP to the blacklist
func (e *NetworkEngine) BlockIP(ip string) {
	e.ipReputation.blacklist[ip] = true
}

// GreylistIP increments the greylist counter for an IP
func (e *NetworkEngine) GreylistIP(ip string) {
	e.ipReputation.greylist[ip]++
}

func getClientIP(req *http.Request) string {
	// Check X-Forwarded-For header
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Use RemoteAddr
	host, _, _ := net.SplitHostPort(req.RemoteAddr)
	if host != "" {
		return host
	}

	return req.RemoteAddr
}

func isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}

	for _, cidr := range privateRanges {
		_, ipnet, _ := net.ParseCIDR(cidr)
		if ipnet != nil && ipnet.Contains(parsedIP) {
			return true
		}
	}

	return false
}
