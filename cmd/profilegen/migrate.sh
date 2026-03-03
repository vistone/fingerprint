#!/bin/bash
# migrate.sh - 批量迁移指纹配置到 YAML
#
# 使用方法:
#   chmod +x cmd/profilegen/migrate.sh
#   ./cmd/profilegen/migrate.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
SPECS_DIR="${PROJECT_ROOT}/profiles/specs"
GENERATED_DIR="${PROJECT_ROOT}/profiles/generated"

echo "=== Fingerprint Profile Migration Tool ==="
echo ""

# 创建目录
mkdir -p "${SPECS_DIR}"
mkdir -p "${GENERATED_DIR}"

# 步骤 1: 解析现有 Go 文件，提取配置
echo "[1/4] 解析现有指纹配置..."
go run "${SCRIPT_DIR}/extract.go" \
    -input "${PROJECT_ROOT}/profiles/internal_browser_profiles.go" \
    -output "${SPECS_DIR}"

# 步骤 2: 验证 YAML 配置
echo "[2/4] 验证 YAML 配置..."
for yaml_file in "${SPECS_DIR}"/*.yaml; do
    if [ -f "$yaml_file" ]; then
        echo "  ✓ $(basename "$yaml_file")"
    fi
done

# 步骤 3: 生成 Go 代码
echo "[3/4] 生成 Go 代码..."
go run "${SCRIPT_DIR}/main.go" \
    -input "${SPECS_DIR}" \
    -output "${GENERATED_DIR}/generated_profiles.go"

# 步骤 4: 验证生成的代码
echo "[4/4] 验证生成的代码..."
cd "${PROJECT_ROOT}"
if go build "${GENERATED_DIR}/generated_profiles.go" 2>/dev/null; then
    echo "  ✓ 生成的代码编译成功"
else
    echo "  ✗ 生成的代码编译失败"
    exit 1
fi

echo ""
echo "=== Migration Complete ==="
echo "生成的文件:"
echo "  - YAML configs: ${SPECS_DIR}/"
echo "  - Go code: ${GENERATED_DIR}/generated_profiles.go"
echo ""
echo "下一步:"
echo "  1. 检查生成的 YAML 文件"
echo "  2. 对比生成的 Go 代码与原始代码"
echo "  3. 逐步替换 profiles/ 目录下的手工代码"
