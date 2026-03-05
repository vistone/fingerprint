// Package contrib 实现指纹构建器
package contrib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/vistone/fingerprint/modules/internal/plugins"
)

// Builder 指纹构建器
type Builder struct {
	data *plugins.FingerprintData
}

// NewBuilder 创建构建器
func NewBuilder(name string) *Builder {
	return &Builder{
		data: &plugins.FingerprintData{
			Metadata: plugins.FingerprintMetadata{
				Name: name,
			},
			Extensions: make(map[string]interface{}),
		},
	}
}

// WithDisplayName 设置显示名称
func (b *Builder) WithDisplayName(name string) *Builder {
	b.data.Metadata.DisplayName = name
	return b
}

// WithCategory 设置分类
func (b *Builder) WithCategory(cat plugins.FingerprintCategory) *Builder {
	b.data.Metadata.Category = cat
	return b
}

// WithBrowser 设置浏览器信息
func (b *Builder) WithBrowser(name, version, os string) *Builder {
	b.data.Metadata.Browser = name
	b.data.Metadata.BrowserVersion = version
	b.data.Metadata.OS = os
	return b
}

// WithVersion 设置版本
func (b *Builder) WithVersion(version string) *Builder {
	b.data.Metadata.Version = version
	return b
}

// WithUserAgent 设置 User-Agent
func (b *Builder) WithUserAgent(ua string) *Builder {
	b.data.UserAgent = ua
	return b
}

// WithTLSVersion 设置 TLS 版本
func (b *Builder) WithTLSVersion(version uint16) *Builder {
	if b.data.ClientHello == nil {
		b.data.ClientHello = &plugins.ClientHelloSpec{}
	}
	b.data.ClientHello.TLSVersion = version
	return b
}

// WithCipherSuites 设置密码套件
func (b *Builder) WithCipherSuites(suites []uint16) *Builder {
	if b.data.ClientHello == nil {
		b.data.ClientHello = &plugins.ClientHelloSpec{}
	}
	b.data.ClientHello.CipherSuites = suites
	return b
}

// WithExtensions 设置扩展
func (b *Builder) WithExtensions(exts []uint16) *Builder {
	if b.data.ClientHello == nil {
		b.data.ClientHello = &plugins.ClientHelloSpec{}
	}
	b.data.ClientHello.Extensions = exts
	return b
}

// WithEllipticCurves 设置椭圆曲线
func (b *Builder) WithEllipticCurves(curves []uint16) *Builder {
	if b.data.ClientHello == nil {
		b.data.ClientHello = &plugins.ClientHelloSpec{}
	}
	b.data.ClientHello.EllipticCurves = curves
	return b
}

// WithAuthor 设置作者
func (b *Builder) WithAuthor(name string) *Builder {
	b.data.Metadata.Author = name
	return b
}

// WithLicense 设置许可证
func (b *Builder) WithLicense(license string) *Builder {
	b.data.Metadata.License = license
	return b
}

// WithTags 设置标签
func (b *Builder) WithTags(tags []string) *Builder {
	b.data.Metadata.Tags = tags
	return b
}

// WithMobile 标记移动设备
func (b *Builder) WithMobile(isMobile bool) *Builder {
	b.data.Metadata.IsMobile = isMobile
	return b
}

// Build 构建插件
func (b *Builder) Build() (plugins.Plugin, error) {
	plugin := plugins.NewBasicPlugin(b.data)
	if err := plugin.Validate(); err != nil {
		return nil, err
	}
	return plugin, nil
}

// SaveToFile 保存到文件
func (b *Builder) SaveToFile(filePath string) error {
	jsonData, err := json.MarshalIndent(b.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, jsonData, 0644)
}

// SaveAndRegister 保存并注册
func (b *Builder) SaveAndRegister(filePath string) (plugins.Plugin, error) {
	if err := b.SaveToFile(filePath); err != nil {
		return nil, err
	}

	plugin, err := b.Build()
	if err != nil {
		return nil, err
	}

	if err := plugins.RegisterPlugin(b.data.Metadata.Name, plugin, plugins.SourceCommunity); err != nil {
		return nil, err
	}

	return plugin, nil
}

// ExampleChrome133 Chrome 133 示例
func ExampleChrome133() *Builder {
	return NewBuilder("chrome_133_example").
		WithDisplayName("Chrome 133").
		WithCategory(plugins.CategoryBrowser).
		WithBrowser("Chrome", "133", "Windows NT 10.0").
		WithVersion("133.0.0.0").
		WithUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36").
		WithTLSVersion(0x0304).
		WithCipherSuites([]uint16{0x1302, 0x1301}).
		WithExtensions([]uint16{0x0000, 0x000d, 0x0033}).
		WithEllipticCurves([]uint16{0x001d, 0x0017}).
		WithAuthor("Fingerprint Contributors").
		WithLicense("BSD 3-Clause").
		WithTags([]string{"browser", "chrome", "tls"})
}

// ExampleMobile 移动设备示例
func ExampleMobile() *Builder {
	return NewBuilder("mobile_example").
		WithDisplayName("Custom Mobile").
		WithCategory(plugins.CategoryMobile).
		WithMobile(true).
		WithVersion("1.0.0")
}
