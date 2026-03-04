// +build ignore

// 训练数据演示
package main

import (
	"encoding/json"
	"fmt"
	
	"github.com/vistone/fingerprint"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
)

func main() {
	fmt.Println("=== Training Data & Extended Profiles Demo ===\n")

	// 1. 显示注册的指纹数量
	fmt.Println("1. Registered Profiles:")
	fmt.Printf("   Total: %d profiles\n", profiles.GetProfileCount())
	
	fmt.Println("   By Browser:")
	for _, browser := range []fingerprint.BrowserType{
		fingerprint.BrowserChrome,
		fingerprint.BrowserFirefox,
		fingerprint.BrowserSafari,
		fingerprint.BrowserEdge,
		fingerprint.BrowserOpera,
	} {
		count := len(profiles.GetProfilesByBrowser(browser))
		if count > 0 {
			fmt.Printf("     %s: %d\n", browser, count)
		}
	}

	// 2. 生成合成训练数据
	fmt.Println("\n2. Generate Synthetic Dataset:")
	dataset := ml.GenerateSyntheticDataset("demo", 100)
	fmt.Printf("   Generated %d samples\n", len(dataset.Samples))
	fmt.Printf("   Protocols: %v\n", dataset.Statistics.ProtocolCounts)
	fmt.Printf("   Families: %v\n", dataset.Statistics.FamilyCounts)

	// 3. 转换为训练数据
	fmt.Println("\n3. Convert to Training Data:")
	trainingData := dataset.ToTrainingData()
	fmt.Printf("   Protocol samples: %d\n", len(trainingData.ProtocolFeatures))
	fmt.Printf("   Family groups: %d\n", len(trainingData.FamilyFeatures))
	fmt.Printf("   Version groups: %d\n", len(trainingData.VersionFeatures))

	// 4. 训练分类器
	fmt.Println("\n4. Train Classifier:")
	classifier := ml.NewHierarchicalClassifier()
	classifier.Initialize()
	
	err := classifier.Train(trainingData)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Println("   Training completed!")
	}

	// 5. 使用训练好的分类器
	fmt.Println("\n5. Classification with Trained Model:")
	
	// 获取一个真实的指纹配置
	profile := fingerprint.GetRandom()
	extractor := fingerprint.NewFeatureExtractor()
	features := extractor.ExtractFromProfile(profile)
	
	result := classifier.Classify(features)
	fmt.Printf("   Protocol: %s (%.2f)\n", result.Protocol, result.ProtocolConfidence)
	fmt.Printf("   Family: %s (%.2f)\n", result.Family, result.FamilyConfidence)
	fmt.Printf("   Version: %s (%.2f)\n", result.Version, result.VersionConfidence)
	fmt.Printf("   Overall: %.2f\n", result.Confidence)

	// 6. 导出模型
	fmt.Println("\n6. Export Pretrained Model:")
	trainer := ml.NewTrainer(classifier)
	trainer.LoadDataset(dataset)
	model := trainer.ExportModel("fingerprint_demo", "1.0.0")
	
	modelInfo := ml.GetModelInfo(model)
	fmt.Printf("   Name: %s v%s\n", modelInfo.Name, modelInfo.Version)
	fmt.Printf("   Protocols: %d\n", modelInfo.ProtocolCount)
	fmt.Printf("   Families: %d\n", modelInfo.FamilyCount)
	fmt.Printf("   Versions: %d\n", modelInfo.VersionCount)
	fmt.Printf("   Total Centers: %d\n", modelInfo.TotalCenters)

	// 7. 显示模型结构
	fmt.Println("\n7. Model Structure:")
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"name": model.Name,
		"version": model.Version,
		"protocols": len(model.ProtocolCenters),
		"families": model.FamilyCenters,
		"versions": model.VersionCenters,
	}, "   ", "  ")
	fmt.Println("   " + string(jsonData))

	// 8. 列出所有 Chrome 版本
	fmt.Println("\n8. Available Chrome Versions:")
	chromeProfiles := profiles.GetProfilesByBrowser(fingerprint.BrowserChrome)
	for _, p := range chromeProfiles {
		fmt.Printf("   - %s: %s\n", p.ID, p.BrowserVersion)
	}

	fmt.Println("\n=== Demo Complete ===")
}
