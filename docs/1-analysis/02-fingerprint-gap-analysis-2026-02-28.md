# 指纹能力缺口分析与补齐路线（2026-02-28）

## 1. 结论摘要

当前项目在 **TLS ClientHello 指纹生成（JA3/JA4）**、**Profile 库管理**、**UA/Headers 组装**、**基础异常识别** 方面已具备可用能力；但与近两年主流防护体系相比，仍存在明显短板：

- 仍以 **单请求静态特征** 为主，缺少跨请求行为信号（inter-request signals）
- 对 **JA4+ 家族**支持不完整，仅覆盖 JA4（TLS Client），未覆盖 JA4S/JA4H/JA4X/JA4T 等
- 缺少 **HTTP/2 帧级签名输出** 与 **HTTP/3/QUIC 传输层签名**
- 对 **UA-CH（Client Hints）协商模型**支持不完整，当前更偏静态头模板
- 被动识别与异常检测规则较浅，缺少可校准、可观测、可学习的评分框架

建议采用“三阶段补齐”：**P0（正确性+可观测）→ P1（协议层扩展）→ P2（行为层与生态接入）**。

---

## 2. 外部技术基线（对标）

### 2.1 JA4/JA4+

- JA4 已成为 JA3 之后更稳健的 TLS 客户端特征（应对扩展顺序随机化）
- 业界实践强调组合使用：JA4 + JA4S + JA4H + JA4X + JA4T（或部分组合）
- 重点从“单个 hash”转向“分段特征与组合推理”

### 2.2 行为信号（Inter-request Signals）

- 主流防护不再仅依赖单次握手
- 更依赖 1h 窗口内聚合指标（协议比例、请求分位数、路径/UA/IP 稳定性）
- 单请求伪装成本低，跨请求一致性伪装成本高

### 2.3 UA Reduction + UA-CH

- UA 字符串逐步降粒度，UA-CH 成为结构化补充
- `Accept-CH`、`Permissions-Policy`、低熵/高熵提示、跨源委托成为关键机制
- 实务要求“无提示也可工作”的降级路径

### 2.4 ECH 与可见性变化

- ECH 使 SNI 可见性下降，依赖 SNI 的策略需要收缩
- 但 ClientHello 其他可见字段、传输层行为、应用层行为仍可形成有效信号

---

## 3. 当前项目能力盘点（基于代码）

### 3.1 已具备

1. JA3 与 JA4（TLS Client）计算
   - `ja3.go`、`ja4.go`
2. 大量浏览器/客户端 profile 管理与随机输出
   - `profiles/profiles.go`、`random.go`
3. HTTP 头模板与 UA 生成
   - `headers.go`、`useragent.go`
4. 基础被动识别/异常识别/矛盾识别
   - `defense.go`
5. HTTP/2 参数在 profile 中有数据结构承载
   - `profiles/profiles.go`（settings、priority、pseudo header order）

### 3.2 主要短板

1. JA4+ 不完整
   - 缺 JA4S / JA4H / JA4X / JA4T 计算能力与 API
2. HTTP/2 指纹“配置存在，但无标准化签名输出”
   - 目前可取到 settings/order 等，但没有 TS1 类 canonical/signature 产出
3. HTTP/3/QUIC 指纹缺少真实实现路径
   - 示例说明为主，缺可执行签名计算与验证接口
4. UA-CH 支持偏静态
   - 头部可生成，但缺 `Accept-CH`/跨源委托/高熵请求策略抽象
5. 行为层能力缺失
   - 无跨请求聚合、无时窗特征、无风险评分器
6. 识别器可解释性与可扩展性不足
   - `defense.go` 规则较硬编码，缺规则版本化与阈值配置化

---

## 4. 缺口矩阵（技术维度）

| 维度 | 现状 | 差距等级 | 影响 |
|---|---|---|---|
| TLS Client 指纹（JA3/JA4） | 已实现 | 低 | 基础可用 |
| JA4S（ServerHello） | 未实现 | 高 | 缺少会话双边关联 |
| JA4H（HTTP 客户端） | 未实现 | 高 | 无法做请求层高保真识别 |
| JA4X（证书生成模式） | 未实现 | 中-高 | 对服务端家族识别不足 |
| JA4T/HTTP2 帧签名 | 未实现（仅配置） | 高 | 抗 bot/抗伪装关键链路缺失 |
| QUIC/HTTP3 签名 | 未实现 | 中-高 | 新流量占比提升后识别退化 |
| UA-CH 协商策略 | 部分 | 中 | 头部仿真真实性受限 |
| 行为时窗信号 | 未实现 | 高 | 容易被单请求伪装绕过 |
| 规则引擎与评分 | 初级硬编码 | 中 | 调参与迭代成本高 |
| 可观测性（指标/回放） | 较弱 | 中 | 难做线上验证与回归 |

---

## 5. 补齐优先级与路线

## P0（1-2 周）：先把“真实性与可验证性”打牢

1. 统一协议特征抽取层（Feature Extractor）
   - 新增内部抽象：TLS、HTTP2、Headers、UA-CH 特征统一结构
2. 识别规则配置化
   - 将 `defense.go` 硬编码阈值改为可配置（默认值 + 版本号）
3. 可观测性最小闭环
   - 输出 debug feature 快照（可开关）与 benchmark 对照
4. 文档收敛
   - README 中“已支持/规划中”明确分层，避免能力误读

## P1（2-4 周）：补协议层核心缺口

1. 新增 JA4S 计算
   - 输入 ServerHello 相关字段，输出标准化摘要
2. 新增 HTTP/2 签名（TS1 风格）
   - 基于 settings、window update、pseudo header order、priority 生成 canonical signature
3. 新增 JA4H 最小版本
   - 对请求方法、关键头、cookie 结构做稳定摘要
4. QUIC/HTTP3 特征入口
   - 提供最小可用接口（先不做全量）

## P2（4-8 周）：上行为层与对抗能力

1. Inter-request Signals
   - 1h 时窗聚合：协议比例、请求频率分位、路径稳定性、IP/JA4 漂移
2. 风险评分器
   - 单请求分 + 时窗分组合，输出 explainable score
3. 指纹库更新机制
   - 增加 profile 数据版本化与自动一致性校验
4. 生态接入
   - 提供 SIEM/WAF 友好的结构化输出（JSON schema + 字段字典）

---

## 6. 建议的代码落点

- 协议抽取层：`internal/` 下新增 `features/`（建议）
- JA4+ 扩展：新增 `ja4s.go`、`ja4h.go`、`http2sig.go`
- 规则与评分：新增 `scoring.go`，并让 `defense.go` 调用
- 配置：新增 `config/` 或 `types.go` 中扩展可配置结构
- 文档：更新 `README.md` 与 `docs/3-references/00-quick-reference.md`

---

## 7. 风险与注意事项

1. 许可边界
   - JA4（TLS Client）与 JA4+ 家族在商用分发场景的许可条款不同，需在发布前确认
2. 误报风险
   - 行为规则必须灰度上线，先做观测再做阻断
3. 伪装对抗
   - 单一特征不可靠，必须做跨层组合（TLS + HTTP2/3 + 行为）
4. 跨浏览器兼容
   - UA-CH 生态支持并不完全一致，需保持 fallback 路径

---

## 8. 下一步执行建议

建议立即进入 P0，并并行设计 P1 的接口草案。优先顺序：

1. Feature 抽取统一结构
2. HTTP/2 签名实现（收益最高）
3. JA4S 最小实现
4. README 能力声明与路线图同步

完成以上后，再进入行为层（P2）。
