# 交接：Docker 基础服务启动与 P0-005 部分验证

- 日期 / 执行者 / 角色：2026-09-26 / 本次会话 Agent / Builder（环境维护）
- 当前基线：未提交工作区；本次仅改动 PROGRESS.md 与本文件
- 相关任务与契约：P0-005；阻塞 B-ENV-001（已解除）、B-ENV-002（新增）

## 已交付

- 启动 Docker Desktop，Linux engine 就绪（Docker 28.4.0、Compose v2.39.4）。
- 按 ENVIRONMENT.md 执行 `npm run doctor`、`npm run infra:config`，均通过。
- 拉取 `postgres:17.6-alpine`、`redis:7.4.5-alpine` 并启动，两容器 healthy。
- PostgreSQL `SELECT 1`、Redis `PING` 通过；API 启动后 `/healthz=200`、`/readyz=200`、`/api/v1/system=200`，随后已停止 API 进程。
- 未改动 `deploy/docker-compose/compose.yaml` 与 `.env`，未删除或改动其他项目的容器/数据卷。

## 验证命令与实际结果

| 命令 / 操作                      | 实际结果                                                                                 |
| -------------------------------- | ---------------------------------------------------------------------------------------- |
| npm run doctor                   | 通过（git / go / compose / docker info / .env）                                          |
| npm run infra:config             | 通过                                                                                     |
| npm run infra:up                 | **失败**：minio 拉取被拒，postgres / redis 未被该次命令启动                              |
| docker compose up postgres redis | 通过：两容器 healthy                                                                     |
| node scripts/infra-smoke.mjs     | **失败于 MinIO**：PG `SELECT 1`、Redis `PING` 已通过，随后 `ECONNREFUSED 127.0.0.1:9000` |
| API /healthz /readyz /system     | 200 / 200 / 200                                                                          |
| npm run infra:down               | 未执行（服务按需保留运行）                                                               |

失败原因为外部依赖变更，不是本机配置：`minio/minio`（含 `minio/mc`）仓库已被 MinIO 从 Docker Hub 移除，社区版源码仓库亦已归档，`docker login` 无法解决。

## 未完成 / 阻塞与解除条件

- MinIO 未启动，P0-005 验收未满（要求三项健康检查全部通过）。登记为 B-ENV-002。
- 用户裁定：暂不处理 MinIO，保持 compose.yaml 原样。
- 解除条件：项目裁定镜像来源（中转镜像源 / 替代 S3 兼容存储 / 本地构建）后，重跑 `npm run infra:up` 与 `node scripts/infra-smoke.mjs`。
- 修改 deploy 属于需 Builder 审阅的范围；若更换对象存储技术栈，按 DEVELOPMENT.md 需新增 ADR。

## 下一位执行者的第一步

读根 PROGRESS.md（B-ENV-002）→ 认领并记录执行者 → 处理 MinIO 镜像来源决策 → 完成 infra-smoke 三项全通过后再把 P0-005 从 blocked 推进。

## 禁止误解

- PG / Redis 可用与 API `/readyz=200` 只证明数据库连通，不代表 MinIO、队列、判题或业务表可用。
- 当前容器是本项目开发依赖（仅绑定 127.0.0.1），不是已验收的选手 Sandbox，也不是生产部署。
- 不要用 `infra:down -v` 或删除数据卷来"修复"启动问题；本机另有其他项目的容器（mental-health-agent），不得删除或改动。
