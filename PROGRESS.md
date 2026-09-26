# OJ 开发进度与任务台账

**唯一实时进度入口** · 更新日期：2026-09-26 · 当前阶段：**Phase 0 / 工程初始化已交付，阶段验收尚未完成**。

核心契约：0.1.0-draft / proposed。原始计划是项目依据，不代表已完成能力。尚未实现真实判题或业务提交。

## 当前交付与验证

| 类别                               | 当前状态                 | 证据 / 说明                                                                  |
| ---------------------------------- | ------------------------ | ---------------------------------------------------------------------------- |
| 正式开发流程、职责、模板、路线图   | 已编写，待独立审阅       | DEVELOPMENT.md、AGENTS.md、ROADMAP.md、docs/templates                        |
| Git / Monorepo / 依赖锁定          | 已建立；依赖已安装       | go.mod/go.sum、package-lock.json；main 分支；未提交、无 remote               |
| API / Worker / 核心类型            | 骨架可编译，基础测试通过 | Worker 仅 idle；核心类型为草案                                               |
| Web / 编辑器                       | 构建通过；浏览器验证通过 | 首页显示 API 连接正常；Monaco 渲染 C++                                       |
| pgx / sqlc                         | 依赖已安装；代码生成通过 | 仅基础 SELECT 1，不存在业务表                                                |
| 格式 / lint / 单元测试 / API smoke | 通过                     | [验证记录](docs/development/VALIDATION.md)                                   |
| npm 依赖检查                       | 通过，当前报告 0 项      | Monaco 已更新到修复版本                                                      |
| Compose 配置                       | 语法验证通过             | PostgreSQL / Redis / MinIO，端口仅绑定本机                                   |
| Compose 服务运行                   | **部分通过**             | engine 已就绪；PostgreSQL / Redis 已启动并健康；MinIO 镜像被官方移除，未启动 |
| 本机 Go race                       | **未能执行**             | 当前 CGO 不可用；CI 中已配置 Ubuntu race                                     |
| 远程 CI / 独立评审                 | **未执行**               | 无 remote；无独立 reviewer                                                   |

## 当前任务

状态词的精确定义见 DEVELOPMENT.md。角色与实际执行者分开记录；以下“本次初始化 Agent”不冒充独立 reviewer。

| ID                                               | 工作                         | 状态      | Owner 角色 / 实际执行者                | Review owner                 | 依赖 / 下一步                                                                                    |
| ------------------------------------------------ | ---------------------------- | --------- | -------------------------------------- | ---------------------------- | ------------------------------------------------------------------------------------------------ |
| [P0-001](docs/tasks/P0-001-foundation.md)        | 工程骨架与依赖               | in_review | Builder + Worker / 本次初始化 Agent    | Builder，待指派              | 检查和 smoke 通过；审阅工程，服务联通拆至 P0-005                                                 |
| [P0-002](docs/tasks/P0-002-workflow.md)          | 流程、职责及统一 MD          | in_review | Architect + Worker / 本次初始化 Agent  | 项目负责人 / Builder，待指派 | 核对流程和台账后登记审阅结论                                                                     |
| [P0-003](docs/tasks/P0-003-contract-draft.md)    | 核心契约草案                 | in_review | Architect + Builder / 本次初始化 Agent | Architect，待指派            | 草案可编译；交 P0-004 评审                                                                       |
| [P0-004](docs/tasks/P0-004-contract-review.md)   | 独立 Architect Review / 冻结 | ready     | Architect reviewer / 未认领            | 项目负责人确认评审结论       | 输入已齐；评审四个重点并写冻结记录                                                               |
| [P0-005](docs/tasks/P0-005-infra-validation.md)  | 本机依赖服务真实联通         | blocked   | Builder / 2026-09-26 会话 Agent        | Builder                      | PG `SELECT 1`、Redis `PING`、API `/readyz=200` 已通过；MinIO 健康检查失败（B-ENV-002），验收未满 |
| [P1-001](docs/tasks/P1-001-mini-judge-design.md) | Mini Judge 设计              | planned   | Architect / 未认领                     | Architect reviewer           | P0-004 + 可用 Linux 开发环境                                                                     |

## 阶段门禁

| Phase           | 状态        | 进入下一阶段前的条件                               |
| --------------- | ----------- | -------------------------------------------------- |
| 0 Foundation    | in_progress | 契约独立评审、基础服务实际联通、工程审阅与阶段验收 |
| 1 Mini Judge    | planned     | Phase 0 验收；Linux runner 设计                    |
| 2 Web OJ        | planned     | Phase 1 验收；DB/授权/API 评审                     |
| 3 Async OJ      | planned     | Phase 2 验收；Queue/recovery 契约冻结              |
| 4 Secure OJ     | planned     | Threat Model 与 Linux 隔离设计、安全评审           |
| 5 Standard OJ   | planned     | Phase 4 安全门禁                                   |
| 6 Contest       | planned     | 普通题稳定、量化负载目标                           |
| 7 Interactive   | planned     | IPC/生命周期 Spec                                  |
| 8 Communication | planned     | 多源码/图/聚合资源 Spec                            |
| 9 Cluster       | planned     | lease/fencing/容量调度 Spec                        |
| 10 Production   | planned     | 安全、负载、恢复与运维门禁                         |

## 阻塞与解除条件

| ID           | 阻塞                               | 当前处理                                                                         | 解除方式                                                                              |
| ------------ | ---------------------------------- | -------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| B-ENV-001    | Docker Desktop Linux engine 未就绪 | **已解除**（2026-09-26）：启动 Docker Desktop 后 engine 28.4.0 就绪，doctor 全绿 | 按 P0-005 完成真实验证（部分完成，见 B-ENV-002）                                      |
| B-ENV-002    | MinIO 官方镜像已从 Docker Hub 移除 | postgres / redis 已启动并验证；compose 未改动，MinIO 未启动；用户裁定暂不处理    | 项目裁定镜像来源（中转镜像源 / 替代 S3 存储 / 本地构建）后重跑 infra:up + infra-smoke |
| B-TEST-001   | Windows race 需要 CGO 工具链       | 普通 Go 测试已通过；未冒充 race 通过                                             | 在 Linux/具备 CGO 的环境运行 race；CI 已定义                                          |
| B-REVIEW-001 | 尚无独立架构评审                   | 所有核心契约保持 proposed                                                        | 认领 P0-004、记录 reviewer 与结论，再冻结                                             |

## 下一次开始

1. 读 [交接](docs/handoffs/2026-09-26-foundation.md)、[基础服务启动交接](docs/handoffs/2026-09-26-infra-restart.md) 和 [验证记录](docs/development/VALIDATION.md)。
2. 优先认领 P0-004；P0-005 已完成 PG / Redis 联通，剩余 MinIO 阻塞见 B-ENV-002。
3. 将核心设计的调整写 ADR 和版本记录；通过后再进入 P1-001。
4. 更新本台账中的实际执行者、日期、证据与状态。不要另建一份“最新进度”。

## 变更日志

| 日期       | 记录                                                                                                                                                |
| ---------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2026-09-26 | 启动 Docker Desktop 解除 B-ENV-001；infra:config、doctor 通过，PostgreSQL / Redis 启动并健康，API /readyz=200；MinIO 镜像被官方移除，登记 B-ENV-002 |
| 2026-09-26 | 归档原文，正式化 Phase 0–10 流程；建立工程和依赖；验证 API、构建、编辑器和 sqlc；记录 Docker/CGO/评审待办。                                         |
