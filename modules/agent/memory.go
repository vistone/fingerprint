package agent

import (
	"sync"
	"time"
)

// Memory 智能体记忆系统
// 存储各客户端的观测历史，支持时间窗口内的快速查询和过期清理。
type Memory struct {
	sessions map[string]*ClientSession
	window   time.Duration
	maxObs   int
	mu       sync.RWMutex

	// 全局计数
	totalObs int
}

// ClientSession 单个客户端的观测会话
type ClientSession struct {
	ClientID     string
	Observations []*Observation
	LastActive   time.Time

	// 缓存的指纹集合（避免重复遍历）
	fingerprintSet map[string]struct{}
}

// NewMemory 创建记忆系统
func NewMemory(window time.Duration, maxObs int) *Memory {
	return &Memory{
		sessions: make(map[string]*ClientSession),
		window:   window,
		maxObs:   maxObs,
	}
}

// Record 记录一条观测
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

	// 超过上限时丢弃旧观测
	if len(session.Observations) > m.maxObs {
		excess := len(session.Observations) - m.maxObs
		session.Observations = session.Observations[excess:]
	}
}

// GetSession 获取客户端会话（只读快照）
func (m *Memory) GetSession(clientID string) *ClientSession {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[clientID]
}

// SessionCount 活跃会话数
func (m *Memory) SessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// TotalObservations 总观测数
func (m *Memory) TotalObservations() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalObs
}

// Cleanup 清理超时会话
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

// RecentObservations 获取指定客户端在时间窗口内的观测
func (m *Memory) RecentObservations(clientID string, window time.Duration) []*Observation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[clientID]
	if !ok {
		return nil
	}

	cutoff := time.Now().Add(-window)
	var result []*Observation
	// 从后向前遍历（最新在后）
	for i := len(session.Observations) - 1; i >= 0; i-- {
		if session.Observations[i].Timestamp.Before(cutoff) {
			break
		}
		result = append(result, session.Observations[i])
	}
	return result
}

// AllSessions 返回所有活跃会话的客户端 ID 列表（用于策略演化遍历）
func (m *Memory) AllSessions() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	return ids
}
