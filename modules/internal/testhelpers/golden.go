package testhelpers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// translated comment
// translated comment
// translated comment
type GoldenFile struct {
	dir      string // translated comment
	filename string // translated comment
	path     string // translated comment
	update   bool   // translated comment
}

// translated comment
// translated comment
// translated comment
// translated comment
func NewGoldenFile(dir, filename string, updateMode bool) *GoldenFile {
	return &GoldenFile{
		dir:      dir,
		filename: filename,
		path:     filepath.Join(dir, filename),
		update:   updateMode,
	}
}

// translated comment
// translated comment
func (gf *GoldenFile) Load(v interface{}) error {
	if gf.update {
		// translated comment
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

// translated comment
// translated comment
func (gf *GoldenFile) Save(v interface{}) error {
	if !gf.update {
		// translated comment
		return nil
	}

	// translated comment
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

// translated comment
// translated comment
func (gf *GoldenFile) LoadFromReader(reader io.Reader) ([]byte, error) {
	if gf.update {
		return nil, nil
	}

	return io.ReadAll(reader)
}

// translated comment
// translated comment
func (gf *GoldenFile) SaveToWriter(data []byte) error {
	if !gf.update {
		return nil
	}

	if err := os.MkdirAll(gf.dir, 0755); err != nil {
		return fmt.Errorf("failed to create golden file directory: %w", err)
	}

	return os.WriteFile(gf.path, data, 0644)
}

// translated comment
func (gf *GoldenFile) Exists() bool {
	_, err := os.Stat(gf.path)
	return err == nil
}

// translated comment
func (gf *GoldenFile) Path() string {
	return gf.path
}

// translated comment
func (gf *GoldenFile) Remove() error {
	return os.Remove(gf.path)
}

// translated comment
func (gf *GoldenFile) Size() (int64, error) {
	info, err := os.Stat(gf.path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
