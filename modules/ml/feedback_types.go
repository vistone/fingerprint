package ml

import (
	"time"

	"github.com/vistone/fingerprint/modules/profiles"
)

// FeedbackSource identifies where a feedback sample originated.
type FeedbackSource string

const (
	FeedbackSourceCrawler FeedbackSource = "crawler"
	FeedbackSourceWAF     FeedbackSource = "waf"
	FeedbackSourceGateway FeedbackSource = "gateway"
)

// CrawlerFeedback represents a feedback event from the crawler subsystem.
type CrawlerFeedback struct {
	ProfileID     string
	Profile       *profiles.ClientProfile
	URL           string
	Success       bool
	Blocked       bool
	BlockReason   string
	DetectionInfo map[string]interface{}
	Duration      time.Duration
	Timestamp     time.Time
}

// ToFeedbackSample converts a CrawlerFeedback to a ML FeedbackSample.
func (cf *CrawlerFeedback) ToFeedbackSample() *FeedbackSample {
	reward := 1.0
	if cf.Blocked {
		reward = 0.0
	} else if !cf.Success {
		reward = 0.3
	}

	return &FeedbackSample{
		Profile:   cf.Profile,
		Reward:    reward,
		Label:     cf.ProfileID,
		Timestamp: cf.Timestamp,
	}
}

// WAFDetectionFeedback represents a feedback event from the WAF subsystem.
type WAFDetectionFeedback struct {
	ClientIP        string
	RiskScore       float64
	DetectionLayers []string
	Blocked         bool
	FingerprintID   string
	AntiBot         string // detected anti-bot system identifier
	Timestamp       time.Time
}

// ToFeedbackSample converts a WAFDetectionFeedback to a ML FeedbackSample.
// High risk detected = the fingerprint was detectable (low reward).
// Low risk = the fingerprint looked legitimate (high reward).
func (wf *WAFDetectionFeedback) ToFeedbackSample() *FeedbackSample {
	reward := 1.0 - wf.RiskScore
	if wf.Blocked {
		reward = 0.0
	}

	return &FeedbackSample{
		Reward:    reward,
		Label:     wf.FingerprintID,
		Timestamp: wf.Timestamp,
	}
}
