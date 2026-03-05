# Examples

本目录包含 fingerprint 库的使用示例。

## 示例列表

| 示例 | 说明 | 运行方式 |
|------|------|---------|
| `simple/` | 最简单的指纹获取 | `cd simple && go run .` |
| `basic/` | 基础 API 使用 | `cd basic && go run .` |
| `random/` | 随机指纹生成 | `cd random && go run .` |
| `advanced/` | 高级功能（网关、安全） | `cd advanced && go run .` |
| `plugin/` | 插件系统示例 | `cd plugin && go run .` |

## 快速开始

```bash
# 简单示例
cd simple
go run .

# 基础示例
cd basic
go run .
```

## 示例结构

每个示例目录包含：
- `main.go` - 示例代码
- `go.mod` - 模块定义（独立模块）

所有示例都使用 workspace 的 replace 指令指向本地模块。
