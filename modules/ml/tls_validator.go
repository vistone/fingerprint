// Package ml — tls_validator.go provides ML-guided TLS fingerprint validation.
//
// The TLSValidator uses the trained forgery detector and cross-layer
// consistency checks to evaluate whether a TLS ClientHello configuration
// looks realistic for its claimed browser.
//
// This is used by the generator to reject unrealistic TLS configurations
// and can be called independently to validate any ClientHello before use.
package ml

import (
	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// =========================================================================
// TLSValidator
// =========================================================================

// TLSValidator validates TLS configurations using ML models.
type TLSValidator struct {
	pipeline *ModelPipeline
}

// NewTLSValidator creates a new TLS validator backed by the given pipeline.
func NewTLSValidator(pipeline *ModelPipeline) *TLSValidator {
	return &TLSValidator{pipeline: pipeline}
}

// TLSValidationResult contains the ML validation of a TLS configuration.
type TLSValidationResult struct {
	// Valid indicates whether the TLS configuration looks realistic.
	Valid bool

	// DetectedBrowser is the browser the ML model thinks this TLS config belongs to.
	DetectedBrowser string

	// Confidence is the classification confidence [0,1].
	Confidence float64

	// ForgeryProb is the probability this is a forged TLS fingerprint [0,1].
	ForgeryProb float64

	// ConsistencyScore is how consistent the TLS parameters are [0,1].
	ConsistencyScore float64

	// CipherSuiteScore rates the cipher suite selection quality [0,1].
	CipherSuiteScore float64

	// ExtensionScore rates the extension configuration quality [0,1].
	ExtensionScore float64

	// Issues lists specific problems found.
	Issues []string
}

// ValidateProfile validates the TLS aspects of a full client profile.
func (v *TLSValidator) ValidateProfile(profile *profiles.ClientProfile) *TLSValidationResult {
	features := EncodeFingerprint(profile)
	crossFeatures := ComputeCrossLayerFeatures(features)
	embedding := v.pipeline.encoder.EncodeSingle(features)
	browser := v.pipeline.classifier.ClassifySingle(embedding)
	forgery := v.pipeline.detector.DetectSingle(features, crossFeatures)

	result := &TLSValidationResult{
		DetectedBrowser:  string(browser.Family),
		Confidence:       browser.Confidence,
		ForgeryProb:      forgery.ForgeryProb,
		ConsistencyScore: crossLayerConsistencyScore(crossFeatures),
	}

	// Evaluate TLS-specific features from the feature vector.
	// Features[0..7] are TLS features.
	if len(features) >= 8 {
		result.CipherSuiteScore = 1.0 - features[1]*0.5 // cipher count normalized
		result.ExtensionScore = 1.0 - features[3]*0.5   // extension anomaly
	}

	// Validate.
	result.Valid = true
	if result.ForgeryProb > 0.3 {
		result.Valid = false
		result.Issues = append(result.Issues, "high forgery probability for TLS configuration")
	}
	if result.ConsistencyScore < 0.7 {
		result.Valid = false
		result.Issues = append(result.Issues, "TLS parameters inconsistent with other layers")
	}

	// Check claimed browser vs detected browser.
	if profile.BrowserType != "" && profile.BrowserType != browser.Family {
		if browser.Confidence > 0.7 {
			result.Issues = append(result.Issues,
				"claimed browser does not match ML-detected browser from TLS fingerprint")
		}
	}

	// Check cipher suite completeness.
	if len(profile.CipherSuites) < 3 {
		result.Issues = append(result.Issues, "too few cipher suites")
	}
	if len(profile.Extensions) < 5 {
		result.Issues = append(result.Issues, "too few TLS extensions")
	}

	return result
}

// ValidateCipherSuites evaluates whether a set of cipher suites is
// consistent with a claimed browser.
func (v *TLSValidator) ValidateCipherSuites(cipherSuites []uint16, claimedBrowser string) *TLSValidationResult {
	// Build a minimal profile for validation.
	profile := &profiles.ClientProfile{
		CipherSuites: cipherSuites,
	}
	if claimedBrowser != "" {
		profile.BrowserType = core.BrowserType(claimedBrowser)
	}
	return v.ValidateProfile(profile)
}

// =========================================================================
// HTTPValidator
// =========================================================================

// HTTPValidator validates HTTP configurations using ML models.
type HTTPValidator struct {
	pipeline *ModelPipeline
}

// NewHTTPValidator creates a new HTTP validator backed by the given pipeline.
func NewHTTPValidator(pipeline *ModelPipeline) *HTTPValidator {
	return &HTTPValidator{pipeline: pipeline}
}

// HTTPValidationResult contains the ML validation of HTTP configuration.
type HTTPValidationResult struct {
	// Valid indicates whether the HTTP configuration looks realistic.
	Valid bool

	// DetectedBrowser is the browser the ML model thinks owns these headers.
	DetectedBrowser string

	// Confidence is the classification confidence [0,1].
	Confidence float64

	// ForgeryProb is the probability this is a forged HTTP fingerprint [0,1].
	ForgeryProb float64

	// ConsistencyScore is how consistent HTTP headers are with TLS [0,1].
	ConsistencyScore float64

	// HeaderOrderScore rates the header ordering quality [0,1].
	HeaderOrderScore float64

	// Issues lists specific problems found.
	Issues []string
}

// ValidateProfile validates the HTTP aspects of a full client profile.
func (v *HTTPValidator) ValidateProfile(profile *profiles.ClientProfile) *HTTPValidationResult {
	features := EncodeFingerprint(profile)
	crossFeatures := ComputeCrossLayerFeatures(features)
	embedding := v.pipeline.encoder.EncodeSingle(features)
	browser := v.pipeline.classifier.ClassifySingle(embedding)
	forgery := v.pipeline.detector.DetectSingle(features, crossFeatures)

	result := &HTTPValidationResult{
		DetectedBrowser:  string(browser.Family),
		Confidence:       browser.Confidence,
		ForgeryProb:      forgery.ForgeryProb,
		ConsistencyScore: crossLayerConsistencyScore(crossFeatures),
	}

	// Evaluate HTTP-specific features.
	// Features[8..13] are HTTP/2 features.
	if len(features) >= 14 {
		result.HeaderOrderScore = 1.0 - features[12]*0.5
	}

	result.Valid = true
	if result.ForgeryProb > 0.3 {
		result.Valid = false
		result.Issues = append(result.Issues, "high forgery probability for HTTP configuration")
	}
	if result.ConsistencyScore < 0.7 {
		result.Valid = false
		result.Issues = append(result.Issues, "HTTP parameters inconsistent with TLS layer")
	}

	// Check HTTP/2 settings presence.
	if profile.HTTP2Settings.InitialWindowSize == 0 && profile.HTTP2Settings.MaxConcurrentStreams == 0 {
		result.Issues = append(result.Issues, "missing HTTP/2 settings")
	}

	// Verify header order matches browser conventions.
	if len(profile.PseudoHeaderOrder) == 0 {
		result.Issues = append(result.Issues, "missing header ordering information")
	}

	return result
}

// =========================================================================
// Combined TLS+HTTP Validator
// =========================================================================

// CrossLayerValidator validates TLS and HTTP together for consistency.
type CrossLayerValidator struct {
	tls  *TLSValidator
	http *HTTPValidator
}

// NewCrossLayerValidator creates a validator that checks TLS+HTTP consistency.
func NewCrossLayerValidator(pipeline *ModelPipeline) *CrossLayerValidator {
	return &CrossLayerValidator{
		tls:  NewTLSValidator(pipeline),
		http: NewHTTPValidator(pipeline),
	}
}

// CrossLayerValidationResult aggregates TLS and HTTP validation.
type CrossLayerValidationResult struct {
	TLS   *TLSValidationResult
	HTTP  *HTTPValidationResult
	Valid bool

	// OverallConsistency is the average consistency across layers.
	OverallConsistency float64

	// BrowserAgreement is true if TLS and HTTP agree on the browser.
	BrowserAgreement bool

	// Issues lists all cross-layer problems.
	Issues []string
}

// ValidateProfile validates both TLS and HTTP layers of a profile.
func (v *CrossLayerValidator) ValidateProfile(profile *profiles.ClientProfile) *CrossLayerValidationResult {
	tlsResult := v.tls.ValidateProfile(profile)
	httpResult := v.http.ValidateProfile(profile)

	result := &CrossLayerValidationResult{
		TLS:                tlsResult,
		HTTP:               httpResult,
		OverallConsistency: (tlsResult.ConsistencyScore + httpResult.ConsistencyScore) / 2,
		BrowserAgreement:   tlsResult.DetectedBrowser == httpResult.DetectedBrowser,
	}

	result.Valid = tlsResult.Valid && httpResult.Valid

	if !result.BrowserAgreement {
		result.Issues = append(result.Issues,
			"TLS and HTTP layers suggest different browsers")
		result.Valid = false
	}

	// Collect all issues
	for _, issue := range tlsResult.Issues {
		result.Issues = append(result.Issues, "[TLS] "+issue)
	}
	for _, issue := range httpResult.Issues {
		result.Issues = append(result.Issues, "[HTTP] "+issue)
	}

	return result
}
