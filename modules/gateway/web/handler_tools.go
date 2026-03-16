package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
	tlsmod "github.com/vistone/fingerprint/modules/tls"
)

// =====================================================================
// Fingerprint tools API
// =====================================================================

// handleToolsJA3 computes JA3 and JA4 fingerprints.
func (h *Handler) handleToolsJA3(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProfileID    string   `json:"profileId"`
		TLSVersion   uint16   `json:"tlsVersion"`
		CipherSuites []uint16 `json:"cipherSuites"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var spec core.ClientHelloSpec
	if req.ProfileID != "" {
		if p, ok := h.findProfile(req.ProfileID); ok {
			spec = core.ClientHelloSpec{
				TLSVersion:      p.TLSVersion,
				CipherSuites:    p.CipherSuites,
				Extensions:      p.Extensions,
				SupportedCurves: p.SupportedCurves,
			}
		} else {
			http.Error(w, "Profile not found", http.StatusNotFound)
			return
		}
	} else {
		spec = core.ClientHelloSpec{
			TLSVersion:   req.TLSVersion,
			CipherSuites: req.CipherSuites,
		}
	}

	ja3 := tlsmod.CalculateJA3(spec)
	ja4 := tlsmod.CalculateJA4(spec)

	result := map[string]interface{}{
		"ja3": map[string]interface{}{
			"hash": ja3.Hash,
			"raw":  ja3.RawString,
		},
		"ja4": map[string]interface{}{
			"fingerprint": ja4.Fingerprint,
		},
		"input": map[string]interface{}{
			"tlsVersion":   spec.TLSVersion,
			"cipherSuites": len(spec.CipherSuites),
			"extensions":   len(spec.Extensions),
			"curves":       len(spec.SupportedCurves),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleToolsValidate validates profile completeness.
func (h *Handler) handleToolsValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProfileID string `json:"profileId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	profile, ok := h.findProfile(req.ProfileID)
	if !ok {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	validator := profiles.NewProfileValidator()
	result := validator.Validate(profile)

	// TCP/IP validation.
	tcpipResult := ""
	if profile.TCPIP != nil {
		tcpipResult = profiles.ValidateTCPIP(profile.TCPIP)
	}

	// Header validation.
	headerResult := map[string]interface{}{}
	if profile.Headers != nil {
		hvr := profiles.ValidateHeaders(profile.Headers)
		headerResult["missing"] = hvr.Missing
		headerResult["empty"] = hvr.Empty
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"profileId":   req.ProfileID,
		"profileName": profile.Name,
		"validation": map[string]interface{}{
			"valid":         result.Valid,
			"errors":        result.Errors,
			"warnings":      result.Warnings,
			"missingFields": result.MissingFields,
		},
		"tcpipValidation":  tcpipResult,
		"headerValidation": headerResult,
	})
}

// handleToolsCompare compares two profiles.
func (h *Handler) handleToolsCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProfileA string `json:"profileA"`
		ProfileB string `json:"profileB"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	profileA, okA := h.findProfile(req.ProfileA)
	profileB, okB := h.findProfile(req.ProfileB)
	if !okA || !okB {
		http.Error(w, "One or both profiles not found", http.StatusNotFound)
		return
	}

	// Extract feature vectors and compute similarity.
	extractor := h.gateway.GetExtractor()
	fvA := extractor.ExtractFromProfile(&profileA)
	fvB := extractor.ExtractFromProfile(&profileB)

	similarity := calculateSimilarity(fvA, fvB)

	// Build detailed comparison payload.
	comparison := map[string]interface{}{
		"a": map[string]interface{}{
			"id": profileA.ID, "name": profileA.Name,
			"browser": profileA.BrowserType, "version": profileA.BrowserVersion,
			"os":         profileA.OS,
			"tlsVersion": profileA.TLSVersion,
			"ciphers":    len(profileA.CipherSuites),
			"extensions": len(profileA.Extensions),
		},
		"b": map[string]interface{}{
			"id": profileB.ID, "name": profileB.Name,
			"browser": profileB.BrowserType, "version": profileB.BrowserVersion,
			"os":         profileB.OS,
			"tlsVersion": profileB.TLSVersion,
			"ciphers":    len(profileB.CipherSuites),
			"extensions": len(profileB.Extensions),
		},
		"similarity": similarity,
		"diffs":      buildProfileDiffs(profileA, profileB),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comparison)
}

// =====================================================================
// Helper functions
// =====================================================================

func (h *Handler) findProfile(id string) (profiles.ClientProfile, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, p := range h.profiles {
		if p.ID == id {
			return p, true
		}
	}
	return profiles.ClientProfile{}, false
}

func calculateSimilarity(a, b *core.FeatureVector) float64 {
	if a == nil || b == nil {
		return 0
	}
	// Collect all feature keys.
	keys := make(map[core.FeatureType]bool)
	for k := range a.Features {
		keys[k] = true
	}
	for k := range b.Features {
		keys[k] = true
	}
	if len(keys) == 0 {
		return 1.0
	}

	matches := 0
	for k := range keys {
		va := a.Get(k)
		vb := b.Get(k)
		if va == vb {
			matches++
		} else if va != 0 && vb != 0 {
			// Treat relative error < 10% as a match.
			ratio := va / vb
			if ratio < 0 {
				ratio = -ratio
			}
			if ratio > 0.9 && ratio < 1.1 {
				matches++
			}
		}
	}
	return float64(matches) / float64(len(keys))
}

func buildProfileDiffs(a, b profiles.ClientProfile) []map[string]interface{} {
	diffs := []map[string]interface{}{}

	if a.TLSVersion != b.TLSVersion {
		diffs = append(diffs, map[string]interface{}{
			"field": "TLS Version", "a": a.TLSVersion, "b": b.TLSVersion,
		})
	}
	if len(a.CipherSuites) != len(b.CipherSuites) {
		diffs = append(diffs, map[string]interface{}{
			"field": "Cipher Suites Count", "a": len(a.CipherSuites), "b": len(b.CipherSuites),
		})
	}
	if len(a.Extensions) != len(b.Extensions) {
		diffs = append(diffs, map[string]interface{}{
			"field": "Extensions Count", "a": len(a.Extensions), "b": len(b.Extensions),
		})
	}
	if string(a.BrowserType) != string(b.BrowserType) {
		diffs = append(diffs, map[string]interface{}{
			"field": "Browser", "a": a.BrowserType, "b": b.BrowserType,
		})
	}
	if string(a.OS) != string(b.OS) {
		diffs = append(diffs, map[string]interface{}{
			"field": "OS", "a": a.OS, "b": b.OS,
		})
	}
	if a.HTTP2Settings.InitialWindowSize != b.HTTP2Settings.InitialWindowSize {
		diffs = append(diffs, map[string]interface{}{
			"field": "H2 InitialWindowSize", "a": a.HTTP2Settings.InitialWindowSize, "b": b.HTTP2Settings.InitialWindowSize,
		})
	}
	if a.HTTP2Settings.MaxConcurrentStreams != b.HTTP2Settings.MaxConcurrentStreams {
		diffs = append(diffs, map[string]interface{}{
			"field": "H2 MaxConcurrentStreams", "a": a.HTTP2Settings.MaxConcurrentStreams, "b": b.HTTP2Settings.MaxConcurrentStreams,
		})
	}
	if len(a.PseudoHeaderOrder) > 0 && len(b.PseudoHeaderOrder) > 0 {
		aOrder := strings.Join(a.PseudoHeaderOrder, ",")
		bOrder := strings.Join(b.PseudoHeaderOrder, ",")
		if aOrder != bOrder {
			diffs = append(diffs, map[string]interface{}{
				"field": "Pseudo Header Order", "a": aOrder, "b": bOrder,
			})
		}
	}

	// TCP/IP comparison.
	if a.TCPIP != nil && b.TCPIP != nil {
		if a.TCPIP.TTL != b.TCPIP.TTL {
			diffs = append(diffs, map[string]interface{}{
				"field": "TCP TTL", "a": a.TCPIP.TTL, "b": b.TCPIP.TTL,
			})
		}
		if a.TCPIP.WindowSize != b.TCPIP.WindowSize {
			diffs = append(diffs, map[string]interface{}{
				"field": "TCP Window Size", "a": a.TCPIP.WindowSize, "b": b.TCPIP.WindowSize,
			})
		}
	}

	return diffs
}
