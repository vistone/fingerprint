package main

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/vistone/fingerprint/modules/gateway/web"
)

// metricsMiddleware records request metrics for API handlers.
func metricsMiddleware(nextHandler http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		startTime := time.Now()

		recorder := &responseRecorder{
			ResponseWriter: writer,
			statusCode:     http.StatusOK,
		}

		nextHandler(recorder, request)

		latencyMilliseconds := time.Since(startTime).Milliseconds()
		success := recorder.statusCode < httpSuccessUpperBound
		clientIP := getClientIP(request)

		record := web.RequestRecord{
			Timestamp:      time.Now(),
			IP:             clientIP,
			Method:         request.Method,
			Path:           request.URL.Path,
			Classification: "",
			JA3:            request.Header.Get("X-Ja3-Fingerprint"),
			Latency:        latencyMilliseconds,
			Status:         recorder.statusCode,
		}

		web.RecordAPIMetrics(record, success, "", "")
	}
}

// getClientIP resolves client IP from forwarded headers or remote address.
func getClientIP(request *http.Request) string {
	if xForwardedFor := request.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		firstIP, _, hasComma := strings.Cut(xForwardedFor, ",")
		if hasComma {
			return strings.TrimSpace(firstIP)
		}

		return xForwardedFor
	}

	if xRealIP := request.Header.Get("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}

	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		return request.RemoteAddr
	}

	return host
}

// responseRecorder wraps http.ResponseWriter to capture status code.
type responseRecorder struct {
	http.ResponseWriter

	statusCode int
}

func (recorder *responseRecorder) WriteHeader(code int) {
	recorder.statusCode = code
	recorder.ResponseWriter.WriteHeader(code)
}
