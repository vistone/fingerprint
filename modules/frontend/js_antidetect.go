// Package frontend - JavaScript anti-detection countermeasure code generation
package frontend

import (
	"encoding/json"
	"fmt"

	"github.com/vistone/fingerprint/modules/profiles"
)

// JSAntiDetectCodeGenerator JavaScript anti-detection code generator
type JSAntiDetectCodeGenerator struct {
	profile *profiles.ClientProfile
}

// NewJSAntiDetectCodeGenerator creates a new anti-detection code generator
func NewJSAntiDetectCodeGenerator(profile *profiles.ClientProfile) *JSAntiDetectCodeGenerator {
	return &JSAntiDetectCodeGenerator{
		profile: profile,
	}
}

// GenerateWebGPUCode generates WebGPU countermeasure code
func (g *JSAntiDetectCodeGenerator) GenerateWebGPUCode() string {
	if g.profile.JSAntiDetection == nil || g.profile.JSAntiDetection.WebGPU == nil {
		return ""
	}

	gpu := g.profile.JSAntiDetection.WebGPU
	if !gpu.Available {
		return webGPUDisabledCode()
	}
	return g.generateWebGPUAvailableCode(gpu)
}

func webGPUDisabledCode() string {
	return `
		// Hide WebGPU support
		if (typeof navigator !== 'undefined') {
			Object.defineProperty(navigator, 'gpu', {
				value: undefined,
				configurable: true
			});
		}
		`
}

func (g *JSAntiDetectCodeGenerator) generateWebGPUAvailableCode(gpu *profiles.WebGPUAntiDetect) string {
	featureFlags := []string{}
	if gpu.FeatureFlags != nil {
		featureFlags = gpu.FeatureFlags
	}

	limitsJSON, _ := json.Marshal(gpu.LimitValues)

	return fmt.Sprintf(`
		// WebGPU countermeasure - GPU device spoofing
		(function() {
			const mockGPU = {
				requestAdapter: async function(options) {
					return {
						requestDevice: async function(descriptor) {
							return {
								label: '%s',
								lost: Promise.resolve(),
								features: new Set(%s),
								limits: %s,
								queue: {
									onSubmittedWorkDone: async function() {},
									submit: function() {},
									writeBuffer: function() {},
									writeTexture: function() {},
									copyExternalImageToTexture: function() {},
									createCommandEncoder: function() {
										return {
											finish: function() {},
											copyBufferToBuffer: function() {},
											copyBufferToTexture: function() {},
											copyTextureToBuffer: function() {},
											copyTextureToTexture: function() {},
											clearBuffer: function() {}
										};
									}
								},
								createShaderModule: function() { return {}; },
								createPipelineLayout: function() { return {}; },
								createComputePipeline: function() { return {}; },
								createRenderPipeline: function() { return {}; },
								createBuffer: function() { return { destroy: function() {}, getMappedRange: function() {} }; },
								createTexture: function() { return { createView: function() { return {}; }, destroy: function() {} }; },
								createSampler: function() { return {}; },
								createQuerySet: function() { return {}; },
								createBindGroup: function() { return {}; },
								createBindGroupLayout: function() { return {}; },
								importExternalTexture: function() { return {}; },
								pushErrorScope: function() {},
								popErrorScope: async function() { return null; }
							};
						},
						get enableFeatures() {
							return new Set(%s);
						}
					};
				},
				wgslLanguageFeatures: new Set(['readonly_and_readwrite_storage_textures', 'packed_4x8_integer_dot_product'])
			};

			if (typeof navigator !== 'undefined') {
				Object.defineProperty(navigator, 'gpu', {
					value: mockGPU,
					configurable: true
				});
			}
		})();
		`, gpu.AdapterName, stringifyList(featureFlags), string(limitsJSON), stringifyList(featureFlags))
}

// GenerateMediaDevicesCode generates MediaDevices countermeasure code
func (g *JSAntiDetectCodeGenerator) GenerateMediaDevicesCode() string {
	if g.profile.JSAntiDetection == nil || g.profile.JSAntiDetection.MediaDevices == nil {
		return ""
	}

	md := g.profile.JSAntiDetection.MediaDevices

	allDevices := buildMediaDeviceList(md)
	devicesJSON, _ := json.Marshal(allDevices)
	return fmt.Sprintf(`
		// MediaDevices countermeasure - device list spoofing
		(function() {
			const mockDevices = %s;

			if (navigator.mediaDevices) {
				const origEnumerate = navigator.mediaDevices.enumerateDevices.bind(navigator.mediaDevices);
				navigator.mediaDevices.enumerateDevices = async function() {
					return mockDevices.map(d => ({
						deviceId: d.deviceId,
						groupId: d.groupId,
						kind: d.kind,
						label: d.label,
						toJSON: function() { return this; }
					}));
				};

				// Ensure getUserMedia returns valid stream
				const origGetUserMedia = navigator.mediaDevices.getUserMedia.bind(navigator.mediaDevices);
				navigator.mediaDevices.getUserMedia = async function(constraints) {
					try {
						return await origGetUserMedia(constraints);
					} catch (e) {
						// If device unavailable, return virtual stream
						const canvas = document.createElement('canvas');
						canvas.width = 1280;
						canvas.height = 720;
						const videoStream = canvas.captureStream(30);
						const audioContext = new (window.AudioContext || window.webkitAudioContext)();
						const audioStream = audioContext.createMediaStreamDestination().stream;
						const combinedStream = new MediaStream();
						videoStream.getVideoTracks().forEach(track => combinedStream.addTrack(track));
						audioStream.getAudioTracks().forEach(track => combinedStream.addTrack(track));
						return combinedStream;
					}
				};
			}
		})();
		`, string(devicesJSON))
}

func buildMediaDeviceList(md *profiles.MediaDevicesAntiDetect) []map[string]string {
	allDevices := make([]map[string]string, 0, len(md.VideoInputs)+len(md.AudioInputs)+len(md.AudioOutputs))
	appendMediaDevicesByKind(&allDevices, md.VideoInputs, "videoinput")
	appendMediaDevicesByKind(&allDevices, md.AudioInputs, "audioinput")
	appendMediaDevicesByKind(&allDevices, md.AudioOutputs, "audiooutput")
	return allDevices
}

func appendMediaDevicesByKind(allDevices *[]map[string]string, devices []*profiles.MediaDeviceInfo, kind string) {
	for _, dev := range devices {
		*allDevices = append(*allDevices, map[string]string{
			"deviceId": dev.DeviceID,
			"groupId":  dev.GroupID,
			"kind":     kind,
			"label":    dev.Label,
		})
	}
}

// GeneratePermissionsCode generates Permissions API countermeasure code
func (g *JSAntiDetectCodeGenerator) GeneratePermissionsCode() string {
	if g.profile.JSAntiDetection == nil || g.profile.JSAntiDetection.Permissions == nil {
		return ""
	}

	perms := g.profile.JSAntiDetection.Permissions

	// Build permission state mapping
	permMap := perms.PermissionState
	if permMap == nil {
		permMap = make(map[string]string)
	}

	// Set default permission states
	if _, ok := permMap["camera"]; !ok && perms.AccessCamera {
		permMap["camera"] = "granted"
	}
	if _, ok := permMap["microphone"]; !ok && perms.AccessMicrophone {
		permMap["microphone"] = "granted"
	}
	if _, ok := permMap["notifications"]; !ok && perms.ShowNotification {
		permMap["notifications"] = "granted"
	}
	if _, ok := permMap["geolocation"]; !ok && perms.Geolocation {
		permMap["geolocation"] = "granted"
	}

	permJSON, _ := json.Marshal(permMap)

	code := fmt.Sprintf(`
		// Permissions API countermeasure - permission state spoofing
		(function() {
			const permissionMap = %s;

			if (navigator.permissions) {
				const origQuery = navigator.permissions.query.bind(navigator.permissions);
				navigator.permissions.query = async function(parameters) {
					const name = parameters.name || parameters;
					const state = permissionMap[name] || 'prompt';
					return {
						state: state,
						onchange: null,
						toString: function() { return '[object PermissionStatus]'; },
						addEventListener: function() {},
						removeEventListener: function() {}
					};
				};
			}
		})();
		`, string(permJSON))

	return code
}

// GenerateAutomationCode generates Automation countermeasure code
