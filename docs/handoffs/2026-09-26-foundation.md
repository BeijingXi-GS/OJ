# 交接：OJ 工程基础初始化

日期：2026-09-26。执行者：本次初始化 Agent（流程/架构草案整理 + Builder/Worker 工程实现）。无独立审阅者。

## 已交付

- 原计划归档及 SHA-256，正式流程、Phase 0–10 路线图、Agent 责任与审阅规则。
- PROGRESS.md 唯一台账；任务、设计、ADR、审阅、交接模板；六项可领取任务卡。
- Go module，chi API 与 pgx readiness，独立 idle Worker，核心领域/执行/沙箱契约草案及最小不变量测试。
- React/Vite/TypeScript/Tailwind/shadcn 基础、TanStack Query、Router、Monaco C++ 预览。
- PostgreSQL / Redis / MinIO Compose，随机本地 .env、跨平台初始化/诊断/检查/启动命令、sqlc 代码生成、CI。
- Git main 初始化。没有提交、远程仓库或发布；.env 与缓存不跟踪。

## 验证与当前状态

[验证记录](../development/VALIDATION.md) 绑定本次代码摘要；[PROGRESS.md](../../PROGRESS.md) 是实时状态来源。格式、lint、普通测试、构建、API smoke、sqlc、依赖检查和浏览器预览已通过。不要将这些结论扩展成真实判题、安全隔离或数据库集成已经通过。

## 未完成与解除条件

- Docker Linux engine 不可连接。主机启动后认领 P0-005，执行 infra:up / infra-smoke，并验证 API readyz。
- 本机缺少配置好的 CGO，race 未执行；Ubuntu CI 定义已包含 race，但 remote 尚未建立。
- 领域模型、接口、状态机保持 proposed。认领 P0-004 做独立审阅，再冻结范围和版本。
- 实际判题从 Phase 1 开始；Worker 当前日志为 idle，不消费 Redis。
- 业务表/migration、bucket/存储适配器、登录权限与提交接口尚未建立。

## 下一位 Agent 的第一步

读根 AGENTS.md / PROGRESS.md → 领取 P0-004 或 P0-005 → 在台账填写执行者与范围 → 按任务卡工作。评审四个重点为 package dependency、core interfaces、JudgeRun state machine、database identity。

后续进入 P1-001 时先写 Linux runner 生命周期 Spec，不要直接实现整个 Judge。如果要改变本次核心草案，记录差异和影响，并再次运行适用检查。

## 本地启动

```sh
npm run api
```

另一个终端：

```sh
npm run dev
```

依赖无需重复安装；搬到新机器后先 setup。基础服务通过 infra:up 启动；infra:down 保留卷。不要为了修复启动问题删除其他项目的容器、数据卷或宿主设置。
