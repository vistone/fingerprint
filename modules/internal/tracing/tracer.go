package tracing

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// translated comment
type TraceID string

// translated comment
type Span struct {
	TraceID   TraceID
	Name      string
	StartTime time.Time
	Duration  time.Duration
	Tags      map[string]interface{}
	mu        sync.RWMutex
}

// translated comment
func NewSpan(traceID TraceID, name string) *Span {
	return &Span{
		TraceID:   traceID,
		Name:      name,
		StartTime: time.Now(),
		Tags:      make(map[string]interface{}),
	}
}

// translated comment
func (s *Span) AddTag(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Tags[key] = value
}

// translated comment
func (s *Span) End() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Duration = time.Since(s.StartTime)
}

// translated comment
type Tracer struct {
	spans map[TraceID][]*Span
	mu    sync.RWMutex
}

// translated comment
func NewTracer() *Tracer {
	return &Tracer{
		spans: make(map[TraceID][]*Span),
	}
}

// translated comment
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

// translated comment
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

// translated comment
func GenerateTraceID() TraceID {
	// translated comment
	return TraceID(fmt.Sprintf("%016x", rand.Int63()))
}
