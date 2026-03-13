// Package ml — evolution.go provides a profile evolution system that
// tracks real-world traffic patterns, detects outdated profiles, and
// automatically triggers model retraining.
//
// The ProfileEvolutionEngine monitors:
//   - Browser version drift (new Chrome versions appearing faster than profiles update)
//   - Traffic distribution shift (e.g., Chrome share increasing, IE decreasing)
//   - Forgery pattern changes (new anti-detect tools appearing)
//   - Profile staleness (profiles not seen in real traffic for a long time)
//
// When significant drift is detected, it orchestrates profile registry
// updates and model evolution.
package ml

import (
	"sort"
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/profiles"
)

// =========================================================================
// EvolutionConfig
// =========================================================================

// EvolutionConfig controls the profile evolution system.
type EvolutionConfig struct {
	// StalenessDays: profiles not seen in traffic for this many days
	// are flagged as stale.
	StalenessDays int

	// MinObservations: minimum traffic observations before making
	// evolution decisions.
	MinObservations int

	// DistributionDriftThreshold: KL divergence above this triggers
	// distribution-based evolution.
	DistributionDriftThreshold float64

	// VersionDriftWindow: how far back to look for version gaps.
	VersionDriftWindow time.Duration

	// MaxProfileAge: profiles older than this are considered for retirement.
	MaxProfileAge time.Duration
}

// DefaultEvolutionConfig provides sensible defaults.
var DefaultEvolutionConfig = &EvolutionConfig{
	StalenessDays:              30,
	MinObservations:            100,
	DistributionDriftThreshold: 0.1,
	VersionDriftWindow:         7 * 24 * time.Hour,
	MaxProfileAge:              180 * 24 * time.Hour,
}

// =========================================================================
// ProfileEvolutionEngine
// =========================================================================

// ProfileEvolutionEngine monitors profile health and triggers evolution.
type ProfileEvolutionEngine struct {
	config *EvolutionConfig
	mu     sync.RWMutex

	// Per-profile tracking
	profileStats map[string]*ProfileStat

	// Browser distribution tracker
	distribution *BrowserDistribution

	// Reference distribution (snapshot at startup or last evolution)
	referenceDistribution map[string]float64

	// Evolution history
	events []EvolutionEvent
}

// ProfileStat tracks the health of a single profile.
type ProfileStat struct {
	ProfileID     string
	BrowserType   string
	LastSeen      time.Time
	FirstSeen     time.Time
	HitCount      int64
	ForgeryRate   float64 // rolling average of forgery probability
	AvgConfidence float64 // rolling average of classification confidence
	IsStale       bool
}

// EvolutionEvent records a single evolution trigger.
type EvolutionEvent struct {
	Timestamp       time.Time
	Trigger         string // "drift", "staleness", "version_gap", "manual"
	Description     string
	ProfilesAdded   int
	ProfilesRetired int
	MetricsBefore   *TrainingMetrics
	MetricsAfter    *TrainingMetrics
}

// NewProfileEvolutionEngine creates a new evolution engine.
func NewProfileEvolutionEngine(config *EvolutionConfig) *ProfileEvolutionEngine {
	if config == nil {
		config = DefaultEvolutionConfig
	}
	return &ProfileEvolutionEngine{
		config:       config,
		profileStats: make(map[string]*ProfileStat),
		distribution: NewBrowserDistribution(),
	}
}

// RecordObservation records a fingerprint observation from real traffic.
func (e *ProfileEvolutionEngine) RecordObservation(profileID string, browserType string, result *PipelineResult) {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()

	stat, ok := e.profileStats[profileID]
	if !ok {
		stat = &ProfileStat{
			ProfileID:   profileID,
			BrowserType: browserType,
			FirstSeen:   now,
		}
		e.profileStats[profileID] = stat
	}

	stat.LastSeen = now
	stat.HitCount++

	if result != nil {
		// Exponential moving average for forgery rate.
		alpha := 0.1
		stat.ForgeryRate = stat.ForgeryRate*(1-alpha) + result.Forgery.ForgeryProb*alpha
		stat.AvgConfidence = stat.AvgConfidence*(1-alpha) + result.Browser.Confidence*alpha
	}

	e.distribution.Record(browserType)
}

// CheckHealth evaluates the overall profile ecosystem health.
func (e *ProfileEvolutionEngine) CheckHealth() *EvolutionHealthReport {
	e.mu.RLock()
	defer e.mu.RUnlock()

	report := &EvolutionHealthReport{
		Timestamp:     time.Now(),
		TotalProfiles: len(e.profileStats),
	}

	now := time.Now()
	staleThreshold := time.Duration(e.config.StalenessDays) * 24 * time.Hour

	for _, stat := range e.profileStats {
		if now.Sub(stat.LastSeen) > staleThreshold {
			stat.IsStale = true
			report.StaleProfiles++
			report.StaleProfileIDs = append(report.StaleProfileIDs, stat.ProfileID)
		}

		if stat.ForgeryRate > 0.5 {
			report.HighForgeryProfiles++
			report.HighForgeryProfileIDs = append(report.HighForgeryProfileIDs, stat.ProfileID)
		}

		if stat.AvgConfidence < 0.3 && stat.HitCount > 10 {
			report.LowConfidenceProfiles++
		}
	}

	// Check distribution drift.
	if e.referenceDistribution != nil && e.distribution.Total() >= int64(e.config.MinObservations) {
		report.DistributionDrift = e.distribution.KLDivergence(e.referenceDistribution)
		report.DriftDetected = report.DistributionDrift > e.config.DistributionDriftThreshold
	}

	// Count active profiles.
	report.ActiveProfiles = report.TotalProfiles - report.StaleProfiles

	return report
}

// EvolutionHealthReport summarises the profile ecosystem health.
type EvolutionHealthReport struct {
	Timestamp             time.Time
	TotalProfiles         int
	ActiveProfiles        int
	StaleProfiles         int
	StaleProfileIDs       []string
	HighForgeryProfiles   int
	HighForgeryProfileIDs []string
	LowConfidenceProfiles int
	DistributionDrift     float64
	DriftDetected         bool
	NeedsEvolution        bool
	Reasons               []string
}

// NeedsEvolution returns whether automatic evolution should be triggered.
func (r *EvolutionHealthReport) ShouldEvolve() bool {
	if r.DriftDetected {
		r.NeedsEvolution = true
		r.Reasons = append(r.Reasons, "distribution drift detected")
	}
	if r.StaleProfiles > r.TotalProfiles/4 {
		r.NeedsEvolution = true
		r.Reasons = append(r.Reasons, "too many stale profiles")
	}
	if r.HighForgeryProfiles > r.TotalProfiles/10 {
		r.NeedsEvolution = true
		r.Reasons = append(r.Reasons, "many profiles flagged as forgery")
	}
	return r.NeedsEvolution
}

// SnapshotDistribution saves the current distribution as reference for
// future drift detection.
func (e *ProfileEvolutionEngine) SnapshotDistribution() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.referenceDistribution = e.distribution.Distribution()
}

// TopStaleProfiles returns up to n stale profiles sorted by staleness.
func (e *ProfileEvolutionEngine) TopStaleProfiles(n int) []ProfileStat {
	e.mu.RLock()
	defer e.mu.RUnlock()

	now := time.Now()
	staleThreshold := time.Duration(e.config.StalenessDays) * 24 * time.Hour

	var stale []ProfileStat
	for _, stat := range e.profileStats {
		if now.Sub(stat.LastSeen) > staleThreshold {
			stale = append(stale, *stat)
		}
	}

	sort.Slice(stale, func(i, j int) bool {
		return stale[i].LastSeen.Before(stale[j].LastSeen)
	})

	if n > 0 && len(stale) > n {
		stale = stale[:n]
	}

	return stale
}

// TopForgeryProfiles returns profiles with the highest forgery rates.
func (e *ProfileEvolutionEngine) TopForgeryProfiles(n int) []ProfileStat {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var all []ProfileStat
	for _, stat := range e.profileStats {
		all = append(all, *stat)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].ForgeryRate > all[j].ForgeryRate
	})

	if n > 0 && len(all) > n {
		all = all[:n]
	}
	return all
}

// InitFromRegistry initialises profile stats from the current registry.
func (e *ProfileEvolutionEngine) InitFromRegistry(registry *profiles.ProfileRegistry) {
	e.mu.Lock()
	defer e.mu.Unlock()

	allProfiles := registry.GetAll()
	now := time.Now()

	for _, p := range allProfiles {
		if _, ok := e.profileStats[p.ID]; !ok {
			e.profileStats[p.ID] = &ProfileStat{
				ProfileID:   p.ID,
				BrowserType: string(p.BrowserType),
				FirstSeen:   now,
				LastSeen:    now,
			}
		}
	}
}

// RecordEvent appends an evolution event to the history log.
func (e *ProfileEvolutionEngine) RecordEvent(event EvolutionEvent) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.events = append(e.events, event)
}

// Events returns the evolution event history.
func (e *ProfileEvolutionEngine) Events() []EvolutionEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	result := make([]EvolutionEvent, len(e.events))
	copy(result, e.events)
	return result
}

// ProfileCount returns the number of tracked profiles.
func (e *ProfileEvolutionEngine) ProfileCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.profileStats)
}
