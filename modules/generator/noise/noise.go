package noise

// Phase 3: basic migration completed; deep optimization remains.
import (
	"github.com/vistone/fingerprint/modules/kit"
)

// NoiseConfig configures fingerprint noise injection.
type NoiseConfig struct {
	// Intensity range: 0.0 - 1.0.
	Intensity float64
	// Enables Canvas noise.
	EnableCanvas bool
	// Enables Audio noise.
	EnableAudio bool
	// Enables WebGL noise.
	EnableWebGL bool
	// Enables Font noise.
	EnableFont bool
	// Enables Screen noise.
	EnableScreen bool
}

// DefaultNoiseConfig is the default noise profile.
var DefaultNoiseConfig = NoiseConfig{
	Intensity:    0.3,
	EnableCanvas: true,
	EnableAudio:  true,
	EnableWebGL:  true,
	EnableFont:   true,
	EnableScreen: false,
}

// NoiseInjector applies random noise for active fingerprint protection.
type NoiseInjector struct {
	config NoiseConfig
}

// NewNoiseInjector creates a new noise injector.
func NewNoiseInjector(config NoiseConfig) *NoiseInjector {
	// Clamp intensity to valid range.
	if config.Intensity < 0.0 {
		config.Intensity = 0.0
	}
	if config.Intensity > 1.0 {
		config.Intensity = 1.0
	}
	return &NoiseInjector{config: config}
}

// CanvasNoise defines Canvas fingerprint perturbations.
type CanvasNoise struct {
	// Pixel offsets for R/G/B channels.
	PixelOffsetR int
	PixelOffsetG int
	PixelOffsetB int
	// Subpixel text rendering offset.
	TextOffsetX float64
	TextOffsetY float64
}

// GenerateCanvasNoise generates Canvas noise settings.
func (n *NoiseInjector) GenerateCanvasNoise() *CanvasNoise {
	if !n.config.EnableCanvas {
		return &CanvasNoise{}
	}
	maxOffset := int(n.config.Intensity * 5)
	if maxOffset < 1 {
		maxOffset = 1
	}
	return &CanvasNoise{
		PixelOffsetR: randIntRange(-maxOffset, maxOffset),
		PixelOffsetG: randIntRange(-maxOffset, maxOffset),
		PixelOffsetB: randIntRange(-maxOffset, maxOffset),
		TextOffsetX:  randFloatRange(-n.config.Intensity*0.5, n.config.Intensity*0.5),
		TextOffsetY:  randFloatRange(-n.config.Intensity*0.5, n.config.Intensity*0.5),
	}
}

// AudioNoise defines Audio fingerprint perturbations.
type AudioNoise struct {
	// Audio sample noise level.
	NoiseLevel float64
	// Frequency offset.
	FrequencyOffset float64
}

// GenerateAudioNoise generates Audio noise settings.
func (n *NoiseInjector) GenerateAudioNoise() *AudioNoise {
	if !n.config.EnableAudio {
		return &AudioNoise{}
	}
	return &AudioNoise{
		NoiseLevel:      n.config.Intensity * 0.01, // Up to 1% noise.
		FrequencyOffset: randFloatRange(-n.config.Intensity*5.0, n.config.Intensity*5.0),
	}
}

// WebGLNoise defines WebGL fingerprint perturbations.
type WebGLNoise struct {
	// Max texture size offset.
	MaxTextureSizeOffset int
	// Max vertex uniform offset.
	MaxVertexUniformsOffset int
	// Renderer string suffix.
	RendererSuffix string
	// Vendor string suffix.
	VendorSuffix string
}

// webGLRendererSuffixes stores candidate renderer suffixes.
var webGLRendererSuffixes = []string{
	"", "", "", // Empty suffix has higher probability.
	" (ANGLE)", " (Direct3D11 vs_5_0 ps_5_0)", " (OpenGL)",
}

// GenerateWebGLNoise generates WebGL noise settings.
func (n *NoiseInjector) GenerateWebGLNoise() *WebGLNoise {
	if !n.config.EnableWebGL {
		return &WebGLNoise{}
	}
	maxOffset := int(n.config.Intensity * 512)
	return &WebGLNoise{
		MaxTextureSizeOffset:    randIntRange(-maxOffset, maxOffset),
		MaxVertexUniformsOffset: randIntRange(-maxOffset/4, maxOffset/4),
		RendererSuffix:          utils.RandomChoiceString(webGLRendererSuffixes),
		VendorSuffix:            "",
	}
}

// FontNoise defines Font fingerprint perturbations.
type FontNoise struct {
	// Width offset in percentage.
	WidthOffsetPercent float64
	// Height offset in percentage.
	HeightOffsetPercent float64
}

// GenerateFontNoise generates Font noise settings.
func (n *NoiseInjector) GenerateFontNoise() *FontNoise {
	if !n.config.EnableFont {
		return &FontNoise{}
	}
	maxOffset := n.config.Intensity * 0.01 // Up to 1%.
	return &FontNoise{
		WidthOffsetPercent:  randFloatRange(-maxOffset, maxOffset),
		HeightOffsetPercent: randFloatRange(-maxOffset, maxOffset),
	}
}

// ScreenNoise defines Screen fingerprint perturbations.
type ScreenNoise struct {
	// Screen width offset.
	WidthOffset int
	// Screen height offset.
	HeightOffset int
	// Color depth offset.
	ColorDepthOffset int
}

// GenerateScreenNoise generates Screen noise settings.
func (n *NoiseInjector) GenerateScreenNoise() *ScreenNoise {
	if !n.config.EnableScreen {
		return &ScreenNoise{}
	}
	maxPixelOffset := int(n.config.Intensity * 10)
	return &ScreenNoise{
		WidthOffset:      randIntRange(-maxPixelOffset, maxPixelOffset),
		HeightOffset:     randIntRange(-maxPixelOffset, maxPixelOffset),
		ColorDepthOffset: 0, // Color depth is intentionally stable.
	}
}

// BrowserNoiseProfile is the full browser noise configuration.
type BrowserNoiseProfile struct {
	Canvas *CanvasNoise
	Audio  *AudioNoise
	WebGL  *WebGLNoise
	Font   *FontNoise
	Screen *ScreenNoise
}

// GenerateFullProfile generates a full browser noise profile.
func (n *NoiseInjector) GenerateFullProfile() *BrowserNoiseProfile {
	return &BrowserNoiseProfile{
		Canvas: n.GenerateCanvasNoise(),
		Audio:  n.GenerateAudioNoise(),
		WebGL:  n.GenerateWebGLNoise(),
		Font:   n.GenerateFontNoise(),
		Screen: n.GenerateScreenNoise(),
	}
}

// GenerateBrowserNoiseProfile generates a full profile with defaults.
func GenerateBrowserNoiseProfile() *BrowserNoiseProfile {
	injector := NewNoiseInjector(DefaultNoiseConfig)
	return injector.GenerateFullProfile()
}

// randIntRange returns a random integer in [min, max].
func randIntRange(min, max int) int {
	if min >= max {
		return min
	}
	n := utils.GetGlobalRandGenerator().Intn(max - min + 1)
	return min + n
}

// randFloatRange returns a random float in [min, max].
func randFloatRange(min, max float64) float64 {
	if min >= max {
		return min
	}
	// Map an integer random value into the float range.
	const precision = 10000
	n := utils.GetGlobalRandGenerator().Intn(precision + 1)
	return min + float64(n)/float64(precision)*(max-min)
}
