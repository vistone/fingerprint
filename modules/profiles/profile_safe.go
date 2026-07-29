// Package profiles provides safe fingerprint profile operations
package profiles

import (
	"github.com/vistone/fingerprint/modules/core"
)

// SafeGetUserAgent safely gets User-Agent (with nil check)
func (p *ClientProfile) SafeGetUserAgent() string {
	if p == nil || p.Headers == nil {
		return ""
	}
	return p.Headers.UserAgent
}

// SafeGetHeader safely gets specified Header
func (p *ClientProfile) SafeGetHeader(key string) string {
	if p == nil || p.Headers == nil {
		return ""
	}

	switch key {
	case "Accept":
		return p.Headers.Accept
	case "Accept-Language":
		return p.Headers.AcceptLanguage
	case "Accept-Encoding":
		return p.Headers.AcceptEncoding
	case "User-Agent":
		return p.Headers.UserAgent
	default:
		if p.Headers.Custom != nil {
			return p.Headers.Custom[key]
		}
		return ""
	}
}

// Validate validates fingerprint profile
func (p *ClientProfile) Validate() error {
	if p == nil {
		return core.NewCodedError(core.ErrCodeNilPointer, "ClientProfile.Validate", nil)
	}

	validator := core.NewValidator()

	// basic field validation
	validator.NotEmpty(p.ID, "profile.ID").
		NotEmpty(p.Name, "profile.Name").
		ValidBrowserType(p.BrowserType, "profile.BrowserType").
		NotEmpty(p.BrowserVersion, "profile.BrowserVersion").
		ValidOS(p.OS, "profile.OS").
		NotEmpty(p.OSVersion, "profile.OSVersion")

	// validate TLS version
	if p.TLSVersion != 0x0301 && p.TLSVersion != 0x0302 && p.TLSVersion != 0x0303 && p.TLSVersion != 0x0304 {
		validator.AddErrorf("profile.TLSVersion is invalid: 0x%04x", p.TLSVersion)
	}

	// validate CipherSuites non-empty
	if len(p.CipherSuites) == 0 {
		validator.AddErrorf("profile.CipherSuites cannot be empty")
	}

	return validator.Error()
}

// IsValid checks whether the fingerprint is valid
func (p *ClientProfile) IsValid() bool {
	return p.Validate() == nil
}

// Clone safely clones fingerprint profile
func (p *ClientProfile) Clone() *ClientProfile {
	if p == nil {
		return nil
	}

	clone := &ClientProfile{
		ID:                p.ID,
		Name:              p.Name,
		Description:       p.Description,
		BrowserType:       p.BrowserType,
		BrowserVersion:    p.BrowserVersion,
		OS:                p.OS,
		OSVersion:         p.OSVersion,
		OSArch:            p.OSArch,
		OSBitness:         p.OSBitness,
		TLSVersion:        p.TLSVersion,
		CipherSuites:      make([]uint16, len(p.CipherSuites)),
		Extensions:        make([]core.TLSExtension, len(p.Extensions)),
		SupportedCurves:   make([]core.CurveID, len(p.SupportedCurves)),
		SupportedPoints:   make([]uint8, len(p.SupportedPoints)),
		HTTP2Settings:     p.HTTP2Settings,
		HTTP2Priorities:   make([]core.HTTP2Priority, len(p.HTTP2Priorities)),
		PseudoHeaderOrder: make([]string, len(p.PseudoHeaderOrder)),
		ConnectionFlow:    p.ConnectionFlow,
		Metadata:          make(map[string]interface{}, len(p.Metadata)),
	}

	// copy slices
	copy(clone.CipherSuites, p.CipherSuites)
	copy(clone.Extensions, p.Extensions)
	copy(clone.SupportedCurves, p.SupportedCurves)
	copy(clone.SupportedPoints, p.SupportedPoints)
	copy(clone.HTTP2Priorities, p.HTTP2Priorities)
	copy(clone.PseudoHeaderOrder, p.PseudoHeaderOrder)

	// copy Headers
	if p.Headers != nil {
		clone.Headers = p.Headers.Clone()
	}
	if p.HTTP3Settings != nil {
		http3Settings := *p.HTTP3Settings
		clone.HTTP3Settings = &http3Settings
	}
	if p.QUICVersions != nil {
		clone.QUICVersions = append([]uint32(nil), p.QUICVersions...)
	}
	if p.TCPIP != nil {
		tcpip := *p.TCPIP
		clone.TCPIP = &tcpip
	}
	if p.JSAntiDetection != nil {
		clone.JSAntiDetection = cloneJSAntiDetection(p.JSAntiDetection)
	}

	// copy Metadata
	for k, v := range p.Metadata {
		clone.Metadata[k] = v
	}

	return clone
}

func cloneJSAntiDetection(src *JSAntiDetection) *JSAntiDetection {
	if src == nil {
		return nil
	}

	clone := *src
	if src.WebGPU != nil {
		webGPU := *src.WebGPU
		if src.WebGPU.FeatureFlags != nil {
			webGPU.FeatureFlags = append([]string(nil), src.WebGPU.FeatureFlags...)
		}
		if src.WebGPU.LimitValues != nil {
			webGPU.LimitValues = make(map[string]uint64, len(src.WebGPU.LimitValues))
			for key, value := range src.WebGPU.LimitValues {
				webGPU.LimitValues[key] = value
			}
		}
		clone.WebGPU = &webGPU
	}
	if src.MediaDevices != nil {
		mediaDevices := *src.MediaDevices
		if src.MediaDevices.VideoInputs != nil {
			mediaDevices.VideoInputs = append([]*MediaDeviceInfo(nil), src.MediaDevices.VideoInputs...)
		}
		if src.MediaDevices.AudioInputs != nil {
			mediaDevices.AudioInputs = append([]*MediaDeviceInfo(nil), src.MediaDevices.AudioInputs...)
		}
		if src.MediaDevices.AudioOutputs != nil {
			mediaDevices.AudioOutputs = append([]*MediaDeviceInfo(nil), src.MediaDevices.AudioOutputs...)
		}
		if src.MediaDevices.UserMediaConstraints != nil {
			mediaDevices.UserMediaConstraints = make(map[string]interface{}, len(src.MediaDevices.UserMediaConstraints))
			for key, value := range src.MediaDevices.UserMediaConstraints {
				mediaDevices.UserMediaConstraints[key] = value
			}
		}
		clone.MediaDevices = &mediaDevices
	}
	if src.Permissions != nil {
		permissions := *src.Permissions
		if src.Permissions.PermissionState != nil {
			permissions.PermissionState = make(map[string]string, len(src.Permissions.PermissionState))
			for key, value := range src.Permissions.PermissionState {
				permissions.PermissionState[key] = value
			}
		}
		clone.Permissions = &permissions
	}
	if src.Automation != nil {
		automation := *src.Automation
		clone.Automation = &automation
	}

	return &clone
}

// GetRegistrySafe safely gets the registry (with nil check)
func GetRegistrySafe() *ProfileRegistry {
	if DefaultRegistry == nil {
		DefaultRegistry = NewProfileRegistry()
	}
	return DefaultRegistry
}

// RegisterSafe safely registers a fingerprint
func RegisterSafe(profile ClientProfile) error {
	if err := profile.Validate(); err != nil {
		return err
	}
	GetRegistrySafe().Register(profile)
	return nil
}

// GetSafe safely gets a fingerprint
func GetSafe(id string) (ClientProfile, error) {
	if id == "" {
		return ClientProfile{}, core.NewCodedError(core.ErrCodeInvalidInput, "GetSafe",
			core.ErrProfileNotFound)
	}

	p, ok := GetRegistrySafe().Get(id)
	if !ok {
		return ClientProfile{}, core.NewCodedErrorf(core.ErrCodeProfileNotFound, "GetSafe",
			"profile not found: %s", id)
	}

	return p, nil
}
