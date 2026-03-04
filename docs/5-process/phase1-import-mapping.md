# Phase 1 TLS 层内化 - Import 变更清单

**执行日期**: 2026-03-10 ~ 2026-03-17  
**影响范围**: ~12 个文件，~50+ 个 import 语句  

---

## 📋 import 路径映射表

### 类型 1: internal/tlsutil -> tls/internal/utils

| 源文件 | 旧 import | 新 import |
| -------- | ---------- | ---------- |
| tls/ja3/ja3.go | `"github.com/vistone/fingerprint/internal/tlsutil"` | `"github.com/vistone/fingerprint/tls/internal/utils"` |
| tls/ja3/errors.go | 同上 | 同上 |
| tls/ja4/ja4.go | 同上 | 同上 |
| tls/ja4/errors.go | 同上 | 同上 |
| tls/ja4s/ja4s.go | 同上 | 同上 |
| tls/ja4s/errors.go | 同上 | 同上 |
| tls/tls.go | 同上 | 同上 |
| internal/extension/*.go | 同上 | 同上 |
| test/*.go | 同上 | 同上 |
| config/bridge.go | (如使用) | 同上 |

### 类型 2: tls/utils -> tls/internal/utils

**注**: 如果有相对 import `../utils` 则改为 `../internal/utils`

| 源文件 | 旧 import | 新 import |
| -------- | ---------- | ---------- |
| tls/*/file.go | 相对 import `../utils` | 相对 import `../internal/utils` |

### 类型 3: tls/ech -> tls/internal/ech

| 源文件 | 旧 import | 新 import |
| -------- | ---------- | ---------- |
| tls/tls.go | `"github.com/vistone/fingerprint/tls/ech"` | `"github.com/vistone/fingerprint/tls/internal/ech"` |
| tls/types.go | 同上 | 同上 |

---

## 🔍 关键文件详细检查

### tls/ja3/ja3.go

```go
// 旧代码
import (
    "..."
    "github.com/vistone/fingerprint/internal/tlsutil"
    "..."
)

// 新代码
import (
    "..."
    "github.com/vistone/fingerprint/tls/internal/utils"
    "..."
)

// 使用方式不变
// tlsutil.GenerateGREASE() -> utils.GenerateGREASE()
```plaintext

### tls/ja3/errors.go

```go
// 同上（如果有 import）
```plaintext

### tls/ja4/ja4.go, ja4s/ja4s.go

```go
// 同上
```plaintext

### tls/tls.go

```go
// 旧代码
import (
    "github.com/vistone/fingerprint/internal/tlsutil"
    "github.com/vistone/fingerprint/tls/ech"  // ECH 导入
)

// 新代码
import (
    "github.com/vistone/fingerprint/tls/internal/utils"
    "github.com/vistone/fingerprint/tls/internal/ech"  // 新位置
)
```plaintext

### internal/extension/canary_framework.go 及其他

```go
// 如果使用 tlsutil
// 旧
import "github.com/vistone/fingerprint/internal/tlsutil"

// 新
import "github.com/vistone/fingerprint/tls/internal/utils"
```plaintext

### test/ 下的文件

测试文件中对 tlsutil 和 ech 的导入都需要更新：

```go
// test/fingerprint_test.go, test/integration_test.go 等
// 旧
import "github.com/vistone/fingerprint/internal/tlsutil"

// 新
import "github.com/vistone/fingerprint/tls/internal/utils"
```plaintext

---

## ⚠️ 特殊情况处理

### 相对 import（如果存在）

某些文件可能使用 **相对 import** 而非绝对 import：

```go
// tls/ja3/ja3.go 中
// 旧（相对）
import "../utils"

// 新（相对）  
import "../internal/utils"
```plaintext

**检查方式**:
```bash
grep -E "^\s*import\s+\"\.\./.*\"" tls/**/*.go
```plaintext

---

## 🔐 验证清单

完成 import 变更后，逐一验证：

### 1. 语法检查

```bash
go build ./tls/...
```plaintext

期望: 全部编译成功

### 2. 单包测试

```bash
go test ./tls/ja3 -v
go test ./tls/ja4 -v
go test ./tls/ja4s -v
```plaintext

期望: 所有测试通过

### 3. 集成测试

```bash
go test ./test/... -v -short
```plaintext

期望: 集成测试通过（特别是涉及 TLS 的部分）

### 4. 整体构建

```bash
go build ./...
```plaintext

期望: 整个项目成功编译

### 5. Import 残留检查

```bash
# 确保没有旧 import 遗留
grep -r "internal/tlsutil" --include="*.go" .

# 不应该有输出（除了本文档）
```plaintext

---

## 📊 文件影响矩阵

```plaintext
文件分类        | 文件数 | import 行数 | 风险等级
─────────────────────────────────────────────────
TLS 主包        | 3     | 5          | 低
TLS ja3/ja4/ja4s| 6     | 12         | 低
配置和扩展      | 2     | 3          | 中
测试文件        | 3+    | 8+         | 低
─────────────────────────────────────────────────
合计            | ~15   | ~30+       | 低
```plaintext

---

## 🎬 执行步骤

### 自动化方式（推荐）

```bash
# 1. 模拟运行（查看变更）
bash scripts/phase1_tls_migration.sh dry-run

# 2. 确认无误后，执行真实迁移
bash scripts/phase1_tls_migration.sh execute

# 3. 验证结果
go mod tidy
go build ./...
go test ./...
```plaintext

### 手工方式

若自动化脚本有问题，可手工执行：

```bash
# Step 1: 创建目录
mkdir -p tls/internal/utils
mkdir -p tls/internal/ech

# Step 2: 移动文件
cp -r tls/utils/*.go tls/internal/utils/
cp -r internal/tlsutil/*.go tls/internal/utils/
cp -r tls/ech/*.go tls/internal/ech/

# Step 3: 批量替换（使用编辑器或 sed）
find . -name "*.go" -type f \
  -exec sed -i 's|"github.com/vistone/fingerprint/internal/tlsutil"|"github.com/vistone/fingerprint/tls/internal/utils"|g' {} +

find . -name "*.go" -type f \
  -exec sed -i 's|"github.com/vistone/fingerprint/tls/ech"|"github.com/vistone/fingerprint/tls/internal/ech"|g' {} +

# Step 4: 删除旧目录
rm -rf tls/utils tls/ech internal/tlsutil

# Step 5: 验证
go mod tidy
go build ./...
go test ./...
```plaintext

---

## 📝 示例: 完整文件替换

### tls/ja3/ja3.go （示例）

**变更前**:
```go
package ja3

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/vistone/fingerprint/internal/tlsutil"    // ← 旧路径
	"github.com/vistone/fingerprint/types"
)

func (j *JA3Generator) Generate(clientHello []byte) (*types.Fingerprint, error) {
	// ... 代码 ...
	
	grease := tlsutil.GenerateGREASE()  // ← 使用不变
	
	// ... 代码 ...
}
```plaintext

**变更后**:
```go
package ja3

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/vistone/fingerprint/tls/internal/utils"   // ← 新路径
	"github.com/vistone/fingerprint/types"
)

func (j *JA3Generator) Generate(clientHello []byte) (*types.Fingerprint, error) {
	// ... 代码 ...
	
	grease := utils.GenerateGREASE()  // ← 函数调用不变
	
	// ... 代码 ...
}
```plaintext

---

## 回滚指令

若出现问题，快速回滚：

```bash
# 方式 1: 使用 git tag
git reset --hard phase1-backup-YYYYMMDD_HHMMSS

# 方式 2: 使用 git reflog 找到分支点
git reflog
git reset --hard <commit-sha>

# 方式 3: 完全撤销（危险操作）
git clean -fd
git reset --hard HEAD
```plaintext

---

## ✅ 完成标志

✓ 所有文件都从 `internal/tlsutil` 改为 `tls/internal/utils`  
✓ 所有文件都从 `tls/ech` 改为 `tls/internal/ech`  
✓ `go build ./...` 成功  
✓ `go test ./tls/...` 通过  
✓ `go test ./test/...` 通过  
✓ 无 import 残留  

---

## 相关文档

- 完整重构计划: [package-restructuring-plan.md](package-restructuring-plan.md)
- Phase 2 HTTP 层: 见完整计划
- Phase 3 pkg API: 见完整计划
