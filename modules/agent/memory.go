package agent

import (
	"sync"
	"time"
)

// Memory agent memory system
// Stores observation history for each client, supports fast queries and expiration cleanup within time window.
type Memory struct {
	sessions map[string]*ClientSession
	window   time.Duration
	maxObs   int
	mu       sync.RWMutex

	// Global count
	totalObs int
}

// ClientSession observation session for a single client
type ClientSession struct {
	ClientID     string
	Observations []*Observation
	LastActive   time.Time

	// Cached fingerprint set (avoid repeated traversal)
	fingerprintSet map[string]struct{}
}

// NewMemory create memory system
func NewMemory(window time.Duration, maxObs int) *Memory {
	return &Memory{
		sessions: make(map[string]*ClientSession),
		window:   window,
		maxObs:   maxObs,
	}
}

// Record log an observation
func (m *Memory) Record(obs *Observation) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[obs.ClientID]
	if !ok {
		session = &ClientSession{
			ClientID:       obs.ClientID,
			Observations:   make([]*Observation, 0, 64),
			fingerprintSet: make(map[string]struct{}),
		}
		m.sessions[obs.ClientID] = session
	}

	session.Observations = append(session.Observations, obs)
	session.LastActive = obs.Timestamp
	session.fingerprintSet[obs.FingerprintHash] = struct{}{}
	m.totalObs++

	// Discard old observations when exceeding limit
	if len(session.Observations) > m.maxObs {
		excess := len(session.Observations) - m.maxObs
		session.Observations = session.Observations[excess:]
	}
}

// GetSession get client session (read-only snapshot)
func (m *Memory) GetSession(clientID string) *ClientSession {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[clientID]
}

// SessionCount active session count
func (m *Memory) SessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// TotalObservations total observation count
func (m *Memory) TotalObservations() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalObs
}

// Cleanup cleanup timed-out sessions
func (m *Memory) Cleanup(timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, session := range m.sessions {
		if now.Sub(session.LastActive) > timeout {
			m.totalObs -= len(session.Observations)
			delete(m.sessions, id)
		}
	}
}

// RecentObservations get observations of specified client within time window
func (m *Memory) RecentObservations(clientID string, window time.Duration) []*Observation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[clientID]
	if !ok {
		return nil
	}

	cutoff := time.Now().Add(-window)
	var result []*Observation
	// Traverse backwards (newest last)
	for i := len(session.Observations) - 1; i >= 0; i-- {
		if session.Observations[i].Timestamp.Before(cutoff) {
			break
		}
		result = append(result, session.Observations[i])
	}
	return result
}

// AllSessions return list of client IDs for all active sessions (for strategy evolution traversal)
func (m *Memory) AllSessions() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	return ids
}
