#!/usr/bin/env python3
"""
Enhanced Chinese to English comment translator with comprehensive phrase dictionary.
Handles complex technical comments for browser fingerprinting fingerprint simulation.

Usage:
    python3 translate_comments_v2.py --scan           # Scan for Chinese comments
    python3 translate_comments_v2.py --fix            # Auto-translate (with review flags)
    python3 translate_comments_v2.py --file <path>    # Fix specific file
    python3 translate_comments_v2.py --check-dict     # Show translation dictionary
"""

import os
import re
import sys
import glob
import argparse
from pathlib import Path

# Comprehensive Chinese → English translation dictionary
# Organized by category for easier maintenance
TRANSLATION_DICT = {
    # ============ AGENT MODULE ============
    "自主安全智能体": "Autonomous Security Agent",
    "智能体": "Agent",
    "范式转移": "Paradigm shift",
    "被动": "Passive",
    "主动": "Active",
    "行为智能体": "Behavioral agent",
    "核心架构": "Core architecture",
    "观察": "Observe",
    "分析": "Analyze",
    "决策": "Decide",
    "执行": "Act",
    "OADA循环": "OADA loop",
    "关键字循环": "keyword loop",
    "Observer": "Observer",
    "持续收集": "Continuously collect",
    "指纹分析事件流": "fingerprint analysis event stream",
    "BehaviorAnalyzer": "BehaviorAnalyzer",
    "构建客户端行为画像": "Build client behavioral profile",
    "识别时序异常": "Identify temporal anomalies",
    "StrategyEngine": "StrategyEngine",
    "自适应策略引擎": "Adaptive strategy engine",
    "威胁模式": "Threat patterns",
    "动态演化": "Dynamically evolve",
    "检测规则": "Detection rules",
    "Memory": "Memory",
    "智能体记忆系统": "Agent memory system",
    "存储学习到的模式": "Store learned patterns",
    "威胁签名": "Threat signatures",
    "不替换": "Does not replace",
    "现有": "Existing",
    "模块": "Module",
    "之上构建": "Build on top of",
    "更高层": "Higher level",
    "自主决策能力": "Autonomous decision-making capability",
    "形成": "Form",
    "完整闭环": "Complete closed loop",
    "感知": "Perception",
    "认知": "Cognition",
    
    # ============ THREAT & ACTION TYPES ============
    "ActionType": "ActionType",
    "智能体动作类型": "Agent action types",
    "放行": "Allow (pass through)",
    "监控": "Monitor",
    "不阻断但加强观察": "No blocking but enhanced observation",
    "挑战": "Challenge",
    "JS验证": "JS Validation",
    "验证码": "CAPTCHA",
    "限速": "Throttle",
    "阻断": "Block",
    "ThreatClass": "ThreatClass",
    "威胁分类": "Threat classification",
    "自动化工具": "Automated tools",
    "爬虫": "Crawler",
    "指纹伪造": "Fingerprint spoofing",
    "会话异常": "Session anomaly",
    "行为异常": "Behavioral anomaly",
    "主动逃避检测": "Active evasion",
    
    # ============ OBSERVATION & ANALYSIS ============
    "Observation": "Observation",
    "单次观测事件": "Single observation event",
    "感知输入": "Perception input",
    "事件时间戳": "Event timestamp",
    "指纹特征": "Fingerprint attributes",
    "浏览器标识": "Browser identifier",
    "平台": "Platform",
    "用户代理": "User Agent",
    "网络特征": "Network characteristics",
    "IP地址": "IP address",
    "地理位置": "Geographic location",
    "时区": "Timezone",
    "端口": "Port",
    "HTTP头": "HTTP headers",
    "TLS参数": "TLS parameters",
    "编码偏好": "Encoding preferences",
    "请求行为": "Request behavior",
    "访问模式": "Access patterns",
    "交互行为": "Interaction behavior",
    "鼠标移动": "Mouse movement",
    "键盘输入": "Keyboard input",
    "滚动速度": "Scroll speed",
    "Anomaly": "Anomaly",
    "异常检测": "Anomaly detection",
    "位置偏移": "Location offset",
    "时间差异": "Time difference",
    "设备切换": "Device switch",
    "指纹切换": "Fingerprint switch",
    "行为模式": "Behavior pattern",
    
    # ============ TLS & CRYPTO ============
    "TLS握手": "TLS handshake",
    "密码套件": "Cipher suites",
    "扩展": "Extensions",
    "签名算法": "Signature algorithms",
    "密钥交换": "Key exchange",
    "椭圆曲线": "Elliptic curves",
    "曲线": "Curve",
    "支持的群": "Supported groups",
    "密钥协议": "Key agreement",
    "协议版本": "Protocol version",
    "加密算法": "Encryption algorithm",
    "HMAC": "HMAC",
    "摘要算法": "Digest algorithm",
    "证书验证": "Certificate verification",
    "证书链": "Certificate chain",
    "CA": "Certificate Authority",
    "自签名": "Self-signed",
    
    # ============ HTTP & NETWORK ============
    "HTTP版本": "HTTP version",
    "HTTP/2": "HTTP/2",
    "HTTP/1.1": "HTTP/1.1",
    "回退": "Fallback",
    "请求头": "Request headers",
    "响应头": "Response headers",
    "Content-Type": "Content-Type",
    "User-Agent": "User-Agent",
    "Accept": "Accept",
    "Referer": "Referer",
    "Cookie": "Cookie",
    "Set-Cookie": "Set-Cookie",
    "Connection": "Connection",
    "Keep-Alive": "Keep-Alive",
    "Encoding": "Encoding",
    "gzip": "gzip",
    "deflate": "deflate",
    "br": "Brotli",
    "压缩": "Compression",
    "DNS": "DNS",
    "TCP": "TCP",
    "UDP": "UDP",
    "IP": "IP",
    "IPv4": "IPv4",
    "IPv6": "IPv6",
    "QUIC": "QUIC",
    "代理": "Proxy",
    "VPN": "VPN",
    "连接池": "Connection pool",
    "SOCKS": "SOCKS",
    "隧道": "Tunnel",
    
    # ============ CLIENT & TRANSPORT ============
    "Client": "Client",
    "客户端": "Client",
    "完整的浏览器指纹客户端": "Complete browser fingerprint client",
    "浏览器指纹": "Browser fingerprint",
    "指纹模拟": "Fingerprint simulation",
    "指纹链路": "Fingerprint chain",
    "传输层": "Transport layer",
    "智能传输": "Smart transport",
    "标准TLS": "Standard TLS",
    "兼容性": "Compatibility",
    "路径": "Path",
    "禁止": "Disallow",
    "确保": "Ensure",
    "始终": "Always",
    "请求": "Request",
    "响应": "Response",
    "发送": "Send",
    "接收": "Receive",
    "执行HTTP请求": "Execute HTTP request",
    "关闭": "Close",
    "追踪": "Trace/Tracing",
    "追踪器": "Tracer",
    "创建": "Create",
    
    # ============ CONFIG & PROFILE ============
    "Config": "Config",
    "配置": "Configuration",
    "Profile": "Profile",
    "档案": "Profile/Document",
    "桥接": "Bridge",
    "转换": "Convert",
    "映射": "Map",
    "选项": "Options",
    "默认": "Default",
    "超时": "Timeout",
    "常量": "Constant/Constants",
    "定义": "Definition",
    "建立连接": "Establish connection",
    "握手": "Handshake",
    "DNS解析": "DNS resolution",
    "读取": "Read",
    "保活间隔": "Keep-alive interval",
    
    # ============ CORE UTILITIES ============
    "Core": "Core",
    "核心": "Core",
    "Logger": "Logger",
    "日志": "Log/Logging",
    "日志记录器": "Logger",
    "日志适配器": "Log adapter",
    "Pool": "Pool",
    "连接池": "Connection pool",
    "对象池": "Object pool",
    "指标": "Metrics",
    "度量": "Measurement",
    "计数器": "Counter",
    "直方图": "Histogram",
    "Trait": "Trait",
    "特征": "Trait/Feature",
    "Type": "Type",
    "类型": "Type",
    "Error": "Error",
    "错误": "Error",
    "异常": "Exception",
    "验证": "Validation/Validate",
    "检查": "Check",
    "常数": "Constant",
    "Utils": "Utils",
    "工具": "Tool/Utility",
    "函数": "Function",
    "方法": "Method",
    "接口": "Interface",
    "实现": "Implementation",
    "包": "Package",
    "模块": "Module",
    
    # ============ DEFENSE & ML ============
    "Defense": "Defense",
    "防御": "Defense",
    "Detector": "Detector",
    "检测器": "Detector",
    "机器学习": "Machine Learning",
    "ML": "ML",
    "Classifier": "Classifier",
    "分类器": "Classifier",
    "模型": "Model",
    "特征": "Feature",
    "特征提取": "Feature extraction",
    "Extractor": "Extractor",
    "提取器": "Extractor",
    "风险评分": "Risk score",
    "威胁": "Threat",
    "检测": "Detection",
    
    # ============ FRONTEND & JS ============
    "Frontend": "Frontend",
    "前端": "Frontend",
    "JavaScript": "JavaScript",
    "JS": "JavaScript",
    "SDK": "SDK",
    "反检测": "Anti-detection",
    "DOM": "DOM",
    "BOM": "BOM",
    "事件": "Event",
    "监听器": "Listener",
    "钩子": "Hook",
    "注入": "Injection/Inject",
    "代理": "Proxy",
    "拦截": "Interception/Intercept",
    
    # ============ TESTING ============
    "Test": "Test",
    "测试": "Test",
    "基准": "Benchmark",
    "基准测试": "Benchmark test",
    "性能": "Performance",
    "单元测试": "Unit test",
    "集成测试": "Integration test",
    "端到端": "End-to-end",
    "成功": "Success",
    "失败": "Failure",
    "错误情况": "Error case",
    "成功情况": "Success case",
    "验证": "Assertion/Verify",
    "期望": "Expected",
    "实际": "Actual",
    
    # ============ GATEWAY & CACHE ============
    "Gateway": "Gateway",
    "网关": "Gateway",
    "缓存": "Cache",
    "Cache": "Cache",
    "Breaker": "Breaker",
    "熔断": "Circuit breaker",
    "Rate Limiter": "Rate limiter",
    "限流": "Rate limiting",
    "Speed limiter": "Speed limiter",
    "速率限制": "Rate limiting",
    "Scanner": "Scanner",
    "扫描": "Scanning",
    "Handler": "Handler",
    "处理器": "Handler",
    
    # ============ GENERAL OPERATIONS ============
    "获取": "Get",
    "设置": "Set",
    "删除": "Delete",
    "更新": "Update",
    "查询": "Query",
    "添加": "Add",
    "移除": "Remove",
    "清除": "Clear",
    "初始化": "Initialize",
    "销毁": "Destroy",
    "复制": "Copy",
    "克隆": "Clone",
    "比较": "Compare",
    "排序": "Sort",
    "过滤": "Filter",
    "映射": "Map",
    "迭代": "Iterate",
    "遍历": "Traverse",
    "计算": "Calculate",
    "转换": "Convert",
    "编码": "Encode",
    "解码": "Decode",
    "序列化": "Serialize",
    "反序列化": "Deserialize",
    "压缩": "Compress",
    "解压": "Decompress",
    "加密": "Encrypt",
    "解密": "Decrypt",
    "哈希": "Hash",
    "校验": "Checksum/Verify",
    
    # ============ SPECIAL DOMAINS ============
    "指纹": "Fingerprint",
    "浏览器": "Browser",
    "设备": "Device",
    "检测": "Detection",
    "逃避": "Evasion",
    "伪造": "Spoofing",
    "模拟": "Simulation",
    "行为": "Behavior",
    "会话": "Session",
    "认证": "Authentication",
    "授权": "Authorization",
    "令牌": "Token",
    "签名": "Signature",
    "证书": "Certificate",
    "密钥": "Key",
    "秘密": "Secret",
    "API": "API",
    "端点": "Endpoint",
    "功能": "Feature/Function",
}

def has_chinese(text):
    """Check if text contains Chinese characters."""
    return bool(re.search(r'[\u4e00-\u9fff]', text))

def translate_phrase(text):
    """Translate Chinese phrase to English."""
    # Try exact match first
    if text in TRANSLATION_DICT:
        return TRANSLATION_DICT[text]
    
    # Try partial matches (longest first)
    matches = [(k, v) for k, v in TRANSLATION_DICT.items() if k in text]
    matches.sort(key=lambda x: len(x[0]), reverse=True)
    
    result = text
    for cn, en in matches:
        result = result.replace(cn, en)
    
    return result

def scan_file(filepath):
    """Scan file for Chinese comments."""
    chinese_lines = []
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            for line_num, line in enumerate(f, 1):
                # Check for Chinese in comments
                if '//' in line:
                    comment = line.split('//', 1)[1]
                    if has_chinese(comment):
                        chinese_lines.append((line_num, line.rstrip()))
                elif '/*' in line or '*' in line:
                    if has_chinese(line):
                        chinese_lines.append((line_num, line.rstrip()))
    except Exception as e:
        print(f"Error scanning {filepath}: {e}", file=sys.stderr)
    return chinese_lines

def translate_file(filepath, dry_run=False):
    """Translate Chinese comments in file."""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            lines = f.readlines()
        
        modified = False
        output_lines = []
        
        for line in lines:
            if has_chinese(line):
                # Translate the line
                translated = line
                for cn, en in sorted(TRANSLATION_DICT.items(), key=lambda x: len(x[0]), reverse=True):
                    translated = translated.replace(cn, en)
                
                # Mark if still has Chinese (partial translation)
                if has_chinese(translated):
                    translated = translated.rstrip() + " [NEEDS REVIEW]\n"
                
                output_lines.append(translated)
                modified = True
            else:
                output_lines.append(line)
        
        if modified and not dry_run:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.writelines(output_lines)
        
        return modified
    except Exception as e:
        print(f"Error translating {filepath}: {e}", file=sys.stderr)
        return False

def main():
    parser = argparse.ArgumentParser(description='Translate Chinese comments to English in Go files')
    parser.add_argument('--scan', action='store_true', help='Scan for Chinese comments')
    parser.add_argument('--fix', action='store_true', help='Auto-translate to English')
    parser.add_argument('--file', help='Process specific file')
    parser.add_argument('--dry-run', action='store_true', help='Preview changes without writing')
    parser.add_argument('--check-dict', action='store_true', help='Show translation dictionary')
    
    args = parser.parse_args()
    
    if args.check_dict:
        print(f"Translation dictionary has {len(TRANSLATION_DICT)} entries:")
        for cn, en in sorted(TRANSLATION_DICT.items()):
            print(f"  '{cn}' → '{en}'")
        return
    
    # Get file list
    if args.file:
        files = [args.file]
    else:
        files = glob.glob('modules/**/*.go', recursive=True) + glob.glob('cmd/**/*.go', recursive=True) + glob.glob('examples/**/*.go', recursive=True)
    
    total_files = 0
    total_lines = 0
    
    for filepath in sorted(files):
        if args.scan:
            chinese_lines = scan_file(filepath)
            if chinese_lines:
                total_files += 1
                total_lines += len(chinese_lines)
                print(f"{filepath}: {len(chinese_lines)} lines)")
                if args.file:  # Show details for single file
                    for line_no, content in chinese_lines[:5]:
                        print(f"  Line {line_no}: {content[:80]}")
                    if len(chinese_lines) > 5:
                        print(f"  ... and {len(chinese_lines) - 5} more lines")
        elif args.fix:
            if translate_file(filepath, dry_run=args.dry_run):
                print(f"{'[DRY RUN] ' if args.dry_run else ''}✓ Fixed: {filepath}")
                total_files += 1
    
    if args.scan:
        print("\n" + "=" * 60)
        print(f"Summary: {total_files} files with {total_lines} Chinese comment lines")
        print("Run with --fix to auto-translate")
    elif args.fix:
        print("\n" + "=" * 60)
        print(f"Translated {total_files} files")
        print("Check manually for [NEEDS REVIEW] tags and correct as needed")

if __name__ == '__main__':
    main()
