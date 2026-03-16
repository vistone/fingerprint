package testhelpers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// GoldenFile manages golden-file recording and replay in tests
// A golden file records external dependency outputs (for example, network responses)
// It is replayed in unit tests to avoid dependency on live network calls
type GoldenFile struct {
	dir      string // golden file storage directory
	filename string // golden file name (without directory)
	path     string // full path
	update   bool   // whether to update the golden file
}

// NewGoldenFile creates a GoldenFile manager
// dir: golden file storage directory
// filename: golden file name (for example, "http_responses.json")
// updateMode: true writes new responses; false reads and verifies
func NewGoldenFile(dir, filename string, updateMode bool) *GoldenFile {
	return &GoldenFile{
		dir:      dir,
		filename: filename,
		path:     filepath.Join(dir, filename),
		update:   updateMode,
	}
}

// Load reads recorded data from golden file
// Returns an error when file is missing and update mode is disabled
func (gf *GoldenFile) Load(v interface{}) error {
	if gf.update {
		// Skip loading existing data in update mode
		return nil
	}

	data, err := os.ReadFile(gf.path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("golden file not found: %s (run with -update flag to create it)", gf.path)
		}
		return fmt.Errorf("failed to read golden file: %w", err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal golden file: %w", err)
	}

	return nil
}

// Save persists data to golden file
// Typically used in update mode
func (gf *GoldenFile) Save(v interface{}) error {
	if !gf.update {
		// Do not modify golden file outside update mode
		return nil
	}

	// Ensure target directory exists
	if err := os.MkdirAll(gf.dir, 0755); err != nil {
		return fmt.Errorf("failed to create golden file directory: %w", err)
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := os.WriteFile(gf.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write golden file: %w", err)
	}

	return nil
}

// LoadFromReader loads golden data from reader
// Useful for streaming data or special formats
func (gf *GoldenFile) LoadFromReader(reader io.Reader) ([]byte, error) {
	if gf.update {
		return nil, nil
	}

	return io.ReadAll(reader)
}

// SaveToWriter writes data into golden file
// Useful for streaming data or special formats
func (gf *GoldenFile) SaveToWriter(data []byte) error {
	if !gf.update {
		return nil
	}

	if err := os.MkdirAll(gf.dir, 0755); err != nil {
		return fmt.Errorf("failed to create golden file directory: %w", err)
	}

	return os.WriteFile(gf.path, data, 0644)
}

// Exists reports whether the golden file exists
func (gf *GoldenFile) Exists() bool {
	_, err := os.Stat(gf.path)
	return err == nil
}

// Path returns the full golden file path
func (gf *GoldenFile) Path() string {
	return gf.path
}

// Remove deletes the golden file (for cleanup or reset)
func (gf *GoldenFile) Remove() error {
	return os.Remove(gf.path)
}

// Size returns the golden file size in bytes
func (gf *GoldenFile) Size() (int64, error) {
	info, err := os.Stat(gf.path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
