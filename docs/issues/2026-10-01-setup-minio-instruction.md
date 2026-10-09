# ISSUE-2026-10-01-006：setup 完成提示漏掉 MinIO 构建前置

- 日期：2026-10-01；来源：独立 Agent `phase0_review` 的工程审阅。
- 状态：**2026-10-01 修复复验及独立复审通过，关闭**；失败与未测范围保留。

## 前因与事实

MinIO 已从 registry 镜像改为本地源码构建，Compose 设置 `pull_policy: never`。根 README / ENVIRONMENT 的初始化步骤已加入 `npm run infra:build-minio`，但 `scripts/setup.mjs:30` 的成功提示仍为：

```text
Dependencies ready. Run npm run doctor, then npm run infra:up.
```

setup 本身只恢复依赖 / 创建 `.env`，没有执行 MinIO 构建。此提示沿用旧流程，漏掉已变更的前置步骤。

## 影响与未验证范围

当前机器已有源码镜像并运行，提示错误不让当前服务失败。但新机器只按控制台提示直接 up，缺少本地镜像时无法启动 MinIO；该冷路径失败本轮没有通过删除现有镜像来人为复现。不能因为镜像已缓存而忽略初始化说明不一致。

## 用户接手与验收

使 setup 提示与当前开发文档一致，明确 doctor → infra:build-minio → infra:up → infra-smoke；不需要让 setup 自动改主机 Docker 或改变已有 `.env`。核对 README、环境说明、构建说明、CI 的顺序与前置条件，验证最终提示包含构建步骤。源码构建的重复 LICENSE 问题必须另行解除，见 [独立构建阻塞](2026-10-01-minio-license-copy.md)。

## 再次授权后的处理

用户再次授权修复并推进后，仅修订 setup 成功提示，补齐 `npm run infra:build-minio` 与 `node scripts/infra-smoke.mjs`。node --check 与工程格式检查通过；本轮按相同顺序实际完成 doctor、两次构建、up、smoke，均通过。没有重新运行 setup / npm ci 或覆盖 `.env`，因此只声明提示一致与现有环境复验，不声明全新机器的依赖下载已测试。

## 关闭记录

用户再次授权修复，根 Agent 完成范围内变更与复验，独立 `/root/phase0_recheck` 的 [独立复审](../reviews/2026-10-01-phase0-recheck.md) 通过，确认原 R1 / R2 可关闭。[阶段验收](../reviews/2026-10-01-phase0-acceptance.md) 已登记；不将现有环境通过冒称冷机器初始化、远程 CI 或安全判题已验证。
