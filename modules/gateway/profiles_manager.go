// Package gateway - Profile Configuration Manager
// translated comment
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

// translated comment
type ProfileManager struct {
	profiles   map[string]*profiles.ClientProfile // ProfileID -> Profile
	defaultID  string                             // translated comment
	configDir  string                             // translated comment
	mu         sync.RWMutex
	autoReload bool // translated comment
}

// translated comment
type ProfileManagerConfig struct {
	ConfigDir  string // translated comment
	DefaultID  string // translated comment
	AutoReload bool   // translated comment
}

// translated comment
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

// translated comment
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

	// translated comment
	if profile.ID == "" {
		return errors.ProfileInvalid(filename, "profile ID is empty")
	}

	pm.profiles[profile.ID] = &profile
	return nil
}

// translated comment
func (pm *ProfileManager) LoadAllProfiles() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// translated comment
	allProfiles := profiles.GetAll()
	for _, p := range allProfiles {
		profile := p // Copy to avoid reference issues
		pm.profiles[profile.ID] = &profile
	}

	// translated comment
	if _, err := os.Stat(pm.configDir); !os.IsNotExist(err) {
		// translated comment
		files, err := filepath.Glob(filepath.Join(pm.configDir, "*.json"))
		if err == nil {
			// translated comment
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

	// translated comment
	if len(pm.profiles) == 0 {
		// translated comment
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

// translated comment
func (pm *ProfileManager) LoadDefaultProfiles() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// translated comment
	chrome134 := &profiles.ClientProfile{
		ID:             "chrome_134_default",
		BrowserType:    core.BrowserChrome,
		BrowserVersion: "134.0.6998.35",
		OS:             core.OSWindows10,
		OSVersion:      "10.0.19045",

		// translated comment
		JSAntiDetection: &profiles.JSAntiDetection{
			// translated comment
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

			// translated comment
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

			// translated comment
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

			// translated comment
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

	// translated comment
	firefox := &profiles.ClientProfile{
		ID:             "firefox_132_default",
		BrowserType:    core.BrowserFirefox,
		BrowserVersion: "132.0",
		OS:             core.OSWindows10,
		OSVersion:      "10.0.19045",

		// translated comment
		JSAntiDetection: &profiles.JSAntiDetection{
			Automation: &profiles.AutomationAntiDetect{
				WebDriver:       false,
				Headless:        false,
				PluginsOverride: true,
			},
		},
	}

	// translated comment
	safari := &profiles.ClientProfile{
		ID:             "safari_17_default",
		BrowserType:    core.BrowserSafari,
		BrowserVersion: "17.2",
		OS:             core.OSMacOS14,
		OSVersion:      "14.2",

		// translated comment
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

// translated comment
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

// translated comment
func (pm *ProfileManager) GetDefaultProfile() (*profiles.ClientProfile, error) {
	return pm.GetProfile(pm.defaultID)
}

// translated comment
func (pm *ProfileManager) SetDefaultProfile(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, ok := pm.profiles[id]; !ok {
		return errors.ProfileNotFound(id)
	}

	pm.defaultID = id
	return nil
}

// translated comment
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

// translated comment
func (pm *ProfileManager) AddProfile(profile *profiles.ClientProfile) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if profile.ID == "" {
		return errors.RequiredField("profile.ID")
	}

	pm.profiles[profile.ID] = profile
	return nil
}

// translated comment
func (pm *ProfileManager) RemoveProfile(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if id == pm.defaultID {
		return errors.ProfileRemoveDefault(id)
	}

	delete(pm.profiles, id)
	return nil
}

// translated comment
func (pm *ProfileManager) SaveProfile(id string) error {
	pm.mu.RLock()
	profile, ok := pm.profiles[id]
	pm.mu.RUnlock()

	if !ok {
		return errors.ProfileNotFound(id)
	}

	// translated comment
	if err := os.MkdirAll(pm.configDir, 0755); err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeProfileSaveFailed,
			"failed to create config dir", err).WithDetail("profile_id", id)
	}

	// translated comment
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeProfileSaveFailed,
			"failed to marshal profile", err).WithDetail("profile_id", id)
	}

	// translated comment
	filename := filepath.Join(pm.configDir, fmt.Sprintf("%s.json", id))
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeProfileSaveFailed,
			"failed to write profile file", err).WithDetail("profile_id", id)
	}

	return nil
}

// translated comment
func (pm *ProfileManager) ExportProfilesExample() error {
	// translated comment
	if err := pm.LoadDefaultProfiles(); err != nil {
		return err
	}

	// translated comment
	for id := range pm.profiles {
		if err := pm.SaveProfile(id); err != nil {
			return err
		}
	}

	return nil
}

// translated comment
func (pm *ProfileManager) ReloadProfile(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// translated comment
	if _, ok := pm.profiles[id]; !ok {
		return errors.ProfileNotFound(id)
	}

	// translated comment
	filename := filepath.Join(pm.configDir, fmt.Sprintf("%s.json", id))
	if _, err := os.Stat(filename); err != nil {
		// translated comment
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

// translated comment
func (pm *ProfileManager) ReloadAll() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// translated comment
	allProfiles := profiles.GetAll()
	for _, p := range allProfiles {
		profile := p
		pm.profiles[profile.ID] = &profile
	}

	// translated comment
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

// translated comment
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

// translated comment
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

// translated comment
func (pm *ProfileManager) Count() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.profiles)
}

// translated comment
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

	// translated comment
	clone := *source
	clone.ID = newID
	// translated comment
	clone.Name = fmt.Sprintf("%s (Clone)", source.Name)

	pm.profiles[newID] = &clone
	return nil
}
