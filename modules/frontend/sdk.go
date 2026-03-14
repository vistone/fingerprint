// Package frontend provides frontend fingerprint SDK functionality
// Including JavaScript code generation and server-side processing
package frontend

import (
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/ml"
)

// SDK frontend SDK manager
type SDK struct {
	config    *SDKConfig
	sessions  map[string]*Session
	extractor *ml.FeatureExtractor
	mu        sync.RWMutex
}

// SDKConfig SDK configuration
type SDKConfig struct {
	// Noise injection configuration
	EnableCanvasNoise bool
	EnableAudioNoise  bool
	EnableWebGLNoise  bool
	EnableTimingNoise bool
	NoiseLevel        float64 // 0.0 - 1.0

	// Collection configuration
	CollectCanvas   bool
	CollectWebGL    bool
	CollectAudio    bool
	CollectFonts    bool
	CollectStorage  bool
	CollectWebRTC   bool
	CollectHardware bool
	CollectTiming   bool

	// Session configuration
	SessionTimeout time.Duration
}

// DefaultSDKConfig default SDK configuration
var DefaultSDKConfig = &SDKConfig{
	EnableCanvasNoise: true,
	EnableAudioNoise:  true,
	EnableWebGLNoise:  true,
	EnableTimingNoise: true,
	NoiseLevel:        0.1,

	CollectCanvas:   true,
	CollectWebGL:    true,
	CollectAudio:    true,
	CollectFonts:    true,
	CollectStorage:  true,
	CollectWebRTC:   true,
	CollectHardware: true,
	CollectTiming:   true,

	SessionTimeout: core.DefaultSessionTimeout,
}

// NewSDK creates a new SDK manager
func NewSDK(config *SDKConfig) *SDK {
	if config == nil {
		config = DefaultSDKConfig
	}
	return &SDK{
		config:    config,
		sessions:  make(map[string]*Session),
		extractor: ml.NewFeatureExtractor(),
	}
}

// Session frontend session
type Session struct {
	ID           string
	CreatedAt    time.Time
	LastAccessed time.Time
	Data         map[string]interface{}
	Fingerprint  *ml.FrontendFingerprintData
}

// GenerateJSInjector generates injection script
