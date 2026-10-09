# ISSUE-2026-10-01-001：MinIO 官方镜像不可取得，尝试固定源码重建

- 日期：2026-10-01（Asia/Shanghai）。
- 关联：B-ENV-002 / P0-005；接续 [9 月 30 日问题](2026-09-30-minio-unavailable.md)。
- 状态：**已修复并复验关闭 B-ENV-002**；方案为本地源码镜像，原 registry 访问问题未宣称修复。
- 执行者：本轮 Agent / Builder；用户本轮已明确授权尝试修复，不沿用此前“暂不处理 MinIO”作为禁止此次修复的条件。

## 前因与顺序

1. 保留 `7a952c6` 基线上的所有未提交改动，数据库宿主机端口仍为 `15432`。
2. 新一轮检查发现 Docker Linux pipe 不存在；Desktop/backend 无进程，WSL 的 docker-desktop 停止。这是已知连接问题复现，并非端口修改失败。
3. 通过 `docker desktop start --timeout 60` 有界启动，返回成功；Server `29.8.1 / linux`，PostgreSQL / Redis 随后 healthy，15432 映射仍在。未重置 Docker / WSL，也未更改主机服务启动类型。
4. 核验原 Docker Hub 固定版本 manifest：TCP 请求超时；核验官方 Quay 同版本：HEAD 401 Unauthorized。没有运行 docker login 或更换账号。
5. 读取官方源码元数据，master 指向固定提交 `7aac2a2c5b7c882e68c1ce017d8256be2feea27f`；Go 下载得到对应 pseudo-version 和校验元数据。
6. 建立源码锁定、本地构建脚本、scratch Dockerfile 与静态健康探针；固定源码编译与镜像构建成功，Compose 已切到本地镜像。

## 脱敏失败

```text
Docker Hub:
HEAD registry-1.docker.io/v2/minio/minio/manifests/RELEASE.2025-09-07T16-13-09Z
dial tcp ...:443: connectex: connection attempt failed / timed out

Quay:
HEAD quay.io/v2/minio/minio/manifests/RELEASE.2025-09-07T16-13-09Z
401 Unauthorized
```

以上能证明本次未取得官方 manifest，不能证明 Docker Hub 根因是标签删除；Quay 401 也不等于 OJ `.env` 中的 MinIO 管理凭据错误。此前“官方移除”的历史推断不作为本轮确定原因。

## 修复方案与影响

保留 MinIO / S3 技术栈、localhost 端口与持久化定义，用固定官方提交构建开发镜像。完整决策见 [ADR-0002](../adr/0002-local-minio-source-image.md)，操作见 [构建说明](../../deploy/minio/README.md)。不会修改 OJ 的 Go 依赖或前端源码。

该方案依赖 Go 模块下载和固定证书来源镜像；当前机器已有所需 CA 来源缓存。它解决本机预编译 MinIO 镜像不可取得的路径，不能宣称修复了所有 registry 网络问题。源码已归档，仅供本地开发，生产需另行裁定。

## 验证与恢复判据

| 检查                              | 实际结果                                                                                      |
| --------------------------------- | --------------------------------------------------------------------------------------------- |
| 固定源码                          | `v0.0.0-20260212201848-7aac2a2c5b7c`，Go checksum 下载成功                                    |
| module sum                        | `h1:/qdoYujxF5MBLDMVLeiYJ/7fBa1HikX39Dvzb9Ji8pA=`                                             |
| go.mod sum                        | `h1:yCWDkwWO9IWpGsT4mreDDN/B/QVmK2zC666uInRAcqE=`                                             |
| Go 交叉编译                       | go1.25.4，linux/amd64、CGO disabled；上游代码未修改                                           |
| MinIO binary SHA-256              | `ab90d4d794fc5826df7766ce54f02c099f75b76550ec26847e4a76b8b7a741cf`                            |
| 首次 Docker build                 | 成功；出现未设 ARG 默认值的非阻塞提示。已在本次构建配置中补固定默认值，再构建无该提示         |
| 运行用户与卷                      | UID/GID 10001；`oj-dev_minio-data:/data`；PG / Redis 仍挂载原命名卷                           |
| `npm run infra:up`                | exit 0，三项服务 healthy；PG 保持运行，Redis 由 Compose 重新创建并保留原卷，首次创建 MinIO 卷 |
| `node scripts/infra-smoke.mjs`    | exit 0；PG SELECT 1 / Redis PING / MinIO liveness 全通过                                      |
| 真实 `.env` + 临时 API            | `/readyz=200`、`{"status":"ready"}`；临时端口 50428，验证后已停止                             |
| `npm run check` / 健康探针 Go vet | 均通过；前端既有 chunk 警告单独记录                                                           |

镜像自报 `DEVELOPMENT.GOGET` 是上游源码构建默认值；固定来源见 OCI revision label 和 `.local/minio-build/build-metadata.json`，不混同原预编译 2025-09-07 版本。未验证生产安全 / 业务 S3 持久化 / Linux Judge 隔离 / 远程 CI。

B-ENV-002（运行不可用）关闭；最终重复构建发生新的 [LICENSE 复制 EPERM](2026-10-01-minio-license-copy.md)，B-BUILD-001 打开，主 Agent 停止修复 / 下一阶段推进，P0-005 blocked。用户随后明确授权独立只读评审，详见 [评审授权与门禁](2026-10-01-independent-review-gate.md)。首建健康成功不能替代重复构建通过；没有删卷、重置主机或修改前端。
