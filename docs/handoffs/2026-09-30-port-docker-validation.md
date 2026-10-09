# 交接：PostgreSQL 端口修复与 Docker 基础服务复验

- 日期 / 执行者 / 角色：2026-09-30 / 本轮 Agent / Builder（开发环境维护）。
- 基线：`7a952c6`；当前分支：`task/P1-001-architect-handoff`；本轮改动未提交 / 未推送。
- 相关任务：P0-005；B-ENV-001 与 B-ENV-003 已复验关闭，B-ENV-002 仍阻塞。
- 用户授权：明确修复 `5432` 冲突，再验证 Docker 并继续操作；其他阻塞仍按用户接管规则处理。前端实际实现未授权，前端详细 MD 保持此前交付。
- 代码 / 配置摘要：52 文件，SHA-256 `a0822ccf5331c1bbf5b2b2969106ea4eeafeb44bf4049c26fd95ebf8bf223ce7`；不含 `.env`、MD、缓存和构建产物。

## 已交付

1. 查明 `5432` 位于 Windows 排除范围 `5384–5483`，绑定测试返回 10013；`15432` 绑定测试通过。
2. `.env.example`、Compose fallback 和本机 `.env` 的 PostgreSQL 宿主机端口同步到 `15432`，连接串端口同步；容器内部保持 `5432`。本机凭据保留且不打印，`.env` 仍被 Git 忽略。
3. 仅重建本项目 PostgreSQL，原命名卷继续挂载；PostgreSQL / Redis 均运行且 healthy。没有停止其他项目服务或删除卷。
4. 环境说明更新端口和旧副本同步方法；AGENTS / DEVELOPMENT 记录本次明确授权；任务卡、PROGRESS 和独立问题档案更新。

## 验证与实际结果

| 检查                                                                     | 结果                                                | 边界                                                                     |
| ------------------------------------------------------------------------ | --------------------------------------------------- | ------------------------------------------------------------------------ |
| `docker --context desktop-linux info`                                    | Server 29.8.1 / linux                               | 使用获得主机执行权限的上下文，受限环境的 pipe 权限错误不作为主机故障结论 |
| `config --quiet` / `npm run infra:config`                                | 通过                                                | 解析成功，不输出配置凭据                                                 |
| Compose `up -d --no-deps --pull never --wait --wait-timeout 60 postgres` | 通过，healthy                                       | 没有拉取镜像或启动 MinIO                                                 |
| 实际端口 / 挂载检查                                                      | `127.0.0.1:15432->5432/tcp`；`oj-dev_postgres-data` | 命名卷保留；未做全量数据审计                                             |
| PostgreSQL `SELECT 1` / Redis `PING`                                     | `1` / `PONG`                                        | 容器内部验证                                                             |
| 真实 `.env` + 临时 API `/readyz`                                         | HTTP 200，`{"status":"ready"}`                      | 证明宿主机连接新数据库端口并完成认证；仍不代表 OJ 业务可用               |
| `npm run doctor`                                                         | 通过                                                | Node、Go、Git、Compose、Docker 和本地依赖文件可用                        |
| `node scripts/infra-smoke.mjs`                                           | **失败**：MinIO `ECONNREFUSED 127.0.0.1:9000`       | PostgreSQL / Redis 校验已通过，完整环境验收未通过                        |
| 前端实现 / 全套 check / race / 远程 CI / 独立审阅                        | **本轮未执行**                                      | 没有业务源码或依赖改动，不把历史结果当本轮证据                           |

临时 API 端口为 `47443`，已在验证结束后停止；开发启动仍使用 `npm run api`。数据库客户端改用 `127.0.0.1:15432`。验证辅助脚本放在工作区 `work/`，不纳入项目正式测试。

必要文档检查已完成：本轮修改的 MD / YAML 使用 Prettier 格式化；50 份 MD 的本地文件链接检查无缺失（未检查锚点和远端网页）；`git diff --check` 通过；`git check-ignore .env` 确认本机凭据文件仍被忽略。这些检查不等于项目全套 `npm run check`。

## 阻塞与停止位置

端口和 Docker 连接问题已解决，详见 [端口修复档案](../issues/2026-09-30-postgres-port.md)。MinIO 仍未运行，独立记录见 [MinIO 问题档案](../issues/2026-09-30-minio-unavailable.md)；本轮没有复验远端镜像来源，因此不沿用历史外部根因作为新事实。

按用户“阻塞则停止”的要求，停止后续开发，仅完成必要 MD 记录和文档检查。P0-005 仍 blocked，Phase 0 尚未验收；P0-004 独立架构评审与 P1-001 Mini Judge 设计没有完成。

## 下一位执行者的第一步

等待用户处理对象存储并通知继续，阅读 PROGRESS / MinIO 问题档案，然后完整复验 `infra:up` 与 infra-smoke。不要重置 Docker、删卷或反复绑定 `5432`。

前端继续以 [框架实施指南](../frontend/IMPLEMENTATION_GUIDE.md) 和 [Agent 任务包](../frontend/AGENT_TASKS.md) 为交接。当前只授权 MD 交付；实际实施前须明确授权并完成 P2-100 评审。不要重复创建 Vite 工程，不把 mock 验收当真实后端验收。
