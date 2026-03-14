package frontend

import (
	"fmt"
	"strings"

	"github.com/vistone/fingerprint/modules/profiles"
)

func (g *JSAntiDetectCodeGenerator) GenerateAutomationCode() string {
	if g.profile.JSAntiDetection == nil || g.profile.JSAntiDetection.Automation == nil {
		return ""
	}

	auto := g.profile.JSAntiDetection.Automation

	code := g.generateWebDriverHiding(auto) +
		g.generateBrowserHiding(auto) +
		g.generateOverrides(auto) +
		g.generateRuntimeHiding(auto)

	return fmt.Sprintf(`
		// Automation countermeasure - hide automation tool detection
		(function() {
			'use strict';
			%s
		})();
		`, code)
}

// generateWebDriverHiding generates code to hide automation tool markers.
// It handles: webdriver property, headless indicators, PhantomJS globals,
// Selenium driver attributes, and Puppeteer/Playwright specific properties.
func (g *JSAntiDetectCodeGenerator) generateWebDriverHiding(auto *profiles.AutomationAntiDetect) string {
	code := ""

	if !auto.WebDriver {
		code += `
		// Hide webdriver marker
		Object.defineProperty(navigator, 'webdriver', {
			get: () => undefined,
			configurable: true
		});
		`
	}

	if auto.Headless {
		code += `
		// Hide headless browser features
		Object.defineProperty(navigator, 'vendor', {
			get: () => 'Google Inc.' || navigator.vendor,
			configurable: true
		});
		`
	}

	if auto.Phantom {
		code += `
		// Hide phantomjs features
		Object.defineProperty(window, 'callPhantom', {
			get: () => undefined,
			configurable: true
		});
		`
	}

	if auto.Selenium {
		code += `
		// Hide selenium driver features
		Object.defineProperty(navigator, 'driverName', {
			get: () => undefined,
			configurable: true
		});
		`
	}

	if auto.Puppeteer || auto.Playwright {
		code += `
		// Hide automation tool features
		window.__puppeteer_eval = undefined;
		window.CDP = undefined;
		`
	}

	return code
}

// generateBrowserHiding overrides navigator.plugins and navigator.mimeTypes
// to mimic a real browser environment and prevent automation detection.
func (g *JSAntiDetectCodeGenerator) generateBrowserHiding(auto *profiles.AutomationAntiDetect) string {
	code := ""

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

	return code
}

// generateOverrides overrides browser properties (language, product, vendor)
// to match the target profile and avoid fingerprint inconsistencies.
func (g *JSAntiDetectCodeGenerator) generateOverrides(auto *profiles.AutomationAntiDetect) string {
	code := ""

	if auto.LanguageOverride && g.profile.Headers != nil {
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

	if auto.ProductOverride {
		code += `
		// Override product attribute
		Object.defineProperty(navigator, 'product', {
			get: () => 'Gecko',
			configurable: true
		});
		`
	}

	if auto.VendorOverride {
		code += `
		// Override vendor attribute
		Object.defineProperty(navigator, 'vendor', {
			get: () => 'Google Inc.',
			configurable: true
		});
		`
	}

	return code
}

// generateRuntimeHiding generates code to hide runtime detection features.
func (g *JSAntiDetectCodeGenerator) generateRuntimeHiding(auto *profiles.AutomationAntiDetect) string {
	if !auto.RuntimeOverride {
		return ""
	}

	return `
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
