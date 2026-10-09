# P0-005：本机基础服务联通验证

- 实时状态：见 PROGRESS.md
- Owner 角色：Builder / 环境维护者
- 实际执行者：2026-10-01 本轮 Agent / Builder（Docker 恢复、MinIO 修复与完整复验）；历史尝试见交接
- 风险：R1
- 允许修改：本项目 Compose、scripts 和开发环境说明
- 禁止：重置宿主 Docker、删除数据卷或其他项目容器、更改生产服务；本次用户明确授权保留原命名卷重建 OJ PostgreSQL 以应用端口配置
- 依赖：Docker Linux engine 可运行

## 验收

```sh
npm run doctor
npm run infra:config
npm run infra:build-minio
npm run infra:up
node scripts/infra-smoke.mjs
```

PostgreSQL SELECT 1、Redis PONG、MinIO health 均通过；启动 API 后 /readyz 返回 200。失败时记录 engine/镜像拉取/端口/认证的实际阻塞项；不能用 Compose 语法通过代替容器启动成功。

结束时可 infra:down 停止本项目服务并保留命名卷。

## 2026-10-01 交付与当前验收

用户已授权尝试修复现有问题。Docker 有界启动成功；MinIO 使用固定官方源码交叉编译 / scratch 本地镜像，新增 source-lock、构建命令、静态健康探针和 CI 构建步骤。localhost 端口保持 15432 / 6379 / 9000 / 9001，PG / Redis 原卷保留；首次创建本项目 MinIO 卷。

`npm run infra:config`、`npm run infra:up`、完整 infra-smoke、真实 `.env` 的 API readiness 均通过，三项服务 healthy。MinIO 实际 UID/GID 10001，OCI revision 与来源锁定一致；`npm run check` 及健康探针 Go vet 通过。远程 CI、Windows race、业务 S3 / 判题隔离测试本轮未执行。

环境首次验收通过，曾进入 in_review；最终重复构建在 LICENSE 复制时 EPERM，现回到 **blocked**，见 [独立构建问题](../issues/2026-10-01-minio-license-copy.md)。运行服务和首次验收证据仍有效，不能用它们证明重复构建成功。独立 Agent 已获用户授权核查基础工程，但不修复新问题；未经解除不标为 done 或进入下一阶段。

首次运行证据见 [修复档案](../issues/2026-10-01-minio-source-rebuild.md) 和 [10 月 1 日交接](../handoffs/2026-10-01-environment-review-handoff.md)。用户授权的独立 reviewer `/root/phase0_review` 已交付 [正式报告](../reviews/2026-10-01-phase0-review.md)：工程需修改，R1 重复构建阻塞验收；独立只读复核运行服务、UID / 来源 / 卷与失败文件属性，不冒称独立重跑全部测试。解除条件是同一已有输出目录连续两次构建成功，随后服务 / 真实 API 复验与受影响工程复审；当前仍 blocked。

## 2026-10-01 再次授权修复

用户在独立评审交付后再次要求修复并继续。实际执行者为根 Agent / Builder；允许修改 `scripts/build-minio.mjs`、`scripts/setup.mjs`、README 与环境 / 评审 / 问题文档，保留源码缓存、凭据和命名卷。已授权的 reviewer 可复审受影响工程范围；连续两次构建、服务 / API 复验及真实独立复审完成前不关闭阻塞。

## 2026-09-30 范围与结果（历史）

用户明确授权解决 `5432` 冲突并继续连接 / 验证 Docker。本次允许修改 `.env.example`、本机 `.env`、Compose PostgreSQL 宿主机端口与环境 MD；保留之前的未提交内容及用户负责处理其他异常的约定。

- Docker Server `29.8.1 / linux`；doctor、Compose config 均通过。
- 宿主机端口修复为 `15432`，容器内部保持 `5432`。PostgreSQL 重建后 healthy，原卷 `oj-dev_postgres-data` 继续挂载；Redis healthy。
- PostgreSQL `SELECT 1` 和 Redis `PING` 通过。临时 API 使用真实 `.env`，`/readyz=200`，检查结束后已停止临时进程。
- 完整 infra-smoke 失败于 MinIO `ECONNREFUSED 127.0.0.1:9000`；任务仍 blocked。没有执行镜像替换或继续下一阶段实现。
- 修复自检不是 Builder 独立审阅，Review owner 待指派；没有全量单元 / race / 前端 / CI 新证据。

详见 [端口问题与关闭证据](../issues/2026-09-30-postgres-port.md)、[MinIO 独立阻塞](../issues/2026-09-30-minio-unavailable.md) 和 [本轮交接](../handoffs/2026-09-30-port-docker-validation.md)。用户解决 MinIO 并通知继续后，再完成本任务剩余验收。

## 最终验收（2026-10-01 再次授权后）

同一输出目录连续两次构建、doctor / config / up / infra-smoke / 真实 API readiness / check / smoke 通过。独立 [独立复审](../reviews/2026-10-01-phase0-recheck.md) 通过，原卷保留，已记录源码仍只读、许可证内容一致；[阶段验收](../reviews/2026-10-01-phase0-acceptance.md) 完成。本任务 done；远程 CI / race / S3 / 安全能力未测范围保持不变。
