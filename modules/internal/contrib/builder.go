// translated comment
package contrib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/vistone/fingerprint/modules/internal/plugins"
)

// translated comment
type Builder struct {
	data *plugins.FingerprintData
}

// translated comment
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

// translated comment
func (b *Builder) WithDisplayName(name string) *Builder {
	b.data.Metadata.DisplayName = name
	return b
}

// translated comment
func (b *Builder) WithCategory(cat plugins.FingerprintCategory) *Builder {
	b.data.Metadata.Category = cat
	return b
}

// translated comment
func (b *Builder) WithBrowser(name, version, os string) *Builder {
	b.data.Metadata.Browser = name
	b.data.Metadata.BrowserVersion = version
	b.data.Metadata.OS = os
	return b
}

// translated comment
func (b *Builder) WithVersion(version string) *Builder {
	b.data.Metadata.Version = version
	return b
}

// translated comment
func (b *Builder) WithUserAgent(ua string) *Builder {
	b.data.UserAgent = ua
	return b
}

// translated comment
func (b *Builder) WithTLSVersion(version uint16) *Builder {
	if b.data.ClientHello == nil {
		b.data.ClientHello = &plugins.ClientHelloSpec{}
	}
	b.data.ClientHello.TLSVersion = version
	return b
}

// translated comment
func (b *Builder) WithCipherSuites(suites []uint16) *Builder {
	if b.data.ClientHello == nil {
		b.data.ClientHello = &plugins.ClientHelloSpec{}
	}
	b.data.ClientHello.CipherSuites = suites
	return b
}

// translated comment
func (b *Builder) WithExtensions(exts []uint16) *Builder {
	if b.data.ClientHello == nil {
		b.data.ClientHello = &plugins.ClientHelloSpec{}
	}
	b.data.ClientHello.Extensions = exts
	return b
}

// translated comment
func (b *Builder) WithEllipticCurves(curves []uint16) *Builder {
	if b.data.ClientHello == nil {
		b.data.ClientHello = &plugins.ClientHelloSpec{}
	}
	b.data.ClientHello.EllipticCurves = curves
	return b
}

// translated comment
func (b *Builder) WithAuthor(name string) *Builder {
	b.data.Metadata.Author = name
	return b
}

// translated comment
func (b *Builder) WithLicense(license string) *Builder {
	b.data.Metadata.License = license
	return b
}

// translated comment
func (b *Builder) WithTags(tags []string) *Builder {
	b.data.Metadata.Tags = tags
	return b
}

// translated comment
func (b *Builder) WithMobile(isMobile bool) *Builder {
	b.data.Metadata.IsMobile = isMobile
	return b
}

// translated comment
func (b *Builder) Build() (plugins.Plugin, error) {
	plugin := plugins.NewBasicPlugin(b.data)
	if err := plugin.Validate(); err != nil {
		return nil, err
	}
	return plugin, nil
}

// translated comment
func (b *Builder) SaveToFile(filePath string) error {
	jsonData, err := json.MarshalIndent(b.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	// translated comment
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, jsonData, 0644)
}

// translated comment
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

// translated comment
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

// translated comment
func ExampleMobile() *Builder {
	return NewBuilder("mobile_example").
		WithDisplayName("Custom Mobile").
		WithCategory(plugins.CategoryMobile).
		WithMobile(true).
		WithVersion("1.0.0")
}
