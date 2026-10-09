# P0-004：独立架构与基础工程评审输入包

- 日期：2026-10-01；整理人：本轮 Agent（实现 / 文档作者）。
- 状态：输入包已交付；用户授权的独立 Agent `phase0_review` 已完成 [正式 REVIEW](2026-10-01-phase0-review.md)，**本文件不是独立评审结论**，未冻结核心。
- 被审基线：`7a952c6` 上的当前工作区；代码 / 配置摘要 56 文件，SHA-256 `95e401d13d7692a1d1c2f5872299ee7ada68226a583ad434233ad9dcb95d584f`。
- 相关任务：P0-001 / P0-002 / P0-003 / P0-004 / P0-005；角色：独立 Architect，基础工程项可同时承担 Builder reviewer。

## 1. 进入评审的具体结果

**最新复验失败：** 最终重复 `npm run infra:build-minio` 在 LICENSE 复制时 EPERM，源 / 输出文件均只读；运行服务仍 healthy。详细证据见 [构建问题](../issues/2026-10-01-minio-license-copy.md)。该新缺陷阻止基础工程验收关闭，不能忽略后只沿用下方首建成功结果。

基础服务阻塞已解除：Docker Server 29.8.1 / linux；PG / Redis / MinIO 全部 healthy；完整 infra-smoke 通过，API 读取真实 `.env` 的 `/readyz=200`。MinIO 改为官方固定源码构建的 localhost 开发镜像，详见 [ADR-0002](../adr/0002-local-minio-source-image.md)。

`npm run check` 通过格式、Go vet、ESLint、Go 单元测试、Go 构建、TypeScript 和 Vite 构建。静态健康探针额外 Go vet 通过。核心业务没有新增，提交端点 / 实际 Judge 仍未实现；Windows race / 远程 CI 未重新执行，不能登记为通过。

前端源码未改；构建存在 Monaco chunk 大小警告，单独记录为非阻塞项。前端方案及任务包仍只是实施说明。

## 2. 独立性与范围

[AGENTS.md](../../AGENTS.md) 要求自检不能登记为独立审阅，[ADR-0001](../adr/0001-foundation-boundaries.md) 要求核心接口没有独立审阅记录前保持 draft。本文件作者是原实现者，不能在下面填“独立审阅通过”。

独立 reviewer 应逐项记录结论、证据、阻塞修改项、冻结范围与排除范围，并使用 [REVIEW 模板](../templates/REVIEW.md)。只有类型形状 / 基础不变量有机会在此阶段冻结；Phase 2 DB / API、Phase 3 outbox / fence、Phase 4 Linux 隔离、Phase 7–9 图 / 调度实现不能被一并宣布完成。

## 3. 核心七项核验

| 核验                                          | 具体输入与应核对事项                                                                                       | 当前明确的边界                                                                                                      |
| --------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| API / Judge 权限分离                          | [架构](../architecture/ARCHITECTURE.md)、apps/api、internal/httpapi、services/judge-worker、internal/judge | API 未导入执行适配器；Worker idle。Phase 2 同步 bridge 必须独立进程，尚无 bridge 实现                               |
| Submission / SourceUnit / Artifact / Run 身份 | [领域模型](../architecture/DOMAIN_MODEL.md)、internal/domain/model.go                                      | SourceUnit 为请求快照，Artifact 按 Run 关联；检查显式身份能否支持 rejudge，DB FK / unique / delete 规则仍待 Phase 2 |
| 生命周期与终态、取消 / 重投                   | [状态机](../judge/STATE_MACHINE.md)、internal/domain/state.go、state_test.go                               | CanTransition 只检查边方向，未持久化 verdict 组合 / fence；检查文档是否清楚划分未来约束                             |
| ProcessSpec / Binding / Graph                 | internal/execution/graph.go、internal/sandbox/sandbox.go                                                   | SourceUnitID / ArtifactID 显式绑定；可信管理进程与 Phase 7–8 细节待设计，当前不允许执行                             |
| 资源单位与计量                                | 领域模型、ResourceLimit / ResourcePolicy、ExecutionResult                                                  | ms / bytes / PIDs、角色 / 聚合 / 编译 / overhead 分离；字段存在不代表已有验证器或强制计量                           |
| 原始执行结果与 verdict                        | Sandbox Runtime / Judge Engine 接口、[安全基线](../sandbox/SECURITY_BASELINE.md)                           | Runtime 负责完整清理与原始证据；Engine 映射结果。没有不安全 fallback，也没有已验收 Sandbox                          |
| 事务 / outbox / fence / 可追溯                | 架构“事务与仓储边界”、领域模型、状态机                                                                     | 事务提交后 ACK、不改写 Run 历史是设计约束；具体 SQL / retry / lease 仍待对应阶段                                    |

若 reviewer 发现设计缺陷，应另建问题档案并遵循用户异常处理约定。不得为了关闭评审而缩小已承诺不变量或隐藏失败证据。

## 4. 基础工程与流程核验

- P0-001：检查 workspace / lock、localhost API、Monaco worker、本地缓存脚本、sqlc 基础查询、CI 命令。API smoke 仍须验证未配置 DB 503、提交路由不存在。
- P0-002：检查任务 ID / 认领 / 依赖 / 审阅 / 问题一事一档 / 停止规则与用户授权补充，确认 PROGRESS 为唯一实时状态。
- P0-005：检查 MinIO source-lock / 校验 / 镜像标签 / UID 10001 / healthcheck / 命名卷 / 冷启动依赖。原卷保留，初次创建 MinIO 卷不构成业务存储完成；镜像非生产基线。
- CI 配置已新增 MinIO 构建步骤，但当前没有远程运行证据；源码交叉构建不等于执行了 Linux Sandbox 测试。

## 5. 可执行复核与交付

在仓库根目录复核适用项；不要输出 `.env` / Compose 完整配置中的凭据：

```powershell
npm run check
npm run smoke
npm run doctor
npm run infra:config
node scripts/infra-smoke.mjs
node scripts/source-digest.mjs
git diff --check
```

Docker 只读复核可能需要主机执行权限；受限环境的 pipe 权限失败不能冒充主机 engine 不可用。现有 build / infra 证据可阅读交接，不要求没有新问题时反复重建镜像。

reviewer 交付实际署名的 REVIEW，指出每项通过 / 需修改 / 无法验证，给出窄范围版本冻结结论。项目负责人据证据确认 Phase 0；之后才可恢复 [P1-001](../tasks/P1-001-mini-judge-design.md) 的生命周期设计。原实现者不代替 reviewer 写通过结论。
