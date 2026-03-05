package ja3

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	tls "github.com/bogdanfinn/utls"
	errdefs "github.com/vistone/fingerprint/modules/errors"
	"github.com/vistone/fingerprint/modules/profiles/legacy"
)

// ============================================================================
// 错误类型判断函数测试
// ============================================================================

// TestErrorTypes 测试所有 JA3 错误类型判断函数
func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checkers map[string]func(error) bool
		want     map[string]bool
	}{
		{
			name: "TestIsInvalidClientHelloSpec",
			err:  ErrInvalidClientHelloSpec,
			checkers: map[string]func(error) bool{
				"IsInvalidClientHelloSpec":      IsInvalidClientHelloSpec,
				"IsJA3ProfileNotFound":          IsJA3ProfileNotFound,
				"IsEmptyProfile":                IsEmptyProfile,
				"IsClientHelloIDNotImplemented": IsClientHelloIDNotImplemented,
			},
			want: map[string]bool{
				"IsInvalidClientHelloSpec":      true,
				"IsJA3ProfileNotFound":          false,
				"IsEmptyProfile":                false,
				"IsClientHelloIDNotImplemented": false,
			},
		},
		{
			name: "TestIsJA3ProfileNotFound",
			err:  ErrProfileNotFound,
			checkers: map[string]func(error) bool{
				"IsInvalidClientHelloSpec":      IsInvalidClientHelloSpec,
				"IsJA3ProfileNotFound":          IsJA3ProfileNotFound,
				"IsEmptyProfile":                IsEmptyProfile,
				"IsClientHelloIDNotImplemented": IsClientHelloIDNotImplemented,
			},
			want: map[string]bool{
				"IsInvalidClientHelloSpec":      false,
				"IsJA3ProfileNotFound":          true,
				"IsEmptyProfile":                false,
				"IsClientHelloIDNotImplemented": false,
			},
		},
		{
			name: "TestIsEmptyProfile",
			err:  ErrEmptyProfile,
			checkers: map[string]func(error) bool{
				"IsInvalidClientHelloSpec":      IsInvalidClientHelloSpec,
				"IsJA3ProfileNotFound":          IsJA3ProfileNotFound,
				"IsEmptyProfile":                IsEmptyProfile,
				"IsClientHelloIDNotImplemented": IsClientHelloIDNotImplemented,
			},
			want: map[string]bool{
				"IsInvalidClientHelloSpec":      false,
				"IsJA3ProfileNotFound":          false,
				"IsEmptyProfile":                true,
				"IsClientHelloIDNotImplemented": false,
			},
		},
		{
			name: "TestIsClientHelloIDNotImplemented",
			err:  ErrClientHelloIDNotImplemented,
			checkers: map[string]func(error) bool{
				"IsInvalidClientHelloSpec":      IsInvalidClientHelloSpec,
				"IsJA3ProfileNotFound":          IsJA3ProfileNotFound,
				"IsEmptyProfile":                IsEmptyProfile,
				"IsClientHelloIDNotImplemented": IsClientHelloIDNotImplemented,
			},
			want: map[string]bool{
				"IsInvalidClientHelloSpec":      false,
				"IsJA3ProfileNotFound":          false,
				"IsEmptyProfile":                false,
				"IsClientHelloIDNotImplemented": true,
			},
		},
		{
			name: "nil error returns false for all",
			err:  nil,
			checkers: map[string]func(error) bool{
				"IsInvalidClientHelloSpec":      IsInvalidClientHelloSpec,
				"IsJA3ProfileNotFound":          IsJA3ProfileNotFound,
				"IsEmptyProfile":                IsEmptyProfile,
				"IsClientHelloIDNotImplemented": IsClientHelloIDNotImplemented,
			},
			want: map[string]bool{
				"IsInvalidClientHelloSpec":      false,
				"IsJA3ProfileNotFound":          false,
				"IsEmptyProfile":                false,
				"IsClientHelloIDNotImplemented": false,
			},
		},
		{
			name: "unrelated error returns false for all",
			err:  errors.New("some random error"),
			checkers: map[string]func(error) bool{
				"IsInvalidClientHelloSpec":      IsInvalidClientHelloSpec,
				"IsJA3ProfileNotFound":          IsJA3ProfileNotFound,
				"IsEmptyProfile":                IsEmptyProfile,
				"IsClientHelloIDNotImplemented": IsClientHelloIDNotImplemented,
			},
			want: map[string]bool{
				"IsInvalidClientHelloSpec":      false,
				"IsJA3ProfileNotFound":          false,
				"IsEmptyProfile":                false,
				"IsClientHelloIDNotImplemented": false,
			},
		},
		{
			name: "wrapped error should match",
			err:  fmt.Errorf("wrapped: %w", ErrInvalidClientHelloSpec),
			checkers: map[string]func(error) bool{
				"IsInvalidClientHelloSpec": IsInvalidClientHelloSpec,
			},
			want: map[string]bool{
				"IsInvalidClientHelloSpec": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for checkerName, checker := range tt.checkers {
				got := checker(tt.err)
				want := tt.want[checkerName]
				if got != want {
					t.Errorf("%s() = %v, want %v for error %v", checkerName, got, want, tt.err)
				}
			}
		})
	}
}

// ============================================================================
// InitMappedTLSClients 测试
// ============================================================================

// TestInitMappedTLSClients 测试初始化 MappedTLSClients
func TestInitMappedTLSClients(t *testing.T) {
	tests := []struct {
		name         string
		clients      interface{}
		expectedLen  int
		expectedKeys []string
	}{
		{
			name: "init with map[string]ClientProfile",
			clients: map[string]ClientProfile{
				"test_profile_1": &mockClientProfile{},
				"test_profile_2": &mockClientProfile{},
			},
			expectedLen:  2,
			expectedKeys: []string{"test_profile_1", "test_profile_2"},
		},
		{
			name: "init with map[string]interface{}",
			clients: map[string]interface{}{
				"test_profile_3": &mockClientProfile{},
				"test_profile_4": &mockClientProfile{},
			},
			expectedLen:  2,
			expectedKeys: []string{"test_profile_3", "test_profile_4"},
		},
		{
			name: "init with empty map",
			clients: map[string]ClientProfile{
				"empty_test": &mockClientProfile{},
			},
			expectedLen:  1,
			expectedKeys: []string{"empty_test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始状态
			originalMapped := MappedTLSClients
			defer func() {
				MappedTLSClients = originalMapped
			}()

			// 初始化
			InitMappedTLSClients(tt.clients)

			// 验证
			if len(MappedTLSClients) != tt.expectedLen {
				t.Errorf("Expected %d profiles, got %d", tt.expectedLen, len(MappedTLSClients))
			}

			for _, key := range tt.expectedKeys {
				if _, ok := MappedTLSClients[key]; !ok {
					t.Errorf("Expected profile %s not found", key)
				}
			}
		})
	}
}

// TestInitMappedTLSClients_Idempotent 测试多次初始化（幂等性）
func TestInitMappedTLSClients_Idempotent(t *testing.T) {
	// 保存原始状态
	originalMapped := MappedTLSClients
	defer func() {
		MappedTLSClients = originalMapped
	}()

	// 第一次初始化
	firstMap := map[string]ClientProfile{
		"profile_a": &mockClientProfile{},
	}
	InitMappedTLSClients(firstMap)

	if len(MappedTLSClients) != 1 {
		t.Fatalf("Expected 1 profile after first init, got %d", len(MappedTLSClients))
	}

	// 第二次初始化（覆盖）
	secondMap := map[string]ClientProfile{
		"profile_b": &mockClientProfile{},
		"profile_c": &mockClientProfile{},
	}
	InitMappedTLSClients(secondMap)

	if len(MappedTLSClients) != 2 {
		t.Errorf("Expected 2 profiles after second init, got %d", len(MappedTLSClients))
	}

	// 验证第一次的 profile 被覆盖
	if _, ok := MappedTLSClients["profile_a"]; ok {
		t.Error("Old profile should be removed after re-initialization")
	}

	if _, ok := MappedTLSClients["profile_b"]; !ok {
		t.Error("New profile_b should exist")
	}
}

// ============================================================================
// InitMappedTLSClientsRaw 测试
// ============================================================================

// TestInitMappedTLSClientsRaw 测试使用原始字节初始化
func TestInitMappedTLSClientsRaw(t *testing.T) {
	tests := []struct {
		name        string
		clients     interface{}
		expectedLen int
		checkKeys   []string
	}{
		{
			name: "init with valid map",
			clients: map[string]*mockClientProfile{
				"raw_profile_1": {},
				"raw_profile_2": {},
			},
			expectedLen: 2,
			checkKeys:   []string{"raw_profile_1", "raw_profile_2"},
		},
		{
			name:        "init with non-map returns nil",
			clients:     "not a map",
			expectedLen: 0,
			checkKeys:   []string{},
		},
		{
			name:        "init with nil",
			clients:     nil,
			expectedLen: 0,
			checkKeys:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始状态
			originalMapped := MappedTLSClients
			defer func() {
				MappedTLSClients = originalMapped
			}()

			// 初始化
			InitMappedTLSClientsRaw(tt.clients)

			// 验证
			if len(MappedTLSClients) != tt.expectedLen {
				t.Errorf("Expected %d profiles, got %d", tt.expectedLen, len(MappedTLSClients))
			}

			for _, key := range tt.checkKeys {
				if _, ok := MappedTLSClients[key]; !ok {
					t.Errorf("Expected profile %s not found", key)
				}
			}
		})
	}
}

// TestInitMappedTLSClientsRaw_WithMethod 测试带有方法的类型
func TestInitMappedTLSClientsRaw_WithMethod(t *testing.T) {
	// 保存原始状态
	originalMapped := MappedTLSClients
	defer func() {
		MappedTLSClients = originalMapped
	}()

	// 使用带有 GetClientHelloSpec 方法的类型
	clients := map[string]*mockClientProfileWithMethod{
		"method_profile": {
			spec: tls.ClientHelloSpec{
				CipherSuites: []uint16{0x1301, 0x1302},
				Extensions:   []tls.TLSExtension{&tls.SNIExtension{}},
			},
		},
	}

	InitMappedTLSClientsRaw(clients)

	if len(MappedTLSClients) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(MappedTLSClients))
	}

	// 验证可以通过 ClientProfile 接口调用 GetClientHelloSpec
	profile, ok := MappedTLSClients["method_profile"]
	if !ok {
		t.Fatal("method_profile not found")
	}

	spec, err := profile.GetClientHelloSpec()
	if err != nil {
		t.Fatalf("GetClientHelloSpec failed: %v", err)
	}

	if len(spec.CipherSuites) != 2 {
		t.Errorf("Expected 2 cipher suites, got %d", len(spec.CipherSuites))
	}
}

// ============================================================================
// GetClientHelloSpec 测试
// ============================================================================

// mockClientProfile 实现了 ClientProfile 接口的 mock
func (m *mockClientProfile) GetClientHelloSpec() (tls.ClientHelloSpec, error) {
	if m.err != nil {
		return tls.ClientHelloSpec{}, m.err
	}
	return m.spec, nil
}

// mockClientProfileWithMethod 带有 GetClientHelloSpec 方法的 mock 类型
type mockClientProfileWithMethod struct {
	spec tls.ClientHelloSpec
	err  error
}

func (m *mockClientProfileWithMethod) GetClientHelloSpec() (tls.ClientHelloSpec, error) {
	return m.spec, m.err
}

type mockClientProfile struct {
	spec tls.ClientHelloSpec
	err  error
}

// TestGetClientHelloSpec 测试通过配置获取 spec
func TestGetClientHelloSpec(t *testing.T) {
	validSpec := tls.ClientHelloSpec{
		CipherSuites: []uint16{0x1301, 0x1302},
		Extensions:   []tls.TLSExtension{&tls.SNIExtension{}},
	}

	tests := []struct {
		name      string
		profile   ClientProfile
		wantErr   bool
		errCheck  func(error) bool
		checkSpec func(*testing.T, tls.ClientHelloSpec)
	}{
		{
			name:    "valid profile returns spec",
			profile: &mockClientProfile{spec: validSpec},
			wantErr: false,
			checkSpec: func(t *testing.T, spec tls.ClientHelloSpec) {
				if len(spec.CipherSuites) != 2 {
					t.Errorf("Expected 2 cipher suites, got %d", len(spec.CipherSuites))
				}
			},
		},
		{
			name:    "profile returns error",
			profile: &mockClientProfile{err: fmt.Errorf("spec error")},
			wantErr: true,
		},
		{
			name:    "spec with nil slices returns error on validation",
			profile: &mockClientProfile{spec: tls.ClientHelloSpec{CipherSuites: nil, Extensions: nil}},
			wantErr: false, // GetClientHelloSpec itself doesn't validate, error happens in ComputeJA3
			checkSpec: func(t *testing.T, spec tls.ClientHelloSpec) {
				if spec.CipherSuites != nil {
					t.Error("Expected nil CipherSuites")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := tt.profile.GetClientHelloSpec()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkSpec != nil {
				tt.checkSpec(t, spec)
			}
		})
	}
}

// ============================================================================
// ComputeJA3FromProfile 测试
// ============================================================================

// TestComputeJA3FromProfile 测试从 Profile 计算 JA3
func TestComputeJA3FromProfile(t *testing.T) {
	tests := []struct {
		name    string
		profile ClientProfile
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid profile",
			profile: &mockClientProfile{
				spec: tls.ClientHelloSpec{
					CipherSuites: []uint16{0x1301, 0x1302},
					Extensions:   []tls.TLSExtension{&tls.SNIExtension{}},
				},
			},
			wantErr: false,
		},
		{
			name: "profile with empty spec",
			profile: &mockClientProfile{
				spec: tls.ClientHelloSpec{
					CipherSuites: []uint16{},
					Extensions:   []tls.TLSExtension{},
				},
			},
			wantErr: true,
			errMsg:  "cipher suites list is empty",
		},
		{
			name:    "profile that returns error",
			profile: &mockClientProfile{err: errors.New("profile error")},
			wantErr: true,
			errMsg:  "profile error",
		},
		{
			name: "spec with only GREASE values",
			profile: &mockClientProfile{
				spec: tls.ClientHelloSpec{
					CipherSuites: []uint16{0x0A0A, 0x1A1A},
					Extensions:   []tls.TLSExtension{&tls.UtlsGREASEExtension{}},
				},
			},
			wantErr: true,
			errMsg:  "no valid cipher suites",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ComputeJA3FromProfile(tt.profile)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected non-nil result")
				return
			}

			// 验证结果是有效的 MD5 哈希
			if len(result.Hash) != 32 {
				t.Errorf("Expected MD5 hash length 32, got %d", len(result.Hash))
			}

			if result.RawString == "" {
				t.Error("Expected non-empty raw string")
			}
		})
	}
}

// ============================================================================
// ComputeJA3ByProfileName 测试
// ============================================================================

// TestComputeJA3ByProfileName 测试通过 Profile 名称计算 JA3
func TestComputeJA3ByProfileName(t *testing.T) {
	// 保存原始状态
	originalMapped := MappedTLSClients
	defer func() {
		MappedTLSClients = originalMapped
	}()

	// 设置测试用的 MappedTLSClients
	validSpec := tls.ClientHelloSpec{
		CipherSuites: []uint16{0x1301, 0x1302},
		Extensions:   []tls.TLSExtension{&tls.SNIExtension{}},
	}

	MappedTLSClients = map[string]ClientProfile{
		"valid_profile":   &mockClientProfile{spec: validSpec},
		"error_profile":   &mockClientProfile{err: errors.New("spec error")},
		"invalid_profile": &mockClientProfile{spec: tls.ClientHelloSpec{CipherSuites: []uint16{}, Extensions: []tls.TLSExtension{}}},
	}

	tests := []struct {
		name        string
		profileName string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "existing valid profile",
			profileName: "valid_profile",
			wantErr:     false,
		},
		{
			name:        "non-existing profile",
			profileName: "non_existing_profile",
			wantErr:     true,
			errMsg:      "不存在",
		},
		{
			name:        "profile that returns error",
			profileName: "error_profile",
			wantErr:     true,
			errMsg:      "spec error",
		},
		{
			name:        "profile with invalid spec",
			profileName: "invalid_profile",
			wantErr:     true,
			errMsg:      "cipher suites list is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ComputeJA3ByProfileName(tt.profileName)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected non-nil result")
				return
			}

			// 验证结果是有效的 MD5 哈希
			if len(result.Hash) != 32 {
				t.Errorf("Expected MD5 hash length 32, got %d", len(result.Hash))
			}
		})
	}
}

// ============================================================================
// FindProfileByJA3_Extended 测试
// ============================================================================

// TestFindProfileByJA3_Extended 扩展现有测试到 100% 覆盖
func TestFindProfileByJA3_Extended(t *testing.T) {
	// 保存原始状态并在测试后恢复
	originalMapped := MappedTLSClients
	originalIndexOnce := ja3ProfileIndexOnce
	originalIndex := ja3ProfileIndex
	defer func() {
		MappedTLSClients = originalMapped
		ja3ProfileIndexOnce = originalIndexOnce
		ja3ProfileIndex = originalIndex
	}()

	// 创建有效的 spec 用于计算 JA3
	validSpec := tls.ClientHelloSpec{
		CipherSuites: []uint16{0x1301, 0x1302},
		Extensions: []tls.TLSExtension{
			&tls.SNIExtension{},
			&tls.SupportedCurvesExtension{Curves: []tls.CurveID{tls.X25519}},
			&tls.SupportedPointsExtension{SupportedPoints: []uint8{0}},
		},
	}

	// 计算预期的 JA3 hash
	result, err := ComputeJA3FromSpec(validSpec)
	if err != nil {
		t.Fatalf("Failed to compute JA3: %v", err)
	}
	expectedHash := result.Hash

	tests := []struct {
		name           string
		setup          func()
		ja3Hash        string
		expectedResult []string
		description    string
	}{
		{
			name: "normal lookup",
			setup: func() {
				// 重置 once 和 index
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = map[string]ClientProfile{
					"test_chrome": &mockClientProfile{spec: validSpec},
				}
			},
			ja3Hash:        expectedHash,
			expectedResult: []string{"test_chrome"},
			description:    "应该找到匹配的 profile",
		},
		{
			name: "uninitialized MappedTLSClients",
			setup: func() {
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = nil
			},
			ja3Hash:        expectedHash,
			expectedResult: nil,
			description:    "未初始化时应该返回 nil",
		},
		{
			name: "invalid hash length - too short",
			setup: func() {
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = map[string]ClientProfile{
					"test_chrome": &mockClientProfile{spec: validSpec},
				}
			},
			ja3Hash:        "12345678",
			expectedResult: nil,
			description:    "短哈希不应该匹配",
		},
		{
			name: "invalid hash length - too long",
			setup: func() {
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = map[string]ClientProfile{
					"test_chrome": &mockClientProfile{spec: validSpec},
				}
			},
			ja3Hash:        "1234567890123456789012345678901234567890",
			expectedResult: nil,
			description:    "长哈希不应该匹配",
		},
		{
			name: "no matching profile",
			setup: func() {
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = map[string]ClientProfile{
					"test_chrome": &mockClientProfile{spec: validSpec},
				}
			},
			ja3Hash:        "00000000000000000000000000000000",
			expectedResult: nil,
			description:    "不存在的哈希应该返回 nil",
		},
		{
			name: "case insensitive match - uppercase",
			setup: func() {
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = map[string]ClientProfile{
					"test_chrome": &mockClientProfile{spec: validSpec},
				}
			},
			ja3Hash:        strings.ToUpper(expectedHash),
			expectedResult: []string{"test_chrome"},
			description:    "大写哈希应该匹配",
		},
		{
			name: "case insensitive match - mixed case",
			setup: func() {
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = map[string]ClientProfile{
					"test_chrome": &mockClientProfile{spec: validSpec},
				}
			},
			ja3Hash:        strings.ToUpper(expectedHash[:16]) + strings.ToLower(expectedHash[16:]),
			expectedResult: []string{"test_chrome"},
			description:    "混合大小写哈希应该匹配",
		},
		{
			name: "multiple profiles with same JA3",
			setup: func() {
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = map[string]ClientProfile{
					"test_chrome_a": &mockClientProfile{spec: validSpec},
					"test_chrome_b": &mockClientProfile{spec: validSpec},
				}
			},
			ja3Hash:        expectedHash,
			expectedResult: []string{"test_chrome_a", "test_chrome_b"},
			description:    "相同 JA3 的多个 profile 应该都被返回",
		},
		{
			name: "empty hash",
			setup: func() {
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = map[string]ClientProfile{
					"test_chrome": &mockClientProfile{spec: validSpec},
				}
			},
			ja3Hash:        "",
			expectedResult: nil,
			description:    "空哈希应该返回 nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result := FindProfileByJA3(tt.ja3Hash)

			if tt.expectedResult == nil {
				if len(result) != 0 {
					t.Errorf("%s: expected nil/empty result, got %v", tt.description, result)
				}
				return
			}

			if len(result) != len(tt.expectedResult) {
				t.Errorf("%s: expected %d results, got %d: %v", tt.description, len(tt.expectedResult), len(result), result)
				return
			}

			// 验证所有期望的结果都在返回的结果中
			for _, expected := range tt.expectedResult {
				found := false
				for _, r := range result {
					if r == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: expected result %s not found in %v", tt.description, expected, result)
				}
			}
		})
	}
}

// TestFindProfileByJA3_WithRealProfiles 使用真实 profile 进行测试
func TestFindProfileByJA3_WithRealProfiles(t *testing.T) {
	// 保存原始状态
	originalMapped := MappedTLSClients
	originalIndexOnce := ja3ProfileIndexOnce
	originalIndex := ja3ProfileIndex
	defer func() {
		MappedTLSClients = originalMapped
		ja3ProfileIndexOnce = originalIndexOnce
		ja3ProfileIndex = originalIndex
	}()

	// 使用真实 profile 初始化
	InitMappedTLSClients(profiles.MappedTLSClients)
	// 重置索引
	ja3ProfileIndexOnce = sync.Once{}
	ja3ProfileIndex = nil

	// 计算一个真实 profile 的 JA3
	profile, ok := MappedTLSClients["chrome_133"]
	if !ok {
		t.Skip("chrome_133 profile not available")
	}

	result, err := ComputeJA3FromProfile(profile)
	if err != nil {
		t.Skipf("chrome_133 does not support spec export: %v", err)
	}

	// 使用计算出的 hash 查找 profile
	foundProfiles := FindProfileByJA3(result.Hash)

	// 至少应该找到 chrome_133
	foundChrome133 := false
	for _, name := range foundProfiles {
		if name == "chrome_133" {
			foundChrome133 = true
			break
		}
	}

	if !foundChrome133 {
		t.Errorf("Expected to find chrome_133 in results, got: %v", foundProfiles)
	}
}

// ============================================================================
// 边界情况和集成测试
// ============================================================================

// TestErrorWrapping 测试错误包装和解包
func TestErrorWrapping(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		isTargetErr  func(error) bool
		expectedMatch bool
	}{
		{
			name:          "ErrInvalidClientHelloSpec wraps ErrInvalidFingerprint",
			err:           ErrInvalidClientHelloSpec,
			isTargetErr:   func(e error) bool { return errors.Is(e, errdefs.ErrInvalidFingerprint) },
			expectedMatch: true,
		},
		{
			name:          "ErrProfileNotFound wraps ErrProfileNotFound",
			err:           ErrProfileNotFound,
			isTargetErr:   func(e error) bool { return errors.Is(e, errdefs.ErrProfileNotFound) },
			expectedMatch: true,
		},
		{
			name:          "ErrEmptyProfile wraps ErrInvalidFingerprint",
			err:           ErrEmptyProfile,
			isTargetErr:   func(e error) bool { return errors.Is(e, errdefs.ErrInvalidFingerprint) },
			expectedMatch: true,
		},
		{
			name:          "ErrClientHelloIDNotImplemented wraps ErrClientHelloSpecNotImplemented",
			err:           ErrClientHelloIDNotImplemented,
			isTargetErr:   func(e error) bool { return errors.Is(e, errdefs.ErrClientHelloSpecNotImplemented) },
			expectedMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.isTargetErr(tt.err)
			if got != tt.expectedMatch {
				t.Errorf("Error wrapping check failed: got %v, want %v", got, tt.expectedMatch)
			}
		})
	}
}
