# 交接：重复构建修复与工程复验

- 日期：2026-10-01（Asia/Shanghai）；执行者：根 Agent / Builder。
- 授权：独立评审交付后，用户再次要求“尝试修复问题并继续推进工程”。
- 任务：P0-005；当前提交基线 `7a952c6`，分支 `task/P0-005-minio-source-2026-10-01`；保留全部已有未提交内容，未提交 / 推送。
- 代码 / 配置摘要：56 文件，SHA-256 `469c5c38da2a199f64c6e1e04d16bce4a3e7493145af1e72f97e43154a86bc00`；MD、缓存和凭据排除。
- 工程独立复审：独立 `/root/phase0_recheck` 已通过，见 [正式复审](../reviews/2026-10-01-phase0-recheck.md)；作者检查不冒充 REVIEW。

## 改动与原因

`scripts/build-minio.mjs` 不再用 copyFileSync 传播 module cache 的只读属性。对于已有构建输出，仅将固定目标 `.local/minio-build/LICENSE` 改为可写，然后读取上游内容、写入本项目输出，mode=0644。没有修改源码 cache、许可证内容、来源锁、Go / npm 依赖或 Docker 数据卷。

`scripts/setup.mjs` 完成提示补齐 doctor → infra:build-minio → infra:up → infra-smoke。没有重新运行 setup 或改变 `.env`，也没有为提示改动重新安装依赖。README 移除已过时的“无 remote”声明，指向实时台账；不宣称远程 CI 已通过。

另整理 [七项基础不变量契约](../architecture/FOUNDATION_INVARIANTS.md)，严格取自已通过的独立原则审阅；已独立复审并窄范围 frozen，完整核心类型继续 draft。

## 实际验证

| 验证                                     | 结果与边界                                                                                                                                         |
| ---------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| Docker 恢复                              | 初始 Linux engine 管道缺失；已知问题范围内有界 start 成功，Server 29.8.1 / linux / x86_64                                                          |
| 连续两次 infra:build-minio               | 同一已有 `.local/minio-build` 下两次 exit 0；第二次实际覆盖已存在的许可证                                                                          |
| LICENSE 属性与 SHA-256                   | 输出 IsReadOnly=false，源缓存 IsReadOnly=true；两份均 34523 bytes，SHA-256 均为 `0D96A4FF68AD6D4B6F1F30F713B18D5184912BA8DD389F86AA7710DB079ABCB0` |
| doctor / infra:config / infra:up         | 均 exit 0；仅 MinIO 因镜像更新重建，三项服务 healthy                                                                                               |
| infra-smoke                              | PG SELECT 1、Redis PING、MinIO liveness 均 PASS                                                                                                    |
| 真实 `.env` 的 API readiness             | HTTP 200 / status=ready，数据库 127.0.0.1:15432；临时 API 端口 41271，检查后已停止                                                                 |
| npm run check                            | PASS：格式、Go vet / ESLint、Go 单元、Go / TS / Vite 构建；现有 Monaco 警告仍在原问题档案内                                                        |
| npm run smoke                            | PASS：健康 / metadata、未配置 DB 的 readiness 503、提交路由不存在；检查后临时 API 已停止                                                           |
| 健康探针 Go vet、两个 JS 的 node --check | exit 0                                                                                                                                             |
| 实际 MinIO 身份 / 来源 / 卷              | UID/GID 10001:10001，revision `7aac2a2c5b7c882e68c1ce017d8256be2feea27f`，`oj-dev_minio-data:/data`                                                |
| 原 PG / Redis 卷                         | `oj-dev_postgres-data` / `oj-dev_redis-data` 继续挂载，没有删卷或初始化数据库                                                                      |
| 运行平台只读信息                         | Linux kernel `6.6.87.2-microsoft-standard-WSL2`，cgroup version 2，driver cgroupfs；不证明已授权 / 已测试 runner 的 cgroup 操作                    |

本轮未运行 Windows race、远程 CI、冷机器初始化、S3 业务持久化、旧 MinIO 数据格式回退、Judge / 队列 / Linux 隔离测试。不用本地通过代替上述证据。

## 当前门禁与后续

构建失败原因得到修复后的连续复验支持，独立 reviewer 已核查并通过；B-BUILD-001 与两项入口说明问题已关闭，七项原则按审阅范围采用，[阶段验收](../reviews/2026-10-01-phase0-acceptance.md) 通过。完整核心草案没有自动冻结；允许进入 P1-001 设计。

后续 P1-001 只产出可实现的生命周期设计、fixture 规范和 Builder / Worker 任务包；当前没有 C++ runner 环境版本锁、CLI 或真实判题实现。本轮前端仍仅交付 MD。
