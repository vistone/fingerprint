// Package frontend - JavaScript 反检测对抗点代码生成
package frontend

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vistone/fingerprint/modules/profiles"
)

// JSAntiDetectCodeGenerator JavaScript 反检测代码生成器
type JSAntiDetectCodeGenerator struct {
	profile *profiles.ClientProfile
}

// NewJSAntiDetectCodeGenerator 创建新的反检测代码生成器
func NewJSAntiDetectCodeGenerator(profile *profiles.ClientProfile) *JSAntiDetectCodeGenerator {
	return &JSAntiDetectCodeGenerator{
		profile: profile,
	}
}

// GenerateWebGPUCode 生成 WebGPU 对抗代码
func (g *JSAntiDetectCodeGenerator) GenerateWebGPUCode() string {
	if g.profile.JSAntiDetection == nil || g.profile.JSAntiDetection.WebGPU == nil {
		return ""
	}

	gpu := g.profile.JSAntiDetection.WebGPU

	// 如果不支持 WebGPU，返回隐藏代码
	if !gpu.Available {
		return `
		// 隐藏 WebGPU 支持
		if (typeof navigator !== 'undefined') {
			Object.defineProperty(navigator, 'gpu', {
				value: undefined,
				configurable: true
			});
		}
		`
	}

	// 构建 WebGPU 适配器和设备信息
	featureFlags := []string{}
	if gpu.FeatureFlags != nil {
		featureFlags = gpu.FeatureFlags
	}

	limitsJSON, _ := json.Marshal(gpu.LimitValues)

	code := fmt.Sprintf(`
		// WebGPU 对抗点 - GPU 设备欺骗
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

// GenerateMediaDevicesCode 生成 MediaDevices 对抗代码
func (g *JSAntiDetectCodeGenerator) GenerateMediaDevicesCode() string {
	if g.profile.JSAntiDetection == nil || g.profile.JSAntiDetection.MediaDevices == nil {
		return ""
	}

	md := g.profile.JSAntiDetection.MediaDevices

	// 构建设备列表
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
		// MediaDevices 对抗点 - 设备列表欺骗
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

				// 确保 getUserMedia 返回有效的流
				const origGetUserMedia = navigator.mediaDevices.getUserMedia.bind(navigator.mediaDevices);
				navigator.mediaDevices.getUserMedia = async function(constraints) {
					try {
						return await origGetUserMedia(constraints);
					} catch (e) {
						// 如果设备不可用，返回虚拟流
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

// GeneratePermissionsCode 生成 Permissions API 对抗代码
func (g *JSAntiDetectCodeGenerator) GeneratePermissionsCode() string {
	if g.profile.JSAntiDetection == nil || g.profile.JSAntiDetection.Permissions == nil {
		return ""
	}

	perms := g.profile.JSAntiDetection.Permissions

	// 构建权限状态映射
	permMap := perms.PermissionState
	if permMap == nil {
		permMap = make(map[string]string)
	}

	// 设置默认权限状态
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
		// Permissions API 对抗点 - 权限状态欺骗
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

// GenerateAutomationCode 生成 Automation 对抗代码
func (g *JSAntiDetectCodeGenerator) GenerateAutomationCode() string {
	if g.profile.JSAntiDetection == nil || g.profile.JSAntiDetection.Automation == nil {
		return ""
	}

	auto := g.profile.JSAntiDetection.Automation

	// 构建对抗代码
	code := ""

	// 1. 隐藏 webdriver 标记
	if !auto.WebDriver {
		code += `
		// 隐藏 webdriver 标记
		Object.defineProperty(navigator, 'webdriver', {
			get: () => undefined,
			configurable: true
		});
		`
	}

	// 2. 隐藏 headless 特征
	if auto.Headless {
		code += `
		// 隐藏 headless 浏览器特征
		Object.defineProperty(navigator, 'vendor', {
			get: () => 'Google Inc.' || navigator.vendor,
			configurable: true
		});
		`
	}

	// 3. 隐藏 phantomjs
	if auto.Phantom {
		code += `
		// 隐藏 phantomjs 特征
		Object.defineProperty(window, 'callPhantom', {
			get: () => undefined,
			configurable: true
		});
		`
	}

	// 4. 隐藏 selenium
	if auto.Selenium {
		code += `
		// 隐藏 selenium 驱动特征
		Object.defineProperty(navigator, 'driverName', {
			get: () => undefined,
			configurable: true
		});
		`
	}

	// 5. 隐藏 puppeteer 和 playwright
	if auto.Puppeteer || auto.Playwright {
		code += `
		// 隐藏自动化工具特征
		window.__puppeteer_eval = undefined;
		window.CDP = undefined;
		`
	}

	// 6. 覆盖 plugins
	if auto.PluginsOverride {
		code += `
		// 覆盖 plugins 数组
		Object.defineProperty(navigator, 'plugins', {
			get: () => [
				{ name: 'Native Client Plugin', description: 'Native Client Executable', version: '', filename: 'internal-nacl-plugin' },
				{ name: 'Chrome PDF Plugin', description: 'Portable Document Format', version: '1', filename: 'internal-pdf-viewer' },
				{ name: 'Chrome PDF Viewer', description: 'Portable Document Format', version: '1', filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai' }
			],
			configurable: true
		});

		// 覆盖 mimeTypes
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

	// 7. 覆盖 language
	if auto.LanguageOverride && g.profile.Headers != nil {
		// 从 UA 或 Accept-Language 中推断语言
		lang := "en-US"
		if g.profile.Headers.AcceptLanguage != "" {
			parts := strings.Split(g.profile.Headers.AcceptLanguage, ",")
			if len(parts) > 0 {
				lang = strings.TrimSpace(parts[0])
			}
		}
		code += fmt.Sprintf(`
		// 覆盖 language 属性
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

	// 8. 覆盖 product
	if auto.ProductOverride {
		code += `
		// 覆盖 product 属性
		Object.defineProperty(navigator, 'product', {
			get: () => 'Gecko',
			configurable: true
		});
		`
	}

	// 9. 覆盖 vendor
	if auto.VendorOverride {
		code += `
		// 覆盖 vendor 属性
		Object.defineProperty(navigator, 'vendor', {
			get: () => 'Google Inc.',
			configurable: true
		});
		`
	}

	// 10. 隐藏运行时检测特征
	if auto.RuntimeOverride {
		code += `
		// 隐藏运行时检测特征
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
		// Automation 对抗点 - 自动化工具检测隐藏
		(function() {
			'use strict';
			%s
		})();
		`, code)
}

// GenerateCrossLayerConsistencyCode 生成跨层一致性校验代码
func (g *JSAntiDetectCodeGenerator) GenerateCrossLayerConsistencyCode() string {
	if g.profile.JSAntiDetection == nil {
		return ""
	}

	// 从配置文件中提取各层信息
	ua := ""
	secChUA := ""

	if g.profile.Headers != nil {
		ua = g.profile.Headers.UserAgent
		secChUA = g.profile.Headers.SecCHUA
	}

	code := fmt.Sprintf(`
		// 跨层一致性校验 - 确保 UA/CH/JS/TCP-IP 一致
		(function() {
			const expectedUA = '%s';
			const expectedSecChUA = '%s';

			// 1. UA 一致性检查
			if (navigator.userAgent !== expectedUA) {
				console.warn('[Fingerprint] UA 层不一致');
			}

			// 2. Client Hints 一致性检查
			if (typeof navigator.userAgentData !== 'undefined') {
				const uaData = navigator.userAgentData;
				// 验证浏览器品牌、版本等信息一致
				const brands = uaData.brands || [];
				const expectedBrands = %s;

				const brandsMatch = brands.length === expectedBrands.length &&
					brands.every((b, i) => b.brand === expectedBrands[i].brand && b.version === expectedBrands[i].version);

				if (!brandsMatch) {
					console.warn('[Fingerprint] CH 层品牌不一致');
				}
			}

			// 3. JavaScript 层一致性检查
			const jsUA = navigator.userAgent;
			const jsProduct = navigator.product;
			const jsVendor = navigator.vendor;
			const jsPlatform = navigator.platform;

			if (jsUA !== expectedUA) {
				console.warn('[Fingerprint] JS 层 UA 不一致');
			}

			// 4. 硬件信息一致性
			const hardwareCells = {
				cores: navigator.hardwareConcurrency,
				memory: navigator.deviceMemory,
				maxTouchPoints: navigator.maxTouchPoints,
				platform: navigator.platform
			};

			// 5. 总体一致性报告
			const consistencyReport = {
				ua: { expected: expectedUA, actual: jsUA, match: expectedUA === jsUA },
				layers: {
					http: { ua, ch: secChUA },
					js: { ua: jsUA, product: jsProduct, vendor: jsVendor, platform: jsPlatform },
					hardware: hardwareCells
				}
			};

			// 将一致性报告附加到窗口对象，供后续上报
			window.__fingerprintConsistency = consistencyReport;
		})();
		`, escapeQuotes(ua), escapeQuotes(secChUA), g.generateBrandsList())

	return code
}

// GenerateFullAntiDetectionCode 生成完整的反检测代码
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

// 辅助函数

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
	// 从 SecCHUA 中解析 brands 列表
	// 用户可以根据需要实现更复杂的解析逻辑
	return "[]"
}
