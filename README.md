# OJ 开发仓库

基于原始开发计划整理的正式工程起点。当前交付是 **Phase 0 工程基础**，包含可启动的 Go API、空闲模式 Judge Worker、React 开发工作台、PostgreSQL / Redis / MinIO 配置及协作流程。尚未实现真实判题或提交持久化。

## 先读这些文档

| 入口                                        | 用途                                                     |
| ------------------------------------------- | -------------------------------------------------------- |
| [PROGRESS.md](PROGRESS.md)                  | **唯一进度台账**：任务、负责人、阻塞项、验证证据、下一步 |
| [AGENTS.md](AGENTS.md)                      | Agent 职责、目录所有权、执行与交接规则                   |
| [DEVELOPMENT.md](DEVELOPMENT.md)            | 正式开发流程、状态流转、验收与变更门禁                   |
| [ROADMAP.md](ROADMAP.md)                    | Phase 0–10 范围、产物和验收标准                          |
| [开发环境](docs/development/ENVIRONMENT.md) | Windows / Linux 使用说明与故障排查                       |
| [架构](docs/architecture/ARCHITECTURE.md)   | 服务边界、依赖方向、契约状态                             |

## 本地开始

要求 Node.js 22.19+（22.x）或 24.x、npm、Go 1.25+、Git。基础服务另需正在运行的 Docker Desktop Linux engine / Docker Engine + Compose v2。

在本目录中运行，PowerShell 与 Linux/macOS 命令相同：

```sh
npm run setup
npm run doctor
npm run infra:up
npm run check
```

分别在两个终端启动：

```sh
npm run api
npm run dev
```

打开 [开发工作台](http://127.0.0.1:5173)。API 存活检查为 [healthz](http://127.0.0.1:8080/healthz)。API 自动读取根目录 .env；数据库不可用时 readiness 返回 503，健康接口仍可使用。

```sh
npm run worker       # 空闲骨架，不消费任务
npm run smoke        # 临时启动 API 验证基础接口，随后自动停止
npm run infra:status
npm run infra:down   # 停止基础服务，保留数据卷
```

首次依赖已经安装的副本可直接运行；其他机器先执行 setup。setup 保留现有 .env，并为新环境生成随机开发凭据。缓存位于 .local/，依赖位于 node_modules/，都不纳入 Git。不得提交 .env。

## 目录

```text
apps/api/                 Go HTTP 控制平面入口
apps/web/                 React + TypeScript + Vite
services/judge-worker/    独立 Worker 入口
internal/domain/         核心类型及状态机草案
internal/httpapi/        HTTP 路由与 readiness
internal/judge/          JudgeEngine 草案
internal/execution/      ExecutionGraph 草案
internal/sandbox/        隔离接口草案（无实现）
internal/platform/       PostgreSQL / sqlc 基础
packages/proto/          OpenAPI
deploy/docker-compose/  本地依赖服务
docs/                    架构、任务模板、验收和交接
scripts/                 跨平台开发命令
```

其余规划模块按阶段建立，详见目录内 README；占位目录不代表已实现。CI 定义在 .github/workflows/ci.yml，推送到支持 Actions 的仓库后才会运行。当前未配置远程仓库。

## 原始依据

原始文件按原样保留在 [docs/reference/OJ_AGENT_DEVELOPMENT_PLAN_REVISED.md](docs/reference/OJ_AGENT_DEVELOPMENT_PLAN_REVISED.md)，只作为来源资料；当前执行规则以本仓库正式文档为准。原文差异与裁定见 [ADR-0001](docs/adr/0001-foundation-boundaries.md)。
