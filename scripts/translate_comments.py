#!/usr/bin/env python3
"""
Translate Chinese comments to English in Go source files.
This tool scans all .go files and converts Chinese comments to English.

Usage:
    python3 translate_comments.py --scan      # Scan for Chinese comments
    python3 translate_comments.py --fix       # Auto-translate to English
    python3 translate_comments.py --file <path> # Fix specific file
"""

import os
import re
import sys
import glob
import argparse
from pathlib import Path

# Simple Chinese to English translation mappings
TRANSLATION_MAP = {
    # Package/module descriptions
    "Package": "Package",
    "提供完整的浏览器指纹模拟客户端": "Provides complete browser fingerprint simulation client",
    "从 TCP/IP 层到 TLS 层到 HTTP 层的全栈模拟": "Full-stack simulation from TCP/IP to TLS to HTTP",
    
    # Comments about constants
    "超时常量定义": "Timeout constants definition",
    "使用 core 包标准值": "uses core package standard values",
    "建立 TCP/IP 连接超时": "TCP/IP connection establishment timeout",
    "TLS 握手超时": "TLS handshake timeout",
    "DNS 解析超时": "DNS resolution timeout",
    "读取 HTTP 响应头超时": "HTTP response header read timeout",
    "单个请求超时": "Single request timeout",
    "总请求超时（包括重定向）": "Total request timeout including redirects",
    "TCP 保活间隔": "TCP keep-alive interval",
    
    # Comments about types
    "完整的浏览器指纹客户端": "Complete browser fingerprint client",
    "客户端选项": "Client options",
    "禁止使用标准 TLS 兼容回退路径，确保请求始终走指纹链路": "Disallow standard TLS compatibility fallback, ensure requests always use fingerprint chain",
    "默认选项": "Default options",
    "创建浏览器指纹客户端": "Create browser fingerprint client",
    "创建智能传输层（支持 HTTP/2 → HTTP/1.1 回退）": "Create smart transport layer (supports HTTP/2 to HTTP/1.1 fallback)",
    "创建 HTTP 客户端（使用 fhttp）": "Create HTTP client (using fhttp)",
    "创建带追踪的客户端": "Create client with tracing",
    "创建请求追踪器": "Create request tracer",
    
    # Common operations
    "执行 HTTP 请求": "Execute HTTP request",
    "关闭客户端": "Close client",
    "发送请求": "Send request",
    "处理响应": "Handle response",
    "验证": "Validate",
    "初始化": "Initialize",
    "检查": "Check",
    "获取": "Get",
    "设置": "Set",
    "删除": "Delete",
    "更新": "Update",
    "查询": "Query",
    "计算": "Calculate",
    "转换": "Convert",
    "编码": "Encode",
    "解码": "Decode",
    
    # Test descriptions
    "测试": "Test",
    "基准测试": "Benchmark",
    "成功": "Success",
    "失败": "Failure",
    "错误": "Error",
    "验证结果": "Verify result",
}

# Extended translations using regex patterns
PATTERN_TRANSLATIONS = [
    (r"// ([\u4e00-\u9fff]+)", lambda m: f"// {translate_phrase(m.group(1))}"),
    (r"/\* ([\u4e00-\u9fff]+)", lambda m: f"/* {translate_phrase(m.group(1))}"),
    (r"\* ([\u4e00-\u9fff]+)", lambda m: f"* {translate_phrase(m.group(1))}"),
]

def translate_phrase(text):
    """Translate a Chinese phrase to English."""
    # Try exact match first
    if text in TRANSLATION_MAP:
        return TRANSLATION_MAP[text]
    
    # Try partial matches
    for cn, en in TRANSLATION_MAP.items():
        if cn in text:
            return text.replace(cn, en)
    
    # Fallback: Use a simple romanization/English approximation
    # This is a placeholder - in production, use a real translation API
    return f"[TRANSLATE: {text}]"

def has_chinese(text):
    """Check if text contains Chinese characters."""
    return bool(re.search(r'[\u4e00-\u9fff]', text))

def scan_file(filepath):
    """Scan a file for Chinese comments."""
    chinese_lines = []
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            for line_num, line in enumerate(f, 1):
                if '//' in line:
                    comment_part = line.split('//', 1)[1]
                    if has_chinese(comment_part):
                        chinese_lines.append((line_num, line.strip()))
    except Exception as e:
        print(f"Error reading {filepath}: {e}", file=sys.stderr)
    
    return chinese_lines

def translate_file(filepath):
    """Translate Chinese comments in a file to English."""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # Track if changes were made
        original_content = content
        
        # Simple regex-based translation
        chinese_pattern = re.compile(r'(//.*?[\u4e00-\u9fff].*?)(\n|$)')
        
        def replace_comment(match):
            comment_line = match.group(1)
            # Try to translate known phrases
            translated = comment_line
            for cn, en in TRANSLATION_MAP.items():
                if cn in translated:
                    translated = translated.replace(cn, en)
            # Mark remaining untranslated Chinese
            if has_chinese(translated):
                translated += " [NEEDS REVIEW]"
            return translated + match.group(2)
        
        content = chinese_pattern.sub(replace_comment, content)
        
        # Write back if changes were made
        if content != original_content:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(content)
            return True
        return False
    except Exception as e:
        print(f"Error translating {filepath}: {e}", file=sys.stderr)
        return False

def main():
    parser = argparse.ArgumentParser(
        description='Translate Chinese comments to English in Go files'
    )
    parser.add_argument('--scan', action='store_true', 
                       help='Scan for Chinese comments')
    parser.add_argument('--fix', action='store_true',
                       help='Auto-translate Chinese comments')
    parser.add_argument('--file', type=str,
                       help='Process specific file')
    parser.add_argument('--dry-run', action='store_true',
                       help='Show what would be fixed without making changes')

    args = parser.parse_args()

    # Get file list
    if args.file:
        files = [args.file]
    else:
        files = glob.glob('modules/**/*.go', recursive=True)
        files += glob.glob('cmd/**/*.go', recursive=True)
        files += glob.glob('examples/**/*.go', recursive=True)

    if args.scan or not (args.fix or args.file):
        # Scan mode
        total_files = 0
        total_lines = 0
        
        for filepath in sorted(files):
            if filepath.startswith('.'):
                continue
            
            chinese_lines = scan_file(filepath)
            if chinese_lines:
                total_files += 1
                total_lines += len(chinese_lines)
                print(f"\n{filepath} ({len(chinese_lines)} lines):")
                for line_num, line in chinese_lines[:3]:  # Show first 3
                    print(f"  {line_num}: {line[:80]}")
                if len(chinese_lines) > 3:
                    print(f"  ... and {len(chinese_lines) - 3} more lines")
        
        print(f"\n{'='*60}")
        print(f"Summary: {total_files} files with {total_lines} Chinese comment lines")
        print(f"Run with --fix to auto-translate")
    
    elif args.fix:
        # Fix mode
        fixed_count = 0
        for filepath in sorted(files):
            if filepath.startswith('.'):
                continue
            
            if translate_file(filepath):
                fixed_count += 1
                print(f"✓ Fixed: {filepath}")
        
        print(f"\n{'='*60}")
        print(f"Fixed {fixed_count} files")
        print("\nNOTE: Manual review is recommended for complex comments!")
        print("Many translated comments are marked with [NEEDS REVIEW]")

if __name__ == '__main__':
    main()
