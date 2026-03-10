// Package gateway - Profile Configuration Manager
// 管理多个 ClientProfile 配置文件
package gateway

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/errors"
	"github.com/vistone/fingerprint/modules/profiles"
)

// ProfileManager Profile 配置管理器
type ProfileManager struct {
	profiles   map[string]*profiles.ClientProfile // ProfileID -> Profile
	defaultID  string                             // 默认 Profile ID
	configDir  string                             // 配置文件目录
	mu         sync.RWMutex
	autoReload bool // 是否自动重新加载配置
}

// ProfileManagerConfig ProfileManager 配置
type ProfileManagerConfig struct {
	ConfigDir  string // 配置文件目录
	DefaultID  string // 默认 Profile ID
	AutoReload bool   // 是否自动重新加载
}

// NewProfileManager 创建新的 ProfileManager
func NewProfileManager(config *ProfileManagerConfig) *ProfileManager {
	if config == nil {
		config = &ProfileManagerConfig{
			ConfigDir:  "./profiles",
			DefaultID:  "chrome_134_default",
			AutoReload: false,
		}
	}

	return &ProfileManager{
		profiles:   make(map[string]*profiles.ClientProfile),
		defaultID:  config.DefaultID,
		configDir:  config.ConfigDir,
		autoReload: config.AutoReload,
	}
}

// LoadProfile 从文件加载单个 Profile
func (pm *ProfileManager) LoadProfile(filename string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	fullPath := filepath.Join(pm.configDir, filename)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeProfileLoadFailed,
			"failed to read profile file", err).WithDetail("filename", filename)
	}

	var profile profiles.ClientProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeProfileInvalid,
			"failed to parse profile file", err).WithDetail("filename", filename)
	}

	// 验证 Profile ID
	if profile.ID == "" {
		return errors.ProfileInvalid(filename, "profile ID is empty")
	}

	pm.profiles[profile.ID] = &profile
	return nil
}

// LoadAllProfiles 从目录加载所有 Profile，并汇总 profiles 模块中的所有 profiles
func (pm *ProfileManager) LoadAllProfiles() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 首先加载 profiles 模块中的所有 profiles（这是完整列表）
	allProfiles := profiles.GetAll()
	for _, p := range allProfiles {
		profile := p // Copy to avoid reference issues
		pm.profiles[profile.ID] = &profile
	}

	// 然后尝试从配置目录加载本地 profile 文件（可以覆盖内置的）
	if _, err := os.Stat(pm.configDir); !os.IsNotExist(err) {
		// 读取目录中的所有 JSON 文件
		files, err := filepath.Glob(filepath.Join(pm.configDir, "*.json"))
		if err == nil {
			// 加载每个文件（这会覆盖同 ID 的内置 profile）
			for _, file := range files {
				if data, err := os.ReadFile(file); err == nil {
					var profile profiles.ClientProfile
					if err := json.Unmarshal(data, &profile); err == nil && profile.ID != "" {
						pm.profiles[profile.ID] = &profile
					}
				}
			}
		}
	}

	// 如果仍然没有加载到任何 profiles（profiles 模块为空），加载默认的 3 个
	if len(pm.profiles) == 0 {
		// 直接在这里初始化 3 个默认 profiles
		defaultChrome := profiles.ClientProfile{
			ID:             "chrome_134_default",
			BrowserType:    core.BrowserChrome,
			BrowserVersion: "134.0.6998.35",
			OS:             core.OSWindows10,
			OSVersion:      "10.0.19045",
			Name:           "Chrome 134",
		}
		pm.profiles["chrome_134_default"] = &defaultChrome

		defaultFirefox := profiles.ClientProfile{
			ID:             "firefox_132_default",
			BrowserType:    core.BrowserFirefox,
			BrowserVersion: "132.0",
			OS:             core.OSWindows10,
			OSVersion:      "10.0.19045",
			Name:           "Firefox 132",
		}
		pm.profiles["firefox_132_default"] = &defaultFirefox

		defaultSafari := profiles.ClientProfile{
			ID:             "safari_17_default",
			BrowserType:    core.BrowserSafari,
			BrowserVersion: "17.0",
			OS:             core.OSMacOS14,
			OSVersion:      "14.0",
			Name:           "Safari 17",
		}
		pm.profiles["safari_17_default"] = &defaultSafari
	}

	return nil
}

// LoadDefaultProfiles 加载默认内置 Profile
func (pm *ProfileManager) LoadDefaultProfiles() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 创建 Chrome 134 默认配置（带完整 P3 反检测）
	chrome134 := &profiles.ClientProfile{
		ID:             "chrome_134_default",
		BrowserType:    core.BrowserChrome,
		BrowserVersion: "134.0.6998.35",
		OS:             core.OSWindows10,
		OSVersion:      "10.0.19045",

		// P3 反检测配置
		JSAntiDetection: &profiles.JSAntiDetection{
			// WebGPU 对抗点
			WebGPU: &profiles.WebGPUAntiDetect{
				Available:   true,
				AdapterName: "ANGLE (Intel, Intel(R) HD Graphics 630, OpenGL 4.5.0 - Build 31.0.101.2127)",
				DeviceType:  "integrated-gpu",
				VendorID:    "0x8086",
				FeatureFlags: []string{
					"depth-clip-control",
					"depth32float-stencil8",
					"indirect-first-instance",
					"shader-f16",
					"timestamp-query",
					"texture-compression-bc",
					"texture-compression-etc2",
					"texture-compression-astc",
				},
				LimitValues: map[string]uint64{
					"maxTextureDimension1D":                     16384,
					"maxTextureDimension2D":                     16384,
					"maxTextureDimension3D":                     2048,
					"maxTextureArrayLayers":                     2048,
					"maxBindGroups":                             4,
					"maxDynamicUniformBuffersPerPipelineLayout": 8,
					"maxDynamicStorageBuffersPerPipelineLayout": 4,
					"maxSampledTexturesPerShaderStage":          16,
					"maxSamplersPerShaderStage":                 16,
					"maxStorageBuffersPerShaderStage":           8,
					"maxStorageTexturesPerShaderStage":          4,
					"maxUniformBuffersPerShaderStage":           12,
					"maxUniformBufferBindingSize":               65536,
					"maxStorageBufferBindingSize":               134217728,
					"maxVertexBuffers":                          8,
					"maxVertexAttributes":                       16,
					"maxVertexBufferArrayStride":                2048,
					"maxComputeWorkgroupStorageSize":            16384,
					"maxComputeInvocationsPerWorkgroup":         256,
					"maxComputeWorkgroupSizeX":                  256,
					"maxComputeWorkgroupSizeY":                  256,
					"maxComputeWorkgroupSizeZ":                  64,
					"maxComputeWorkgroupsPerDimension":          65535,
				},
				BackendType: "d3d12",
			},

			// MediaDevices 对抗点
			MediaDevices: &profiles.MediaDevicesAntiDetect{
				VideoInputs: []*profiles.MediaDeviceInfo{
					{
						DeviceID:  "default",
						GroupID:   "group_video_default",
						Kind:      "videoinput",
						Label:     "Integrated Camera (04f2:b6d9)",
						VendorID:  "04f2",
						ProductID: "b6d9",
					},
					{
						DeviceID:  "videoinput_2",
						GroupID:   "group_video_2",
						Kind:      "videoinput",
						Label:     "USB Camera (046d:0825)",
						VendorID:  "046d",
						ProductID: "0825",
					},
				},
				AudioInputs: []*profiles.MediaDeviceInfo{
					{
						DeviceID: "default",
						GroupID:  "group_audio_default",
						Kind:     "audioinput",
						Label:    "Microphone Array (Realtek(R) Audio)",
					},
				},
				AudioOutputs: []*profiles.MediaDeviceInfo{
					{
						DeviceID: "default",
						GroupID:  "group_audio_output",
						Kind:     "audiooutput",
						Label:    "Speakers (Realtek(R) Audio)",
					},
				},
			},

			// Permissions 对抗点
			Permissions: &profiles.PermissionsAntiDetect{
				PermissionState: map[string]string{
					"camera":        "prompt",
					"microphone":    "prompt",
					"geolocation":   "prompt",
					"notifications": "granted",
				},
				AccessCamera:     true,
				AccessMicrophone: true,
				ShowNotification: true,
			},

			// Automation 对抗点
			Automation: &profiles.AutomationAntiDetect{
				WebDriver:        false,
				Headless:         false,
				ChromeDebugPort:  false,
				Phantom:          false,
				Selenium:         true,
				Puppeteer:        true,
				Playwright:       true,
				PluginsOverride:  true,
				LanguageOverride: true,
				ProductOverride:  true,
				VendorOverride:   true,
				RuntimeOverride:  true,
			},
		},
	}

	// 创建 Firefox 默认配置（轻量级）
	firefox := &profiles.ClientProfile{
		ID:             "firefox_132_default",
		BrowserType:    core.BrowserFirefox,
		BrowserVersion: "132.0",
		OS:             core.OSWindows10,
		OSVersion:      "10.0.19045",

		// Firefox 只启用基础 Automation 对抗
		JSAntiDetection: &profiles.JSAntiDetection{
			Automation: &profiles.AutomationAntiDetect{
				WebDriver:       false,
				Headless:        false,
				PluginsOverride: true,
			},
		},
	}

	// 创建 Safari 默认配置（macOS）
	safari := &profiles.ClientProfile{
		ID:             "safari_17_default",
		BrowserType:    core.BrowserSafari,
		BrowserVersion: "17.2",
		OS:             core.OSMacOS14,
		OSVersion:      "14.2",

		// Safari 只启用基础对抗
		JSAntiDetection: &profiles.JSAntiDetection{
			Permissions: &profiles.PermissionsAntiDetect{
				PermissionState: map[string]string{
					"camera":      "prompt",
					"microphone":  "prompt",
					"geolocation": "prompt",
				},
			},
		},
	}

	pm.profiles["chrome_134_default"] = chrome134
	pm.profiles["firefox_132_default"] = firefox
	pm.profiles["safari_17_default"] = safari

	return nil
}

// GetProfile 获取指定 ID 的 Profile（返回副本，防止外部修改内部状态）
func (pm *ProfileManager) GetProfile(id string) (*profiles.ClientProfile, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	profile, ok := pm.profiles[id]
	if !ok {
		return nil, errors.ProfileNotFound(id)
	}
	clone := *profile
	return &clone, nil
}

// GetDefaultProfile 获取默认 Profile
func (pm *ProfileManager) GetDefaultProfile() (*profiles.ClientProfile, error) {
	return pm.GetProfile(pm.defaultID)
}

// SetDefaultProfile 设置默认 Profile ID
func (pm *ProfileManager) SetDefaultProfile(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, ok := pm.profiles[id]; !ok {
		return errors.ProfileNotFound(id)
	}

	pm.defaultID = id
	return nil
}

// ListProfiles 列出所有 Profile ID
func (pm *ProfileManager) ListProfiles() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	ids := make([]string, 0, len(pm.profiles))
	for id := range pm.profiles {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// AddProfile 动态添加 Profile
func (pm *ProfileManager) AddProfile(profile *profiles.ClientProfile) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if profile.ID == "" {
		return errors.RequiredField("profile.ID")
	}

	pm.profiles[profile.ID] = profile
	return nil
}

// RemoveProfile 移除 Profile
func (pm *ProfileManager) RemoveProfile(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if id == pm.defaultID {
		return errors.ProfileRemoveDefault(id)
	}

	delete(pm.profiles, id)
	return nil
}

// SaveProfile 保存 Profile 到文件
func (pm *ProfileManager) SaveProfile(id string) error {
	pm.mu.RLock()
	profile, ok := pm.profiles[id]
	pm.mu.RUnlock()

	if !ok {
		return errors.ProfileNotFound(id)
	}

	// 确保目录存在
	if err := os.MkdirAll(pm.configDir, 0755); err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeProfileSaveFailed,
			"failed to create config dir", err).WithDetail("profile_id", id)
	}

	// 序列化 Profile
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeProfileSaveFailed,
			"failed to marshal profile", err).WithDetail("profile_id", id)
	}

	// 写入文件
	filename := filepath.Join(pm.configDir, fmt.Sprintf("%s.json", id))
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeProfileSaveFailed,
			"failed to write profile file", err).WithDetail("profile_id", id)
	}

	return nil
}

// ExportProfilesExample 导出示例配置文件（用于文档和初始化）
func (pm *ProfileManager) ExportProfilesExample() error {
	// 加载默认配置
	if err := pm.LoadDefaultProfiles(); err != nil {
		return err
	}

	// 导出所有默认配置
	for id := range pm.profiles {
		if err := pm.SaveProfile(id); err != nil {
			return err
		}
	}

	return nil
}

// ReloadProfile 重新加载指定 Profile
func (pm *ProfileManager) ReloadProfile(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 检查是否是内置 profile
	if _, ok := pm.profiles[id]; !ok {
		return errors.ProfileNotFound(id)
	}

	// 尝试从文件重新加载
	filename := filepath.Join(pm.configDir, fmt.Sprintf("%s.json", id))
	if _, err := os.Stat(filename); err != nil {
		// 文件不存在，返回当前内存中的版本
		return nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeProfileLoadFailed,
			"failed to read profile file", err).WithDetail("profile_id", id)
	}

	var profile profiles.ClientProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeProfileInvalid,
			"failed to parse profile", err).WithDetail("profile_id", id)
	}

	pm.profiles[id] = &profile
	return nil
}

// ReloadAll 重新加载所有 Profile
func (pm *ProfileManager) ReloadAll() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 重新加载内置 profiles
	allProfiles := profiles.GetAll()
	for _, p := range allProfiles {
		profile := p
		pm.profiles[profile.ID] = &profile
	}

	// 从配置目录加载本地 profile 文件
	if _, err := os.Stat(pm.configDir); !os.IsNotExist(err) {
		files, _ := filepath.Glob(filepath.Join(pm.configDir, "*.json"))
		for _, file := range files {
			if data, err := os.ReadFile(file); err == nil {
				var profile profiles.ClientProfile
				if err := json.Unmarshal(data, &profile); err == nil && profile.ID != "" {
					pm.profiles[profile.ID] = &profile
				}
			}
		}
	}

	return nil
}

// GetProfilesByBrowser 按浏览器类型获取 Profiles
func (pm *ProfileManager) GetProfilesByBrowser(browser core.BrowserType) []*profiles.ClientProfile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var result []*profiles.ClientProfile
	for _, p := range pm.profiles {
		if p.BrowserType == browser {
			result = append(result, p)
		}
	}
	return result
}

// GetProfilesByOS 按操作系统获取 Profiles
func (pm *ProfileManager) GetProfilesByOS(os core.OperatingSystem) []*profiles.ClientProfile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var result []*profiles.ClientProfile
	for _, p := range pm.profiles {
		if p.OS == os {
			result = append(result, p)
		}
	}
	return result
}

// Count 返回 Profile 数量
func (pm *ProfileManager) Count() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.profiles)
}

// CloneProfile 克隆 Profile
func (pm *ProfileManager) CloneProfile(sourceID, newID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if newID == "" {
		return errors.RequiredField("newID")
	}

	if _, exists := pm.profiles[newID]; exists {
		return fmt.Errorf("profile ID %q already exists", newID)
	}

	source, ok := pm.profiles[sourceID]
	if !ok {
		return errors.ProfileNotFound(sourceID)
	}

	// 创建副本
	clone := *source
	clone.ID = newID
	// 修改名称以区分
	clone.Name = fmt.Sprintf("%s (Clone)", source.Name)

	pm.profiles[newID] = &clone
	return nil
}
