package noise

// translated comment
import (
	"github.com/vistone/fingerprint/modules/kit"
)

// translated comment
type NoiseConfig struct {
	// translated comment
	Intensity float64
	// translated comment
	EnableCanvas bool
	// translated comment
	EnableAudio bool
	// translated comment
	EnableWebGL bool
	// translated comment
	EnableFont bool
	// translated comment
	EnableScreen bool
}

// translated comment
var DefaultNoiseConfig = NoiseConfig{
	Intensity:    0.3,
	EnableCanvas: true,
	EnableAudio:  true,
	EnableWebGL:  true,
	EnableFont:   true,
	EnableScreen: false,
}

// translated comment
// translated comment
type NoiseInjector struct {
	config NoiseConfig
}

// translated comment
func NewNoiseInjector(config NoiseConfig) *NoiseInjector {
	// translated comment
	if config.Intensity < 0.0 {
		config.Intensity = 0.0
	}
	if config.Intensity > 1.0 {
		config.Intensity = 1.0
	}
	return &NoiseInjector{config: config}
}

// translated comment
type CanvasNoise struct {
	// translated comment
	PixelOffsetR int
	PixelOffsetG int
	PixelOffsetB int
	// translated comment
	TextOffsetX float64
	TextOffsetY float64
}

// translated comment
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

// translated comment
type AudioNoise struct {
	// translated comment
	NoiseLevel float64
	// translated comment
	FrequencyOffset float64
}

// translated comment
func (n *NoiseInjector) GenerateAudioNoise() *AudioNoise {
	if !n.config.EnableAudio {
		return &AudioNoise{}
	}
	return &AudioNoise{
		NoiseLevel:      n.config.Intensity * 0.01, // translated comment
		FrequencyOffset: randFloatRange(-n.config.Intensity*5.0, n.config.Intensity*5.0),
	}
}

// translated comment
type WebGLNoise struct {
	// translated comment
	MaxTextureSizeOffset int
	// translated comment
	MaxVertexUniformsOffset int
	// translated comment
	RendererSuffix string
	// translated comment
	VendorSuffix string
}

// translated comment
var webGLRendererSuffixes = []string{
	"", "", "", // translated comment
	" (ANGLE)", " (Direct3D11 vs_5_0 ps_5_0)", " (OpenGL)",
}

// translated comment
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

// translated comment
type FontNoise struct {
	// translated comment
	WidthOffsetPercent float64
	// translated comment
	HeightOffsetPercent float64
}

// translated comment
func (n *NoiseInjector) GenerateFontNoise() *FontNoise {
	if !n.config.EnableFont {
		return &FontNoise{}
	}
	maxOffset := n.config.Intensity * 0.01 // translated comment
	return &FontNoise{
		WidthOffsetPercent:  randFloatRange(-maxOffset, maxOffset),
		HeightOffsetPercent: randFloatRange(-maxOffset, maxOffset),
	}
}

// translated comment
type ScreenNoise struct {
	// translated comment
	WidthOffset int
	// translated comment
	HeightOffset int
	// translated comment
	ColorDepthOffset int
}

// translated comment
func (n *NoiseInjector) GenerateScreenNoise() *ScreenNoise {
	if !n.config.EnableScreen {
		return &ScreenNoise{}
	}
	maxPixelOffset := int(n.config.Intensity * 10)
	return &ScreenNoise{
		WidthOffset:      randIntRange(-maxPixelOffset, maxPixelOffset),
		HeightOffset:     randIntRange(-maxPixelOffset, maxPixelOffset),
		ColorDepthOffset: 0, // translated comment
	}
}

// translated comment
type BrowserNoiseProfile struct {
	Canvas *CanvasNoise
	Audio  *AudioNoise
	WebGL  *WebGLNoise
	Font   *FontNoise
	Screen *ScreenNoise
}

// translated comment
func (n *NoiseInjector) GenerateFullProfile() *BrowserNoiseProfile {
	return &BrowserNoiseProfile{
		Canvas: n.GenerateCanvasNoise(),
		Audio:  n.GenerateAudioNoise(),
		WebGL:  n.GenerateWebGLNoise(),
		Font:   n.GenerateFontNoise(),
		Screen: n.GenerateScreenNoise(),
	}
}

// translated comment
func GenerateBrowserNoiseProfile() *BrowserNoiseProfile {
	injector := NewNoiseInjector(DefaultNoiseConfig)
	return injector.GenerateFullProfile()
}

// translated comment
func randIntRange(min, max int) int {
	if min >= max {
		return min
	}
	n := utils.GetGlobalRandGenerator().Intn(max - min + 1)
	return min + n
}

// translated comment
func randFloatRange(min, max float64) float64 {
	if min >= max {
		return min
	}
	// translated comment
	const precision = 10000
	n := utils.GetGlobalRandGenerator().Intn(precision + 1)
	return min + float64(n)/float64(precision)*(max-min)
}
