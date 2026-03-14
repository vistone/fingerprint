package crawler

import (
	"math/rand"
	"time"
)

// StealthEngine - Anti-detection engine
type StealthEngine struct {
	config *CrawlerConfig
}

// NewStealthEngine - Create anti-detection engine
func NewStealthEngine(config *CrawlerConfig) *StealthEngine {
	return &StealthEngine{config: config}
}

// HumanBehavior - Human behavior simulation
type HumanBehavior struct {
	MouseMovements []MouseMovement
	ScrollPattern  ScrollPattern
	TypingPattern  TypingPattern
}

// MouseMovement - Mouse movement
type MouseMovement struct {
	X, Y      int
	Timestamp time.Time
}

// ScrollPattern - Scroll pattern
type ScrollPattern struct {
	Intervals []time.Duration
	Distances []int
}

// TypingPattern - Typing pattern
type TypingPattern struct {
	Intervals []time.Duration
	Errors    int
}

// GenerateMousePath - Generate mouse movement path
func (s *StealthEngine) GenerateMousePath(fromX, fromY, toX, toY int) []MouseMovement {
	// Use Bezier curves to simulate human mouse movement
	movements := make([]MouseMovement, 0)

	distance := distance(fromX, fromY, toX, toY)
	steps := distance/10 + rand.Intn(5)

	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)

		// Add random jitter
		jitterX := rand.Intn(10) - 5
		jitterY := rand.Intn(10) - 5

		x := int(lerp(float64(fromX), float64(toX), t)) + jitterX
		y := int(lerp(float64(fromY), float64(toY), t)) + jitterY

		movements = append(movements, MouseMovement{
			X:         x,
			Y:         y,
			Timestamp: time.Now().Add(time.Duration(i) * time.Millisecond),
		})
	}

	return movements
}

// GenerateScrollPattern - Generate scroll pattern
func (s *StealthEngine) GenerateScrollPattern() ScrollPattern {
	pattern := ScrollPattern{
		Intervals: make([]time.Duration, 0),
		Distances: make([]int, 0),
	}

	// Simulate human scrolling behavior: fast-slow-pause-fast
	segments := []struct {
		count    int
		speed    time.Duration
		distance int
	}{
		{3, 100 * time.Millisecond, 300},
		{2, 200 * time.Millisecond, 150},
		{1, 500 * time.Millisecond, 0}, // Pause
		{2, 150 * time.Millisecond, 200},
	}

	for _, seg := range segments {
		for i := 0; i < seg.count; i++ {
			pattern.Intervals = append(pattern.Intervals,
				seg.speed+time.Duration(rand.Intn(50))*time.Millisecond)
			pattern.Distances = append(pattern.Distances,
				seg.distance+rand.Intn(50))
		}
	}

	return pattern
}

// GenerateTypingPattern - Generate typing pattern
func (s *StealthEngine) GenerateTypingPattern(text string) TypingPattern {
	pattern := TypingPattern{
		Intervals: make([]time.Duration, len(text)),
	}

	for i := range text {
		// Simulate typing speed variation
		baseDelay := 80 + rand.Intn(120) // 80-200ms

		// Occasionally add longer pauses (simulate thinking)
		if rand.Float32() < 0.1 {
			baseDelay += 500 + rand.Intn(1000)
		}

		pattern.Intervals[i] = time.Duration(baseDelay) * time.Millisecond
	}

	// Simulate occasional typing errors
	if len(text) > 10 && rand.Float32() < 0.2 {
		pattern.Errors = 1 + rand.Intn(2)
	}

	return pattern
}

// GetRandomViewport - Get random viewport
func (s *StealthEngine) GetRandomViewport() (width, height int) {
	viewports := []struct {
		w, h int
	}{
		{1920, 1080},
		{1366, 768},
		{1440, 900},
		{1536, 864},
		{1280, 720},
		{1680, 1050},
	}

	vp := viewports[rand.Intn(len(viewports))]
	return vp.w, vp.h
}

// GetRandomTimezone - Get random timezone
func (s *StealthEngine) GetRandomTimezone() string {
	timezones := []string{
		"America/New_York",
		"America/Los_Angeles",
		"Europe/London",
		"Europe/Paris",
		"Asia/Tokyo",
		"Asia/Shanghai",
		"Australia/Sydney",
	}
	return timezones[rand.Intn(len(timezones))]
}

// lerp - Linear interpolation
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// distance - Calculate distance between two points
func distance(x1, y1, x2, y2 int) int {
	dx := x2 - x1
	dy := y2 - y1
	return int(float64(dx*dx + dy*dy))
}
