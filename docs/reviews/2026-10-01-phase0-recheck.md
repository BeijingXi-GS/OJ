# Phase 0 独立工程复审

- 被审版本：提交基线 `7a952c6` 的当前未提交工作区；56 个代码 / 配置文件，SHA-256 `469c5c38da2a199f64c6e1e04d16bce4a3e7493145af1e72f97e43154a86bc00`。本 reviewer 实际执行 `node scripts/source-digest.mjs` 核对一致；摘要排除 MD、凭据、缓存和构建输出。
- 审阅类型：Builder / Architect / Phase 0 适用工程验收复审。
- 实际审阅者：独立子 Agent `/root/phase0_recheck`；实现者为根 Agent，**独立于实现者：是**。
- 日期：2026-10-01（Asia/Shanghai）。用户此前明确授权独立评审，本轮再次授权修复并推进；本 reviewer 不实现后续阶段、不冻结契约、不改源码 / 文件属性 / 服务 / 台账。
- 输入：[原独立报告](2026-10-01-phase0-review.md)、[本轮修复交接](../handoffs/2026-10-01-license-fix-validation.md)、[作者阶段验收表](2026-10-01-phase0-acceptance.md)、AGENTS、DEVELOPMENT、ROADMAP、P0-001–005 任务卡及实际代码 / 配置。
- 结论：**受影响基础工程复审通过；原 R1 / R2 可关闭。七项基础不变量可窄范围采用。Phase 0 可在根 Agent 登记采用记录、问题关闭和最终验收后通过；本报告没有自行更改阶段状态。** 完整 `0.1.0-draft` 类型、Graph、Runtime、schema、资源强制语义和事务接口继续 proposed。

## 1. R1 / R2 复审结论

### R1：重复覆盖只读 LICENSE——修复通过，可关闭

`scripts/build-minio.mjs` 已移除许可证的 `copyFileSync`。只对固定构建目标 `.local/minio-build/LICENSE` 在存在时使用 `chmodSync(..., 0o644)`，再读取上游文件原始字节、写入本项目输出，mode=0644。没有修改 module cache 或许可证内容，也没有为了修复而更改来源锁、Go / npm 依赖、数据库或 Docker 数据卷。首次创建和已有只读输出两条路径均与目标一致。

作者在同一既有构建目录连续两次 `npm run infra:build-minio` exit 0，并完成 doctor / Compose 配置 / up / 完整 infra-smoke / 真实数据库 API readiness / check / smoke / 探针 vet。该连续构建证据属于作者，本 reviewer 阅读交接并检查实际实现，没有重新构建镜像或冒称独立重跑。

本 reviewer 独立只读核对：

| 项目                         | 实际结果                                                                             |
| ---------------------------- | ------------------------------------------------------------------------------------ |
| 上游 cache 的 LICENSE        | 34523 bytes，IsReadOnly=true，ReadOnly + Archive                                     |
| `.local/minio-build/LICENSE` | 34523 bytes，IsReadOnly=false，Archive                                               |
| 两份 LICENSE SHA-256         | 均为 `0D96A4FF68AD6D4B6F1F30F713B18D5184912BA8DD389F86AA7710DB079ABCB0`              |
| MinIO 二进制 SHA-256         | `ab90d4d794fc5826df7766ce54f02c099f75b76550ec26847e4a76b8b7a741cf`，与构建元数据一致 |
| 语法核对                     | 本 reviewer 执行两个受影响脚本的 `node --check`，exit 0                              |

这些证据满足原报告 R1 的具体解除条件，先前只读输出覆盖解释已得到实际修复后的复验支持。不能由此推断任意 ACL / 文件锁问题均已测试。[独立问题档案](../issues/2026-10-01-minio-license-copy.md) 可由根 Agent 按本报告关闭，保留失败历史。

### R2：两个入口说明——修正通过，可关闭

- README 已移除过时的“当前未配置远程仓库”，将实际 remote / CI 状态指向 PROGRESS；本 reviewer 只读核对 origin 存在，文字明确没有把工作流配置等同于远程 CI 已通过。
- setup 的成功提示现为 doctor → infra:build-minio → infra:up → infra-smoke，与本地源码镜像的构建前置一致。README、开发环境说明、MinIO 构建说明和 CI 均包含构建前置；保留已有 `.env` 的实现未改动。

对应 [README 问题](../issues/2026-10-01-readme-remote-status.md) 与 [setup 问题](../issues/2026-10-01-setup-minio-instruction.md) 可关闭。冷机器首次初始化没有执行，因此此结论只覆盖入口一致性与现有环境复验，不是全新机器下载成功的证明。

## 2. 运行环境、来源与数据卷

本 reviewer 使用经过授权的只读 Docker 查询，并只输出必要的身份、标签、端口和卷字段，没有读取 `.env` 或输出服务环境凭据，没有重建 / 启停服务。

| 检查                        | 实际结果                                                                                              |
| --------------------------- | ----------------------------------------------------------------------------------------------------- |
| 当前 engine                 | Server 29.8.1，linux / x86_64                                                                         |
| 当前服务                    | PostgreSQL、Redis、MinIO 均 healthy                                                                   |
| 暴露端口                    | 仅 127.0.0.1；PG 15432→5432、Redis 6379、MinIO 9000 / 9001                                            |
| MinIO 运行镜像              | `sha256:57acc20067eda4fc3bf7984d7a7245dc49bf9e66c61e0b798f1fd3d2d0977056`，linux/amd64                |
| MinIO 用户                  | 10001:10001                                                                                           |
| MinIO OCI source / revision | `https://github.com/minio/minio` / `7aac2a2c5b7c882e68c1ce017d8256be2feea27f`，与来源锁完全一致       |
| 挂载                        | `oj-dev_minio-data:/data`、`oj-dev_postgres-data:/var/lib/postgresql/data`、`oj-dev_redis-data:/data` |
| 实际 Go checksum 配置       | 本 reviewer 读取 `go env GOSUMDB` 为 `sum.golang.org`；构建元数据保留 module / go.mod sums            |

镜像 ID 与原报告不同，当前 revision、平台、用户、卷和二进制 hash 一致，符合本轮修复后重建的事实；没有把开发源码镜像声明为原官方预编译 release。固定源码和 CA digest、scratch 最终层及 localhost 限制仍保留。原报告关于任意使用者 Go 校验配置、旧数据卷兼容和生产维护的边界继续适用，本轮没有验证这些范围。

## 3. 七项基础不变量的窄范围审阅

独立逐项比较 [FOUNDATION_INVARIANTS.md](../architecture/FOUNDATION_INVARIANTS.md)、原报告第 3–4 节、ARCHITECTURE / DOMAIN_MODEL / STATE_MACHINE / SECURITY_BASELINE 与实际核心草案。新文档整理的是已审原则及原报告明确要求的关系 / 安全边界，没有把尚未设计的接口形状升级为 frozen。受审文档为 `foundation-invariants/1.0.0`、当前 reviewed；本次读取的文件 SHA-256 为 `CB818F293111FEF2F6C9AE83E67E91AAF7AEFF74E7A812B8904FCED9FDE9CAAC`。根 Agent 后续仅登记状态 / 采用链接时不应把语义一起扩大。

| 原则                                       | 独立判断与范围                                                                                                                               |
| ------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| 1. API / Judge 分离                        | 与原已审边界一致，可冻结禁止 API 执行源码及独立 bridge 原则；没有实际部署权限或 bridge 实现验收。                                            |
| 2. 不可变快照、Run 产物及重判              | 与已审身份关系和版本追溯一致；不冻结 ID / FK / 存储 schema、版本升级规则或删除策略。                                                         |
| 3. 显式 Source / Artifact / Process 绑定   | 同一 Submission 快照约束是原报告明确要求；可冻结身份关系，不冻结完整 Graph / 管理进程表示。                                                  |
| 4. lifecycle / verdict 分离与终态          | 与状态文档及原已审原则一致；FINISHED 的合法 verdict 约束是原则，不表示 helper 已实现全部组合 / 竞争校验。                                    |
| 5. 资源分类、单位与聚合峰值                | 与原已审原则一致；不冻结默认值、CPU / wall 起止、输出 / PID / temp 强制执行或竞争归因。                                                      |
| 6. Runtime 证据 / 清理与 Engine 映射       | 清理失败归基础设施错误、无不安全降级和可信 fixture 边界与原审阅及安全基线一致；不冻结 `(result, error)` 接口、runner 实现或 Linux 安全能力。 |
| 7. 至少一次 / commit 后 ACK / fence / 追溯 | 与原已审事务原则一致；不声明 outbox、lease、条件 SQL、取消竞争或恢复算法已实现。                                                             |

**建议只对这七项原则完成窄范围采用 / 冻结记录，引用原报告和本复审。** 完整 structs、Graph / SourceBinding、ResourcePolicy 可执行语义、Runtime / Engine 错误契约、DB / API schema、授权、outbox / fencing / lease、Linux 隔离和 Phase 7–9 细节均继续 draft / proposed，实施前仍需相应 DESIGN 与独立审阅。本报告不研究或审阅 P1 新设计。

## 4. Phase 0 适用门禁与未测范围

根据 ROADMAP 的 Phase 0 行、P0-001–005 和 DEVELOPMENT 的适用检查要求：本阶段验收对象是骨架、基础服务、可编译草案、流程和独立原则审阅。作者的全套 check / smoke、真实服务 / 数据库 readiness 和连续构建证据，加上原报告的基础工程审阅及本次受影响工程复审，足以支持该范围的工程验收。

| 门禁                               | 证据及裁定                                                                                                                                                                                                                 |
| ---------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 骨架 / API smoke / 前端构建        | 作者本轮 check / smoke 通过，原报告基础范围通过；前端源码未实施新框架，真实提交路由不存在、Worker idle。                                                                                                                   |
| 核心 package / 状态 / 身份独立评审 | 原报告七项独立回答保留，本次窄范围原则核对通过；类型保持草案，不冒充运行能力。                                                                                                                                             |
| 基础服务 / 复现                    | 作者连续构建、config / up / infra-smoke / API readiness 通过；本 reviewer 独立核对现有身份、来源、健康和卷，R1 可关闭。                                                                                                    |
| 流程 / 入口                        | 原报告基础范围通过；R2 修正，唯一台账、异常接管和独立性规则保留。                                                                                                                                                          |
| 并发 / 安全 / 业务一致性专项       | 本阶段无 Queue / Judge / Sandbox 或业务事务实现；相应专项测试不在此次工程验收范围，不登记为已通过。                                                                                                                        |
| Windows race / 远程 CI             | **未执行**。当前普通工程修复未引入并发运行实现，ROADMAP / P0 任务未将远程 CI 或本机 race 单独列为 Phase 0 必须成功的退出项；不能将 Ubuntu race 配置冒称结果。Linux runner 实现阶段仍须执行适用测试 / race 与生命周期矩阵。 |
| 冷机器 / S3 / 旧卷 / 生产恢复      | **未执行**；本次 localhost 新 MinIO 卷健康及缓存复现不证明上述能力。                                                                                                                                                       |
| 前端性能                           | 现有 Monaco 体积警告已单独归档，非阻塞；不因此宣称产品前端已搭建或性能验收完成。                                                                                                                                           |

作者阶段验收表对上述适用范围的解释合理，明确区分作者测试与独立 review，并保留未测边界。文档采用、问题关闭、P0 任务状态与阶段最终结论仍由根 Agent 按实际完成记录同步；不能以本报告存在代替这些交接。

## 5. 独立性与最终交接

本 reviewer 实际独立执行：源码摘要、两个 JS 的语法检查、LICENSE 属性 / SHA-256、MinIO 二进制 hash 与元数据、Go checksum 配置、Docker 当前平台 / 服务 / 端口 / UID / revision / 命名卷核对；独立静态读取变更与原则 / 流程 / 任务卡。未独立重跑作者的构建、doctor、完整 check / smoke、真实数据库 API、race 或远程 CI，没有修改源码 / 属性、重建镜像、启停服务或登记冻结。

本次没有发现新的阻塞。根 Agent 可据本报告关闭已授权修复的 R1 / R2，登记七项基础不变量的窄范围采用，更新阶段验收 / 任务 / 台账；这些记录完成后，Phase 0 工程可以验收，并依用户授权进入 P1-001 **设计**。完整核心与后续阶段设计不因本报告自动冻结；前端仍仅交付 MD。新问题继续按用户要求一事一档并在阻塞时停止。
