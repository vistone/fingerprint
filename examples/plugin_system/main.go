package main

import (
	"fmt"
	"log"

	"github.com/vistone/fingerprint/internal/contrib"
	"github.com/vistone/fingerprint/internal/plugins"
)

func main() {
	fmt.Println("=== 指纹插件系统示例 ===\n")

	// 例子 1: 创建 Chrome 133 指纹
	fmt.Println("1️⃣  创建 Chrome 133 指纹:")
	builder := contrib.ExampleChrome133()
	plugin, err := builder.Build()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 创建成功: %s\n", plugin.Metadata().DisplayName)

	// 例子 2: 注册指纹
	fmt.Println("\n2️⃣  注册指纹:")
	meta := plugin.Metadata()
	if err := plugins.RegisterPlugin(meta.Name, plugin, plugins.SourceCommunity); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 注册成功: %s\n", meta.Name)

	// 例子 3: 获取指纹
	fmt.Println("\n3️⃣  获取指纹:")
	retrieved, err := plugins.GetPlugin(meta.Name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 获取成功: %s\n", retrieved.GetUserAgent()[:50]+"...")

	// 例子 4: 转换为 TLS spec
	fmt.Println("\n4️⃣  转换为 TLS ClientHelloSpec:")
	spec, err := plugin.ToClientHelloSpec()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 转换成功:\n")
	fmt.Printf("  - TLS 版本: 0x%04x\n", spec.TLSVersMin)
	fmt.Printf("  - 密码套件数: %d\n", len(spec.CipherSuites))

	// 例子 5: 创建移动设备指纹
	fmt.Println("\n5️⃣  创建移动设备指纹:")
	mobileBuilder := contrib.ExampleMobile().
		WithUserAgent("Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X)").
		WithTLSVersion(0x0304).
		WithCipherSuites([]uint16{0x1301})

	mobilePlugin, err := mobileBuilder.Build()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 创建成功: %s (移动设备: %v)\n",
		mobilePlugin.Metadata().DisplayName,
		mobilePlugin.Metadata().IsMobile)

	fmt.Println("\n✨ 所有示例完成!")
}
