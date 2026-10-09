# ADR-0002：本地开发 MinIO 使用固定官方源码构建

- 日期：2026-10-01。
- 状态：本地开发方案已采用，首次运行验收通过；独立审阅认可 localhost 来源方案，但构建脚本需修改；不是生产存储裁定。
- 任务：P0-005 / B-ENV-002。
- 作者：本轮 Agent；授权：用户要求尝试修复现有问题并继续推进工程。

## 背景与证据

当前 Compose 使用 `minio/minio:RELEASE.2025-09-07T16-13-09Z`，本机没有该镜像 / MinIO 容器。2026-10-01 恢复 Docker 后核验：Docker Hub manifest 请求连接超时；官方 Quay 同版本 HEAD 返回 401。两种错误不同，不能据此推断全球镜像仓库均已删除，也不能把本次失败归因于本机 MinIO 凭据。

MinIO [官方仓库](https://github.com/minio/minio) 已归档，README 给出了从源码构建的方式。当前上游 [Dockerfile](https://github.com/minio/minio/blob/7aac2a2c5b7c882e68c1ce017d8256be2feea27f/Dockerfile) 依赖 `minio/minio:latest`，不能用于解决本机无法取得该镜像的循环依赖。

## 决策

1. 保留 MinIO 技术栈、S3 / Console 端口、现有凭据和 `minio-data` 持久化定义；不替换为新的 S3 产品。
2. 固定官方提交 `7aac2a2c5b7c882e68c1ce017d8256be2feea27f`，对应 Go module `v0.0.0-20260212201848-7aac2a2c5b7c`。使用 Go 校验数据库验证模块，记录 module sum / go.mod sum，不修改上游源码。
3. `npm run infra:build-minio` 使用本机 Go 交叉编译静态 `linux/amd64` 二进制；源码 / 依赖 / 产物存于被忽略的 `.local/`。不把 MinIO 加进 OJ 的 `go.mod`。
4. 从 scratch 构建本地镜像 `oj/minio:source-7aac2a2c5b7c-linux-amd64`；仅从固定摘要的官方 Redis 镜像复制 CA 文件，不运行 Redis。当前机器已有该镜像，已只读确认 CA 文件存在。
5. 镜像以 UID/GID `10001:10001` 运行；独立静态探针请求 `/minio/health/live`，不依赖 shell / curl。保留上游 AGPL LICENSE 并标记源码仓库与提交。
6. 本地构建元数据记录编译器版本、源码校验值、二进制 SHA-256 和镜像引用；实际构建结果以交接为准。固定源码与校验值不等于已验证跨编译器逐字节一致。

## 备选与限制

继续使用原官方预编译镜像需要先恢复 registry 可访问性；当前请求未成功。第三方重打包、替代 S3 实现和主机直接运行 MinIO 均未采用。本方案在已有 Go / Docker 环境下保留服务边界，并使构建过程可检查。

首次编译会下载较多模块；新机器还需取得固定 CA 来源镜像。现有源码仓库已归档，不能宣称持续获得上游安全修复。因此本方案限定 localhost 开发用途；生产对象存储选型、维护与安全审阅须另行裁定，不由本次环境修复替代。

## 兼容、数据与回退

端口仍为 `9000/9001`，数据库宿主机端口仍为 `15432`。Compose project 和卷名保持一致；不运行 `down -v` / prune，不删现有卷。此处没有 OJ 业务 schema / bucket 迁移。

若恢复官方镜像来源，可将 Compose image / healthcheck 恢复到经验证的定义并重新创建 MinIO，保持卷挂载；降级前必须确认存储格式兼容并备份数据，不能在已写入新格式的卷上盲目回退。本轮不执行降级。

## 验收与审阅

要求固定源码编译、镜像构建、非 root 用户 / 挂载核对、完整 Compose `up --wait`、PG SELECT 1 / Redis PING / MinIO live 全通过，并确认宿主机 API readiness。CI 配置跟随新的构建步骤，但远程执行结果单独记录。

审阅者应检查来源锁定、校验数据库、scratch 文件集、卷权限、冷启动依赖和回退约束。没有独立审阅前不把本方案标为生产已冻结。

实际运行用户 `10001:10001`，MinIO / PG / Redis 全部 healthy，完整 infra-smoke 和真实数据库 API readiness 通过。运行二进制的 MinIO 自报版本为 `DEVELOPMENT.GOGET`（源码构建默认值）；来源以 OCI revision label、固定 module version / checksum 和构建元数据为证，不将其冒称某个官方预编译 release。

2026-10-01 独立 reviewer `/root/phase0_review` 的 [正式报告](../reviews/2026-10-01-phase0-review.md) 已交付：来源方案可用于 localhost 开发；LICENSE 重复复制 EPERM 阻塞工程验收，见 [问题档案](../issues/2026-10-01-minio-license-copy.md)。独立审阅不代表重复构建成功、Phase 0 通过或生产方案冻结。

## 修复后的独立复审

LICENSE 重复构建修复与服务复验通过，独立 [独立复审](../reviews/2026-10-01-phase0-recheck.md) 确认本地工程可验收；[阶段验收](../reviews/2026-10-01-phase0-acceptance.md) 已完成。本 ADR 采用范围仍为 localhost 开发，未转换为生产裁定。
