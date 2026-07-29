package waf

import (
	"net"
	"net/http"
	"strings"
)

func extractClientIP(req *http.Request, trustedProxies []string) string {
	remoteIP := remoteAddrIP(req.RemoteAddr)
	if isTrustedProxy(remoteIP, trustedProxies) {
		if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
			if idx := strings.Index(xff, ","); idx != -1 {
				return strings.TrimSpace(xff[:idx])
			}
			return strings.TrimSpace(xff)
		}
		if xri := req.Header.Get("X-Real-IP"); xri != "" {
			return strings.TrimSpace(xri)
		}
	}
	return remoteIP
}

func remoteAddrIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil && host != "" {
		return host
	}
	return remoteAddr
}

func isTrustedProxy(remoteIP string, trustedProxies []string) bool {
	for _, candidate := range trustedProxies {
		if strings.TrimSpace(candidate) == remoteIP {
			return true
		}
	}
	return false
}
