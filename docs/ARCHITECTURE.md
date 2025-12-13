# 架构设计文档

## 项目结构

```
/workspace/
├── bin/                    # 编译后的二进制文件
├── docs/                   # 文档目录
│   ├── API.md             # API 参考
│   ├── ARCHITECTURE.md    # 架构设计
│   └── CONTRIBUTING.md    # 贡献指南
├── examples/              # 示例代码
│   ├── basic/            # 基础示例
│   ├── complete/         # 完整示例
│   ├── headers/          # Headers 示例
│   └── ...
├── internal/              # 内部工具包
│   └── utils/            # 通用工具函数
│       ├── strings.go    # 字符串工具
│       ├── rand.go       # 随机数生成器
│       └── useragent.go  # User-Agent 工具
├── profiles/              # 指纹配置
│   ├── profiles.go       # Profile 定义
│   ├── internal_browser_profiles.go
│   ├── internal_custom_profiles.go
│   ├── contributed_browser_profiles.go
│   └── contributed_custom_profiles.go
├── test/                  # 测试文件
│   ├── fingerprint_test.go
│   └── network_test.go
├── go.mod                 # Go 模块文件
├── go.sum                 # 依赖锁定文件
├── LICENSE                # 许可证
├── README.md              # 项目说明
├── CHANGELOG.md           # 变更日志
├── types.go               # 共享类型定义
├── headers.go             # HTTP Headers 处理
├── useragent.go           # User-Agent 生成
├── useragent_helper.go    # User-Agent 辅助函数
├── random.go              # 随机指纹选择
└── profiles.go            # Profile 类型别名
```

## 模块划分

### 1. 核心模块

#### types.go
- 定义所有共享类型：`BrowserType`、`OperatingSystem`、`HTTPHeaders`、`FingerprintResult` 等
- 避免循环依赖

#### profiles.go
- `ClientProfile` 类型别名
- 导出 `profiles` 子包的主要类型和变量

### 2. 功能模块

#### headers.go
- HTTP Headers 生成和管理
- 支持 30+ 种全球语言
- 自动生成标准浏览器 Headers

#### useragent.go
- User-Agent 生成器核心实现
- 支持所有主流浏览器和版本
- 模板系统

#### useragent_helper.go
- User-Agent 辅助函数
- 简化 API 调用

#### random.go
- 随机指纹选择
- 支持按浏览器类型筛选
- 集成 User-Agent 和 Headers 生成

### 3. 工具模块（internal/utils）

#### strings.go
- 统一的字符串操作工具
- 使用标准库实现，性能优化

#### rand.go
- 统一的随机数生成器
- 线程安全
- 单例模式

#### useragent.go
- User-Agent 相关工具函数
- 版本提取、平台识别等

### 4. 配置模块（profiles/）

包含所有浏览器指纹配置：
- 主流浏览器指纹（Chrome、Firefox、Safari、Opera）
- 移动端指纹
- 自定义指纹

## 设计原则

### 1. 职责单一

每个模块只负责一个核心功能：
- `headers.go` 只处理 HTTP Headers
- `useragent.go` 只处理 User-Agent 生成
- `random.go` 只处理随机选择逻辑

### 2. 低耦合

- 使用 `types.go` 定义共享类型，避免循环依赖
- 工具函数放在 `internal/utils`，统一管理
- 各模块相互独立，可单独使用

### 3. 高内聚

- 相关功能集中在一个模块内
- 减少跨模块调用

### 4. 依赖方向

```
random.go → useragent.go → types.go
         ↘             ↗
         headers.go → types.go
              ↓
      internal/utils
```

## 线程安全

### 随机数生成器

- 使用 `internal/utils.RandGenerator` 统一管理
- 所有随机操作都是线程安全的
- 单例模式，避免多次初始化

### 全局变量

- `MappedTLSClients` 是只读的，线程安全
- `Languages`、`OperatingSystems` 是只读的，线程安全

## 性能优化

### 1. 使用标准库

- 所有字符串操作使用标准库（有汇编优化）
- 避免自定义实现

### 2. 减少内存分配

- 使用对象池（如需要）
- 克隆操作时只克隆必要的数据

### 3. 延迟初始化

- User-Agent 模板在首次使用时初始化
- 避免不必要的启动开销

## 扩展性

### 添加新浏览器

1. 在 `profiles/` 中添加新的指纹配置
2. 在 `useragent.go` 的 `initTemplates()` 中添加 User-Agent 模板
3. 更新 `README.md` 的支持列表

### 添加新功能

1. 在 `types.go` 中定义新的类型（如需要）
2. 在对应模块中实现功能
3. 在 `test/` 中添加测试
4. 更新文档

## 测试策略

### 单元测试

- 每个模块都有对应的测试
- 覆盖核心功能和边界情况

### 集成测试

- 测试完整的工作流程
- 测试真实网络请求（可选）

### 基准测试

- 测试关键路径的性能
- 确保优化有效
