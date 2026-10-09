# ISSUE-2026-09-30-002：完整基础服务验收被 MinIO 不可用阻塞

- 日期：2026-09-30（Asia/Shanghai）
- 状态（当日）：阻塞，停止后续开发；**2026-10-01 用户授权修复后，固定源码镜像运行及完整 smoke 通过，B-ENV-002 已关闭**，详见 [后续修复](2026-10-01-minio-source-rebuild.md)。
- 关联：B-ENV-002 / P0-005；承接 [9 月 26 日交接](../handoffs/2026-09-26-infra-restart.md)。
- 记录人：本轮 Agent；处理负责人：用户。
- 本次仅复验现状，没有拉取 / 替换 MinIO 镜像、修改其 Compose 定义或改为其他存储。

## 背景与前因

9 月 26 日曾记录 MinIO 镜像拉取失败，交接中保留了用户“暂不处理 MinIO”的裁定。之后 Docker engine 连接失败，无法复验整个环境。

本轮用户授权修复 `5432` 冲突并重试 Docker。实际已确认 Linux engine `29.8.1` 可连接，并修复 OJ 数据库端口为 `15432`；PostgreSQL / Redis healthy，宿主机 API 数据库 readiness 为 200。随后执行完整基础服务 smoke，定位到仍未满足的验收项。

## 操作顺序与实际结果

| 操作                                      | 实际结果                                                   | 能说明什么                               |
| ----------------------------------------- | ---------------------------------------------------------- | ---------------------------------------- |
| `docker --context desktop-linux info`     | Server 29.8.1 / linux                                      | 当前已不是 engine 连接故障               |
| 项目 Compose `ps -a`                      | 只有 PostgreSQL / Redis，两者 healthy；无 MinIO 容器       | 对象存储未作为本项目服务运行             |
| 本机 Docker 镜像列表                      | PostgreSQL / Redis 等，未看到配置中的 MinIO 镜像           | 本机缺少该镜像；没有据此判定远程仓库状态 |
| `npm run doctor` / `npm run infra:config` | 均通过                                                     | 工具和配置解析正常，不代表所有服务健康   |
| `node scripts/infra-smoke.mjs`            | exit 1，失败于 `scripts/infra-smoke.mjs:40` 的 MinIO fetch | 完整 P0-005 验收未通过                   |

该脚本先执行 PostgreSQL `SELECT 1`、Redis `PING` 并校验结果，再请求 MinIO。此次已到达 MinIO fetch，前两项校验通过；此前单独执行也得到 `1` 和 `PONG`。

脱敏失败原文：

```text
TypeError: fetch failed
cause: connect ECONNREFUSED 127.0.0.1:9000
code: ECONNREFUSED
address: 127.0.0.1
port: 9000
```

## 已知原因与尚未确认的原因

直接失败条件是本机 MinIO 健康端点不可连接，且没有本项目 MinIO 容器。当前配置仍为 `minio/minio:RELEASE.2025-09-07T16-13-09Z`，S3 端口 `9000`、控制台 `9001`。

历史交接曾将拉取失败归因于外部镜像来源变化；本轮**没有重新请求远端 registry**，因此不把旧文“官方移除”等描述当作本轮确认的外部事实。当前还不能区分镜像标签 / 来源 / 访问权限 / 网络等远程原因，也没有新的 pull 错误原文。`docker login` 或更换镜像是否有效，均未验证。

## 后果与停止位置

- `5432` 修复成功，B-ENV-003 关闭；Docker 连接成功，B-ENV-001 关闭。PostgreSQL 保留原命名卷并运行，Redis 保持运行。
- B-ENV-002 保持打开；P0-005 仍 blocked，Phase 0 不能完成环境验收。
- 全量 `infra:up`、S3 联调、下一阶段实现尚未完成。本次没有再次运行完整 up 去触发同类拉取失败。
- 按用户“一事一档、交由用户处理、阻塞则停止”的要求，结束必要记录后停止。没有利用数据库已恢复绕过此停止要求；没有继续 P1-001 或前端实现。

## 用户接手事项与恢复条件

1. 先裁定对象存储镜像来源：检查现有标签能否拉取，或明确批准经过核验的镜像来源 / 本地构建方案。凭据留在本机，不写入问题报告。
2. 若采用其他 S3 实现，需要记录技术裁定 / ADR，并核对 S3 行为、健康检查、端口和持久化目录；不能只替换镜像名就宣布兼容。
3. 在仓库根目录按裁定后的配置执行 `npm run infra:config`、`npm run infra:up`、`node scripts/infra-smoke.mjs`，要求三个依赖健康且 smoke exit 0。
4. 处理后通知 Agent 继续。复验并记录实际结果后，才能关闭 B-ENV-002 / 推进 P0-005；独立架构评审仍需单独完成。

本轮端口修复和全部验证的交接见 [9 月 30 日交接](../handoffs/2026-09-30-port-docker-validation.md)。
