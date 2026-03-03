# 包结构重构计划 (Package Restructuring Plan)

**执行时间**：Week 5-8 (2026-03-03 ~ 2026-03-24)  
**风险等级**：⚠️ 高风险（涉及全局 import 路径变更）  
**回滚难度**：困难（需要逐一确认测试）  

---

## 📊 重构概览

### 目标结构

```
current:                       target:
├── tls/                       ├── tls/
│   ├── ja3/                   │   ├── ja3/
│   ├── ja4/                   │   ├── ja4/
│   ├── ja4s/                  │   ├── ja4s/
│   ├── utils/        💾       │   └── internal/          # TLS 内部工具 + utils合并
│   ├── ech/                   │       ├── utils/         
│   ├── tls.go                 │       ├── ech/
│   └── types.go               │       ├── tls.go
│                              │       └── types.go
├── http/                      ├── http/
│   ├── headers/               │   ├── headers/
│   ├── http2/                 │   ├── http2/
│   ├── clienthints/           │   ├── clienthints/
│   ├── ja4h/                  │   ├── ja4h/
│   ├── useragent/             │   ├── policy/
│   ├── policy/                │   ├── errors.go
│   ├── errors.go              │   ├── types.go
│   └── types.go               │   ├── http.go
│                              │   └── internal/          # HTTP 内部工具
├── internal/                  │       ├── utils/
│   ├── utils/        🔄       │       ├── builder/
│   ├── errors/       🔄       │       ├── caching/
│   ├── tlsutil/      🔄       │       └── policy/
│   ├── config/                ├── internal/
│   ├── extension/             │   ├── utils/            # 全局工具（精简）
│   ├── metrics/               │   ├── errors/           # 错误体系（保留）
│   ├── monitor/               │   ├── tlsutil/          # TLS -> tls/internal/utils
│   └── ...                    │   ├── httputil/         # HTTP -> http/internal/utils
│                              │   ├── cache/            # 统一缓存实现
│                              │   ├── config/
│                              │   ├── extension/
│                              │   └── ...
│                              ├── pkg/                  # 新增：公共 API 暴露
│                              │   ├── fingerprint/
│                              │   ├── profiling/
│                              │   └── telemetry/
└── profiles/                  └── profiles/
```

---

## 🔄 三阶段迁移策略

### **Phase 1: TLS 层内化** (Week 5-6)

#### 1.1 目标
- 创建 `tls/internal/` 子目录
- 将 TLS 私有工具从 `internal/tlsutil` 移动到 `tls/internal/utils`
- 整合 `tls/utils/` 到 `tls/internal/`
- 更新所有 TLS 相关 imports

#### 1.2 具体步骤

**Step 1: 目录创建**
```bash
mkdir -p tls/internal/utils
mkdir -p tls/internal/ech
```

**Step 2: 文件迁移**
```bash
# 迁移 TLS 特定工具
cp -r tls/utils/* tls/internal/utils/           # 合并现有 utils
cp internal/tlsutil/* tls/internal/utils/       # GREASE, ext handling
cp tls/ech/* tls/internal/ech/                  # ECH 相关

# 重新组织
rm -rf tls/utils
rm -rf tls/ech
# ech 内容移入 tls/internal/ech
```

**Step 3: Import 更新清单**

| 文件 | 旧路径 | 新路径 |
|------|--------|--------|
| tls/ja3/*.go | `internal/tlsutil` → | `../internal/utils` |
| tls/ja4/*.go | `internal/tlsutil` → | `../internal/utils` |
| tls/ja4s/*.go | `internal/tlsutil` → | `../internal/utils` |
| tls/*.go | `internal/tlsutil` → | `./internal/utils` |
| test/*.go | `"github.com/vistone/fingerprint/internal/tlsutil"` → | `"github.com/vistone/fingerprint/tls/internal/utils"` |

**Step 4: 代码搜索和替换**
```bash
# 搜索所有 import tlsutil 的文件
grep -r "internal/tlsutil" --include="*.go" .

# 批量替换（需要使用 sed 或手工确认）
# 例如：tls/ja3/ja3.go
#   import "github.com/vistone/fingerprint/internal/tlsutil"
# 改为：
#   import "github.com/vistone/fingerprint/tls/internal/utils"
```

**Step 5: 验证**
```bash
go mod tidy
go build ./...
go test ./tls/...
```

#### 1.3 关键影响文件
- tls/ja3/ja3.go, errors.go
- tls/ja4/ja4.go, errors.go
- tls/ja4s/ja4s.go, errors.go
- tls/tls.go, types.go
- internal/tlsutil/grease.go, converter.go
- internal/extension/ (可能有 TLS 工具调用)
- test/ (测试文件中的 imports)

#### 1.4 预期风险
- ✅ **低风险**：TLS 包相对独立，内聚性强
- ⚠️ **风险**: test/ 目录中对 tlsutil 的引用

#### 1.5 测试清单
- [ ] `go test ./tls/...` 全部通过
- [ ] `go test ./test/...` 全部通过
- [ ] `go build -o fingerprint .` 成功

---

### **Phase 2: HTTP 层内化** (Week 6-7)

#### 2.1 目标
- 创建 `http/internal/` 子目录
- 隔离 HTTP 私有工具（headers builder、useragent generator）
- 整合 HTTP 相关缓存逻辑
- 复用 Phase 1 的经验

#### 2.2 具体步骤

**Step 1: 目录创建**
```bash
mkdir -p http/internal/headers
mkdir -p http/internal/useragent
mkdir -p http/internal/caching
mkdir -p http/internal/policy
```

**Step 2: 文件迁移**
```bash
# 迁移 HTTP 的私有实现（通常包含 helpers、builders）
# headers/builder.go 的内部帮助函数等
mv http/headers/*.go http/internal/headers/
mv http/useragent/*.go http/internal/useragent/
# 保留公共暴露的接口在 http/ 直接

# 注意：需要保留 public types 在 http/types.go
```

**Step 3: 创建内部公共接口**

原则：`http/` 保留公共接口，`http/internal/` 放内部实现

```go
// http/headers.go (public facade)
package http

import "github.com/vistone/fingerprint/http/internal/headers"

// 暴露的接口保留原样
type HeaderBuilder = headers.HeaderBuilder

// http/useragent.go (public facade)
package http

import "github.com/vistone/fingerprint/http/internal/useragent"

type UAGenerator = useragent.UAGenerator
```

**Step 4: 更新 imports 清单**

| 包 | 旧导入 | 新导入 |
|-----|--------|---------|
| http/headers/ | 无跨包依赖 | `../internal/headers` |
| http/useragent/ | 无跨包依赖 | `../internal/useragent` |
| http/clienthints/ | `"github.com/vistone/fingerprint/http/..."`  | 保持 + `../internal/useragent` |
| internal/extension/ | `"github.com/vistone/fingerprint/http"` | 保持外部接口 |

**Step 5: 验证**
```bash
go mod tidy
go build ./...
go test ./http/...
```

#### 2.3 关键影响文件
- http/headers/builder.go, headers.go, interfaces.go
- http/useragent/generator.go, useragent.go, interfaces.go
- http/clienthints/ (可能依赖 useragent)
- http/http.go, types.go
- internal/extension/ (HTTP 适配器)
- test/ (HTTP 集成测试)

#### 2.4 预期风险
- ⚠️ **中等风险**：HTTP 包中有循环依赖风险（clienthints ← → useragent）
- ⚠️ **风险**: 需要确保 public vs internal 的边界清晰

#### 2.5 测试清单
- [ ] `go test ./http/...` 全部通过（特别是 clienthints）
- [ ] `go test ./test/...` 中的 HTTP 测试通过
- [ ] `go build ./...` 成功
- [ ] 检查是否有对 `http/internal` 的外部导入

---

### **Phase 3: 共享工具整合 + pkg 接口** (Week 7-8)

#### 3.1 目标
- 优化 `internal/utils/`（只保留真正全局的工具）
- 新建 `internal/cache/`（统一缓存实现）
- 新建 `internal/httputil/`（HTTP 层标准工具，独立于 http/internal/）
- 新建 `pkg/` 暴露公共 API，隐藏内部细节

#### 3.2 具体步骤

**Step 1: internal/utils 精简**

当前的 `internal/utils` 包含：
- `useragent.go` - 应该迁移到 `http/internal/useragent/`
- `strings.go` - 保留（全局使用）
- `rand.go` - 保留（全局使用）

```bash
# 操作
rm internal/utils/useragent.go     # 已迁移到 http/internal/useragent
# strings.go, rand.go 保留
```

**Step 2: internal/cache 新建**

当前缓存逻辑分散在各处，整合到统一接口：

```bash
mkdir -p internal/cache

# cache/interface.go (统一接口)
# cache/memory.go (内存实现)
# cache/ttl.go (TTL 支持)
```

```go
// internal/cache/cache.go
package cache

type Cache interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{})
    Delete(key string)
    Clear()
}

type TTLCache interface {
    Cache
    SetWithTTL(key string, value interface{}, ttl time.Duration)
}

// 单例
var (
    profileCache TTLCache
    headerCache TTLCache
)
```

**Step 3: internal/httputil 新建**

HTTP 层宏观工具（不属于某个具体 HTTP 子包）：

```bash
mkdir -p internal/httputil

# httputil/normalization.go (header 规范化等)
# httputil/validation.go (HTTP 参数验证)
# httputil/constants.go (HTTP 常量)
```

**Step 4: pkg 公共 API 暴露**

新建 pkg 目录，定义稳定的公共接口：

```bash
mkdir -p pkg/fingerprint
mkdir -p pkg/profiling  
mkdir -p pkg/telemetry

# pkg/fingerprint/fingerprint.go (主 API)
# pkg/profiling/generator.go
# pkg/telemetry/metrics.go
```

关键原则：
- pkg/ 中的接口不依赖 internal/
- internal/ 变更不影响 pkg/ 使用者
- pkg/ 可以被外部项目安全导入

```go
// pkg/fingerprint/fingerprint.go
package fingerprint

import (
    "github.com/vistone/fingerprint/types"
)

// 公共 API
type Fingerprinter interface {
    Generate(ctx context.Context) (*types.Fingerprint, error)
    GenerateWithProfile(ctx context.Context, profile string) (*types.Fingerprint, error)
}

// 获取默认实现
func NewFingerprinter() Fingerprinter {
    // 返回内部实现（对用户透明）
    return newDefaultFingerprinter()
}
```

#### 3.3 import 变更映射

| 模块 | 旧导入 | 新导入 |
|------|--------|--------|
| internal/extension | `"github.com/vistone/.../internal/utils"` | 按需保留或改用 cache/httputil |
| http/ | (无) | `"github.com/vistone/.../internal/cache"` (可选) |
| tls/ | (无) | (无) |
| 外部用户 | 多种临时接口 | `"github.com/vistone/.../pkg/fingerprint"` |

#### 3.4 关键影响文件
- internal/utils/useragent.go (删除)
- internal/ 所有子包 (可能需要调整 import)
- extension/ (可能重构）
- 所有对 internal/ 的外部导入

#### 3.5 预期风险
- ⚠️ **高风险**：涉及最广泛的 import 变更
- 🔴 **严重**：pkg/ 定义不清可能导致 API 不稳定

#### 3.6 测试清单
- [ ] `go test ./...` 全部通过
- [ ] `go test ./internal/cache/...` (新包)
- [ ] `go test ./pkg/...` (新包)
- [ ] `go build ./...`
- [ ] `go mod verify` (模块完整性)
- [ ] 写一个示例，仅导入 pkg/ 验证可用性

---

## 🛠 自动化工具和脚本

### 脚本 1: import 分析工具

```bash
# scripts/analyze_imports.sh
#!/bin/bash

echo "=== Analyzing imports by package ==="

# 统计各包的导入源
for pkg in tls http internal; do
    echo ""
    echo "--- Package: $pkg ---"
    grep -r "import" $pkg --include="*.go" | grep -oE '".*"' | sort | uniq -c | sort -rn | head -10
done

# 检测循环依赖
echo ""
echo "=== Checking circular dependencies ==="
go mod graph | grep -E "internal|http|tls" | sort | uniq
```

### 脚本 2: 批量替换工具

```bash
# scripts/migrate_imports.sh
#!/bin/bash

PHASE=${1:-1}

case $PHASE in
    1)
        echo "Phase 1: Migrating tlsutil -> tls/internal/utils"
        # 遍历所有 .go 文件
        find . -name "*.go" -type f | while read file; do
            # 替换 import "github.com/vistone/fingerprint/internal/tlsutil"
            # 为     import "github.com/vistone/fingerprint/tls/internal/utils"
            sed -i 's|"github.com/vistone/fingerprint/internal/tlsutil"|"github.com/vistone/fingerprint/tls/internal/utils"|g' "$file"
        done
        ;;
    2)
        echo "Phase 2: Migrating HTTP internals"
        find . -name "*.go" -type f | while read file; do
            # HTTP 层交叉引用
            sed -i 's|github.com/vistone/fingerprint/http/headers|github.com/vistone/fingerprint/http/internal/headers|g' "$file"
            sed -i 's|github.com/vistone/fingerprint/http/useragent|github.com/vistone/fingerprint/http/internal/useragent|g' "$file"
        done
        ;;
    3)
        echo "Phase 3: Finalizing pkg API"
        # 指导性脚本，不自动替换（需要人工审核）
        echo "Manual review needed for pkg/ migration"
        ;;
esac

echo "Running go mod tidy..."
go mod tidy

echo "Verifying builds..."
go build ./...
```

### 脚本 3: 验证脚本

```bash
# scripts/verify_restructuring.sh
#!/bin/bash

PHASE=${1:-1}

echo "=== Phase $PHASE Verification ==="

# 基础检查
echo ""
echo "1. Build check..."
if go build ./...; then
    echo "✅ Build succeeded"
else
    echo "❌ Build failed"
    exit 1
fi

# 测试检查
echo ""
echo "2. Test check..."
case $PHASE in
    1)
        if go test ./tls/...; then
            echo "✅ TLS tests passed"
        else
            echo "❌ TLS tests failed"
            exit 1
        fi
        ;;
    2)
        if go test ./http/...; then
            echo "✅ HTTP tests passed"
        else
            echo "❌ HTTP tests failed"
            exit 1
        fi
        ;;
    3)
        if go test ./...; then
            echo "✅ All tests passed"
        else
            echo "❌ Some tests failed"
            exit 1
        fi
        ;;
esac

# 循环依赖检查
echo ""
echo "3. Circular dependency check..."
if go mod graph | grep -E "internal|http|tls" | head -1 > /dev/null; then
    echo "⚠️ Possible circular dependencies detected"
else
    echo "✅ No obvious circular dependencies"  
fi

# Import 清洁度检查
echo ""
echo "4. Import cleanliness check..."
for file in $(find . -name "*.go" -type f); do
    # 检查是否还有旧的导入路径
    if grep -q "internal/tlsutil" "$file"; then
        echo "❌ Found old import in $file: internal/tlsutil"
    fi
done
echo "✅ Import check complete"
```

---

## 📋 回滚计划

### 预防措施
1. **每个 Phase 前备份**
   ```bash
   git tag -a phase-1-backup -m "Backup before Phase 1 restructuring"
   git push origin phase-1-backup
   ```

2. **分支工作流**
   ```bash
   git checkout -b restructure/phase1
   # 进行所有变更
   git push origin restructure/phase1
   # PR 审查后 merge
   ```

### 快速回滚
```bash
# 如果构建失败
git reset --hard phase-1-backup

# 恢复特定文件
git checkout phase-1-backup -- tls/
```

---

## 📅 时间线

| 周次 | Phase | 关键里程碑 | 验证 |
|------|-------|----------|------|
| Week 5-6 | Phase 1 TLS | tls/internal/ 完毕 | go test ./tls/... ✅ |
| Week 6-7 | Phase 2 HTTP | http/internal/ 完毕 | go test ./http/... ✅ |
| Week 7-8 | Phase 3 pkg | pkg/ API 暴露 | 外部导入示例 ✅ |

---

## ✅ 最终检查清单

- [ ] 所有 Phase 的 go test 通过
- [ ] 所有 Phase 的 go build 成功
- [ ] go mod tidy 整理完毕
- [ ] 无未引用的旧目录残留
- [ ] 文档更新（README, contributing guide）
- [ ] 示例代码更新（使用新的 pkg/ API）
- [ ] CI/CD 流程验证（如有）

---

## 额外说明

### 为什么这个结构更好？

1. **模块化**：TLS, HTTP 各自独立，修改隔离
2. **清晰的边界**：internal/ 不再是"垃圾桶"
3. **性能**：编译时依赖更清晰，可独立编译子项目
4. **可维护性**：新开发者易于理解代码组织
5. **API 稳定性**：pkg/ 的公共 API 与 internal 变更解耦

### 与灰度推出的协调

- Phase 1 (TLS): Week 5-6 与灰度 Day 1-4 并行（无冲突）
- Phase 2 (HTTP): Week 6-7 与灰度 Day 5-7 并行（无冲突）
- Phase 3 (pkg): Week 7-8 在灰度推出完成后开始

---

**下一步**：确认是否立即开始 Phase 1？
