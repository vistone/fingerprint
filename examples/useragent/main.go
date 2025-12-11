package main

import (
	"fmt"
	"log"

	"github.com/vistone/fingerprint"
)

func main() {
	fmt.Println("=== User-Agent 示例 ===")

	// 示例 1: 根据 profile 名称获取 User-Agent（操作系统随机）
	fmt.Println("1. 根据 profile 名称获取 User-Agent（操作系统随机）:")
	profiles := []string{
		"chrome_133",
		"chrome_120",
		"firefox_135",
		"safari_16_0",
		"opera_91",
	}

	for _, name := range profiles {
		ua, err := fingerprint.GetUserAgentByProfileName(name)
		if err != nil {
			log.Printf("错误: %v", err)
			continue
		}
		fmt.Printf("  %s: %s\n", name, ua)
	}

	fmt.Println("\n2. 指定操作系统获取 User-Agent:")
	profileName := "chrome_133"
	osList := []fingerprint.OperatingSystem{
		fingerprint.OSWindows10,
		fingerprint.OSWindows11,
		fingerprint.OSMacOS15,
		fingerprint.OSLinux,
	}

	for _, os := range osList {
		ua, err := fingerprint.GetUserAgentByProfileNameWithOS(profileName, os)
		if err != nil {
			log.Printf("错误: %v", err)
			continue
		}
		fmt.Printf("  %s: %s\n", os, ua)
	}

	fmt.Println("\n3. 从 ClientProfile 对象获取 User-Agent:")
	profile, ok := fingerprint.MappedTLSClients["chrome_133"]
	if !ok {
		log.Fatal("profile chrome_133 not found")
	}

	ua, err := fingerprint.GetUserAgentFromProfile(profile)
	if err != nil {
		log.Fatalf("错误: %v", err)
	}
	fmt.Printf("  chrome_133 profile: %s\n", ua)

	fmt.Println("\n4. 随机操作系统示例:")
	for i := 0; i < 3; i++ {
		os := fingerprint.RandomOS()
		ua, err := fingerprint.GetUserAgentByProfileNameWithOS("firefox_135", os)
		if err != nil {
			log.Printf("错误: %v", err)
			continue
		}
		fmt.Printf("  随机 OS %d: %s\n", i+1, ua)
	}

	fmt.Println("\n5. 移动端 User-Agent（不需要操作系统）:")
	mobileProfiles := []string{
		"safari_ios_18_0",
		"zalando_android_mobile",
		"nike_ios_mobile",
	}

	for _, name := range mobileProfiles {
		ua, err := fingerprint.GetUserAgentByProfileName(name)
		if err != nil {
			log.Printf("错误: %v", err)
			continue
		}
		fmt.Printf("  %s: %s\n", name, ua)
	}
}

