# ISSUE-2026-10-01-004：MinIO 重复构建无法覆盖只读 LICENSE

- 日期：2026-10-01（Asia/Shanghai）。
- 状态：**2026-10-01 修复复验及独立复审通过，关闭**；失败与未测范围保留。
- 关联：B-BUILD-001 / P0-005；执行者：本轮 Agent；问题处理负责人：用户。
- 当前服务状态：首次构建的 MinIO / PG / Redis 已健康运行，完整健康 smoke 与数据库 API readiness 已通过；本问题不等于运行服务重新宕机。

## 1. 前因、操作与后果

本轮用户授权修复已知 MinIO 阻塞。固定官方源码首次编译、LICENSE 复制和 Docker build 成功，随后三个依赖 healthy，`npm run check` 和 `npm run smoke` 均通过。

为了验证最终构建脚本的缓存复现路径，再执行 `npm run infra:build-minio`。源码下载 / 两个 Go 二进制构建已完成，但脚本在 `scripts/build-minio.mjs:97` 复制许可证时 exit 1。其后的 Docker build / doctor / Compose up / infra-smoke **未执行**；不能把上一轮成功或本次其他回归通过当作重复构建通过。

脱敏原文：

```text
Error: EPERM: operation not permitted, copyfile
  .local/go/pkg/mod/github.com/minio/minio@v0.0.0-20260212201848-7aac2a2c5b7c/LICENSE
  -> .local/minio-build/LICENSE
at scripts/build-minio.mjs:97
Node.js v22.19.0
```

## 2. 已核实事实与原因边界

仅做只读核实，没有修改文件属性或重试覆盖：

| 文件                         | 实际属性 / 内容                                                         |
| ---------------------------- | ----------------------------------------------------------------------- |
| Go module cache 原始 LICENSE | `IsReadOnly=true`，Attributes=33（ReadOnly + Archive），34523 bytes     |
| 本地构建输出 LICENSE         | `IsReadOnly=true`，Attributes=33，34523 bytes                           |
| 两者 SHA-256                 | 同为 `0D96A4FF68AD6D4B6F1F30F713B18D5184912BA8DD389F86AA7710DB079ABCB0` |

最有证据的解释：首次 `copyFileSync` 使构建输出保留只读属性，重复构建尝试覆盖该输出时被 Windows 拒绝。未做清除属性后的排除性试验，也未检查所有文件锁 / ACL，因此把该原因标为高可信解释，而非已复验的修复结论。

这是新构建脚本在已有输出目录下的幂等性缺陷，不是 Go 源码校验失败、MinIO root 凭据错误或 Docker 引擎故障。原 Go module cache 保持只读是正常行为，不应通过修改上游缓存来“修复”。

## 3. 首次失败时的停止位置与评审授权（历史）

P0-005 从暂时 in_review 改回 blocked；运行服务维持现状，没有删卷 / reset / chmod / 再次构建。MinIO运行不可用的 B-ENV-002 已解决，但新的 B-BUILD-001 尚未解决。

用户在确认框中随后明确授权独立 Agent 评审，因此本轮仅另行完成这项已授权的只读评审：reviewer 可以核查现有契约和失败证据，不能修复此问题或绕过它进入下一阶段。这项授权已保留，恢复后不需再次请求委派确认。

独立 reviewer `/root/phase0_review` 已交付 [正式报告](../reviews/2026-10-01-phase0-review.md)，只读确认源 / 输出 LICENSE 均只读且内容 hash 一致，并将本问题列为 R1 阻塞：基础工程需修改，Phase 0 不可验收。Reviewer 未试修、重建镜像或启停服务，因此没有新增修复通过证据。

## 4. 给用户的处理方案

以下为首次发现时的建议，**当时未执行**；后来再次授权后的实际修复见第 5 节：

1. 只清除本项目构建输出 `.local/minio-build/LICENSE` 的只读属性，不修改 `.local/go/pkg/mod`。在仓库根目录可使用：

   ```powershell
   $buildLicense = Get-Item -LiteralPath '.local/minio-build/LICENSE'
   $buildLicense.IsReadOnly = $false
   ```

2. 修订构建脚本的许可证输出方式：已有输出时处理其可写权限，再读取原许可证内容写入构建输出；新输出使用可写文件模式，避免再次把 module cache 的只读属性复制过去。保持上游 LICENSE 内容和 checksum 不变。
3. 仅手动清属性是暂时措施；现有 `copyFileSync` 再次传播属性后，下次仍可能失败。必须验证最终脚本的连续执行。
4. 在同一已有 `.local/minio-build` 下连续执行两次 `npm run infra:build-minio`，都要求 exit 0，再执行 `npm run infra:config`、`npm run infra:up`、完整 infra-smoke 与 API readiness；核对许可证 SHA / 源码校验、实际卷和 UID 未改变。

处理完后通知继续，或明确授权 Agent 修复这个新问题。独立评审有其他阻塞发现时，另行记录并一起处理；不能只依据运行容器 healthy 关闭本构建问题。

## 5. 用户再次授权后的修复与复验

用户再次要求修复并推进，根 Agent 处理该已记录问题。脚本只对固定构建输出 chmod 0644，再用读取 / 写入保留原始字节，不改变上游源码缓存。已有只读输出可修复，新输出保持可写。

同一已有目录连续两次 `npm run infra:build-minio` 均 exit 0；输出 IsReadOnly=false、源 IsReadOnly=true，SHA-256 与上表完全一致。Docker / doctor / config / up / 完整 infra-smoke / 真实 API readiness / check / smoke 全通过。故先前只读覆盖解释已经获得修复后的复验证据；仍未宣称任何 ACL / 文件锁情形均已排除。

完整证据见 [修复交接](../handoffs/2026-10-01-license-fix-validation.md)。独立工程复审结论尚未交付前不关闭最终门禁。

## 关闭记录

用户再次授权修复，根 Agent 完成范围内变更与复验，独立 `/root/phase0_recheck` 的 [独立复审](../reviews/2026-10-01-phase0-recheck.md) 通过，确认原 R1 / R2 可关闭。[阶段验收](../reviews/2026-10-01-phase0-acceptance.md) 已登记；不将现有环境通过冒称冷机器初始化、远程 CI 或安全判题已验证。
