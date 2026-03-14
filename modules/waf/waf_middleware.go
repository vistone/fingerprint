package waf

import (
	"fmt"
	"net/http"
)

// Middleware HTTP middleware
func (w *WAF) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		result := w.Analyze(req.Context(), req)

		switch result.Action {
		case ActionAllow:
			next.ServeHTTP(rw, req)

		case ActionMonitor:
			// Add monitoring headers
			rw.Header().Set("X-WAF-Monitor", "1")
			rw.Header().Set("X-WAF-Risk", fmt.Sprintf("%.2f", result.RiskScore))
			next.ServeHTTP(rw, req)

		case ActionThrottle:
			rw.WriteHeader(http.StatusTooManyRequests)
			rw.Write([]byte("Rate limit exceeded"))

		case ActionChallenge:
			w.serveChallenge(rw, req, result)

		case ActionBlock:
			w.serveBlock(rw, req, result)
		}
	})
}
