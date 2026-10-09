# P0-001：工程基础与开发环境

- 实时状态：见 PROGRESS.md
- Owner 角色 / 实际执行者：Builder + Worker / 本次初始化 Agent
- Review owner：独立 Builder reviewer `/root/phase0_review`（2026-10-01）
- 输入：归档的原始计划；ADR-0001 工程范围
- 风险：R1；核心类型草案另由 P0-004 评审
- 允许修改：apps、services/judge-worker、internal/httpapi、platform、scripts、deploy/docker-compose、CI 和根工程配置
- 禁止：真实执行选手代码、开放外部服务、未经评审的业务表、生产部署
- 依赖：本机 Go / Node / Docker engine

## 输出

可启动 HTTP API、独立空闲 Worker、React/Vite/Tailwind/shadcn/TanStack/Router/Monaco 基础；OpenAPI、pgx/sqlc；Compose、跨平台命令、依赖锁文件和 CI。

## 验收

| 操作                           | 预期                                               | 证据                                                                                              |
| ------------------------------ | -------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| npm run check                  | 格式/lint/测试/前后端构建通过                      | docs/development/VALIDATION.md                                                                    |
| npm run smoke                  | API 能启动；未配置 DB 不假报 ready；无提交执行路由 | 同上                                                                                              |
| npm run sqlc                   | 生成代码可编译                                     | 同上                                                                                              |
| npm run infra:up + infra-smoke | 三项真实依赖可用                                   | 同上；未完成时拆到 P0-005                                                                         |
| Builder Review                 | 确认工程边界和复现路径                             | [2026-10-01 独立报告](../reviews/2026-10-01-phase0-review.md)；原骨架基础范围通过，新增构建需修改 |

## 回退

停止本项目进程，infra:down 保留数据卷；回退配置版本，不删除他人的代码、.env 或持久数据。

## 2026-10-01 独立审阅

用户授权的 `/root/phase0_review` 已独立阅读骨架、流程、配置和 CI，并只读核对实际服务。原骨架基础范围通过；新增 MinIO 构建存在 [重复复制 LICENSE 阻塞](../issues/2026-10-01-minio-license-copy.md)，整体基础工程结论需修改。当前 check / smoke 通过证据与未测范围见 [交接](../handoffs/2026-10-01-environment-review-handoff.md)；本任务维持 in_review，服务构建复现与阶段最终验收仍未完成。

## 最终验收（2026-10-01）

此前 in_review 为历史状态。重复构建已修复并连续验证，独立 [独立复审](../reviews/2026-10-01-phase0-recheck.md) 通过，[阶段验收](../reviews/2026-10-01-phase0-acceptance.md) 完成。本任务工程骨架范围 done，不代表真实判题、产品前端或远程 CI 已完成。
