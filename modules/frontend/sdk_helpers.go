package frontend

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
)

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
