package testhelpers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// GoldenFile 用于管理测试中的 golden file 记录和回放
// Golden file 是指记录真实外部依赖（如网络响应）的文件，
// 用于在单元测试中重放，避免每次测试都依赖真实网络
type GoldenFile struct {
	dir      string // golden file 存储目录
	filename string // golden file 文件名（不含目录）
	path     string // 完整路径
	update   bool   // 是否更新 golden file
}

// NewGoldenFile 创建一个新的 GoldenFile 管理器
// dir: golden file 存储目录
// filename: golden file 文件名（例如 "http_responses.json"）
// updateMode: 如果为 true，会写入新的响应；如果为 false，会读取并验证
func NewGoldenFile(dir, filename string, updateMode bool) *GoldenFile {
	return &GoldenFile{
		dir:      dir,
		filename: filename,
		path:     filepath.Join(dir, filename),
		update:   updateMode,
	}
}

// Load 从 golden file 中加载记录的数据
// 如果文件不存在且不在更新模式下，返回错误
func (gf *GoldenFile) Load(v interface{}) error {
	if gf.update {
		// 更新模式下，不加载已有数据
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

// Save 将数据保存到 golden file
// 通常在更新模式下使用
func (gf *GoldenFile) Save(v interface{}) error {
	if !gf.update {
		// 非更新模式下，不修改 golden file
		return nil
	}

	// 确保目录存在
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

// LoadFromReader 从读取器中加载 golden file 中的数据
// 用于处理流式数据或特殊格式
func (gf *GoldenFile) LoadFromReader(reader io.Reader) ([]byte, error) {
	if gf.update {
		return nil, nil
	}

	return io.ReadAll(reader)
}

// SaveToWriter 将数据写入到 golden file
// 用于处理流式数据或特殊格式
func (gf *GoldenFile) SaveToWriter(data []byte) error {
	if !gf.update {
		return nil
	}

	if err := os.MkdirAll(gf.dir, 0755); err != nil {
		return fmt.Errorf("failed to create golden file directory: %w", err)
	}

	return os.WriteFile(gf.path, data, 0644)
}

// Exists 检查 golden file 是否存在
func (gf *GoldenFile) Exists() bool {
	_, err := os.Stat(gf.path)
	return err == nil
}

// Path 返回 golden file 的完整路径
func (gf *GoldenFile) Path() string {
	return gf.path
}

// Remove 删除 golden file（用于脑洁或重置）
func (gf *GoldenFile) Remove() error {
	return os.Remove(gf.path)
}

// Size 返回 golden file 的大小（字节）
func (gf *GoldenFile) Size() (int64, error) {
	info, err := os.Stat(gf.path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
