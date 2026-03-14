package crawler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// FeedbackCollector - Data feedback collector
type FeedbackCollector struct {
	endpoint    string
	buffer      []*CrawlResult
	mu          sync.RWMutex
	batchSize   int
	flushTicker *time.Ticker
	stopChan    chan struct{}
}

// FeedbackEntry - Data structure for feedback transmission
type FeedbackEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	URL         string                 `json:"url"`
	Blocked     bool                   `json:"blocked"`
	BlockReason string                 `json:"block_reason,omitempty"`
	Fingerprint string                 `json:"fingerprint_id"`
	ProxyUsed   string                 `json:"proxy_used,omitempty"`
	Detection   map[string]interface{} `json:"detection_info"`
	Headers     map[string]string      `json:"response_headers,omitempty"`
	ContentType string                 `json:"content_type,omitempty"`
}

// NewFeedbackCollector - Create feedback collector
func NewFeedbackCollector(endpoint string) *FeedbackCollector {
	fc := &FeedbackCollector{
		endpoint:  endpoint,
		buffer:    make([]*CrawlResult, 0),
		batchSize: 10,
		stopChan:  make(chan struct{}),
	}

	// Start periodic flush
	if endpoint != "" {
		fc.flushTicker = time.NewTicker(30 * time.Second)
		go fc.flushLoop()
	}

	return fc
}

// Collect - Collect crawl results
func (fc *FeedbackCollector) Collect(result *CrawlResult) {
	fc.mu.Lock()
	fc.buffer = append(fc.buffer, result)
	shouldFlush := len(fc.buffer) >= fc.batchSize
	fc.mu.Unlock()

	if shouldFlush {
		fc.Flush()
	}
}

// flushLoop - Periodic flush loop
func (fc *FeedbackCollector) flushLoop() {
	for {
		select {
		case <-fc.flushTicker.C:
			fc.Flush()
		case <-fc.stopChan:
			return
		}
	}
}

// Flush - Flush buffer immediately
func (fc *FeedbackCollector) Flush() error {
	fc.mu.Lock()
	if len(fc.buffer) == 0 {
		fc.mu.Unlock()
		return nil
	}

	batch := make([]*CrawlResult, len(fc.buffer))
	copy(batch, fc.buffer)
	fc.buffer = fc.buffer[:0]
	fc.mu.Unlock()

	// Convert to feedback entries
	entries := make([]FeedbackEntry, 0, len(batch))
	for _, r := range batch {
		entry := FeedbackEntry{
			Timestamp:   r.Timestamp,
			URL:         r.URL,
			Blocked:     r.Blocked,
			BlockReason: r.BlockReason,
			ProxyUsed:   r.ProxyUsed,
			Detection:   r.DetectionInfo,
			ContentType: r.ContentType,
		}

		if r.Fingerprint != nil {
			entry.Fingerprint = r.Fingerprint.ID
		}

		// Simplify response headers
		entry.Headers = make(map[string]string)
		for k, v := range r.Headers {
			if len(v) > 0 {
				entry.Headers[k] = v[0]
			}
		}

		entries = append(entries, entry)
	}

	// Send data
	return fc.send(entries)
}

// send - Send data to feedback endpoint
func (fc *FeedbackCollector) send(entries []FeedbackEntry) error {
	if fc.endpoint == "" {
		return nil
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	resp, err := http.Post(fc.endpoint, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Close - Close collector
func (fc *FeedbackCollector) Close() {
	if fc.flushTicker != nil {
		fc.flushTicker.Stop()
	}
	close(fc.stopChan)
	fc.Flush()
}
