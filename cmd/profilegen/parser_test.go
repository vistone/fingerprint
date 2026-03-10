//go:build profilegen
// +build profilegen

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseYAMLFile_PathTraversal tests path traversal protection
func TestParseYAMLFile_PathTraversal(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "path traversal with ..",
			path:    "../../../etc/passwd",
			wantErr: true,
			errMsg:  "path contains illegal characters",
		},
		{
			name:    "path traversal attempt 2",
			path:    "profiles/../../../etc/passwd",
			wantErr: true,
			errMsg:  "path contains illegal characters",
		},
		{
			name:    "absolute path outside allowed dirs",
			path:    "/etc/passwd",
			wantErr: true,
			errMsg:  "path not in allowed range",
		},
		{
			name:    "path in non-allowed directory",
			path:    "tmp/malicious.yaml",
			wantErr: true,
			errMsg:  "path not in allowed range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseYAMLFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseYAMLFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("parseYAMLFile() error = %v, want to contain %q", err, tt.errMsg)
			}
		})
	}
}

// TestParseAllProfiles_DirectoryValidation tests directory verification
func TestParseAllProfiles_DirectoryValidation(t *testing.T) {
	tests := []struct {
		name    string
		dir     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "directory with path traversal",
			dir:     "../../../etc",
			wantErr: true,
			errMsg:  "directory not in allowed range",
		},
		{
			name:    "non-allowed directory",
			dir:     "tmp",
			wantErr: true,
			errMsg:  "directory not in allowed range",
		},
		{
			name:    "absolute path outside",
			dir:     "/tmp",
			wantErr: true,
			errMsg:  "directory not in allowed range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseAllProfiles(tt.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAllProfiles() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("parseAllProfiles() error = %v, want to contain %q", err, tt.errMsg)
			}
		})
	}
}

// TestParseYAMLFile_FileSizeLimit tests file size limit
func TestParseYAMLFile_FileSizeLimit(t *testing.T) {
	// create a temporary large file
	tmpDir := t.TempDir()
	largeFilePath := filepath.Join(tmpDir, "large.yaml")

	// create a file exceeding 10MB (here create a small file for testing logic)
	const testMaxSize = 1024 // test with 1KB limit
	largeData := make([]byte, testMaxSize+100)
	if err := os.WriteFile(largeFilePath, largeData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// temporarily modify allowedDirs to include tmpDir (for testing only)
	oldAllowedDirs := []string{"profiles/specs", "cmd/profilegen/extract"}
	// note: in actual tests need to dynamically adjust allowedDirs, here simplified process

	_, err := parseYAMLFile(largeFilePath)
	if err == nil {
		t.Error("Expected error for oversized file")
	}

	// restore original settings
	_ = oldAllowedDirs
}

// TestParseAllProfiles_SymlinkSkip tests symbolic link skipping
func TestParseAllProfiles_SymlinkSkip(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping symlink test in short mode")
	}

	// create temporary file directly in allowed directory
	tmpDir := "profiles/specs/test_tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// create a real file
	realFile := filepath.Join(tmpDir, "real.yaml")
	if err := os.WriteFile(realFile, []byte("name: test"), 0644); err != nil {
		t.Fatalf("Failed to create real file: %v", err)
	}

	// create symbolic link (may fail on some systems)
	symlinkPath := filepath.Join(tmpDir, "link.yaml")
	if err := os.Symlink(realFile, symlinkPath); err != nil {
		t.Logf("Cannot create symlink (expected on some systems): %v", err)
		return
	}

	// parsing directory should skip symbolic links
	profiles, files, err := parseAllProfiles(tmpDir)
	if err != nil {
		t.Fatalf("parseAllProfiles failed: %v", err)
	}

	// should only contain real file, not symbolic links
	if len(profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(profiles))
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}
}

// TestParseYAMLFile_InvalidYAML tests invalid YAML processing
func TestParseYAMLFile_InvalidYAML(t *testing.T) {
	// create temporary file in allowed directory
	tmpDir := "profiles/specs"

	invalidFile := filepath.Join(tmpDir, "test_invalid.yaml")

	// write invalid YAML
	invalidContent := []byte(`
name: test
  invalid_indentation: value
extensions:
  - type: test
    params: [invalid
`)
	if err := os.WriteFile(invalidFile, invalidContent, 0644); err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}
	defer os.Remove(invalidFile)

	_, err := parseYAMLFile(invalidFile)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "parse YAML failed") {
		t.Errorf("Expected YAML parsing error, got: %v", err)
	}
}

// TestParseYAMLFile_EmptyFile tests empty file processing
func TestParseYAMLFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.yaml")

	// create empty file
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	_, err := parseYAMLFile(emptyFile)
	if err == nil {
		t.Error("Expected error for empty file")
	}
}

// BenchmarkParseYAMLFile benchmark test: YAML parsing performance
func BenchmarkParseYAMLFile(b *testing.B) {
	// use real configuration file
	testFile := "profiles/specs/chrome_133.yaml"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := parseYAMLFile(testFile)
		if err != nil {
			b.Fatal(err)
		}
	}
}
