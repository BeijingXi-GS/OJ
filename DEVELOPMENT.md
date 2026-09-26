# 正式开发流程

版本：1.0 · 基准日期：2026-09-26。此文件描述流程；实时进度由 PROGRESS.md 维护。原始计划是输入材料，正式裁定记录在 ADR。

## 工作单元与状态

使用三层结构：Phase（阶段验收）→ Task（可独立审阅的工作单元）→ Evidence（绑定版本的验证记录）。每个任务使用 P<阶段>-<三位编号>，详细任务卡放 docs/tasks/。

允许状态：

```text
planned → ready → in_progress → in_review → done
                    ↓              ↓
                  blocked ←────────┘
                    ↓
                  ready
in_review → in_progress（审阅退回）
planned / ready / in_progress / blocked → cancelled（写明原因）
```

- planned：范围或依赖尚未明确。
- ready：输入契约、负责人角色、修改范围、验收和依赖均明确。
- in_progress：由台账中的实际执行者认领，写明开始日期。
- blocked：写明原因、解除条件、依赖人/环境；保留其他可推进的独立任务。
- in_review：实现及证据已交付，等待指定审阅；审阅者未指派也必须标明。
- done：适用检查通过，审阅已记录，文档与交接已更新。
- cancelled：不再实施且记录去向。

阶段状态根据其门禁判定，不根据代码量或日期估算百分比。没有独立审阅时保留 in_review；阶段未验收不得宣称完成。环境配置就绪、依赖已安装、服务已启动、功能已实现分开记录。

## 从需求到交付

| 步骤             | 负责人                 | 输入 → 产物                           | 退出条件                                         |
| ---------------- | ---------------------- | ------------------------------------- | ------------------------------------------------ |
| 0 范围进入       | 项目负责人 + Architect | 上阶段验收 → Goals / Non-goals / 依赖 | 当前阶段边界、风险、验收明确                     |
| 1 设计与风险     | Architect              | 需求 → DESIGN、ADR、威胁/故障分析     | 核心对象、接口、资源/安全边界可审阅              |
| 2 契约评审冻结   | Architect reviewer     | 草案 → 版本号 + 审阅记录              | 不变量、API、状态、DB、兼容策略已确认            |
| 3 拆分与认领     | Builder lead           | 冻结契约 → 任务卡 + 台账              | 单一 Owner、路径范围、依赖、验收可执行           |
| 4 隔离实现       | 各 Owner               | task 分支 → 最小闭环                  | 范围内正常及失败路径完成                         |
| 5 辅助工程       | Worker                 | 明确接口 → UI / fixture / 文档等      | 禁止修改项未被突破                               |
| 6 自动检查       | Owner                  | 当前工作区 → 检查记录                 | format、lint、test、build 通过；适用专项测试通过 |
| 7 分级审阅       | Builder / Architect    | diff + 证据 → REVIEW                  | 依据 AGENTS.md 完成契约和风险检查                |
| 8 集成与故障注入 | Builder + Architect    | 模块 → 集成版本                       | crash、重复、超时、OOM 等适用场景通过            |
| 9 阶段验收       | 项目负责人 + Architect | 阶段验收表 → 明确结论                 | 功能/失败/不变量/安全/观测/运维齐全              |
| 10 合并与归档    | 集成人                 | 已审版本 → release notes + 证据       | 记录协议/schema/runtime 版本和回退方案           |
| 11 回归与变更    | Architect + 相关 Owner | 变更申请 → 新 ADR / 新契约            | 兼容、迁移、回退、重判影响明确                   |

Phase 0 可先落实不影响核心契约的工程初始化；核心类型只是可编译草案。后续功能实现必须使用已冻结版本，不因为有草案代码就跳过步骤 2。

## Definition of Ready

任务必须具备：目标、非目标、风险等级、Owner 角色、实际执行者（认领时填写）、Review owner、允许路径、禁止改动、依赖 ID、输入契约版本、输出物、验收命令及预期、失败路径。使用 [TASK 模板](docs/templates/TASK.md)。

风险分级：R0 文档/格式；R1 普通业务和 UI；R2 状态/一致性/并发/资源；R3 安全隔离、授权、不可逆数据变更。R2/R3 必须 Architect review，R3 同时提交威胁分析和专项安全证据。

## Definition of Done

- 输出与任务范围一致，正常/失败路径已覆盖。
- 应用检查 `npm run check` 通过；基础接口 `npm run smoke` 通过。
- 改动涉及数据库、Queue、Sandbox 时增加真实集成、竞态或隔离测试；不以 mock 代替真实资源边界证据。
- 记录命令、OS、工具版本、时间、代码 commit 或工作区摘要，以及实际结果。
- 适用审阅已通过，未解决项不能影响验收。
- 进度台账、任务卡、交接及运维说明同步。
- 测试暂未具备环境时，写 blocked 或 in_review，不能登记 done。

无需为纯排版等低影响变更堆砌测试；执行与变更风险相称的检查。当前 CI 通过只能证明 Phase 0 检查集，不能证明真实判题安全。

## 契约变更与分支

契约状态使用 proposed → reviewed → frozen → superseded。冻结记录包括版本、日期、审阅者、证据链接。未填审阅者的文档保持 proposed。变更使用 ADR，写明 Why、Affected modules、Compatibility、Migration、Rollback、Rejudge / Rebuild scope。

建议 main → phase-N-integration → task/Pn-xxx-slug。任务完成后按 contract / migration → core → adapters → UI / docs 顺序集成。禁止强推覆盖协作成果；产生冲突按已审契约解决。无远程仓库时可先本地工作，不伪造 PR 或 CI 结果。

## 阶段验收与回退

验收依据 ROADMAP 对应行及该阶段 DESIGN 的可量化指标；Phase 6 吞吐量/延迟、Phase 9 容量和恢复时限必须在实施前定义。模板见 REVIEW。数据库变更明确可逆 migration 或 forward-fix；历史源码、产物与 JudgeRun 审计记录不删除。

本次基础环境回退：停止本项目进程，运行 `npm run infra:down`；保留数据卷和 .env。删除数据卷、重置数据库或更改生产环境不属于日常清理。

## 通信题专项顺序（Phase 8）

Role 校验 → 冻结 SourceUnit 快照 → 逐份解析语言/版本 → 在 CompileBudget 下编译 → 持久化产物 hash → 显式 SourceBinding → 图与资源验证 → 预留总资源 → Job 级资源容器 → 启动进程 → 监督通道和聚合内存 → 失败传播及完整清理 → 结果事务提交 → ACK。

量测同时记录单 Role 和 Job 聚合值，资源计量不可信归类 SYSTEM_ERROR。此流程为后续阶段验收基线，本阶段无执行实现。
