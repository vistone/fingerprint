#!/bin/bash

# Phase 1 TLS 层内化自动迁移脚本
# 执行: bash scripts/phase1_tls_migration.sh [dry-run|execute]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
DRY_RUN=${1:-dry-run}

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# ============================================================================
# Phase 1: TLS 层内化
# ============================================================================

log_info "=== Phase 1: TLS 层内化迁移 ==="
log_info "模式: $DRY_RUN"
echo ""

# 步骤 1: 备份和版本控制
log_info "Step 1: 创建备份标签"
if [ "$DRY_RUN" == "execute" ]; then
    git tag -a "phase1-backup-$(date +%Y%m%d_%H%M%S)" -m "Backup before Phase 1 TLS restructuring" || log_warn "标签创建失败（可能已存在）"
    git stash || log_warn "stash 失败（工作区干净）"
    log_success "备份完成"
else
    log_info "[模拟] git tag -a phase1-backup-YYYYMMDD_HHMMSS ..."
fi

echo ""

# 步骤 2: 目录创建
log_info "Step 2: 创建目录结构"
DIRS_TO_CREATE=(
    "tls/internal"
    "tls/internal/utils"
    "tls/internal/ech"
)

for dir in "${DIRS_TO_CREATE[@]}"; do
    if [ "$DRY_RUN" == "execute" ]; then
        mkdir -p "$REPO_ROOT/$dir"
        log_success "创建: $dir"
    else
        log_info "[模拟] mkdir -p $dir"
    fi
done

echo ""

# 步骤 3: 文件迁移预检
log_info "Step 3: 分析文件迁移清单"

# 找出所有需要迁移的文件
declare -A FILE_MIGRATIONS

# tls/utils -> tls/internal/utils
if [ -d "$REPO_ROOT/tls/utils" ]; then
    for file in "$REPO_ROOT"/tls/utils/*.go; do
        filename=$(basename "$file")
        FILE_MIGRATIONS["$file"]="tls/internal/utils/$filename"
        log_info "  tls/utils/$filename → tls/internal/utils/$filename"
    done
fi

# internal/tlsutil -> tls/internal/utils
if [ -d "$REPO_ROOT/internal/tlsutil" ]; then
    for file in "$REPO_ROOT"/internal/tlsutil/*.go; do
        filename=$(basename "$file")
        FILE_MIGRATIONS["$file"]="tls/internal/utils/$filename"
        log_info "  internal/tlsutil/$filename → tls/internal/utils/$filename"
    done
fi

# tls/ech -> tls/internal/ech
if [ -d "$REPO_ROOT/tls/ech" ]; then
    for file in "$REPO_ROOT"/tls/ech/*.go; do
        filename=$(basename "$file")
        FILE_MIGRATIONS["$file"]="tls/internal/ech/$filename"
        log_info "  tls/ech/$filename → tls/internal/ech/$filename"
    done
fi

echo ""

# 步骤 4: 执行文件迁移
log_info "Step 4: 执行文件迁移"

if [ "$DRY_RUN" == "execute" ]; then
    for src in "${!FILE_MIGRATIONS[@]}"; do
        dst="${FILE_MIGRATIONS[$src]}"
        cp -v "$src" "$REPO_ROOT/$dst"
        log_success "迁移: $(basename $src)"
    done
    log_success "文件迁移完成"
else
    for src in "${!FILE_MIGRATIONS[@]}"; do
        dst="${FILE_MIGRATIONS[$src]}"
        log_info "[模拟] cp $src → $dst"
    done
fi

echo ""

# 步骤 5: Import 路径分析
log_info "Step 5: 分析 Import 变更清单"
echo ""
echo "需要更新的 import 路径:"
echo "────────────────────────────────────────────────"

# 检查所有 go 文件
IMPORT_FILES=$(grep -r "internal/tlsutil\|\".*tls/utils" --include="*.go" "$REPO_ROOT" | cut -d: -f1 | sort | uniq)

echo "$IMPORT_FILES" | while read file; do
    if [ -n "$file" ] && [ -f "$file" ]; then
        rel_path="${file#$REPO_ROOT/}"
        grep_result=$(grep "internal/tlsutil\|\".*tls/utils" "$file" || true)
        if [ -n "$grep_result" ]; then
            echo ""
            echo "文件: $rel_path"
            echo "$grep_result" | while read line; do
                echo "  旧: $(echo $line | grep -oE '\"[^\"]*\"')"
            done
        fi
    fi
done

echo ""
echo "────────────────────────────────────────────────"

echo ""

# 步骤 6: 批量替换 Import
log_info "Step 6: 执行 Import 路径替换"

if [ "$DRY_RUN" == "execute" ]; then
    # 替换 1: internal/tlsutil -> tls/internal/utils
    find "$REPO_ROOT" -name "*.go" -type f | while read file; do
        if grep -q "internal/tlsutil" "$file"; then
            sed -i 's|"github.com/vistone/fingerprint/internal/tlsutil"|"github.com/vistone/fingerprint/tls/internal/utils"|g' "$file"
            log_success "更新: $(basename $file)"
        fi
    done
    
    # 替换 2: tls/utils -> tls/internal/utils（绝对路径）
    find "$REPO_ROOT" -name "*.go" -type f | while read file; do
        if grep -q '"github.com/vistone/fingerprint/tls/utils"' "$file"; then
            sed -i 's|"github.com/vistone/fingerprint/tls/utils"|"github.com/vistone/fingerprint/tls/internal/utils"|g' "$file"
            log_success "更新: $(basename $file)"
        fi
    done
    
    # 替换 3: tls/ech -> tls/internal/ech
    find "$REPO_ROOT" -name "*.go" -type f | while read file; do
        if grep -q '"github.com/vistone/fingerprint/tls/ech"' "$file"; then
            sed -i 's|"github.com/vistone/fingerprint/tls/ech"|"github.com/vistone/fingerprint/tls/internal/ech"|g' "$file"
            log_success "更新: $(basename $file)"
        fi
    done
    
    log_success "Import 路径替换完成"
else
    log_info "[模拟] sed -i replace internal/tlsutil -> tls/internal/utils"
    log_info "[模拟] sed -i replace tls/utils -> tls/internal/utils"
    log_info "[模拟] sed -i replace tls/ech -> tls/internal/ech"
fi

echo ""

# 步骤 7: 删除旧目录
log_info "Step 7: 清理旧目录"

DIRS_TO_REMOVE=(
    "tls/utils"
    "tls/ech"
    "internal/tlsutil"
)

for dir in "${DIRS_TO_REMOVE[@]}"; do
    if [ -d "$REPO_ROOT/$dir" ]; then
        if [ "$DRY_RUN" == "execute" ]; then
            rm -rf "$REPO_ROOT/$dir"
            log_success "删除: $dir"
        else
            log_info "[模拟] rm -rf $dir"
        fi
    fi
done

echo ""

# 步骤 8: 重新整理模块
log_info "Step 8: 整理模块依赖"

if [ "$DRY_RUN" == "execute" ]; then
    cd "$REPO_ROOT"
    go mod tidy
    log_success "go mod tidy 完成"
else
    log_info "[模拟] cd $REPO_ROOT && go mod tidy"
fi

echo ""

# 步骤 9: 构建验证
log_info "Step 9: 构建验证"

if [ "$DRY_RUN" == "execute" ]; then
    cd "$REPO_ROOT"
    if go build ./...; then
        log_success "go build 成功"
    else
        log_error "go build 失败！"
        exit 1
    fi
else
    log_info "[模拟] cd $REPO_ROOT && go build ./..."
fi

echo ""

# 步骤 10: 测试验证
log_info "Step 10: 测试验证"

if [ "$DRY_RUN" == "execute" ]; then
    cd "$REPO_ROOT"
    
    # TLS 包测试
    if go test ./tls/... -v; then
        log_success "TLS 包测试通过"
    else
        log_error "TLS 包测试失败！"
        exit 1
    fi
    
    # 集成测试
    if go test ./test/... -v -short; then
        log_success "集成测试通过"
    else
        log_error "集成测试失败！"
        exit 1
    fi
else
    log_info "[模拟] cd $REPO_ROOT && go test ./tls/... -v"
    log_info "[模拟] cd $REPO_ROOT && go test ./test/... -v -short"
fi

echo ""

# 完成
if [ "$DRY_RUN" == "execute" ]; then
    log_success "=== Phase 1 迁移完成！==="
    echo ""
    echo "后续步骤："
    echo "1. 验证所有功能: make test"
    echo "2. 提交变更: git commit -m 'refactor: Phase 1 TLS layer internalization'"
    echo "3. 创建 PR: git push origin restructure/phase1"
else
    log_info "=== 模拟运行完成 ==="
    echo ""
    echo "确认无误后，执行:"
    echo "  bash $(basename $0) execute"
fi
