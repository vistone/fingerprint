package tracing

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// TraceID 追踪 ID
type TraceID string

// Span 追踪 Span
type Span struct {
	TraceID   TraceID
	Name      string
	StartTime time.Time
	Duration  time.Duration
	Tags      map[string]interface{}
	mu        sync.RWMutex
}

// NewSpan 创建新 Span
func NewSpan(traceID TraceID, name string) *Span {
	return &Span{
		TraceID:   traceID,
		Name:      name,
		StartTime: time.Now(),
		Tags:      make(map[string]interface{}),
	}
}

// AddTag 添加标签
func (s *Span) AddTag(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Tags[key] = value
}

// End 结束 Span
func (s *Span) End() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Duration = time.Since(s.StartTime)
}

// Tracer 追踪器
type Tracer struct {
	spans map[TraceID][]*Span
	mu    sync.RWMutex
}

// NewTracer 创建追踪器
func NewTracer() *Tracer {
	return &Tracer{
		spans: make(map[TraceID][]*Span),
	}
}

// StartSpan 开始 Span
func (t *Tracer) StartSpan(traceID TraceID, name string) *Span {
	span := NewSpan(traceID, name)
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, exists := t.spans[traceID]; !exists {
		t.spans[traceID] = make([]*Span, 0)
	}
	t.spans[traceID] = append(t.spans[traceID], span)
	return span
}

// GetTrace 获取追踪
func (t *Tracer) GetTrace(traceID TraceID) []*Span {
	t.mu.RLock()
	defer t.mu.RUnlock()
	spans, exists := t.spans[traceID]
	if !exists {
		return make([]*Span, 0)
	}
	result := make([]*Span, len(spans))
	copy(result, spans)
	return result
}

// GenerateTraceID 生成追踪 ID
func GenerateTraceID() TraceID {
	// 生成16位十六进制字符串
	return TraceID(fmt.Sprintf("%016x", rand.Int63()))
}
