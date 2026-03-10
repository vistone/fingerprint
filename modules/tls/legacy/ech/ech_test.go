package ech

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// translated comment
func TestParseECHExtension(t *testing.T) {
	tests := []struct {
		name      string
		extType   uint16
		data      []byte
		wantErr   bool
		checkFunc func(t *testing.T, ech *ECHExtension)
	}{
		{
			name:    "inner_hello_draft13",
			extType: ExtensionEncryptedClientHello,
			data: func() []byte {
				data := make([]byte, 3)
				data[0] = 0xfe
				data[1] = 0x0d // Version Draft 13
				data[2] = 0x01 // Inner Hello (per ECH spec: inner=1)
				return data
			}(),
			wantErr: false,
			checkFunc: func(t *testing.T, ech *ECHExtension) {
				if ech.Version != ECHVersionDraft13 {
					t.Errorf("Version = 0x%04x, want 0x%04x", ech.Version, ECHVersionDraft13)
				}
				if !ech.IsInnerHello() {
					t.Error("Expected inner hello")
				}
			},
		},
		{
			name:    "outer_hello_draft13",
			extType: ExtensionEncryptedClientHello,
			data: func() []byte {
				// Version (2) + Type (1) + Cipher Suite (4) + Config ID (1) + Length (2) + Encoded CH
				data := make([]byte, 10)
				data[0] = 0xfe
				data[1] = 0x0d // Version Draft 13
				data[2] = 0x00 // Outer Hello (per ECH spec: outer=0)
				data[3] = 0x00
				data[4] = 0x01 // KDF ID
				data[5] = 0x00
				data[6] = 0x01 // AEAD ID
				data[7] = 0x01 // Config ID
				data[8] = 0x00
				data[9] = 0x00 // Encoded CH Length = 0
				return data
			}(),
			wantErr: false,
			checkFunc: func(t *testing.T, ech *ECHExtension) {
				if !ech.IsOuterHello() {
					t.Error("Expected outer hello")
				}
				if ech.ConfigID != 1 {
					t.Errorf("ConfigID = %d, want 1", ech.ConfigID)
				}
			},
		},
		{
			name:    "grease_ech",
			extType: ExtensionEncryptedClientHello,
			data: func() []byte {
				data := make([]byte, 2)
				data[0] = 0x00
				data[1] = 0x00 // GREASE version
				return data
			}(),
			wantErr: false,
			checkFunc: func(t *testing.T, ech *ECHExtension) {
				if !ech.IsGREASE() {
					t.Error("Expected GREASE ECH")
				}
			},
		},
		{
			name:    "too_short",
			extType: ExtensionEncryptedClientHello,
			data:    []byte{0xfe},
			wantErr: true,
		},
		{
			name:    "truncated_outer",
			extType: ExtensionEncryptedClientHello,
			data:    []byte{0xfe, 0x0d, 0x00},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ech, err := ParseECHExtension(tt.extType, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseECHExtension() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, ech)
			}
		})
	}
}

// translated comment
func TestECHExtensionSerialization(t *testing.T) {
	t.Run("inner_hello", func(t *testing.T) {
		ech := &ECHExtension{
			Version:         ECHVersionDraft13,
			ClientHelloType: ECHClientHelloTypeInner,
		}

		data, err := ech.Serialize()
		if err != nil {
			t.Fatalf("Serialize() error = %v", err)
		}

		// translated comment
		parsed, err := ParseECHExtension(ExtensionEncryptedClientHello, data)
		if err != nil {
			t.Fatalf("ParseECHExtension() error = %v", err)
		}

		if parsed.Version != ech.Version {
			t.Errorf("Version mismatch: got 0x%04x, want 0x%04x", parsed.Version, ech.Version)
		}
		if parsed.ClientHelloType != ech.ClientHelloType {
			t.Errorf("ClientHelloType mismatch: got %d, want %d", parsed.ClientHelloType, ech.ClientHelloType)
		}
	})

	t.Run("outer_hello", func(t *testing.T) {
		ech := &ECHExtension{
			Type:            ExtensionEncryptedClientHello,
			Version:         ECHVersionDraft13,
			ClientHelloType: ECHClientHelloTypeOuter,
			CipherSuite: KEMCipherSuite{
				KDFID:  0x0001,
				AEADID: 0x0001,
			},
			ConfigID:  42,
			EncodedCH: []byte("test encrypted data"),
		}
		ech.EncodedCHLength = uint16(len(ech.EncodedCH))

		data, err := ech.Serialize()
		if err != nil {
			t.Fatalf("Serialize() error = %v", err)
		}

		parsed, err := ParseECHExtension(ExtensionEncryptedClientHello, data)
		if err != nil {
			t.Fatalf("ParseECHExtension() error = %v", err)
		}

		if !bytes.Equal(parsed.EncodedCH, ech.EncodedCH) {
			t.Errorf("EncodedCH mismatch: got %v, want %v", parsed.EncodedCH, ech.EncodedCH)
		}
		if parsed.ConfigID != ech.ConfigID {
			t.Errorf("ConfigID mismatch: got %d, want %d", parsed.ConfigID, ech.ConfigID)
		}
	})
}

// translated comment
func TestECHAnalyzer(t *testing.T) {
	analyzer := NewECHAnalyzer()

	t.Run("no_ech", func(t *testing.T) {
		data := ClientHelloData{
			TLSVersion:         0x0303,
			CipherSuites:       []uint16{0x1301, 0x1302},
			Extensions:         []ExtensionData{},
			CompressionMethods: []uint8{0},
			HasSNI:             true,
			SNI:                "example.com",
		}

		result, err := analyzer.AnalyzeClientHello(data)
		if err != nil {
			t.Fatalf("AnalyzeClientHello() error = %v", err)
		}

		if result.ECHPresent {
			t.Error("Expected ECH not present")
		}
		if result.Impact.ImpactLevel != "none" {
			t.Errorf("ImpactLevel = %s, want none", result.Impact.ImpactLevel)
		}
		if !result.Impact.SNIVisible {
			t.Error("Expected SNI visible when no ECH")
		}
	})

	t.Run("with_ech_outer", func(t *testing.T) {
		// translated comment
		echData := make([]byte, 10)
		echData[0] = 0xfe
		echData[1] = 0x0d // Draft 13
		echData[2] = 0x00 // Outer (per ECH spec: outer=0)
		echData[3] = 0x00
		echData[4] = 0x01 // KDF
		echData[5] = 0x00
		echData[6] = 0x01 // AEAD
		echData[7] = 0x01 // Config ID
		echData[8] = 0x00
		echData[9] = 0x02 // Length = 2

		data := ClientHelloData{
			TLSVersion:   0x0303,
			CipherSuites: []uint16{0x1301, 0x1302},
			Extensions: []ExtensionData{
				{Type: ExtensionEncryptedClientHello, Data: append(echData, []byte{0x00, 0x00}...)},
			},
			CompressionMethods: []uint8{0},
			HasSNI:             false, // translated comment
		}

		result, err := analyzer.AnalyzeClientHello(data)
		if err != nil {
			t.Fatalf("AnalyzeClientHello() error = %v", err)
		}

		if !result.ECHPresent {
			t.Error("Expected ECH present")
		}
		if result.ECHType != "outer" {
			t.Errorf("ECHType = %s, want outer", result.ECHType)
		}
		if result.Impact.ImpactLevel != "high" {
			t.Errorf("ImpactLevel = %s, want high", result.Impact.ImpactLevel)
		}
	})

	t.Run("with_grease_ech", func(t *testing.T) {
		// GREASE ECH
		echData := []byte{0x00, 0x00} // Version 0 = GREASE

		data := ClientHelloData{
			TLSVersion:   0x0303,
			CipherSuites: []uint16{0x1301},
			Extensions: []ExtensionData{
				{Type: ExtensionEncryptedClientHello, Data: echData},
			},
			CompressionMethods: []uint8{0},
		}

		result, err := analyzer.AnalyzeClientHello(data)
		if err != nil {
			t.Fatalf("AnalyzeClientHello() error = %v", err)
		}

		if !result.ECHPresent {
			t.Error("Expected ECH present")
		}
		if result.ECHType != "grease" {
			t.Errorf("ECHType = %s, want grease", result.ECHType)
		}
		if result.Impact.ImpactLevel != "low" {
			t.Errorf("ImpactLevel = %s, want low for GREASE", result.Impact.ImpactLevel)
		}
	})
}

// translated comment
func TestECHProfile(t *testing.T) {
	t.Run("default_profile", func(t *testing.T) {
		profile := DefaultECHProfile()

		if !profile.Enabled {
			t.Error("Default profile should be enabled")
		}
		if profile.Version != ECHVersionDraft13 {
			t.Errorf("Version = 0x%04x, want 0x%04x", profile.Version, ECHVersionDraft13)
		}
		if profile.PublicName == "" {
			t.Error("PublicName should not be empty")
		}
	})

	t.Run("validate_valid", func(t *testing.T) {
		profile := DefaultECHProfile()
		if err := profile.Validate(); err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})

	t.Run("validate_invalid_version", func(t *testing.T) {
		profile := DefaultECHProfile()
		profile.Version = 0x9999
		if err := profile.Validate(); err == nil {
			t.Error("Expected error for invalid version")
		}
	})

	t.Run("validate_empty_cipher_suites", func(t *testing.T) {
		profile := DefaultECHProfile()
		profile.CipherSuites = []KEMCipherSuite{}
		if err := profile.Validate(); err == nil {
			t.Error("Expected error for empty cipher suites")
		}
	})

	t.Run("generate_inner_extension", func(t *testing.T) {
		profile := DefaultECHProfile()
		ech, err := profile.GenerateECHExtension(true)
		if err != nil {
			t.Fatalf("GenerateECHExtension() error = %v", err)
		}

		if !ech.IsInnerHello() {
			t.Error("Expected inner hello")
		}
	})

	t.Run("generate_outer_extension", func(t *testing.T) {
		profile := DefaultECHProfile()
		profile.ConfigID = 123
		ech, err := profile.GenerateECHExtension(false)
		if err != nil {
			t.Fatalf("GenerateECHExtension() error = %v", err)
		}

		if !ech.IsOuterHello() {
			t.Error("Expected outer hello")
		}
		if ech.ConfigID != 123 {
			t.Errorf("ConfigID = %d, want 123", ech.ConfigID)
		}
	})
}

// translated comment
func TestConfigGenerator(t *testing.T) {
	opts := DefaultConfigOptions()
	generator := NewConfigGenerator(opts)

	t.Run("generate_config", func(t *testing.T) {
		config, err := generator.GenerateECHConfig()
		if err != nil {
			t.Fatalf("GenerateECHConfig() error = %v", err)
		}

		if config.Version != ECHVersionDraft13 {
			t.Errorf("Version = 0x%04x, want 0x%04x", config.Version, ECHVersionDraft13)
		}
		if config.PublicName != opts.PublicName {
			t.Errorf("PublicName = %s, want %s", config.PublicName, opts.PublicName)
		}
		if len(config.Contents) == 0 {
			t.Error("Contents should not be empty")
		}
	})

	t.Run("generate_config_list", func(t *testing.T) {
		list, err := generator.GenerateECHConfigList(3)
		if err != nil {
			t.Fatalf("GenerateECHConfigList() error = %v", err)
		}

		if len(list.Configs) != 3 {
			t.Errorf("len(Configs) = %d, want 3", len(list.Configs))
		}
	})

	t.Run("serialize_and_parse", func(t *testing.T) {
		list, err := generator.GenerateECHConfigList(1)
		if err != nil {
			t.Fatalf("GenerateECHConfigList() error = %v", err)
		}

		data, err := SerializeECHConfigList(list)
		if err != nil {
			t.Fatalf("SerializeECHConfigList() error = %v", err)
		}

		// translated comment
		parsed, err := ParseECHConfigList(data)
		if err != nil {
			t.Fatalf("ParseECHConfigList() error = %v", err)
		}

		if len(parsed.Configs) != 1 {
			t.Errorf("len(Configs) = %d, want 1", len(parsed.Configs))
		}
	})
}

// translated comment
func TestECHKeyManager(t *testing.T) {
	manager := NewECHKeyManager()
	opts := DefaultConfigOptions()

	t.Run("generate_key", func(t *testing.T) {
		keySet, err := manager.GenerateNewKey(1, opts)
		if err != nil {
			t.Fatalf("GenerateNewKey() error = %v", err)
		}

		if keySet.ConfigID != 1 {
			t.Errorf("ConfigID = %d, want 1", keySet.ConfigID)
		}
		if keySet.PublicConfig == nil {
			t.Error("PublicConfig should not be nil")
		}
		if len(keySet.PrivateKey) == 0 {
			t.Error("PrivateKey should not be empty")
		}
	})

	t.Run("get_public_config", func(t *testing.T) {
		_, err := manager.GenerateNewKey(2, opts)
		if err != nil {
			t.Fatalf("GenerateNewKey() error = %v", err)
		}

		config, err := manager.GetPublicConfig(2)
		if err != nil {
			t.Fatalf("GetPublicConfig() error = %v", err)
		}

		if config.Version != ECHVersionDraft13 {
			t.Errorf("Version = 0x%04x, want 0x%04x", config.Version, ECHVersionDraft13)
		}
	})

	t.Run("get_public_config_not_found", func(t *testing.T) {
		_, err := manager.GetPublicConfig(255)
		if err == nil {
			t.Error("Expected error for non-existent config")
		}
	})
}

// translated comment
func TestProfileFromBrowser(t *testing.T) {
	tests := []struct {
		browser  string
		expected string
	}{
		{"chrome", "google-ech.cloudflareresearch.com"},
		{"Chrome", "google-ech.cloudflareresearch.com"},
		{"firefox", "firefox-ech.cloudflareresearch.com"},
		{"Firefox", "firefox-ech.cloudflareresearch.com"},
		{"safari", ""},
		{"Safari", ""},
		{"unknown", "cloudflare-ech.com"},
	}

	for _, tt := range tests {
		t.Run(tt.browser, func(t *testing.T) {
			profile := ECHProfileFromBrowser(tt.browser)
			if profile == nil {
				t.Fatal("Profile should not be nil")
			}
			// translated comment
			if tt.expected != "" && profile.PublicName != tt.expected {
				t.Errorf("PublicName = %s, want %s", profile.PublicName, tt.expected)
			}
		})
	}
}

// translated comment
func BenchmarkECHExtensionParse(b *testing.B) {
	data := make([]byte, 10)
	data[0] = 0xfe
	data[1] = 0x0d
	data[2] = 0x00 // Outer (per ECH spec: outer=0)
	data[3] = 0x00
	data[4] = 0x01
	data[5] = 0x00
	data[6] = 0x01
	data[7] = 0x01
	data[8] = 0x00
	data[9] = 0x00

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ParseECHExtension(ExtensionEncryptedClientHello, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// translated comment
func BenchmarkECHAnalyzer(b *testing.B) {
	analyzer := NewECHAnalyzer()
	data := ClientHelloData{
		TLSVersion:         0x0303,
		CipherSuites:       []uint16{0x1301, 0x1302},
		Extensions:         []ExtensionData{},
		CompressionMethods: []uint8{0},
		HasSNI:             true,
		SNI:                "example.com",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeClientHello(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// translated comment
func TestECHExtensionValidation(t *testing.T) {
	tests := []struct {
		name    string
		ech     *ECHExtension
		wantErr bool
	}{
		{
			name: "valid_inner",
			ech: &ECHExtension{
				Version:         ECHVersionDraft13,
				ClientHelloType: ECHClientHelloTypeInner,
			},
			wantErr: false,
		},
		{
			name: "valid_outer",
			ech: &ECHExtension{
				Version:         ECHVersionDraft13,
				ClientHelloType: ECHClientHelloTypeOuter,
				EncodedCH:       []byte("encrypted"),
			},
			wantErr: false,
		},
		{
			name: "zero_version",
			ech: &ECHExtension{
				Version:         0,
				ClientHelloType: ECHClientHelloTypeInner,
			},
			wantErr: true,
		},
		{
			name: "outer_without_encoded_ch",
			ech: &ECHExtension{
				Version:         ECHVersionDraft13,
				ClientHelloType: ECHClientHelloTypeOuter,
				EncodedCH:       nil,
			},
			wantErr: true,
		},
		{
			name: "invalid_hello_type",
			ech: &ECHExtension{
				Version:         ECHVersionDraft13,
				ClientHelloType: 99,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ech.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// translated comment
func TestVersionHelpers(t *testing.T) {
	t.Run("get_supported_versions", func(t *testing.T) {
		versions := GetSupportedECHVersions()
		if len(versions) == 0 {
			t.Error("Should return supported versions")
		}
		found := false
		for _, v := range versions {
			if v == ECHVersionDraft13 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Draft 13 should be in supported versions")
		}
	})

	t.Run("version_name", func(t *testing.T) {
		tests := []struct {
			version uint16
			want    string
		}{
			{ECHVersionDraft13, "Draft 13"},
			{ECHVersionDraft14, "Draft 14"},
			{ECHVersionDraft15, "Draft 15"},
			{0x9999, "Unknown(0x9999)"},
		}

		for _, tt := range tests {
			got := ECHVersionName(tt.version)
			if got != tt.want {
				t.Errorf("ECHVersionName(0x%04x) = %s, want %s", tt.version, got, tt.want)
			}
		}
	})
}

// translated comment
func TestMergeWithBase(t *testing.T) {
	profile := DefaultECHProfile()

	t.Run("merge_without_sni", func(t *testing.T) {
		base := []uint16{10, 11, 13} // translated comment
		merged := profile.MergeWithBase(base)

		found := false
		for _, ext := range merged {
			if ext == ExtensionEncryptedClientHello {
				found = true
				break
			}
		}
		if !found {
			t.Error("ECH extension should be added")
		}
	})

	t.Run("merge_with_sni", func(t *testing.T) {
		base := []uint16{0, 10, 11} // translated comment
		merged := profile.MergeWithBase(base)

		// translated comment
		sniIndex := -1
		echIndex := -1
		for i, ext := range merged {
			if ext == 0 {
				sniIndex = i
			}
			if ext == ExtensionEncryptedClientHello {
				echIndex = i
			}
		}

		if sniIndex == -1 {
			t.Error("SNI extension should be present")
		}
		if echIndex == -1 {
			t.Error("ECH extension should be present")
		}
		if echIndex <= sniIndex {
			t.Error("ECH should be after SNI")
		}
	})

	t.Run("ech_already_present", func(t *testing.T) {
		base := []uint16{0, ExtensionEncryptedClientHello, 10}
		merged := profile.MergeWithBase(base)

		// translated comment
		count := 0
		for _, ext := range merged {
			if ext == ExtensionEncryptedClientHello {
				count++
			}
		}
		if count != 1 {
			t.Errorf("ECH extension should appear exactly once, got %d", count)
		}
	})
}

// translated comment
func TestAnalyzeECH(t *testing.T) {
	data := ClientHelloData{
		TLSVersion:         0x0303,
		CipherSuites:       []uint16{0x1301, 0x1302},
		Extensions:         []ExtensionData{},
		CompressionMethods: []uint8{0},
		HasSNI:             true,
	}

	result, err := AnalyzeECH(data)
	if err != nil {
		t.Fatalf("AnalyzeECH() error = %v", err)
	}

	if result.ECHPresent {
		t.Error("Expected no ECH")
	}
}

// translated comment
func TestImpactSummary(t *testing.T) {
	t.Run("no_ech", func(t *testing.T) {
		result := &ECHAnalysisResult{
			ECHPresent: false,
		}
		summary := result.GetImpactSummary()
		if summary == "" {
			t.Error("Impact summary should not be empty")
		}
	})

	t.Run("with_ech", func(t *testing.T) {
		result := &ECHAnalysisResult{
			ECHPresent: true,
			ECHType:    "outer",
			Impact: ECHImpact{
				ImpactLevel: "high",
				SNIVisible:  false,
			},
		}
		summary := result.GetImpactSummary()
		if summary == "" {
			t.Error("Impact summary should not be empty")
		}
	})
}

// translated comment
func mustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// translated comment
func TestParseECHConfigList(t *testing.T) {
	t.Run("empty_list", func(t *testing.T) {
		// translated comment
		data := []byte{0x00, 0x00}
		list, err := ParseECHConfigList(data)
		if err != nil {
			t.Fatalf("ParseECHConfigList() error = %v", err)
		}
		if len(list.Configs) != 0 {
			t.Errorf("len(Configs) = %d, want 0", len(list.Configs))
		}
	})

	t.Run("too_short", func(t *testing.T) {
		data := []byte{0x00} // translated comment
		_, err := ParseECHConfigList(data)
		if err == nil {
			t.Error("Expected error for too short data")
		}
	})
}
