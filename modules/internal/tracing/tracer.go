package tracing

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// TraceID identifies a trace
type TraceID string

// Span represents a trace span
type Span struct {
	TraceID   TraceID
	Name      string
	StartTime time.Time
	Duration  time.Duration
	Tags      map[string]interface{}
	mu        sync.RWMutex
}

// NewSpan creates a new span
func NewSpan(traceID TraceID, name string) *Span {
	return &Span{
		TraceID:   traceID,
		Name:      name,
		StartTime: time.Now(),
		Tags:      make(map[string]interface{}),
	}
}

// AddTag adds a tag
func (s *Span) AddTag(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Tags[key] = value
}

// End closes the span
func (s *Span) End() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Duration = time.Since(s.StartTime)
}

// Tracer stores spans by trace ID
type Tracer struct {
	spans map[TraceID][]*Span
	mu    sync.RWMutex
}

// NewTracer creates a tracer
func NewTracer() *Tracer {
	return &Tracer{
		spans: make(map[TraceID][]*Span),
	}
}

// StartSpan starts and stores a span
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

// GetTrace returns spans of a trace
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

// GenerateTraceID creates a trace ID
func GenerateTraceID() TraceID {
	// Generate a 16-digit hexadecimal string
	return TraceID(fmt.Sprintf("%016x", rand.Int63()))
}
