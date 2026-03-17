package waf

import (
	"net/http"
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/core"
)

// BehaviorEngine analyzes behavioral patterns
type BehaviorEngine struct {
	sessions *SessionManager
	window   time.Duration
}

// SessionManager manages client sessions
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// Session represents a client session
type Session struct {
	ID              string
	IP              string
	UserAgent       string
	RequestCount    int
	FirstRequest    time.Time
	LastRequest     time.Time
	RequestHistory  []RequestRecord
	SuspiciousScore float64
}

// RequestRecord stores a single request record
type RequestRecord struct {
	Timestamp time.Time
	Path      string
	Method    string
}

// BehaviorResult contains behavior analysis results
type BehaviorResult struct {
	Score          float64
	Factors        []core.RiskFactor
	RequestRate    float64 // requests per second
	BurstDetected  bool
	PatternAnomaly bool
}

// NewBehaviorEngine creates a new behavior analysis engine
func NewBehaviorEngine() *BehaviorEngine {
	return &BehaviorEngine{
		sessions: &SessionManager{
			sessions: make(map[string]*Session),
		},
		window: 1 * time.Minute,
	}
}

// Analyze performs behavior analysis for a client
func (e *BehaviorEngine) Analyze(clientID string, req *http.Request) *BehaviorResult {
	result := &BehaviorResult{
		Score:   0,
		Factors: make([]core.RiskFactor, 0),
	}

	now := time.Now()
	session := e.getOrCreateSession(clientID, req, now)
	e.recordRequest(session, req, now)
	result.RequestRate = calculateRequestRate(session, now)
	e.applyBehaviorRiskRules(result, session, now)
	return result
}

func (e *BehaviorEngine) getOrCreateSession(clientID string, req *http.Request, now time.Time) *Session {
	e.sessions.mu.Lock()
	defer e.sessions.mu.Unlock()

	session, exists := e.sessions.sessions[clientID]
	if !exists {
		session = &Session{
			ID:             clientID,
			IP:             clientID,
			UserAgent:      req.UserAgent(),
			FirstRequest:   now,
			RequestHistory: make([]RequestRecord, 0),
		}
		e.sessions.sessions[clientID] = session
	}
	return session
}

func (e *BehaviorEngine) recordRequest(session *Session, req *http.Request, now time.Time) {
	e.sessions.mu.Lock()
	defer e.sessions.mu.Unlock()

	session.RequestCount++
	session.LastRequest = now
	session.RequestHistory = append(session.RequestHistory, RequestRecord{
		Timestamp: now,
		Path:      req.URL.Path,
		Method:    req.Method,
	})

	// Trim old history (keep last 100 requests)
	if len(session.RequestHistory) > 100 {
		session.RequestHistory = session.RequestHistory[len(session.RequestHistory)-100:]
	}
}

func calculateRequestRate(session *Session, now time.Time) float64 {
	duration := now.Sub(session.FirstRequest).Seconds()
	if duration > 0 {
		return float64(session.RequestCount) / duration
	}
	return 0
}

func (e *BehaviorEngine) applyBehaviorRiskRules(result *BehaviorResult, session *Session, now time.Time) {
	if result.RequestRate > 10 {
		result.Score += 0.5
		result.BurstDetected = true
		result.Factors = append(result.Factors, core.RiskFactor{
			Name:        "high_request_rate",
			Weight:      0.5,
			Description: "High request rate detected",
		})
	}

	if countRecentRequests(session, now.Add(-e.window)) > 50 {
		result.Score += 0.4
		result.BurstDetected = true
		result.Factors = append(result.Factors, core.RiskFactor{
			Name:        "request_burst",
			Weight:      0.4,
			Description: "Request burst detected",
		})
	}

	if e.detectPathTraversal(session) {
		result.Score += 0.6
		result.PatternAnomaly = true
		result.Factors = append(result.Factors, core.RiskFactor{
			Name:        "path_traversal_pattern",
			Weight:      0.6,
			Description: "Suspicious path traversal pattern",
		})
	}

	if e.detectSequentialAccess(session) {
		result.Score += 0.3
		result.PatternAnomaly = true
		result.Factors = append(result.Factors, core.RiskFactor{
			Name:        "sequential_access",
			Weight:      0.3,
			Description: "Sequential access pattern (possible crawler)",
		})
	}
}

func countRecentRequests(session *Session, windowStart time.Time) int {
	recentCount := 0
	for _, record := range session.RequestHistory {
		if record.Timestamp.After(windowStart) {
			recentCount++
		}
	}
	return recentCount
}

func (e *BehaviorEngine) detectPathTraversal(session *Session) bool {
	// Check for common path traversal attempts
	suspiciousPaths := []string{"../", "..\\", "/etc/passwd", "/.env", "/config"}

	for _, record := range session.RequestHistory {
		for _, pattern := range suspiciousPaths {
			if contains(record.Path, pattern) {
				return true
			}
		}
	}
	return false
}

func (e *BehaviorEngine) detectSequentialAccess(session *Session) bool {
	if len(session.RequestHistory) < 10 {
		return false
	}

	// Check if requests follow sequential patterns like /item/1, /item/2, etc.
	sequentialCount := 0
	for i := 1; i < len(session.RequestHistory); i++ {
		prev := session.RequestHistory[i-1].Path
		curr := session.RequestHistory[i].Path
		if isSequential(prev, curr) {
			sequentialCount++
		}
	}

	// If more than 70% are sequential, likely a crawler
	return float64(sequentialCount)/float64(len(session.RequestHistory)-1) > 0.7
}

func isSequential(prev, curr string) bool {
	// Simple check for numeric increment
	// This is a basic implementation - real-world would need more sophisticated logic
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
