# 交接：环境修复、工程回归与独立评审结果

本文件保留首次修复 / 失败的历史记录。用户随后再次授权，最新 [修复交接](2026-10-01-license-fix-validation.md)、[独立复审](../reviews/2026-10-01-phase0-recheck.md) 与 [阶段验收](../reviews/2026-10-01-phase0-acceptance.md) 已解除该构建阻塞；实时状态见 PROGRESS。

- 日期 / 执行者 / 角色：2026-10-01 / 本轮 Agent / Builder（环境修复），Architect（评审输入整理，非独立 reviewer）。
- 基线：`7a952c6`；分支 `task/P0-005-minio-source-2026-10-01`（从原工作区建立，保留全部未提交内容）；本轮未提交 / 推送。
- 代码 / 配置摘要：56 文件，SHA-256 `95e401d13d7692a1d1c2f5872299ee7ada68226a583ad434233ad9dcb95d584f`；MD 与本地缓存 / 凭据不计入摘要。
- 任务：P0-005 首建验收通过，但最终重复构建失败，当前 blocked；P0-004 独立报告已交付，进入 in_review 等待结论确认 / 窄范围冻结；完整核心仍 proposed。

## 已交付

1. `docker desktop start --timeout 60` 恢复 Linux engine 29.8.1；数据库宿主机 `15432` 端口保持有效。
2. MinIO 官方预编译镜像 manifest 不可取得，改为固定官方提交 `7aac2a2c5b7c882e68c1ce017d8256be2feea27f` 的开发源码镜像；保留技术栈 / 凭据 / 端口与持久化定义。
3. 新增 `deploy/minio/source-lock.json`、Dockerfile、构建说明、`scripts/build-minio.mjs`、静态健康探针及 `npm run infra:build-minio`；CI 新增对应构建步骤。源码和 CA 来源有固定摘要，Go checksum 保持启用，上游许可证随镜像保留。
4. MinIO 以 UID/GID 10001 运行；PG / Redis 原卷继续挂载，MinIO 首次创建本项目命名卷。所有端口仅绑定 localhost。
5. 根环境入口 / 授权 / 任务卡 / 台账更新；问题拆分成 MinIO 修复、前端非阻塞警告、独立评审门禁。整理 [P0-004 评审输入包](../reviews/P0-004-review-package.md)。

## 验证证据（首建与工程回归）

| 检查                                             | 实际结果                                                                            |
| ------------------------------------------------ | ----------------------------------------------------------------------------------- |
| 固定源码编译与 Docker build                      | 成功，Go1.25.4 → linux/amd64；构建元数据在 `.local/minio-build/build-metadata.json` |
| Compose config / 完整 up                         | 通过，三项服务 healthy                                                              |
| 完整 infra-smoke                                 | PASS：PG SELECT 1 / Redis PING / MinIO liveness                                     |
| 宿主机真实 `.env` + 临时 API                     | readiness HTTP 200；临时 API 已停止，基础服务保留运行                               |
| `npm run check`                                  | PASS：格式 / vet / ESLint / Go 测试 / Go + TypeScript + Vite build                  |
| `node scripts/go.mjs vet ./scripts/minio-health` | PASS；探针静态二进制已实际用于容器健康检查                                          |
| Go / npm 依赖锁文件                              | 未改变；MinIO 依赖在独立上游 module 与本地缓存，不注入 OJ go.mod                    |
| 独立 Architect + Builder 审阅                    | 已交付；基础不变量通过，工程需修改；未独立重跑作者的全套测试                        |
| 远程 CI / race / S3 业务持久化 / 判题隔离        | 本轮未执行，不宣称通过                                                              |

初次构建的 ARG 默认值警告已在本次 Dockerfile 中修正，后续镜像构建无该提示。前端现有 Monaco 2.7 MB chunk 警告仍存在，构建 exit 0，已单独记录；本轮没有前端源码改动。

随后 `npm run smoke` 通过，证明基础 API、未配置 DB 的 readiness=503 和不存在提交路由。最终重复 `npm run infra:build-minio` **失败**：LICENSE 复制 EPERM，源 / 输出均只读且 hash 相同；其后的 doctor / up / infra-smoke 未执行。此次失败不能被上表首建成功抵消，详见 [新阻塞](../issues/2026-10-01-minio-license-copy.md)。

MinIO 自报 `DEVELOPMENT.GOGET`，来源应核对 OCI revision label、module version / sum 和 binary SHA-256，不冒称原官方 release。固定源码仓库已归档，限定本地开发；生产存储需另行裁定，见 [ADR-0002](../adr/0002-local-minio-source-image.md)。

## 当前停止位置

运行环境恢复并首次验收通过，最终复验出现新的 [许可证重复复制阻塞](../issues/2026-10-01-minio-license-copy.md)，不将其混同原 MinIO 服务不可用。主 Agent 已按用户规则停止修复 / 下一阶段推进，仅核实失败和记录原因。

用户随后在确认框明确授权独立 Agent。`/root/phase0_review` 已只读核查核心契约与基础工程并交付 [正式独立报告](../reviews/2026-10-01-phase0-review.md)，这是此次另行明确授权的交付；reviewer 未修复新问题或绕过它。相同委派授权保留，恢复后不再请求相同确认。

七项基础原则审阅通过，只建议窄范围冻结；完整 Graph / Runtime / 资源计量 / 事务接口仍是 draft。本轮没有创建冻结记录或改动契约版本。基础工程结论为“需修改”：重复构建 LICENSE EPERM 阻塞 P0-005 与 Phase 0 验收。独立只读复核确认三项容器 healthy、MinIO 的 UID / 来源 / 卷正确、代码配置摘要一致；不将这些复核冒称全套测试独立重跑。

评审发现的 [README remote](../issues/2026-10-01-readme-remote-status.md) / [setup 初始化提示](../issues/2026-10-01-setup-minio-instruction.md) 不一致，已分别归档，由用户安排处理。

用户解除构建问题后，须验证同一输出目录下连续两次构建成功，再完成完整服务 / API 验证、受影响工程范围复审、明确列项的基础不变量冻结及 Phase 0 验收，才能恢复 P1-001。后续设计边界与未冻结项见独立报告第 3–4 节。

前端仍按 [实施指南](../frontend/IMPLEMENTATION_GUIDE.md) / [任务包](../frontend/AGENT_TASKS.md) 交接，实际实施范围需要另行明确，不能从此次修复授权推定已经开始搭建。
