#!/bin/bash

# 批量验证多个profiles的指纹准确性
# 生成完整的验证报告

set -e

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
WORKSPACE_DIR=$(cd "${SCRIPT_DIR}/.." && pwd)
REPORT_DIR="/tmp/fingerprint_verification_reports"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

mkdir -p "${REPORT_DIR}"

echo "============================================================"
echo "批量指纹验证工具 - Batch Fingerprint Verification"
echo "============================================================"
echo "工作目录: ${WORKSPACE_DIR}"
echo "报告目录: ${REPORT_DIR}"
echo "时间戳: ${TIMESTAMP}"
echo ""

# 定义要测试的profiles
PROFILES=(
    "chrome_134"
    "chrome_131"
    "chrome_120"
    "firefox_133"
    "firefox_120"
    "safari_17_6"
    "safari_16_6"
    "edge_131"
    "edge_120"
    "brave_1_72"
    "brave_1_60"
    "opera_114"
)

TOTAL=${#PROFILES[@]}
PASSED=0
FAILED=0
FAILED_PROFILES=()

echo "测试 ${TOTAL} 个 profiles..."
echo ""

for i in "${!PROFILES[@]}"; do
    PROFILE="${PROFILES[$i]}"
    INDEX=$((i + 1))

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "[$INDEX/$TOTAL] 验证 Profile: ${PROFILE}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    OUTPUT_JSON="${REPORT_DIR}/verify_${PROFILE}_${TIMESTAMP}.json"

    cd "${WORKSPACE_DIR}"

    if timeout 120s go run ./scripts/verify_complete.go \
        -profile "${PROFILE}" \
        -out "${OUTPUT_JSON}" \
        -verbose=false 2>&1 | tee "${REPORT_DIR}/verify_${PROFILE}_${TIMESTAMP}.log"; then

        echo "✓ ${PROFILE}: PASS"
        PASSED=$((PASSED + 1))
    else
        EXIT_CODE=$?
        echo "✗ ${PROFILE}: FAIL (exit code: ${EXIT_CODE})"
        FAILED=$((FAILED + 1))
        FAILED_PROFILES+=("${PROFILE}")
    fi

    echo ""

    # Rate limiting between tests
    if [ $INDEX -lt $TOTAL ]; then
        sleep 3
    fi
done

echo ""
echo "============================================================"
echo "批量验证完成 - Batch Verification Complete"
echo "============================================================"
echo "总计: ${TOTAL}"
echo "通过: ${PASSED}"
echo "失败: ${FAILED}"
echo "通过率: $(awk "BEGIN {printf \"%.1f\", ${PASSED}/${TOTAL}*100}")%"
echo ""

if [ $FAILED -gt 0 ]; then
    echo "失败的 Profiles:"
    for PROFILE in "${FAILED_PROFILES[@]}"; do
        echo "  - ${PROFILE}"
    done
    echo ""
fi

# 生成汇总报告
SUMMARY_FILE="${REPORT_DIR}/batch_summary_${TIMESTAMP}.txt"

cat > "${SUMMARY_FILE}" << EOF
指纹验证批量测试报告
Batch Fingerprint Verification Report
============================================================
时间: $(date)
工作目录: ${WORKSPACE_DIR}

测试结果统计:
  总计 Profiles: ${TOTAL}
  通过 (PASS): ${PASSED}
  失败 (FAIL): ${FAILED}
  通过率: $(awk "BEGIN {printf \"%.1f\", ${PASSED}/${TOTAL}*100}")%

测试的 Profiles:
$(for p in "${PROFILES[@]}"; do echo "  - $p"; done)

EOF

if [ $FAILED -gt 0 ]; then
    cat >> "${SUMMARY_FILE}" << EOF
失败的 Profiles:
$(for p in "${FAILED_PROFILES[@]}"; do echo "  - $p"; done)

EOF
fi

cat >> "${SUMMARY_FILE}" << EOF
详细报告文件:
  目录: ${REPORT_DIR}
  JSON报告: verify_*_${TIMESTAMP}.json
  日志文件: verify_*_${TIMESTAMP}.log

验证内容:
  ✓ User-Agent 是否与设计的profile一致
  ✓ Sec-CH-UA 是否与设计的profile一致（Chromium系）
  ✓ HTTP Headers 发送顺序和内容
  ✓ TLS 指纹（通过ALPN和Cipher Suite）
  ✓ HTTP协议版本（HTTP/1.1 vs HTTP/2）

结论:
$(if [ $FAILED -eq 0 ]; then
    echo "  ✓ 所有profiles都正确发送了我们设计的指纹"
    echo "  ✓ 没有使用系统默认的网络栈"
    echo "echo "  ✓ 虚拟TCP/IP+浏览器指纹工作正常"
else
    echo "  ⚠ 部分profiles验证失败，需要检查"
    echo "  ⚠ 查看详细日志文件进行排查"
fi)

EOF

echo "汇总报告已保存: ${SUMMARY_FILE}"
echo ""

cat "${SUMMARY_FILE}"

echo ""
echo "所有报告文件位于: ${REPORT_DIR}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "✓✓✓ 全部通过! All profiles verified successfully! ✓✓✓"
    exit 0
else
    echo "✗✗✗ 部分失败! Some profiles failed verification! ✗✗✗"
    exit 1
fi
