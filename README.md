# OJ 开发仓库

**2026-10-07：通信题开发版本，本地完整工程检查通过，真实 Docker 验收待完成。** 当前有 10 道内置题，支持普通文件判题、交互题、内置构造特判、教师 C++ 特判和发送端/接收端通信题。学生看整体结果，教师上传题面和数据、查看逐例结果，工程师查看调试信息；模式仅供本机手动切换，没有真实登录。

通信 A+B 的 [学生题面](examples/communication-a-plus-b/PROBLEM.md)、[教师数据包](examples/communication-a-plus-b/README.md) 与 [环境指南](docs/judge/COMMUNICATION_ENVIRONMENT.md) 提供两份 C++ 模板、四对数据、上传规则和真实验收脚本。已有依赖的副本按 `npm run compiler:build` → `npm run compiler:up` → `npm run judge:build` → `npm run mvp` 启动，再进入 `/problems/communication-a-plus-b`。必须先恢复正常 Docker Linux 访问并通过严格预检，不能直接在宿主运行选手代码。

本轮修复 Windows 题目发布及只读产物清理兼容问题、测试失败现场收尾；[当前验证](docs/judge/RECOVERY_VALIDATION_2026-10-07.md)、[审批原因与对策](docs/issues/2026-10-07-approval-diagnosis-and-recovery.md)、[GitHub 交付及后续职责](docs/handoffs/2026-10-07-recovery-delivery.md) 已整理。GitHub 写接口 403、普通 Git 连接重置，上传未成功，见 [恢复办法](docs/issues/2026-10-07-github-write-access.md)。真实通信矩阵、页面和旧 SPJ 最终待办尚未完成，实时状态见 [PROGRESS](PROGRESS.md)。以下旧版验收结果只作历史参考。

## 已通过的三模式基线（历史）

基于原始开发计划整理的开发仓库。当前实现**三模式判题MVP**：原5题/24组普通文件用例，新增交互猜数字/3组秘密数据；Monaco输入C++、编译检查、真实标准/交互判题，以及教师题面和数据上传。学生看整体结果，教师看具体用例，工程师看调试证据；模式在本机手动切换，没有真实登录。本轮实现、真实回归与独立本机产品验收已通过，最新状态以PROGRESS为准。API只做HTTP控制，编译和执行在独立服务/受限容器中；完整Phase1/生产安全/提交持久化未完成。

已有本机依赖可运行 `npm run compiler:build` → `npm run compiler:up` → `npm run judge:build` → `npm run mvp`，打开 [题库](http://127.0.0.1:5173)。完整启动、文件配置、判定语义、环境边界和Agent分工见 [文件判题使用说明](docs/judge/README.md)；[实际验证](docs/judge/VALIDATION.md) 与 PROGRESS记录当前验收状态。旧[编译说明](docs/mvp/README.md)保留历史范围。

模式切换、教师普通/交互题上传、协议与新Agent具体职责见 [三模式使用说明](docs/judge/MODES_USER_GUIDE.md)，局部受审契约见 [MODES_INTERACTIVE_CONTRACT](docs/judge/MODES_INTERACTIVE_CONTRACT.md)。

本轮完整[验证](docs/judge/MODES_VALIDATION_2026-10-03.md)、[浏览器记录](docs/judge/BROWSER_QA_2026-10-03.md)、[独立审阅](docs/reviews/2026-10-02-modes-interactive-review.md)和[交接](docs/handoffs/2026-10-03-modes-interactive.md)已交付；当前9题含3道持久教师上传题，手动模式不提供真实身份认证。Windows退出提示的非阻塞运维问题单独记录并交用户，见PROGRESS。

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

完整 Web OJ 的后续规划见 [前端框架实施指南](docs/frontend/IMPLEMENTATION_GUIDE.md) 和 [Agent 详细任务包](docs/frontend/AGENT_TASKS.md)。最新用户已扩展文件运行判题，详见 [判题契约](docs/judge/JUDGE_MVP_CONTRACT.md)；账号 /提交历史等完整能力仍以 PROGRESS.md 为准。

PostgreSQL 宿主机默认端口已改为 `15432`，容器内部仍为 `5432`；已有 `.env` 需同步端口和 `DATABASE_URL`，setup 不会覆盖它。详见 [端口修复与基础服务复验交接](docs/handoffs/2026-09-30-port-docker-validation.md)。执行下列完整启动流程前先检查 PROGRESS.md 中的当前阻塞；不能把部分服务健康当作完整环境验收。

要求 Node.js 22.19+（22.x）或 24.x、npm、Go 1.25+、Git。基础服务另需正在运行的 Docker Desktop Linux engine / Docker Engine + Compose v2。

在本目录中运行，PowerShell 与 Linux/macOS 命令相同：

```sh
npm run setup
npm run doctor
npm run infra:build-minio
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
services/compile-worker/  只编译链接的独立服务
services/judge-mvp/       本机Docker OCI判题控制服务
services/judge-compile-helper/ 固定编译并流式导出产物
internal/domain/         核心类型及状态机草案
internal/httpapi/        HTTP 路由与 readiness
internal/compilation/    编译DTO与HTTP客户端
internal/compilerworker/ 固定G++执行、超时与输出清理
internal/filejudge/      文件快照、checker、Docker监督和HTTP桥接
testdata/problems/       服务器.in/.out与manifest
internal/judge/          JudgeEngine 草案
internal/execution/      ExecutionGraph 草案
internal/sandbox/        隔离接口草案（无实现）
internal/platform/       PostgreSQL / sqlc 基础
packages/proto/          OpenAPI
deploy/docker-compose/  本地依赖服务
deploy/compiler/        本机编译MVP镜像与独立Compose
docs/                    架构、任务模板、验收和交接
scripts/                 跨平台开发命令
```

其余规划模块按阶段建立，详见目录内 README；占位目录不代表已实现。CI 定义在 .github/workflows/ci.yml，推送到支持 Actions 的仓库后才会运行。实际远程仓库与 CI 验证状态见 [PROGRESS.md](PROGRESS.md)，配置工作流不代表远程检查已通过。

## 下一阶段

下一阶段可信fixture的Mini Judge设计包见 [Phase 1 DESIGN](docs/judge/PHASE1_DESIGN.md) 与 [Agent任务包](docs/judge/PHASE1_AGENT_TASKS.md)。当前仅交付待审设计，尚无CLI或真实判题；基础原则冻结范围与阶段结果以PROGRESS为准。

## 原始依据

原始文件按原样保留在 [docs/reference/OJ_AGENT_DEVELOPMENT_PLAN_REVISED.md](docs/reference/OJ_AGENT_DEVELOPMENT_PLAN_REVISED.md)，只作为来源资料；当前执行规则以本仓库正式文档为准。原文差异与裁定见 [ADR-0001](docs/adr/0001-foundation-boundaries.md)。
