// Package gateway - Profile Configuration Manager
// Manages multiple ClientProfile configuration files
package gateway

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/errors"
	"github.com/vistone/fingerprint/modules/profiles"
)

// ProfileManager is the Profile configuration manager
type ProfileManager struct {
	profiles   map[string]*profiles.ClientProfile // ProfileID -> Profile
	defaultID  string                             // Default Profile ID
	configDir  string                             // Configuration file directory
	mu         sync.RWMutex
	autoReload bool // Whether to automatically reload configuration
}

// ProfileManagerConfig is the ProfileManager configuration
type ProfileManagerConfig struct {
	ConfigDir  string // Configuration file directory
	DefaultID  string // Default Profile ID
	AutoReload bool   // Whether to automatically reload
}

// NewProfileManager creates a new ProfileManager
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

// LoadProfile loads a single Profile from a file
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

	// Validate Profile ID
	if profile.ID == "" {
		return errors.ProfileInvalid(filename, "profile ID is empty")
	}

	pm.profiles[profile.ID] = &profile
	return nil
}

// LoadAllProfiles loads all Profiles from directory and aggregates all profiles from the profiles module
func (pm *ProfileManager) LoadAllProfiles() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// First load all profiles from the profiles module (this is the complete list)
	allProfiles := profiles.GetAll()
	for _, p := range allProfiles {
		profile := p // Copy to avoid reference issues
		pm.profiles[profile.ID] = &profile
	}

	// Then try to load local profile files from config directory (can override built-in ones)
	if _, err := os.Stat(pm.configDir); !os.IsNotExist(err) {
		// Read all JSON files in the directory
		files, err := filepath.Glob(filepath.Join(pm.configDir, "*.json"))
		if err == nil {
			// Load each file (this overrides built-in profiles with the same ID)
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

	// If still no profiles loaded (profiles module is empty), load 3 defaults
	if len(pm.profiles) == 0 {
		// Initialize 3 default profiles directly here
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

// LoadDefaultProfiles loads default built-in Profiles
func (pm *ProfileManager) LoadDefaultProfiles() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Create Chrome 134 default configuration (with full P3 anti-detection)
	chrome134 := &profiles.ClientProfile{
		ID:             "chrome_134_default",
		BrowserType:    core.BrowserChrome,
		BrowserVersion: "134.0.6998.35",
		OS:             core.OSWindows10,
		OSVersion:      "10.0.19045",

		// P3 anti-detection configuration
		JSAntiDetection: &profiles.JSAntiDetection{
			// WebGPU anti-detection point
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

			// MediaDevices anti-detection point
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

			// Permissions anti-detection point
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

			// Automation anti-detection point
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

	// Create Firefox default configuration (lightweight)
	firefox := &profiles.ClientProfile{
		ID:             "firefox_132_default",
		BrowserType:    core.BrowserFirefox,
		BrowserVersion: "132.0",
		OS:             core.OSWindows10,
		OSVersion:      "10.0.19045",

		// Firefox only enables basic Automation anti-detection
		JSAntiDetection: &profiles.JSAntiDetection{
			Automation: &profiles.AutomationAntiDetect{
				WebDriver:       false,
				Headless:        false,
				PluginsOverride: true,
			},
		},
	}

	// Create Safari default configuration (macOS)
	safari := &profiles.ClientProfile{
		ID:             "safari_17_default",
		BrowserType:    core.BrowserSafari,
		BrowserVersion: "17.2",
		OS:             core.OSMacOS14,
		OSVersion:      "14.2",

		// Safari only enables basic anti-detection
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

// GetProfile returns the Profile with the specified ID (returns a copy to prevent external modification of internal state)
