// Package frontend - JavaScript anti-detection countermeasure code generation
package frontend

import (
	"encoding/json"
	"fmt"
	"strings"

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

	// If WebGPU not supported, return hiding code
	if !gpu.Available {
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

	// Build WebGPU adapter and device info
	featureFlags := []string{}
	if gpu.FeatureFlags != nil {
		featureFlags = gpu.FeatureFlags
	}

	limitsJSON, _ := json.Marshal(gpu.LimitValues)

	code := fmt.Sprintf(`
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

	return code
}

// GenerateMediaDevicesCode generates MediaDevices countermeasure code
func (g *JSAntiDetectCodeGenerator) GenerateMediaDevicesCode() string {
	if g.profile.JSAntiDetection == nil || g.profile.JSAntiDetection.MediaDevices == nil {
		return ""
	}

	md := g.profile.JSAntiDetection.MediaDevices

	// Build device list
	const (
		videoInput  = "videoinput"
		audioInput  = "audioinput"
		audioOutput = "audiooutput"
	)

	var allDevices []map[string]string
	for _, dev := range md.VideoInputs {
		allDevices = append(allDevices, map[string]string{
			"deviceId": dev.DeviceID,
			"groupId":  dev.GroupID,
			"kind":     videoInput,
			"label":    dev.Label,
		})
	}
	for _, dev := range md.AudioInputs {
		allDevices = append(allDevices, map[string]string{
			"deviceId": dev.DeviceID,
			"groupId":  dev.GroupID,
			"kind":     audioInput,
			"label":    dev.Label,
		})
	}
	for _, dev := range md.AudioOutputs {
		allDevices = append(allDevices, map[string]string{
			"deviceId": dev.DeviceID,
			"groupId":  dev.GroupID,
			"kind":     audioOutput,
			"label":    dev.Label,
		})
	}

	devicesJSON, _ := json.Marshal(allDevices)

	code := fmt.Sprintf(`
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

	return code
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
func (g *JSAntiDetectCodeGenerator) GenerateAutomationCode() string {
	if g.profile.JSAntiDetection == nil || g.profile.JSAntiDetection.Automation == nil {
		return ""
	}

	auto := g.profile.JSAntiDetection.Automation

	// Build countermeasure code
	code := ""

	// 1. Hide webdriver marker
	if !auto.WebDriver {
		code += `
		// Hide webdriver marker
		Object.defineProperty(navigator, 'webdriver', {
			get: () => undefined,
			configurable: true
		});
		`
	}

	// 2. Hide headless feature
	if auto.Headless {
		code += `
		// Hide headless browser features
		Object.defineProperty(navigator, 'vendor', {
			get: () => 'Google Inc.' || navigator.vendor,
			configurable: true
		});
		`
	}

	// 3. Hide phantomjs
	if auto.Phantom {
		code += `
		// Hide phantomjs features
		Object.defineProperty(window, 'callPhantom', {
			get: () => undefined,
			configurable: true
		});
		`
	}

	// 4. Hide selenium
	if auto.Selenium {
		code += `
		// Hide selenium driver features
		Object.defineProperty(navigator, 'driverName', {
			get: () => undefined,
			configurable: true
		});
		`
	}

	// 5. Hide puppeteer and playwright
	if auto.Puppeteer || auto.Playwright {
		code += `
		// Hide automation tool features
		window.__puppeteer_eval = undefined;
		window.CDP = undefined;
		`
	}

	// 6. Override plugins
	if auto.PluginsOverride {
		code += `
		// Override plugins array
		Object.defineProperty(navigator, 'plugins', {
			get: () => [
				{ name: 'Native Client Plugin', description: 'Native Client Executable', version: '', filename: 'internal-nacl-plugin' },
				{ name: 'Chrome PDF Plugin', description: 'Portable Document Format', version: '1', filename: 'internal-pdf-viewer' },
				{ name: 'Chrome PDF Viewer', description: 'Portable Document Format', version: '1', filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai' }
			],
			configurable: true
		});

		// Override mimeTypes
		Object.defineProperty(navigator, 'mimeTypes', {
			get: () => [
				{ type: 'application/x-nacl', description: 'Native Client Executable', suffixes: 'nexe' },
				{ type: 'application/x-pnacl', description: 'Portable Native Client Executable', suffixes: 'pexe' },
				{ type: 'application/pdf', description: 'Portable Document Format', suffixes: 'pdf' }
			],
			configurable: true
		});
		`
	}

	// 7. Override language
	if auto.LanguageOverride && g.profile.Headers != nil {
		// Infer language from UA or Accept-Language
		lang := "en-US"
		if g.profile.Headers.AcceptLanguage != "" {
			parts := strings.Split(g.profile.Headers.AcceptLanguage, ",")
			if len(parts) > 0 {
				lang = strings.TrimSpace(parts[0])
			}
		}
		code += fmt.Sprintf(`
		// Override language attribute
		Object.defineProperty(navigator, 'language', {
			get: () => '%s',
			configurable: true
		});
		Object.defineProperty(navigator, 'languages', {
			get: () => ['%s'],
			configurable: true
		});
		`, lang, lang)
	}

	// 8. Override product
	if auto.ProductOverride {
		code += `
		// Override product attribute
		Object.defineProperty(navigator, 'product', {
			get: () => 'Gecko',
			configurable: true
		});
		`
	}

	// 9. Override vendor
	if auto.VendorOverride {
		code += `
		// Override vendor attribute
		Object.defineProperty(navigator, 'vendor', {
			get: () => 'Google Inc.',
			configurable: true
		});
		`
	}

	// 10. Hide runtime detection features
	if auto.RuntimeOverride {
		code += `
		// Hide runtime detection features
		const OriginalFunction = Function;
		const OriginalGeneratorFunction = (function*(){}).constructor;
		const OriginalAsyncFunction = (async function(){}).constructor;

		try {
			Function = new Proxy(OriginalFunction, {
				construct() {
					arguments[arguments.length - 1] = arguments[arguments.length - 1].replace(/headless/, '');
					return Reflect.construct(OriginalFunction, arguments);
				}
			});
		} catch (e) {}
		`
	}

	return fmt.Sprintf(`
		// Automation countermeasure - hide automation tool detection
		(function() {
			'use strict';
			%s
		})();
		`, code)
}

// GenerateCrossLayerConsistencyCode generates cross-layer consistency validation code
func (g *JSAntiDetectCodeGenerator) GenerateCrossLayerConsistencyCode() string {
	if g.profile.JSAntiDetection == nil {
		return ""
	}

	// Extract layer info from configuration file
	ua := ""
	secChUA := ""

	if g.profile.Headers != nil {
		ua = g.profile.Headers.UserAgent
		secChUA = g.profile.Headers.SecCHUA
	}

	code := fmt.Sprintf(`
		// Cross-layer consistency validation - ensure UA/CH/JS/TCP-IP consistency
		(function() {
			const expectedUA = '%s';
			const expectedSecChUA = '%s';

			// 1. UA consistency check
			if (navigator.userAgent !== expectedUA) {
				console.warn('[Fingerprint] UA layer inconsistency');
			}

			// 2. Client Hints consistency check
			if (typeof navigator.userAgentData !== 'undefined') {
				const uaData = navigator.userAgentData;
				// Verify browser brand, version and other info consistency
				const brands = uaData.brands || [];
				const expectedBrands = %s;

				const brandsMatch = brands.length === expectedBrands.length &&
					brands.every((b, i) => b.brand === expectedBrands[i].brand && b.version === expectedBrands[i].version);

				if (!brandsMatch) {
					console.warn('[Fingerprint] CH layer brand inconsistency');
				}
			}

			// 3. JavaScript layer consistency check
			const jsUA = navigator.userAgent;
			const jsProduct = navigator.product;
			const jsVendor = navigator.vendor;
			const jsPlatform = navigator.platform;

			if (jsUA !== expectedUA) {
				console.warn('[Fingerprint] JS layer UA inconsistency');
			}

			// 4. Hardware info consistency
			const hardwareCells = {
				cores: navigator.hardwareConcurrency,
				memory: navigator.deviceMemory,
				maxTouchPoints: navigator.maxTouchPoints,
				platform: navigator.platform
			};

			// 5. Overall consistency report
			const consistencyReport = {
				ua: { expected: expectedUA, actual: jsUA, match: expectedUA === jsUA },
				layers: {
					http: { ua, ch: secChUA },
					js: { ua: jsUA, product: jsProduct, vendor: jsVendor, platform: jsPlatform },
					hardware: hardwareCells
				}
			};

			// Attach consistency report to window object for later reporting
			window.__fingerprintConsistency = consistencyReport;
		})();
		`, escapeQuotes(ua), escapeQuotes(secChUA), g.generateBrandsList())

	return code
}

// GenerateFullAntiDetectionCode generates complete anti-detection code
func (g *JSAntiDetectCodeGenerator) GenerateFullAntiDetectionCode() string {
	var parts []string

	if code := g.GenerateWebGPUCode(); code != "" {
		parts = append(parts, code)
	}

	if code := g.GenerateMediaDevicesCode(); code != "" {
		parts = append(parts, code)
	}

	if code := g.GeneratePermissionsCode(); code != "" {
		parts = append(parts, code)
	}

	if code := g.GenerateAutomationCode(); code != "" {
		parts = append(parts, code)
	}

	if code := g.GenerateCrossLayerConsistencyCode(); code != "" {
		parts = append(parts, code)
	}

	return strings.Join(parts, "\n")
}

// Auxiliary functions

func stringifyList(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = fmt.Sprintf("'%s'", escapeQuotes(item))
	}
	return fmt.Sprintf("[%s]", strings.Join(quoted, ", "))
}

func escapeQuotes(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	return s
}

func (g *JSAntiDetectCodeGenerator) generateBrandsList() string {
	// Parse brands list from SecCHUA
	// Users can implement more complex parsing logic as needed
	return "[]"
}
