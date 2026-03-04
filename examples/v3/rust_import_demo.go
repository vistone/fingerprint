// +build ignore

// Rust 数据导入示例
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vistone/fingerprint/modules/ml"
)

func main() {
	fmt.Println("=== Rust Data Compatibility Demo ===\n")

	// 1. 创建示例 Rust 格式数据集
	fmt.Println("1. Create Sample Rust Dataset:")
	rustDataset := createSampleRustDataset()
	fmt.Printf("   Created %d fingerprints\n", len(rustDataset.Fingerprints))

	// 2. 保存为 JSON
	tempDir := os.TempDir()
	rustPath := filepath.Join(tempDir, "rust_dataset.json")
	saveRustDataset(rustDataset, rustPath)
	fmt.Printf("   Saved to: %s\n", rustPath)

	// 3. 导入 Rust 数据集
	fmt.Println("\n2. Import Rust Dataset:")
	importer := ml.NewRustImporter(tempDir)
	dataset, err := importer.ImportDataset("rust_dataset.json")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return
	}
	fmt.Printf("   Imported %d samples\n", len(dataset.Samples))

	// 4. 显示统计信息
	fmt.Println("\n3. Dataset Statistics:")
	fmt.Printf("   Name: %s\n", dataset.Name)
	fmt.Printf("   Version: %s\n", dataset.Version)
	fmt.Printf("   Total: %d\n", dataset.Statistics.TotalSamples)
	fmt.Printf("   Protocols: %v\n", dataset.Statistics.ProtocolCounts)
	fmt.Printf("   Families: %v\n", dataset.Statistics.FamilyCounts)

	// 5. 转换为训练数据并训练
	fmt.Println("\n4. Train Classifier:")
	trainingData := dataset.ToTrainingData()
	
	classifier := ml.NewHierarchicalClassifier()
	classifier.Initialize()
	
	err = classifier.Train(trainingData)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return
	}
	fmt.Println("   Training completed!")

	// 6. 测试分类
	fmt.Println("\n5. Test Classification:")
	if len(dataset.Samples) > 0 {
		sample := dataset.Samples[0]
		result := classifier.Classify(sample.Features)
		
		fmt.Printf("   Sample ID: %s\n", sample.ID)
		fmt.Printf("   Expected: %s / %s\n", sample.Label.Protocol, sample.Label.Family)
		fmt.Printf("   Predicted: %s / %s\n", result.Protocol, result.Family)
		fmt.Printf("   Confidence: %.2f\n", result.Confidence)
	}

	// 7. 导出回 Rust 格式
	fmt.Println("\n6. Export Back to Rust Format:")
	exportPath := filepath.Join(tempDir, "exported_rust_dataset.json")
	err = dataset.SaveToRustFormat(exportPath)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return
	}
	fmt.Printf("   Exported to: %s\n", exportPath)

	// 8. 兼容性检查
	fmt.Println("\n7. Feature Compatibility:")
	checker := ml.NewCompatibilityChecker()
	compat := checker.CheckFeatureCompatibility()
	
	compatible := 0
	for feature, status := range compat {
		if status == "compatible" {
			compatible++
		} else {
			fmt.Printf("   %s: %s\n", feature, status)
		}
	}
	fmt.Printf("   Total compatible features: %d/%d\n", compatible, len(compat))

	// 清理
	os.Remove(rustPath)
	os.Remove(exportPath)

	fmt.Println("\n=== Demo Complete ===")
}

func createSampleRustDataset() *ml.RustDataset {
	return &ml.RustDataset{
		Name:        "sample_rust_dataset",
		Version:     "1.0",
		Description: "Sample dataset for demo",
		Fingerprints: []ml.RustFingerprint{
			{
				ID:      "chrome_133_win",
				Browser: "chrome",
				Version: "133",
				OS:      "Windows 10",
				TLS: ml.RustTLSData{
					Version:      0x0303,
					CipherSuites: []uint16{0x1301, 0x1302, 0x1303},
					Extensions:   []uint16{0x0000, 0x0017, 0xff01},
				},
				Features: map[string]float64{
					"tls_version":    0x0303,
					"cipher_suites":  8,
					"extensions":     11,
					"http2_settings": 65536,
					"user_agent":     100,
					"entropy":        15.5,
				},
			},
			{
				ID:      "firefox_133_win",
				Browser: "firefox",
				Version: "133",
				OS:      "Windows 10",
				TLS: ml.RustTLSData{
					Version:      0x0303,
					CipherSuites: []uint16{0x1301, 0x1302},
					Extensions:   []uint16{0x0000, 0x0017, 0x0015},
				},
				Features: map[string]float64{
					"tls_version":    0x0303,
					"cipher_suites":  11,
					"extensions":     11,
					"http2_settings": 131072,
					"user_agent":     200,
					"entropy":        16.2,
				},
			},
			{
				ID:      "safari_18_mac",
				Browser: "safari",
				Version: "18",
				OS:      "macOS 15",
				TLS: ml.RustTLSData{
					Version:      0x0303,
					CipherSuites: []uint16{0x1301, 0x1302, 0x1303},
					Extensions:   []uint16{0x0000, 0x0017},
				},
				Features: map[string]float64{
					"tls_version":    0x0303,
					"cipher_suites":  8,
					"extensions":     9,
					"http2_settings": 2097152,
					"user_agent":     300,
					"entropy":        14.8,
				},
			},
		},
	}
}

func saveRustDataset(dataset *ml.RustDataset, path string) {
	data, _ := json.MarshalIndent(dataset, "", "  ")
	os.WriteFile(path, data, 0644)
}
