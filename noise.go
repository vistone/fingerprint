package fingerprint

import (
	"github.com/vistone/fingerprint/internal/utils"
)

// NoiseConfig 噪声注入配置
type NoiseConfig struct {
	// 强度（0.0 - 1.0）
	Intensity float64
	// 是否启用 Canvas 噪声
	EnableCanvas bool
	// 是否启用 Audio 噪声
	EnableAudio bool
	// 是否启用 WebGL 噪声
	EnableWebGL bool
	// 是否启用 Font 噪声
	EnableFont bool
	// 是否启用 Screen 噪声
	EnableScreen bool
}

// DefaultNoiseConfig 默认噪声配置
var DefaultNoiseConfig = NoiseConfig{
	Intensity:    0.3,
	EnableCanvas: true,
	EnableAudio:  true,
	EnableWebGL:  true,
	EnableFont:   true,
	EnableScreen: false,
}

// NoiseInjector 噪声注入器，用于主动保护指纹
// 通过添加随机噪声来防止精确的指纹追踪
type NoiseInjector struct {
	config NoiseConfig
}

// NewNoiseInjector 创建新的噪声注入器
func NewNoiseInjector(config NoiseConfig) *NoiseInjector {
	// 确保强度在有效范围内
	if config.Intensity < 0.0 {
		config.Intensity = 0.0
	}
	if config.Intensity > 1.0 {
		config.Intensity = 1.0
	}
	return &NoiseInjector{config: config}
}

// CanvasNoise Canvas 指纹噪声参数
type CanvasNoise struct {
	// 像素偏移量（R/G/B 各通道的随机偏移）
	PixelOffsetR int
	PixelOffsetG int
	PixelOffsetB int
	// 文字渲染偏移（亚像素级别）
	TextOffsetX float64
	TextOffsetY float64
}

// GenerateCanvasNoise 生成 Canvas 指纹噪声参数
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

// AudioNoise Audio 指纹噪声参数
type AudioNoise struct {
	// 音频样本偏移（噪声级别）
	NoiseLevel float64
	// 频率偏移
	FrequencyOffset float64
}

// GenerateAudioNoise 生成 Audio 指纹噪声参数
func (n *NoiseInjector) GenerateAudioNoise() *AudioNoise {
	if !n.config.EnableAudio {
		return &AudioNoise{}
	}
	return &AudioNoise{
		NoiseLevel:      n.config.Intensity * 0.01, // 最大 1% 噪声
		FrequencyOffset: randFloatRange(-n.config.Intensity*5.0, n.config.Intensity*5.0),
	}
}

// WebGLNoise WebGL 指纹噪声参数
type WebGLNoise struct {
	// 最大纹理尺寸偏移
	MaxTextureSizeOffset int
	// 最大顶点 Uniform 偏移
	MaxVertexUniformsOffset int
	// 渲染器字符串混淆
	RendererSuffix string
	// 厂商字符串混淆
	VendorSuffix string
}

// webGLRendererSuffixes 可用的渲染器后缀
var webGLRendererSuffixes = []string{
	"", "", "", // 更高概率的空后缀
	" (ANGLE)", " (Direct3D11 vs_5_0 ps_5_0)", " (OpenGL)",
}

// GenerateWebGLNoise 生成 WebGL 指纹噪声参数
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

// FontNoise Font 指纹噪声参数
type FontNoise struct {
	// 字体宽度偏移（百分比）
	WidthOffsetPercent float64
	// 字体高度偏移（百分比）
	HeightOffsetPercent float64
}

// GenerateFontNoise 生成 Font 指纹噪声参数
func (n *NoiseInjector) GenerateFontNoise() *FontNoise {
	if !n.config.EnableFont {
		return &FontNoise{}
	}
	maxOffset := n.config.Intensity * 0.01 // 最大 1%
	return &FontNoise{
		WidthOffsetPercent:  randFloatRange(-maxOffset, maxOffset),
		HeightOffsetPercent: randFloatRange(-maxOffset, maxOffset),
	}
}

// ScreenNoise Screen 指纹噪声参数
type ScreenNoise struct {
	// 屏幕宽度偏移
	WidthOffset int
	// 屏幕高度偏移
	HeightOffset int
	// 色深偏移
	ColorDepthOffset int
}

// GenerateScreenNoise 生成 Screen 指纹噪声参数
func (n *NoiseInjector) GenerateScreenNoise() *ScreenNoise {
	if !n.config.EnableScreen {
		return &ScreenNoise{}
	}
	maxPixelOffset := int(n.config.Intensity * 10)
	return &ScreenNoise{
		WidthOffset:      randIntRange(-maxPixelOffset, maxPixelOffset),
		HeightOffset:     randIntRange(-maxPixelOffset, maxPixelOffset),
		ColorDepthOffset: 0, // 色深不建议修改
	}
}

// BrowserNoiseProfile 完整的浏览器噪声配置
type BrowserNoiseProfile struct {
	Canvas *CanvasNoise
	Audio  *AudioNoise
	WebGL  *WebGLNoise
	Font   *FontNoise
	Screen *ScreenNoise
}

// GenerateFullProfile 生成完整的浏览器噪声配置
func (n *NoiseInjector) GenerateFullProfile() *BrowserNoiseProfile {
	return &BrowserNoiseProfile{
		Canvas: n.GenerateCanvasNoise(),
		Audio:  n.GenerateAudioNoise(),
		WebGL:  n.GenerateWebGLNoise(),
		Font:   n.GenerateFontNoise(),
		Screen: n.GenerateScreenNoise(),
	}
}

// GenerateBrowserNoiseProfile 使用默认配置生成完整的浏览器噪声配置
func GenerateBrowserNoiseProfile() *BrowserNoiseProfile {
	injector := NewNoiseInjector(DefaultNoiseConfig)
	return injector.GenerateFullProfile()
}

// randIntRange 生成 [min, max] 范围内的随机整数
func randIntRange(min, max int) int {
	if min >= max {
		return min
	}
	n := utils.GetGlobalRandGenerator().Intn(max - min + 1)
	return min + n
}

// randFloatRange 生成 [min, max] 范围内的随机浮点数
func randFloatRange(min, max float64) float64 {
	if min >= max {
		return min
	}
	// 使用整数随机数映射到浮点数范围
	const precision = 10000
	n := utils.GetGlobalRandGenerator().Intn(precision + 1)
	return min + float64(n)/float64(precision)*(max-min)
}
