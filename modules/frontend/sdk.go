// Package frontend provides frontend fingerprint SDK functionality
// Including JavaScript code generation and server-side processing
package frontend

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
)

// SDK frontend SDK manager
type SDK struct {
	config    *SDKConfig
	sessions  map[string]*Session
	extractor *ml.FeatureExtractor
	mu        sync.RWMutex
}

// SDKConfig SDK configuration
type SDKConfig struct {
	// Noise injection configuration
	EnableCanvasNoise bool
	EnableAudioNoise  bool
	EnableWebGLNoise  bool
	EnableTimingNoise bool
	NoiseLevel        float64 // 0.0 - 1.0

	// Collection configuration
	CollectCanvas   bool
	CollectWebGL    bool
	CollectAudio    bool
	CollectFonts    bool
	CollectStorage  bool
	CollectWebRTC   bool
	CollectHardware bool
	CollectTiming   bool

	// Session configuration
	SessionTimeout time.Duration
}

// DefaultSDKConfig default SDK configuration
var DefaultSDKConfig = &SDKConfig{
	EnableCanvasNoise: true,
	EnableAudioNoise:  true,
	EnableWebGLNoise:  true,
	EnableTimingNoise: true,
	NoiseLevel:        0.1,

	CollectCanvas:   true,
	CollectWebGL:    true,
	CollectAudio:    true,
	CollectFonts:    true,
	CollectStorage:  true,
	CollectWebRTC:   true,
	CollectHardware: true,
	CollectTiming:   true,

	SessionTimeout: core.DefaultSessionTimeout,
}

// NewSDK creates a new SDK manager
func NewSDK(config *SDKConfig) *SDK {
	if config == nil {
		config = DefaultSDKConfig
	}
	return &SDK{
		config:    config,
		sessions:  make(map[string]*Session),
		extractor: ml.NewFeatureExtractor(),
	}
}

// Session frontend session
type Session struct {
	ID           string
	CreatedAt    time.Time
	LastAccessed time.Time
	Data         map[string]interface{}
	Fingerprint  *ml.FrontendFingerprintData
}

// GenerateJSCore generates core JavaScript code
func (sdk *SDK) GenerateJSCore() string {
	return fmt.Sprintf(`
// Fingerprint SDK Core v1.0
(function(global) {
    'use strict';

    const CONFIG = %s;
    const SESSION_ID = '%s';

    // Noise generator
    class NoiseGenerator {
        constructor(level) {
            this.level = level;
            this.seed = Math.random();
        }

        generateCanvasNoise() {
            if (!CONFIG.EnableCanvasNoise) return null;
            return {
                r: (Math.random() - 0.5) * this.level * 2,
                g: (Math.random() - 0.5) * this.level * 2,
                b: (Math.random() - 0.5) * this.level * 2,
                a: 0
            };
        }

        generateAudioNoise() {
            if (!CONFIG.EnableAudioNoise) return null;
            return (Math.random() - 0.5) * this.level * 0.001;
        }

        generateTimingNoise() {
            if (!CONFIG.EnableTimingNoise) return null;
            return Math.floor(Math.random() * this.level * 10);
        }
    }

    // Fingerprint collector
    class FingerprintCollector {
        constructor() {
            this.noiseGen = new NoiseGenerator(CONFIG.NoiseLevel);
            this.data = {};
        }

        async collectAll() {
            return {
                canvas: CONFIG.CollectCanvas ? await this.collectCanvas() : null,
                webgl: CONFIG.CollectWebGL ? await this.collectWebGL() : null,
                audio: CONFIG.CollectAudio ? await this.collectAudio() : null,
                fonts: CONFIG.CollectFonts ? await this.collectFonts() : null,
                storage: CONFIG.CollectStorage ? this.collectStorage() : null,
                webrtc: CONFIG.CollectWebRTC ? await this.collectWebRTC() : null,
                hardware: CONFIG.CollectHardware ? this.collectHardware() : null,
                timing: CONFIG.CollectTiming ? this.collectTiming() : null,
                timestamp: Date.now(),
                sessionId: SESSION_ID
            };
        }

        async collectCanvas() {
            try {
                const canvas = document.createElement('canvas');
                canvas.width = 200;
                canvas.height = 50;
                const ctx = canvas.getContext('2d');

                // Draw
                ctx.textBaseline = 'alphabetic';
                ctx.fillStyle = '#f60';
                ctx.fillRect(0, 0, 200, 50);
                ctx.fillStyle = '#069';
                ctx.font = '11pt "Times New Roman"';
                ctx.fillText('Fingerprint v1.0', 2, 15);

                // Add noise
                const noise = this.noiseGen.generateCanvasNoise();
                if (noise) {
                    const imageData = ctx.getImageData(0, 0, 200, 50);
                    const data = imageData.data;
                    for (let i = 0; i < data.length; i += 4) {
                        data[i] = Math.min(255, Math.max(0, data[i] + noise.r));
                        data[i+1] = Math.min(255, Math.max(0, data[i+1] + noise.g));
                        data[i+2] = Math.min(255, Math.max(0, data[i+2] + noise.b));
                    }
                    ctx.putImageData(imageData, 0, 0);
                }

                const dataURL = canvas.toDataURL();
                return {
                    hash: this.hashString(dataURL),
                    entropy: this.calculateEntropy(dataURL)
                };
            } catch (e) {
                return { error: e.message };
            }
        }

        async collectWebGL() {
            try {
                const canvas = document.createElement('canvas');
                const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
                if (!gl) return { error: 'WebGL not supported' };

                const debugInfo = gl.getExtension('WEBGL_debug_renderer_info');
                const vendor = debugInfo ? gl.getParameter(debugInfo.UNMASKED_VENDOR_WEBGL) : 'unknown';
                const renderer = debugInfo ? gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL) : 'unknown';
                const extensions = gl.getSupportedExtensions() || [];

                return {
                    vendor: vendor,
                    renderer: renderer,
                    extensions: extensions.slice(0, 20),
                    entropy: extensions.length / 100
                };
            } catch (e) {
                return { error: e.message };
            }
        }

        async collectAudio() {
            try {
                const AudioContext = window.AudioContext || window.webkitAudioContext;
                if (!AudioContext) return { error: 'AudioContext not supported' };

                const ctx = new AudioContext();
                const oscillator = ctx.createOscillator();
                const analyser = ctx.createAnalyser();
                const gainNode = ctx.createGain();

                oscillator.type = 'triangle';
                oscillator.frequency.value = 10000;

                gainNode.gain.value = 0;

                oscillator.connect(analyser);
                analyser.connect(gainNode);
                gainNode.connect(ctx.destination);

                oscillator.start();

                const buffer = new Float32Array(analyser.frequencyBinCount);
                analyser.getFloatFrequencyData(buffer);

                oscillator.stop();

                // Add noise
                const noise = this.noiseGen.generateAudioNoise();
                if (noise) {
                    for (let i = 0; i < buffer.length; i++) {
                        buffer[i] += noise;
                    }
                }

                const hash = this.hashBuffer(buffer);

                return {
                    sampleRate: ctx.sampleRate,
                    hash: hash,
                    entropy: this.calculateBufferEntropy(buffer)
                };
            } catch (e) {
                return { error: e.message };
            }
        }

        async collectFonts() {
            const baseFonts = ['monospace', 'sans-serif', 'serif'];
            const testFonts = [
                'Arial', 'Courier New', 'Georgia', 'Times New Roman',
                'Verdana', 'Helvetica', 'Tahoma', 'Trebuchet MS'
            ];

            const available = [];
            const testString = 'mmmmmmmmmwwwwwww';
            const testSize = '72px';

            const canvas = document.createElement('canvas');
            const ctx = canvas.getContext('2d');

            for (const base of baseFonts) {
                ctx.font = testSize + ' ' + base;
                const baseWidth = ctx.measureText(testString).width;

                for (const font of testFonts) {
                    ctx.font = testSize + ' "' + font + '", ' + base;
                    const width = ctx.measureText(testString).width;

                    if (width !== baseWidth) {
                        available.push(font);
                    }
                }
            }

            return {
                list: available,
                count: available.length
            };
        }

        collectStorage() {
            try {
                return {
                    localStorageSize: JSON.stringify(localStorage).length,
                    sessionStorageSize: JSON.stringify(sessionStorage).length,
                    indexedDB: 'indexedDB' in window
                };
            } catch (e) {
                return { error: e.message };
            }
        }

        async collectWebRTC() {
            try {
                const pc = new RTCPeerConnection({
                    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
                });

                const ips = [];
                pc.createDataChannel('');

                const offer = await pc.createOffer();
                await pc.setLocalDescription(offer);

                return new Promise((resolve) => {
                    setTimeout(() => {
                        const lines = offer.sdp.split('\n');
                        for (const line of lines) {
                            if (line.includes('candidate')) {
                                const match = line.match(/([0-9]{1,3}\.){3}[0-9]{1,3}/);
                                if (match && !ips.includes(match[0])) {
                                    ips.push(match[0]);
                                }
                            }
                        }

                        pc.close();
                        resolve({
                            localIPs: ips,
                            ipLeaked: ips.length > 0
                        });
                    }, 500);
                });
            } catch (e) {
                return { error: e.message };
            }
        }

        collectHardware() {
            return {
                cores: navigator.hardwareConcurrency || 'unknown',
                memory: navigator.deviceMemory || 'unknown',
                touchPoints: navigator.maxTouchPoints || 0,
                platform: navigator.platform,
                userAgent: navigator.userAgent
            };
        }

        collectTiming() {
            const noise = this.noiseGen.generateTimingNoise();
            const start = performance.now() + noise;

            // Execute some calculations
            let sum = 0;
            for (let i = 0; i < 1000000; i++) {
                sum += i;
            }

            const end = performance.now() + noise;

            return {
                precision: end - start,
                timestamp: Date.now()
            };
        }

        hashString(str) {
            let hash = 0;
            for (let i = 0; i < str.length; i++) {
                const char = str.charCodeAt(i);
                hash = ((hash << 5) - hash) + char;
                hash = hash & hash;
            }
            return hash.toString(16);
        }

        hashBuffer(buffer) {
            let hash = 0;
            for (let i = 0; i < Math.min(buffer.length, 1000); i++) {
                hash = ((hash << 5) - hash) + buffer[i];
                hash = hash & hash;
            }
            return hash.toString(16);
        }

        calculateEntropy(str) {
            const freq = {};
            for (const char of str) {
                freq[char] = (freq[char] || 0) + 1;
            }

            let entropy = 0;
            const len = str.length;
            for (const count of Object.values(freq)) {
                const p = count / len;
                entropy -= p * Math.log2(p);
            }
            return entropy;
        }

        calculateBufferEntropy(buffer) {
            const freq = {};
            for (const val of buffer) {
                const key = Math.floor(val * 10) / 10;
                freq[key] = (freq[key] || 0) + 1;
            }

            let entropy = 0;
            const len = buffer.length;
            for (const count of Object.values(freq)) {
                const p = count / len;
                entropy -= p * Math.log2(p);
            }
            return entropy;
        }
    }

    // Main API
    global.FingerprintSDK = {
        version: '1.0.0',

        collect: async function() {
            const collector = new FingerprintCollector();
            return await collector.collectAll();
        },

        send: async function(url, data) {
            const response = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Session-ID': SESSION_ID
                },
                body: JSON.stringify(data)
            });
            return response.json();
        },

        init: function() {
            // Auto collect and send
            this.collect().then(data => {
                if (CONFIG.endpoint) {
                    this.send(CONFIG.endpoint, data);
                }
            });
        }
    };

})(window);
`, sdk.toJSON(), sdk.generateSessionID())
}

// GenerateJSInjector generates injection script
func (sdk *SDK) GenerateJSInjector(endpoint string) string {
	core := sdk.GenerateJSCore()
	return fmt.Sprintf(`%s

// Auto-init
window.FingerprintSDK.init();
`, core)
}

// HandleCollect HTTP handler function
func (sdk *SDK) HandleCollect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data ml.FrontendFingerprintData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		sessionID = sdk.generateSessionID()
	}

	// Store session
	sdk.mu.Lock()
	sdk.sessions[sessionID] = &Session{
		ID:           sessionID,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		Fingerprint:  &data,
	}
	sdk.mu.Unlock()

	// Extract features
	features := sdk.extractor.ExtractFromFrontend(data)

	response := map[string]interface{}{
		"session_id": sessionID,
		"features":   features.Features,
		"status":     "collected",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSession gets session
func (sdk *SDK) GetSession(id string) (*Session, bool) {
	sdk.mu.RLock()
	defer sdk.mu.RUnlock()

	session, ok := sdk.sessions[id]
	return session, ok
}

// CleanupSessions cleans up expired sessions
func (sdk *SDK) CleanupSessions() {
	sdk.mu.Lock()
	defer sdk.mu.Unlock()

	now := time.Now()
	for id, session := range sdk.sessions {
		if now.Sub(session.LastAccessed) > sdk.config.SessionTimeout {
			delete(sdk.sessions, id)
		}
	}
}

// toJSON converts configuration to JSON
func (sdk *SDK) toJSON() string {
	data, _ := json.Marshal(sdk.config)
	return string(data)
}

// generateSessionID generates session ID
func (sdk *SDK) generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// NoiseGenerator noise generator implementation
type NoiseGenerator struct {
	Level float64
}

// GenerateCanvasNoise generates Canvas noise
func (ng *NoiseGenerator) GenerateCanvasNoise(seed int64) map[string]float64 {
	return map[string]float64{
		"r": (float64(seed%100)/100.0 - 0.5) * ng.Level * 2,
		"g": (float64((seed/100)%100)/100.0 - 0.5) * ng.Level * 2,
		"b": (float64((seed/10000)%100)/100.0 - 0.5) * ng.Level * 2,
	}
}

// Generate implements interface
func (ng *NoiseGenerator) Generate(seed int64) interface{} {
	return ng.GenerateCanvasNoise(seed)
}

// CombinedFingerprint combined fingerprint (server + frontend)
type CombinedFingerprint struct {
	Server   *ml.ServerFingerprintData
	Frontend *ml.FrontendFingerprintData
	Combined *core.FeatureVector
}

// Combine merges server and frontend fingerprints
func (sdk *SDK) Combine(server *ml.ServerFingerprintData, frontend *ml.FrontendFingerprintData) *CombinedFingerprint {
	combined := sdk.extractor.ExtractCombined(*server, *frontend)

	return &CombinedFingerprint{
		Server:   server,
		Frontend: frontend,
		Combined: combined,
	}
}

// GenerateAntiDetectionCode generates complete JavaScript anti-detection code (P3 high entropy)
// Including WebGPU, MediaDevices, Permissions, Automation countermeasures
func (sdk *SDK) GenerateAntiDetectionCode(profile *profiles.ClientProfile) string {
	generator := NewJSAntiDetectCodeGenerator(profile)
	return generator.GenerateFullAntiDetectionCode()
}

// GenerateConsistencyValidationCode generates cross-layer consistency validation code
func (sdk *SDK) GenerateConsistencyValidationCode(profile *profiles.ClientProfile) string {
	generator := NewJSAntiDetectCodeGenerator(profile)
	return generator.GenerateCrossLayerConsistencyCode()
}
