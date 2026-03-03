# 测试文档

本目录包含 fingerprint 项目的测试相关文件和文档。

## 目录结构

```
test/
├── README.md              # 本文档
├── benchmark_baseline.txt # 性能基线数据
└── profile_validation_test.go # Profile 验证测试
```

## 性能基准

### 运行基准测试

```bash
# 运行所有基准测试
go test -bench=. -benchmem ./...

# 运行特定包的基准测试
go test -bench=. -benchmem ./internal/tcpip/...
go test -bench=. -benchmem ./internal/config/...
go test -bench=. -benchmem ./profiles/...
```

### 当前性能基线

| 包 | 基准测试 | 每次操作时间 | 内存分配 | 分配次数 |
|---|---|---|---|---|
| internal/tcpip | BenchmarkComputeTCPSignature | ~2.1μs | 144 B | 8 |
| internal/tcpip | BenchmarkMatchOSSignature | ~300ns | 0 B | 0 |
| internal/tcpip | BenchmarkAnalyzeNetworkBehavior | ~720ns | 336 B | 2 |
| internal/config | BenchmarkManagedConfigClone | ~1.4μs | 581 B | 13 |
| internal/config | BenchmarkFeatureExtractionConfigClone | ~270ns | 112 B | 2 |
| profiles | BenchmarkGetClientHelloSpec | ~2.8μs | 1104 B | 30 |
| profiles | BenchmarkGetSettings | ~0.4ns | 0 B | 0 |
| profiles | BenchmarkGetPseudoHeaderOrder | ~0.4ns | 0 B | 0 |

### 性能回归测试

```bash
# 生成当前基准
go test -bench=. -benchmem ./... > benchmark_current.txt

# 与基线比较
go run ./scripts/compare_benchmarks.go \
  -base test/benchmark_baseline.txt \
  -current benchmark_current.txt
```

## CI/CD 配置

参见 `.github/workflows/ci.yml` 获取完整的 CI 配置，包括：

- 单元测试
- 基准测试
- 代码覆盖率
- 模糊测试
- 静态分析
- 安全扫描

## 覆盖率要求

- 目标：> 60%
- 关键路径（profiles, internal/tcpip）：> 80%

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```
