package clienthints

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/vistone/fingerprint/modules/profiles/legacy"
	"github.com/vistone/fingerprint/modules/core/types"
)

// ============ client_hints.go Tests ============

func TestNewClientHintsPolicy(t *testing.T) {
	tests := []struct {
		name        types.BrowserType
		wantLowEntropy bool
		wantDelegation bool
		highEntropyLen int
	}{
		{
			name:           types.BrowserChrome,
			wantLowEntropy: true,
			wantDelegation: true,
			highEntropyLen: 6,
		},
		{
			name:           types.BrowserEdge,
			wantLowEntropy: true,
			wantDelegation: true,
			highEntropyLen: 5,
		},
		{
			name:           types.BrowserFirefox,
			wantLowEntropy: true,
			wantDelegation: false,
			highEntropyLen: 0,
		},
		{
			name:           types.BrowserSafari,
			wantLowEntropy: true,
			wantDelegation: false,
			highEntropyLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			policy := NewClientHintsPolicy(tt.name)

			if policy.SendLowEntropyHints != tt.wantLowEntropy {
				t.Errorf("SendLowEntropyHints = %v, want %v", policy.SendLowEntropyHints, tt.wantLowEntropy)
			}

			if policy.SupportsCrossOriginDelegation != tt.wantDelegation {
				t.Errorf("SupportsCrossOriginDelegation = %v, want %v", policy.SupportsCrossOriginDelegation, tt.wantDelegation)
			}

			if len(policy.HighEntropyHints) != tt.highEntropyLen {
				t.Errorf("HighEntropyHints length = %d, want %d", len(policy.HighEntropyHints), tt.highEntropyLen)
			}

			// Verify PermissionsPolicy is initialized
			if policy.PermissionsPolicy == nil {
				t.Error("PermissionsPolicy should not be nil")
			}

			// Chrome-specific checks
			if tt.name == types.BrowserChrome {
				if policy.PermissionsPolicy["ch-ua"] != "self" {
					t.Error("Chrome should have ch-ua='self' in PermissionsPolicy")
				}
			}
		})
	}
}

func TestGenerateClientHintsFromProfile(t *testing.T) {
	tests := []struct {
		name           string
		profile        *profiles.ClientProfile
		policy         *ClientHintsPolicy
		wantSecCHUA    bool
		wantSecCHMobile string
		wantSecCHPlatform bool
		wantHighEntropy bool
	}{
		{
			name: "Chrome Desktop",
			profile: &profiles.ClientProfile{
				BrowserType:    "chrome",
				BrowserVersion: "120.0.6099.109",
				OS:             "Windows NT 10.0; Win64; x64",
				OSVersion:      "10.0.19045",
				OSArch:         "x86",
				OSBitness:      "64",
				IsMobile:       false,
				DeviceModel:    "",
			},
			policy: &ClientHintsPolicy{
				SendLowEntropyHints: true,
				HighEntropyHints: []string{
					"Sec-CH-UA-Arch",
					"Sec-CH-UA-Bitness",
					"Sec-CH-UA-Full-Version-List",
					"Sec-CH-UA-Platform-Version",
				},
			},
			wantSecCHUA:       true,
			wantSecCHMobile:   "?0",
			wantSecCHPlatform: true,
			wantHighEntropy:   true,
		},
		{
			name: "Edge Desktop",
			profile: &profiles.ClientProfile{
				BrowserType:    "edge",
				BrowserVersion: "120.0.0.0",
				OS:             "macOS",
				OSVersion:      "14.0",
				OSArch:         "arm",
				OSBitness:      "64",
				IsMobile:       false,
				DeviceModel:    "",
			},
			policy: &ClientHintsPolicy{
				SendLowEntropyHints: true,
				HighEntropyHints: []string{
					"Sec-CH-UA-Arch",
					"Sec-CH-UA-Full-Version-List",
				},
			},
			wantSecCHUA:       true,
			wantSecCHMobile:   "?0",
			wantSecCHPlatform: true,
			wantHighEntropy:   true,
		},
		{
			name: "Chrome Mobile",
			profile: &profiles.ClientProfile{
				BrowserType:    "chrome",
				BrowserVersion: "120.0.0.0",
				OS:             "Android",
				OSVersion:      "14",
				OSArch:         "arm",
				OSBitness:      "64",
				IsMobile:       true,
				DeviceModel:    "Pixel 7",
			},
			policy: &ClientHintsPolicy{
				SendLowEntropyHints: true,
				HighEntropyHints: []string{
					"Sec-CH-UA-Model",
				},
			},
			wantSecCHUA:       true,
			wantSecCHMobile:   "?1",
			wantSecCHPlatform: true,
			wantHighEntropy:   true,
		},
		{
			name: "Disabled Low Entropy",
			profile: &profiles.ClientProfile{
				BrowserType:    "chrome",
				BrowserVersion: "120.0.0.0",
				OS:             "Windows",
			},
			policy: &ClientHintsPolicy{
				SendLowEntropyHints: false,
				HighEntropyHints:    []string{"Sec-CH-UA-Arch"},
			},
			wantSecCHUA:       false,
			wantSecCHMobile:   "",
			wantSecCHPlatform: false,
			wantHighEntropy:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hints := GenerateClientHintsFromProfile(tt.profile, tt.policy)

			// Check Sec-CH-UA
			if tt.wantSecCHUA && hints.SecCHUA == "" {
				t.Error("expected SecCHUA to be set")
			}
			if !tt.wantSecCHUA && hints.SecCHUA != "" {
				t.Error("expected SecCHUA to be empty")
			}

			// Check Sec-CH-UA-Mobile
			if tt.wantSecCHMobile != "" && hints.SecCHUAMobile != tt.wantSecCHMobile {
				t.Errorf("SecCHUAMobile = %s, want %s", hints.SecCHUAMobile, tt.wantSecCHMobile)
			}

			// Check Sec-CH-UA-Platform
			if tt.wantSecCHPlatform && hints.SecCHUAPlatform == "" {
				t.Error("expected SecCHUAPlatform to be set")
			}

			// Check high entropy hints
			if tt.wantHighEntropy {
				if contains(tt.policy.HighEntropyHints, "Sec-CH-UA-Arch") && hints.SecCHUAArch == "" {
					t.Error("expected SecCHUAArch to be set")
				}
				if contains(tt.policy.HighEntropyHints, "Sec-CH-UA-Full-Version-List") && hints.SecCHUAFullVersionList == "" {
					t.Error("expected SecCHUAFullVersionList to be set")
				}
			}

			// Check mobile model
			if tt.profile.IsMobile && contains(tt.policy.HighEntropyHints, "Sec-CH-UA-Model") {
				if hints.SecCHUAModel == "" {
					t.Error("expected SecCHUAModel to be set for mobile device")
				}
			}
		})
	}
}

func TestClientHintsPolicy_ProcessAcceptCH(t *testing.T) {
	tests := []struct {
		name          string
		initialHints  []string
		acceptCHValue string
		wantHints     []string
		wantLen       int
	}{
		{
			name:          "Empty Accept-CH",
			initialHints:  []string{"Sec-CH-UA-Arch"},
			acceptCHValue: "",
			wantHints:     []string{"Sec-CH-UA-Arch"},
			wantLen:       1,
		},
		{
			name:          "Single hint",
			initialHints:  []string{},
			acceptCHValue: "Sec-CH-UA-Bitness",
			wantHints:     []string{"Sec-CH-UA-Bitness"},
			wantLen:       1,
		},
		{
			name:          "Multiple hints",
			initialHints:  []string{"Sec-CH-UA-Arch"},
			acceptCHValue: "Sec-CH-UA-Bitness, Sec-CH-UA-Platform-Version",
			wantLen:       3,
		},
		{
			name:          "Duplicate hint not added",
			initialHints:  []string{"Sec-CH-UA-Arch"},
			acceptCHValue: "Sec-CH-UA-Arch, Sec-CH-UA-Bitness",
			wantLen:       2,
		},
		{
			name:          "Unsupported hint ignored",
			initialHints:  []string{},
			acceptCHValue: "Sec-CH-UA-Arch, X-Custom-Header, Sec-CH-UA-Bitness",
			wantLen:       2,
		},
		{
			name:          "With whitespace",
			initialHints:  []string{},
			acceptCHValue: "  Sec-CH-UA-Arch  ,  Sec-CH-UA-Bitness  ",
			wantLen:       2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &ClientHintsPolicy{
				HighEntropyHints: tt.initialHints,
			}

			policy.ProcessAcceptCH(tt.acceptCHValue)

			if len(policy.HighEntropyHints) != tt.wantLen {
				t.Errorf("HighEntropyHints length = %d, want %d", len(policy.HighEntropyHints), tt.wantLen)
			}
		})
	}
}

func TestClientHintsData_ApplyToHeaders(t *testing.T) {
	tests := []struct {
		name    string
		hints   *ClientHintsData
		wantLen int
		checks  map[string]string
	}{
		{
			name: "All low entropy hints",
			hints: &ClientHintsData{
				SecCHUA:         `"Google Chrome";v="120"`,
				SecCHUAMobile:   "?0",
				SecCHUAPlatform: `"Windows"`,
			},
			wantLen: 3,
			checks: map[string]string{
				"Sec-CH-UA":         `"Google Chrome";v="120"`,
				"Sec-CH-UA-Mobile":  "?0",
				"Sec-CH-UA-Platform": `"Windows"`,
			},
		},
		{
			name: "With high entropy hints",
			hints: &ClientHintsData{
				SecCHUA:               `"Google Chrome";v="120"`,
				SecCHUAMobile:         "?0",
				SecCHUAPlatform:       `"Windows"`,
				SecCHUAArch:           `"x86"`,
				SecCHUABitness:        `"64"`,
				SecCHUAFullVersionList: `"Not A(Brand";v="8.0.0.0", "Chromium";v="120.0.6099.109"`,
				SecCHUAPlatformVersion: `"10.0.19045"`,
				SecCHUAModel:          `""`,
				SecCHUAWoW64:          "?0",
			},
			wantLen: 9,
			checks: map[string]string{
				"Sec-CH-UA-Arch":              `"x86"`,
				"Sec-CH-UA-Bitness":           `"64"`,
				"Sec-CH-UA-Full-Version-List": `"Not A(Brand";v="8.0.0.0", "Chromium";v="120.0.6099.109"`,
			},
		},
		{
			name: "With network hints",
			hints: &ClientHintsData{
				SecCHUA:         `"Google Chrome";v="120"`,
				DeviceMemory:    "8",
				DPR:             "1.5",
				ViewportWidth:   "1920",
				DownlinkSpeed:   "10",
				ECT:             "4g",
				RTT:             "50",
				SaveData:        "off",
			},
			wantLen: 8,
			checks: map[string]string{
				"Device-Memory":  "8",
				"DPR":            "1.5",
				"Viewport-Width": "1920",
				"Downlink":       "10",
				"ECT":            "4g",
				"RTT":            "50",
				"Save-Data":      "off",
			},
		},
		{
			name:    "Empty hints",
			hints:   &ClientHintsData{},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := make(map[string]string)
			tt.hints.ApplyToHeaders(headers)

			if len(headers) != tt.wantLen {
				t.Errorf("headers length = %d, want %d", len(headers), tt.wantLen)
			}

			for key, wantValue := range tt.checks {
				if got, ok := headers[key]; !ok {
					t.Errorf("header %s not found", key)
				} else if got != wantValue {
					t.Errorf("header %s = %s, want %s", key, got, wantValue)
				}
			}
		})
	}
}

// ============ Helper Function Tests ============

func TestGenerateSecCHUA(t *testing.T) {
	tests := []struct {
		name    string
		profile *profiles.ClientProfile
		want    string
	}{
		{
			name: "Chrome",
			profile: &profiles.ClientProfile{
				BrowserType:    "chrome",
				BrowserVersion: "120.0.6099.109",
			},
			want: `"Not A(Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`,
		},
		{
			name: "Edge",
			profile: &profiles.ClientProfile{
				BrowserType:    "edge",
				BrowserVersion: "120.0.0.0",
			},
			want: `"Not A(Brand";v="8", "Chromium";v="120", "Microsoft Edge";v="120"`,
		},
		{
			name: "Unknown browser",
			profile: &profiles.ClientProfile{
				BrowserType:    "firefox",
				BrowserVersion: "120.0",
			},
			want: "",
		},
		{
			name: "Case insensitive chrome",
			profile: &profiles.ClientProfile{
				BrowserType:    "Chrome",
				BrowserVersion: "120.0.0.0",
			},
			want: `"Not A(Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSecCHUA(tt.profile)
			if got != tt.want {
				t.Errorf("generateSecCHUA() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateSecCHUAMobile(t *testing.T) {
	tests := []struct {
		name    string
		profile *profiles.ClientProfile
		want    string
	}{
		{
			name: "Desktop",
			profile: &profiles.ClientProfile{
				IsMobile: false,
			},
			want: "?0",
		},
		{
			name: "Mobile",
			profile: &profiles.ClientProfile{
				IsMobile: true,
			},
			want: "?1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSecCHUAMobile(tt.profile)
			if got != tt.want {
				t.Errorf("generateSecCHUAMobile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateSecCHUAPlatform(t *testing.T) {
	tests := []struct {
		name    string
		profile *profiles.ClientProfile
		want    string
	}{
		{
			name: "Windows",
			profile: &profiles.ClientProfile{
				OS: "Windows NT 10.0",
			},
			want: `"Windows"`,
		},
		{
			name: "macOS",
			profile: &profiles.ClientProfile{
				OS: "Macintosh; Intel Mac OS X 14_0_0",
			},
			want: `"macOS"`,
		},
		{
			name: "Mac case insensitive",
			profile: &profiles.ClientProfile{
				OS: "Mac OS X",
			},
			want: `"macOS"`,
		},
		{
			name: "Linux",
			profile: &profiles.ClientProfile{
				OS: "X11; Linux x86_64",
			},
			want: `"Linux"`,
		},
		{
			name: "Android",
			profile: &profiles.ClientProfile{
				OS: "Android 14",
			},
			want: `"Android"`,
		},
		{
			name: "iOS",
			profile: &profiles.ClientProfile{
				OS: "iPhone; CPU iPhone OS 17_0",
			},
			// Note: iPhone OS string doesn't contain "iOS" substring
			// so it falls through to Unknown based on current implementation
			want: `"Unknown"`,
		},
		{
			name: "Unknown OS",
			profile: &profiles.ClientProfile{
				OS: "Unknown OS",
			},
			want: `"Unknown"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSecCHUAPlatform(tt.profile)
			if got != tt.want {
				t.Errorf("generateSecCHUAPlatform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateFullVersionList(t *testing.T) {
	tests := []struct {
		name    string
		profile *profiles.ClientProfile
		want    string
	}{
		{
			name: "Chrome",
			profile: &profiles.ClientProfile{
				BrowserType:    "chrome",
				BrowserVersion: "120.0.6099.109",
			},
			want: `"Not A(Brand";v="8.0.0.0", "Chromium";v="120.0.6099.109", "Google Chrome";v="120.0.6099.109"`,
		},
		{
			name: "Edge",
			profile: &profiles.ClientProfile{
				BrowserType:    "edge",
				BrowserVersion: "120.0.0.0",
			},
			want: `"Not A(Brand";v="8.0.0.0", "Chromium";v="120.0.0.0", "Microsoft Edge";v="120.0.0.0"`,
		},
		{
			name: "Unknown browser",
			profile: &profiles.ClientProfile{
				BrowserType:    "firefox",
				BrowserVersion: "120.0",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateFullVersionList(tt.profile)
			if got != tt.want {
				t.Errorf("generateFullVersionList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSupportedHighEntropyHint(t *testing.T) {
	tests := []struct {
		hint string
		want bool
	}{
		{"Sec-CH-UA-Arch", true},
		{"Sec-CH-UA-Bitness", true},
		{"Sec-CH-UA-Full-Version-List", true},
		{"Sec-CH-UA-Platform-Version", true},
		{"Sec-CH-UA-Model", true},
		{"Sec-CH-UA-WoW64", true},
		{"Device-Memory", true},
		{"DPR", true},
		{"Viewport-Width", true},
		{"Downlink", true},
		{"ECT", true},
		{"RTT", true},
		{"Save-Data", true},
		{"X-Custom-Header", false},
		{"Sec-CH-UA-Unknown", false},
		{"", false},
		// Case insensitive
		{"sec-ch-ua-arch", true},
		{"SEC-CH-UA-ARCH", true},
		// With whitespace
		{"  Sec-CH-UA-Arch  ", true},
	}

	for _, tt := range tests {
		t.Run(tt.hint, func(t *testing.T) {
			got := isSupportedHighEntropyHint(tt.hint)
			if got != tt.want {
				t.Errorf("isSupportedHighEntropyHint(%q) = %v, want %v", tt.hint, got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		slice []string
		item  string
		want  bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
		{[]string{"a"}, "a", true},
		// Case insensitive
		{[]string{"Apple", "Banana"}, "apple", true},
		{[]string{"Apple", "Banana"}, "APPLE", true},
		// Empty slice
		{nil, "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.item, func(t *testing.T) {
			got := contains(tt.slice, tt.item)
			if got != tt.want {
				t.Errorf("contains(%v, %q) = %v, want %v", tt.slice, tt.item, got, tt.want)
			}
		})
	}
}

// ============ lifecycle.go Tests ============

func TestNewCHLifecycleManager(t *testing.T) {
	manager := NewCHLifecycleManager()

	if manager == nil {
		t.Fatal("NewCHLifecycleManager() returned nil")
	}

	if manager.lifecycles == nil {
		t.Error("lifecycles map should be initialized")
	}

	if manager.negotiationAnalyzer == nil {
		t.Error("negotiationAnalyzer should be initialized")
	}

	if manager.policyAnalyzer == nil {
		t.Error("policyAnalyzer should be initialized")
	}
}

func TestCHLifecycleManager_StartLifecycle(t *testing.T) {
	manager := NewCHLifecycleManager()

	tests := []struct {
		name           string
		originURL      string
		initialHints   []string
		wantPhase      CHPhase
		wantEventCount int
	}{
		{
			name:           "New lifecycle with hints",
			originURL:      "https://example.com",
			initialHints:   []string{"Sec-CH-UA", "Sec-CH-UA-Mobile"},
			wantPhase:      PHASE_INITIAL_REQUEST,
			wantEventCount: 1,
		},
		{
			name:           "New lifecycle without hints",
			originURL:      "https://test.com",
			initialHints:   []string{},
			wantPhase:      PHASE_INITIAL_REQUEST,
			wantEventCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lifecycle := manager.StartLifecycle(tt.originURL, tt.initialHints)

			if lifecycle == nil {
				t.Fatal("StartLifecycle() returned nil")
			}

			if lifecycle.PrimaryOriginURL != tt.originURL {
				t.Errorf("PrimaryOriginURL = %s, want %s", lifecycle.PrimaryOriginURL, tt.originURL)
			}

			if lifecycle.CurrentPhase != tt.wantPhase {
				t.Errorf("CurrentPhase = %v, want %v", lifecycle.CurrentPhase, tt.wantPhase)
			}

			if len(lifecycle.EventLog) != tt.wantEventCount {
				t.Errorf("EventLog length = %d, want %d", len(lifecycle.EventLog), tt.wantEventCount)
			}

			if len(lifecycle.ActiveHints) != len(tt.initialHints) {
				t.Errorf("ActiveHints length = %d, want %d", len(lifecycle.ActiveHints), len(tt.initialHints))
			}

			// Check discovered origins
			if len(lifecycle.DiscoveredOrigins) != 1 || lifecycle.DiscoveredOrigins[0] != tt.originURL {
				t.Errorf("DiscoveredOrigins = %v, want [%s]", lifecycle.DiscoveredOrigins, tt.originURL)
			}

			// Check event details
			if len(lifecycle.EventLog) > 0 {
				event := lifecycle.EventLog[0]
				if event.Type != PHASE_INITIAL_REQUEST {
					t.Errorf("Event type = %v, want PHASE_INITIAL_REQUEST", event.Type)
				}
				hintCount, ok := event.Details["hint_count"].(int)
				if !ok || hintCount != len(tt.initialHints) {
					t.Errorf("Event hint_count = %v, want %d", event.Details["hint_count"], len(tt.initialHints))
				}
			}
		})
	}
}

func TestCHLifecycleManager_ProcessServerResponse(t *testing.T) {
	manager := NewCHLifecycleManager()

	// Start a lifecycle first
	originURL := "https://example.com"
	manager.StartLifecycle(originURL, []string{"Sec-CH-UA"})

	tests := []struct {
		name              string
		acceptCHValue     string
		permissionsPolicy string
		wantPhase         CHPhase
		wantErr           bool
	}{
		{
			name:          "Valid response with Accept-CH",
			acceptCHValue: "Sec-CH-UA-Arch, Sec-CH-UA-Bitness",
			wantPhase:     PHASE_SERVER_RESPONSE,
			wantErr:       false,
		},
		{
			name:              "Valid response with both headers",
			acceptCHValue:     "Sec-CH-UA-Platform-Version",
			permissionsPolicy: "ch-ua=(self), ch-ua-arch=()",
			wantPhase:         PHASE_SERVER_RESPONSE,
			wantErr:           false,
		},
		{
			name:          "Empty Accept-CH",
			acceptCHValue: "",
			wantPhase:     PHASE_SERVER_RESPONSE,
			wantErr:       false,
		},
		{
			name:          "Non-existent lifecycle",
			acceptCHValue: "Sec-CH-UA-Arch",
			wantPhase:     PHASE_SERVER_RESPONSE,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.wantErr {
				// Test with non-existent lifecycle
				err = manager.ProcessServerResponse("https://nonexistent.com", tt.acceptCHValue, tt.permissionsPolicy)
				if err == nil {
					t.Error("expected error for non-existent lifecycle")
				}
				return
			}

			err = manager.ProcessServerResponse(originURL, tt.acceptCHValue, tt.permissionsPolicy)
			if err != nil {
				t.Errorf("ProcessServerResponse() error = %v", err)
				return
			}

			lifecycle, _ := manager.GetLifecycleReport(originURL)
			if lifecycle.CurrentPhase != tt.wantPhase {
				t.Errorf("CurrentPhase = %v, want %v", lifecycle.CurrentPhase, tt.wantPhase)
			}

			// Check event log
			if len(lifecycle.EventLog) < 2 {
				t.Error("expected at least 2 events")
				return
			}

			serverEvent := lifecycle.EventLog[len(lifecycle.EventLog)-1]
			if serverEvent.Type != PHASE_SERVER_RESPONSE {
				t.Errorf("last event type = %v, want PHASE_SERVER_RESPONSE", serverEvent.Type)
			}

			// Check NegotiationStrategy
			if lifecycle.NegotiationStrategy == nil {
				t.Error("NegotiationStrategy should be set")
			}

			// Check PermissionsPolicy
			if tt.permissionsPolicy != "" && lifecycle.PermissionsPolicy == nil {
				t.Error("PermissionsPolicy should be set")
			}
		})
	}
}

func TestCHLifecycleManager_ProcessSubsequentRequest(t *testing.T) {
	manager := NewCHLifecycleManager()

	// Start lifecycle and process server response
	originURL := "https://example.com"
	manager.StartLifecycle(originURL, []string{"Sec-CH-UA"})
	manager.ProcessServerResponse(originURL, "Sec-CH-UA-Arch", "")

	tests := []struct {
		name          string
		requestOrigin string
		includedHints []string
		wantPhase     CHPhase
		wantErr       bool
	}{
		{
			name:          "Same origin request",
			requestOrigin: "https://example.com",
			includedHints: []string{"Sec-CH-UA", "Sec-CH-UA-Arch"},
			wantPhase:     PHASE_SUBSEQUENT_REQUESTS,
			wantErr:       false,
		},
		{
			name:          "Cross origin request",
			requestOrigin: "https://cdn.example.com",
			includedHints: []string{"Sec-CH-UA"},
			wantPhase:     PHASE_CROSS_ORIGIN_SUB_REQUESTS,
			wantErr:       false,
		},
		{
			name:          "Non-existent lifecycle",
			requestOrigin: "https://example.com",
			includedHints: []string{"Sec-CH-UA"},
			wantPhase:     PHASE_SUBSEQUENT_REQUESTS,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.wantErr {
				err = manager.ProcessSubsequentRequest("https://nonexistent.com", tt.requestOrigin, tt.includedHints)
				if err == nil {
					t.Error("expected error for non-existent lifecycle")
				}
				return
			}

			err = manager.ProcessSubsequentRequest(originURL, tt.requestOrigin, tt.includedHints)
			if err != nil {
				t.Errorf("ProcessSubsequentRequest() error = %v", err)
				return
			}

			lifecycle, _ := manager.GetLifecycleReport(originURL)
			if lifecycle.CurrentPhase != tt.wantPhase {
				t.Errorf("CurrentPhase = %v, want %v", lifecycle.CurrentPhase, tt.wantPhase)
			}

			// Check if origin was discovered
			found := false
			for _, origin := range lifecycle.DiscoveredOrigins {
				if origin == tt.requestOrigin {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("request origin %s should be in DiscoveredOrigins", tt.requestOrigin)
			}
		})
	}
}

func TestCHLifecycleManager_TerminateLifecycle(t *testing.T) {
	manager := NewCHLifecycleManager()

	// Start and process lifecycle
	originURL := "https://example.com"
	manager.StartLifecycle(originURL, []string{"Sec-CH-UA"})
	manager.ProcessServerResponse(originURL, "Sec-CH-UA-Arch", "")

	tests := []struct {
		name    string
		origin  string
		wantErr bool
	}{
		{
			name:    "Valid termination",
			origin:  originURL,
			wantErr: false,
		},
		{
			name:    "Non-existent lifecycle",
			origin:  "https://nonexistent.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terminated, err := manager.TerminateLifecycle(tt.origin)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error for non-existent lifecycle")
				}
				return
			}

			if err != nil {
				t.Errorf("TerminateLifecycle() error = %v", err)
				return
			}

			if terminated.CurrentPhase != PHASE_TERMINATED {
				t.Errorf("CurrentPhase = %v, want PHASE_TERMINATED", terminated.CurrentPhase)
			}

			// Check termination event
			if len(terminated.EventLog) == 0 {
				t.Error("expected at least one event")
				return
			}

			lastEvent := terminated.EventLog[len(terminated.EventLog)-1]
			if lastEvent.Type != PHASE_TERMINATED {
				t.Errorf("last event type = %v, want PHASE_TERMINATED", lastEvent.Type)
			}

			// Check duration is recorded
			if _, ok := lastEvent.Details["duration_seconds"]; !ok {
				t.Error("duration_seconds should be in event details")
			}
		})
	}
}

func TestCHLifecycleManager_GetLifecycleReport(t *testing.T) {
	manager := NewCHLifecycleManager()
	originURL := "https://example.com"
	manager.StartLifecycle(originURL, []string{"Sec-CH-UA"})

	tests := []struct {
		name    string
		origin  string
		wantErr bool
	}{
		{
			name:    "Existing lifecycle",
			origin:  originURL,
			wantErr: false,
		},
		{
			name:    "Non-existent lifecycle",
			origin:  "https://nonexistent.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lifecycle, err := manager.GetLifecycleReport(tt.origin)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error for non-existent lifecycle")
				}
				return
			}

			if err != nil {
				t.Errorf("GetLifecycleReport() error = %v", err)
				return
			}

			if lifecycle == nil {
				t.Error("GetLifecycleReport() returned nil lifecycle")
				return
			}

			if lifecycle.PrimaryOriginURL != tt.origin {
				t.Errorf("PrimaryOriginURL = %s, want %s", lifecycle.PrimaryOriginURL, tt.origin)
			}
		})
	}
}

func TestCHLifecycleManager_GetLifecycleMetrics(t *testing.T) {
	manager := NewCHLifecycleManager()
	originURL := "https://example.com"
	lifecycle := manager.StartLifecycle(originURL, []string{"Sec-CH-UA"})
	manager.ProcessServerResponse(originURL, "Sec-CH-UA-Arch", "ch-ua=(self)")

	metrics := manager.GetLifecycleMetrics(lifecycle)

	// Check required fields
	requiredFields := []string{
		"primary_origin",
		"current_phase",
		"total_events",
		"discovered_origins",
		"active_hints",
		"risk_score",
		"duration_seconds",
		"integrity_flags",
		"negotiated_hints",
		"rejected_hints",
		"policy_features",
		"policy_anomalies",
	}

	for _, field := range requiredFields {
		if _, ok := metrics[field]; !ok {
			t.Errorf("metrics missing required field: %s", field)
		}
	}

	// Check values
	if metrics["primary_origin"] != originURL {
		t.Errorf("primary_origin = %v, want %s", metrics["primary_origin"], originURL)
	}

	if metrics["current_phase"] != PHASE_SERVER_RESPONSE {
		t.Errorf("current_phase = %v, want PHASE_SERVER_RESPONSE", metrics["current_phase"])
	}

	if metrics["total_events"] != 2 {
		t.Errorf("total_events = %v, want 2", metrics["total_events"])
	}
}

func TestCHLifecycleManager_GetSummary(t *testing.T) {
	manager := NewCHLifecycleManager()
	originURL := "https://example.com"
	lifecycle := manager.StartLifecycle(originURL, []string{"Sec-CH-UA"})

	summary := manager.GetSummary(lifecycle)

	// Check that summary contains expected information
	if !strings.Contains(summary, originURL) {
		t.Error("summary should contain origin URL")
	}

	if !strings.Contains(summary, "初始请求") {
		t.Error("summary should contain phase name in Chinese")
	}

	if !strings.Contains(summary, "发现源") {
		t.Error("summary should contain discovered origins info")
	}

	if !strings.Contains(summary, "事件") {
		t.Error("summary should contain events info")
	}

	if !strings.Contains(summary, "风险分数") {
		t.Error("summary should contain risk score info")
	}
}

func TestCHLifecycleManager_calculateFinalRiskScore(t *testing.T) {
	manager := NewCHLifecycleManager()

	tests := []struct {
		name           string
		setupLifecycle func(*ClientHintsLifecycle)
		wantMinRisk    float64
		wantMaxRisk    float64
		wantFlags      []string
	}{
		{
			name: "Normal lifecycle",
			setupLifecycle: func(l *ClientHintsLifecycle) {
				// Default lifecycle
			},
			wantMinRisk: 0.0,
			wantMaxRisk: 0.3,
		},
		{
			name: "Excessive anomalies",
			setupLifecycle: func(l *ClientHintsLifecycle) {
				l.EventLog = []CHLifecycleEvent{
					{RiskIndicators: []string{"a", "b", "c", "d", "e", "f"}},
				}
			},
			wantMinRisk: 0.2,
			wantMaxRisk: 0.5,
			wantFlags:   []string{"EXCESSIVE_ANOMALIES_DETECTED"},
		},
		{
			name: "Excessive cross-origin discovery",
			setupLifecycle: func(l *ClientHintsLifecycle) {
				l.DiscoveredOrigins = make([]string, 15)
				for i := range l.DiscoveredOrigins {
					l.DiscoveredOrigins[i] = fmt.Sprintf("https://example%d.com", i)
				}
			},
			wantMinRisk: 0.15,
			wantMaxRisk: 0.45,
			wantFlags:   []string{"EXCESSIVE_CROSS_ORIGIN_DISCOVERY"},
		},
		{
			name: "Unusually short lifecycle",
			setupLifecycle: func(l *ClientHintsLifecycle) {
				l.StartTime = time.Now().Add(500 * time.Millisecond)
			},
			wantMinRisk: 0.1,
			wantMaxRisk: 0.4,
			wantFlags:   []string{"UNUSUALLY_SHORT_LIFECYCLE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originURL := fmt.Sprintf("https://test-%s.com", strings.ReplaceAll(tt.name, " ", "-"))
			lifecycle := manager.StartLifecycle(originURL, []string{"Sec-CH-UA"})
			tt.setupLifecycle(lifecycle)
			_ = lifecycle // Use the variable

			// Update start time for short lifecycle test
			if tt.name == "Unusually short lifecycle" {
				lifecycle.StartTime = time.Now()
			}

			// Sleep a bit to ensure time passes for the short lifecycle test
			if strings.Contains(tt.name, "short") {
				time.Sleep(10 * time.Millisecond)
			}

			manager.calculateFinalRiskScore(lifecycle)

			if lifecycle.RiskScore < tt.wantMinRisk || lifecycle.RiskScore > tt.wantMaxRisk {
				t.Errorf("RiskScore = %f, want between %f and %f", lifecycle.RiskScore, tt.wantMinRisk, tt.wantMaxRisk)
			}

			for _, flag := range tt.wantFlags {
				found := false
				for _, f := range lifecycle.IntegrityFlags {
					if f == flag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected flag %s not found in IntegrityFlags: %v", flag, lifecycle.IntegrityFlags)
				}
			}
		})
	}
}

// ============ negotiation.go Tests ============

func TestNewCHNegotiationAnalyzer(t *testing.T) {
	analyzer := NewCHNegotiationAnalyzer()

	if analyzer == nil {
		t.Fatal("NewCHNegotiationAnalyzer() returned nil")
	}

	if analyzer.strategies == nil {
		t.Error("strategies map should be initialized")
	}
}

func TestCHNegotiationAnalyzer_InitializeFromAcceptCH(t *testing.T) {
	analyzer := NewCHNegotiationAnalyzer()

	tests := []struct {
		name          string
		acceptCHValue string
		origin        string
		wantState     NegotiationState
		wantLowLen    int
		wantHighLen   int
		wantAnomalies bool
	}{
		{
			name:          "Empty Accept-CH",
			acceptCHValue: "",
			origin:        "https://example.com",
			wantState:     NEGOTIATION_REQUESTED,
			wantLowLen:    0,
			wantHighLen:   0,
			wantAnomalies: false,
		},
		{
			name:          "Low entropy hints only",
			acceptCHValue: "Sec-CH-UA, Sec-CH-UA-Mobile, Sec-CH-UA-Platform",
			origin:        "https://example.com",
			wantState:     NEGOTIATION_REQUESTED,
			wantLowLen:    3,
			wantHighLen:   0,
			wantAnomalies: false,
		},
		{
			name:          "Mixed hints",
			acceptCHValue: "Sec-CH-UA, Sec-CH-UA-Arch, Sec-CH-UA-Bitness",
			origin:        "https://example.com",
			wantState:     NEGOTIATION_REQUESTED,
			wantLowLen:    1,
			wantHighLen:   2,
			wantAnomalies: false,
		},
		{
			name:          "Excessive high entropy hints",
			acceptCHValue: "Sec-CH-UA-Arch, Sec-CH-UA-Bitness, Sec-CH-UA-Full-Version, Sec-CH-UA-Platform-Version, Sec-CH-UA-Model, DPR",
			origin:        "https://example.com",
			wantState:     NEGOTIATION_REQUESTED,
			wantLowLen:    0,
			wantHighLen:   6,
			wantAnomalies: true,
		},
		{
			name:          "With whitespace and quotes",
			acceptCHValue: ` "Sec-CH-UA" , "Sec-CH-UA-Arch" `,
			origin:        "https://example.com",
			wantState:     NEGOTIATION_REQUESTED,
			wantLowLen:    1,
			wantHighLen:   1,
			wantAnomalies: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := analyzer.InitializeFromAcceptCH(tt.acceptCHValue, tt.origin)

			if strategy == nil {
				t.Fatal("InitializeFromAcceptCH() returned nil")
			}

			if strategy.State != tt.wantState {
				t.Errorf("State = %v, want %v", strategy.State, tt.wantState)
			}

			if strategy.ServerPrefs == nil {
				t.Fatal("ServerPrefs should not be nil")
			}

			if len(strategy.ServerPrefs.LowEntropyHints) != tt.wantLowLen {
				t.Errorf("LowEntropyHints length = %d, want %d", len(strategy.ServerPrefs.LowEntropyHints), tt.wantLowLen)
			}

			if len(strategy.ServerPrefs.HighEntropyHints) != tt.wantHighLen {
				t.Errorf("HighEntropyHints length = %d, want %d", len(strategy.ServerPrefs.HighEntropyHints), tt.wantHighLen)
			}

			if tt.wantAnomalies && len(strategy.AnomalyFlags) == 0 {
				t.Error("expected anomalies but got none")
			}

			// Check strategy is stored
			if _, ok := analyzer.strategies[tt.origin]; !ok {
				t.Error("strategy should be stored in analyzer")
			}
		})
	}
}

func TestCHNegotiationAnalyzer_DecideHints(t *testing.T) {
	analyzer := NewCHNegotiationAnalyzer()

	tests := []struct {
		name            string
		userPreference  string
		wantHintCount   int
		wantRejected    int
		wantState       NegotiationState
		wantAnomalyFlag bool
	}{
		{
			name:           "No user consent - only low entropy",
			userPreference: "default",
			wantHintCount:  2,
			wantRejected:   2,
			wantState:      NEGOTIATION_ACCEPTED,
		},
		{
			name:           "With user consent - allow all",
			userPreference: "allow-all",
			wantHintCount:  4,
			wantRejected:   0,
			wantState:      NEGOTIATION_ACCEPTED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset strategy for each test
			testStrategy := &NegotiationStrategy{
				State: NEGOTIATION_REQUESTED,
				ServerPrefs: &ServerPreferences{
					LowEntropyHints:  []string{"sec-ch-ua", "sec-ch-ua-mobile"},
					HighEntropyHints: []string{"sec-ch-ua-arch", "sec-ch-ua-bitness"},
				},
				ClientCaps: &ClientCapabilities{
					SupportedLowEntropy:  []string{"sec-ch-ua", "sec-ch-ua-mobile"},
					SupportedHighEntropy: []string{"sec-ch-ua-arch", "sec-ch-ua-bitness"},
					UserConsent:          tt.userPreference == "allow-all",
				},
				NegotiatedHints: []string{},
				RejectedHints:   []string{},
				AnomalyFlags:    []string{},
			}

			decided := analyzer.DecideHints(testStrategy, tt.userPreference)

			if len(decided) != tt.wantHintCount {
				t.Errorf("decided hints count = %d, want %d", len(decided), tt.wantHintCount)
			}

			if len(testStrategy.RejectedHints) != tt.wantRejected {
				t.Errorf("rejected hints count = %d, want %d", len(testStrategy.RejectedHints), tt.wantRejected)
			}

			if testStrategy.State != tt.wantState {
				t.Errorf("State = %v, want %v", testStrategy.State, tt.wantState)
			}

			if testStrategy.State == NEGOTIATION_ACCEPTED {
				if len(testStrategy.NegotiatedHints) != len(decided) {
					t.Error("NegotiatedHints should match decided hints")
				}
			}
		})
	}
}

func TestCHNegotiationAnalyzer_HandleCrossOriginDelegation(t *testing.T) {
	analyzer := NewCHNegotiationAnalyzer()

	tests := []struct {
		name           string
		strategy       *NegotiationStrategy
		delegateOrigin string
		delegateHints  []string
		wantErr        bool
		wantAnomaly    string
	}{
		{
			name: "Authorized delegation",
			strategy: &NegotiationStrategy{
				NegotiatedHints: []string{"sec-ch-ua", "sec-ch-ua-arch"},
				ServerPrefs: &ServerPreferences{
					DelegateToOrigins: []string{"https://cdn.example.com"},
				},
			},
			delegateOrigin: "https://cdn.example.com",
			delegateHints:  []string{"sec-ch-ua"},
			wantErr:        false,
		},
		{
			name: "Wildcard delegation",
			strategy: &NegotiationStrategy{
				NegotiatedHints: []string{"sec-ch-ua"},
				ServerPrefs: &ServerPreferences{
					DelegateToOrigins: []string{"*"},
				},
			},
			delegateOrigin: "https://any-origin.com",
			delegateHints:  []string{"sec-ch-ua"},
			wantErr:        false,
		},
		{
			name: "Unauthorized delegation",
			strategy: &NegotiationStrategy{
				NegotiatedHints: []string{"sec-ch-ua"},
				ServerPrefs: &ServerPreferences{
					DelegateToOrigins: []string{"https://cdn.example.com"},
				},
				AnomalyFlags: []string{},
			},
			delegateOrigin: "https://unauthorized.com",
			delegateHints:  []string{"sec-ch-ua"},
			wantErr:        true,
			wantAnomaly:    "UNAUTHORIZED_DELEGATION_ATTEMPT",
		},
		{
			name: "Unauthorized hint in delegation",
			strategy: &NegotiationStrategy{
				NegotiatedHints: []string{"sec-ch-ua"},
				ServerPrefs: &ServerPreferences{
					DelegateToOrigins: []string{"https://cdn.example.com"},
				},
				AnomalyFlags: []string{},
			},
			delegateOrigin: "https://cdn.example.com",
			delegateHints:  []string{"sec-ch-ua", "sec-ch-ua-bitness"},
			wantErr:        false,
			wantAnomaly:    "UNAUTHORIZED_HINT_DELEGATION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := analyzer.HandleCrossOriginDelegation(tt.strategy, tt.delegateOrigin, tt.delegateHints)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if tt.wantAnomaly != "" {
				found := false
				for _, flag := range tt.strategy.AnomalyFlags {
					if strings.Contains(flag, tt.wantAnomaly) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected anomaly flag containing %s, got %v", tt.wantAnomaly, tt.strategy.AnomalyFlags)
				}
			}

			// Check state is set to delegated on success
			if !tt.wantErr && tt.strategy.State != NEGOTIATION_DELEGATED {
				t.Errorf("State = %v, want NEGOTIATION_DELEGATED", tt.strategy.State)
			}
		})
	}
}

func TestCHNegotiationAnalyzer_evaluateNegotiationRisk(t *testing.T) {
	analyzer := NewCHNegotiationAnalyzer()

	tests := []struct {
		name           string
		setupStrategy  func(*NegotiationStrategy)
		wantMinRisk    float64
		wantMaxRisk    float64
		wantAnomaly    string
	}{
		{
			name: "Normal strategy",
			setupStrategy: func(s *NegotiationStrategy) {
				s.AnomalyFlags = []string{}
				s.ServerPrefs.HighEntropyHints = []string{"sec-ch-ua-arch"}
			},
			wantMinRisk: 0.0,
			wantMaxRisk: 0.1,
		},
		{
			name: "Excessive anomalies",
			setupStrategy: func(s *NegotiationStrategy) {
				s.AnomalyFlags = []string{"a", "b", "c", "d"}
			},
			wantMinRisk: 0.2,
			wantMaxRisk: 0.5,
		},
		{
			name: "Excessive fingerprinting attempt",
			setupStrategy: func(s *NegotiationStrategy) {
				s.AnomalyFlags = []string{}
				s.ServerPrefs.HighEntropyHints = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
			},
			wantMinRisk: 0.3,
			wantMaxRisk: 0.6,
			wantAnomaly: "EXCESSIVE_FINGERPRINTING_ATTEMPT",
		},
		{
			name: "Excessive domain delegation",
			setupStrategy: func(s *NegotiationStrategy) {
				s.AnomalyFlags = []string{}
				s.ServerPrefs.DelegateToOrigins = []string{"a", "b", "c", "d", "e", "f"}
			},
			wantMinRisk: 0.2,
			wantMaxRisk: 0.5,
			wantAnomaly: "EXCESSIVE_DOMAIN_DELEGATION",
		},
		{
			name: "High entropy over insecure",
			setupStrategy: func(s *NegotiationStrategy) {
				s.AnomalyFlags = []string{}
				s.ServerPrefs.AllowInsecure = true
				s.ServerPrefs.HighEntropyHints = []string{"sec-ch-ua-arch"}
			},
			wantMinRisk: 0.4,
			wantMaxRisk: 0.7,
			wantAnomaly: "HIGH_ENTROPY_OVER_INSECURE",
		},
		{
			name: "Excessive cache duration",
			setupStrategy: func(s *NegotiationStrategy) {
				s.AnomalyFlags = []string{}
				s.ServerPrefs.CacheDuration = 40000000 // More than 1 year
			},
			wantMinRisk: 0.15,
			wantMaxRisk: 0.45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := &NegotiationStrategy{
				ServerPrefs: &ServerPreferences{
					HighEntropyHints:  []string{},
					DelegateToOrigins: []string{},
				},
				AnomalyFlags: []string{},
			}

			tt.setupStrategy(strategy)
			analyzer.evaluateNegotiationRisk(strategy)

			if strategy.RiskScore < tt.wantMinRisk || strategy.RiskScore > tt.wantMaxRisk {
				t.Errorf("RiskScore = %f, want between %f and %f", strategy.RiskScore, tt.wantMinRisk, tt.wantMaxRisk)
			}

			if tt.wantAnomaly != "" {
				found := false
				for _, flag := range strategy.AnomalyFlags {
					if flag == tt.wantAnomaly {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected anomaly %s not found in %v", tt.wantAnomaly, strategy.AnomalyFlags)
				}
			}
		})
	}
}

func TestCHNegotiationAnalyzer_GetNegotiationSummary(t *testing.T) {
	analyzer := NewCHNegotiationAnalyzer()

	tests := []struct {
		name           string
		state          NegotiationState
		negotiatedLen  int
		rejectedLen    int
		anomalyLen     int
		riskScore      float64
		wantInSummary  []string
	}{
		{
			name:          "Initial state",
			state:         NEGOTIATION_INIT,
			negotiatedLen: 0,
			rejectedLen:   0,
			anomalyLen:    0,
			riskScore:     0.0,
			wantInSummary: []string{"初始", "0", "0.00"},
		},
		{
			name:          "Accepted state",
			state:         NEGOTIATION_ACCEPTED,
			negotiatedLen: 3,
			rejectedLen:   1,
			anomalyLen:    0,
			riskScore:     0.1,
			wantInSummary: []string{"已接受", "3", "1", "0", "0.10"},
		},
		{
			name:          "Delegated state",
			state:         NEGOTIATION_DELEGATED,
			negotiatedLen: 5,
			rejectedLen:   2,
			anomalyLen:    1,
			riskScore:     0.25,
			wantInSummary: []string{"已委托", "5", "2", "1", "0.25"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := &NegotiationStrategy{
				State:           tt.state,
				NegotiatedHints: make([]string, tt.negotiatedLen),
				RejectedHints:   make([]string, tt.rejectedLen),
				AnomalyFlags:    make([]string, tt.anomalyLen),
				RiskScore:       tt.riskScore,
			}

			summary := analyzer.GetNegotiationSummary(strategy)

			for _, want := range tt.wantInSummary {
				if !strings.Contains(summary, want) {
					t.Errorf("summary should contain %q, got: %s", want, summary)
				}
			}
		})
	}
}

// ============ Integration Tests ============

func TestClientHintsIntegration(t *testing.T) {
	// Create a complete flow test
	manager := NewCHLifecycleManager()
	originURL := "https://example.com"

	// Step 1: Start lifecycle
	lifecycle := manager.StartLifecycle(originURL, []string{"Sec-CH-UA", "Sec-CH-UA-Mobile"})
	if lifecycle.CurrentPhase != PHASE_INITIAL_REQUEST {
		t.Error("Expected PHASE_INITIAL_REQUEST")
	}

	// Step 2: Process server response
	err := manager.ProcessServerResponse(originURL, "Sec-CH-UA-Arch, Sec-CH-UA-Bitness", "ch-ua=(self)")
	if err != nil {
		t.Fatalf("ProcessServerResponse failed: %v", err)
	}
	if lifecycle.CurrentPhase != PHASE_SERVER_RESPONSE {
		t.Error("Expected PHASE_SERVER_RESPONSE")
	}

	// Step 3: Process same-origin subsequent request
	err = manager.ProcessSubsequentRequest(originURL, originURL, []string{"Sec-CH-UA", "Sec-CH-UA-Mobile", "Sec-CH-UA-Arch"})
	if err != nil {
		t.Fatalf("ProcessSubsequentRequest failed: %v", err)
	}
	if lifecycle.CurrentPhase != PHASE_SUBSEQUENT_REQUESTS {
		t.Error("Expected PHASE_SUBSEQUENT_REQUESTS")
	}

	// Step 4: Process cross-origin request
	crossOrigin := "https://cdn.example.com"
	err = manager.ProcessSubsequentRequest(originURL, crossOrigin, []string{"Sec-CH-UA"})
	if err != nil {
		t.Fatalf("ProcessSubsequentRequest (cross-origin) failed: %v", err)
	}
	if lifecycle.CurrentPhase != PHASE_CROSS_ORIGIN_SUB_REQUESTS {
		t.Error("Expected PHASE_CROSS_ORIGIN_SUB_REQUESTS")
	}

	// Verify discovered origins
	foundCrossOrigin := false
	for _, origin := range lifecycle.DiscoveredOrigins {
		if origin == crossOrigin {
			foundCrossOrigin = true
			break
		}
	}
	if !foundCrossOrigin {
		t.Error("Cross-origin should be in DiscoveredOrigins")
	}

	// Step 5: Get metrics
	metrics := manager.GetLifecycleMetrics(lifecycle)
	if metrics["total_events"].(int) != 4 {
		t.Errorf("Expected 4 events, got %d", metrics["total_events"])
	}

	// Step 6: Get summary
	summary := manager.GetSummary(lifecycle)
	if summary == "" {
		t.Error("Summary should not be empty")
	}

	// Step 7: Terminate
	terminated, err := manager.TerminateLifecycle(originURL)
	if err != nil {
		t.Fatalf("TerminateLifecycle failed: %v", err)
	}
	if terminated.CurrentPhase != PHASE_TERMINATED {
		t.Error("Expected PHASE_TERMINATED")
	}
}

func TestClientHintsProfileGeneration(t *testing.T) {
	// Test full profile generation flow
	profile := &profiles.ClientProfile{
		BrowserType:    "chrome",
		BrowserVersion: "120.0.6099.109",
		OS:             "Windows NT 10.0; Win64; x64",
		OSVersion:      "10.0.19045",
		OSArch:         "x86",
		OSBitness:      "64",
		IsMobile:       false,
		DeviceModel:    "",
	}

	policy := NewClientHintsPolicy(types.BrowserChrome)
	hints := GenerateClientHintsFromProfile(profile, policy)

	// Apply to headers
	headers := make(map[string]string)
	hints.ApplyToHeaders(headers)

	// Verify headers
	if headers["Sec-CH-UA"] == "" {
		t.Error("Sec-CH-UA header should be set")
	}

	if headers["Sec-CH-UA-Mobile"] != "?0" {
		t.Error("Sec-CH-UA-Mobile should be ?0 for desktop")
	}

	if !strings.Contains(headers["Sec-CH-UA-Platform"], "Windows") {
		t.Error("Sec-CH-UA-Platform should contain Windows")
	}

	// Chrome policy should include high entropy hints
	if headers["Sec-CH-UA-Arch"] == "" {
		t.Error("Sec-CH-UA-Arch should be set for Chrome")
	}

	if headers["Sec-CH-UA-Bitness"] != `"64"` {
		t.Errorf("Sec-CH-UA-Bitness = %s, want \"64\"", headers["Sec-CH-UA-Bitness"])
	}
}

// Benchmark tests

func BenchmarkGenerateClientHintsFromProfile(b *testing.B) {
	profile := &profiles.ClientProfile{
		BrowserType:    "chrome",
		BrowserVersion: "120.0.6099.109",
		OS:             "Windows NT 10.0; Win64; x64",
		OSVersion:      "10.0.19045",
		OSArch:         "x86",
		OSBitness:      "64",
		IsMobile:       false,
		DeviceModel:    "",
	}
	policy := NewClientHintsPolicy(types.BrowserChrome)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateClientHintsFromProfile(profile, policy)
	}
}

func BenchmarkProcessAcceptCH(b *testing.B) {
	policy := &ClientHintsPolicy{
		HighEntropyHints: []string{},
	}
	acceptCHValue := "Sec-CH-UA-Arch, Sec-CH-UA-Bitness, Sec-CH-UA-Full-Version-List, Sec-CH-UA-Platform-Version"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := &ClientHintsPolicy{HighEntropyHints: []string{}}
		p.ProcessAcceptCH(acceptCHValue)
	}
	_ = policy // Avoid unused variable warning
}

func BenchmarkCHLifecycleManager_ProcessServerResponse(b *testing.B) {
	manager := NewCHLifecycleManager()
	originURL := "https://example.com"
	manager.StartLifecycle(originURL, []string{"Sec-CH-UA"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.ProcessServerResponse(originURL, "Sec-CH-UA-Arch", "ch-ua=(self)")
	}
}
