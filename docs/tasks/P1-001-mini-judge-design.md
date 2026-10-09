# P1-001：Mini Judge 生命周期设计（下一阶段首项）

- 实时状态：见 PROGRESS.md
- Owner 角色：Architect
- 实际执行者：根 Agent / Architect，2026-10-01 Phase 0 验收后认领
- Review owner：独立 Architect，P1-002；Phase 0 reviewer 不自动作为本设计的审阅者
- 风险：R2
- 依赖：Phase 0 窄范围契约及本地工程验收已通过；Linux engine 可用，runner 权限 / 编译器 profile 在 P1-010 实际锁定后才进入实现
- 允许修改：Phase 1 DESIGN、生命周期/失败语义 ADR、任务拆分和 fixture 规范
- 禁止：一次实现全阶段；API 执行代码；接收外部恶意程序

## 目标

把 C++ 编译→受控本地 runner→stdin/stdout→Exact Checker→CLI 结果定义成可实现的 Spec，明确编译/运行分界、timeout/exit race、进程组清理、信号、输出上限、错误归类和可信 fixture 边界。

## 输出与验收

DESIGN 包含 AC/WA/CE/RE/TLE 五类成功复现标准、子进程残留检测、编译超时、平台要求、回退方式，以及 Builder/Worker 后续小任务。实际运行环境和编译器版本明确后才能进入实现。

## 本轮输出（2026-10-01）

已交付 [PHASE1_DESIGN](../judge/PHASE1_DESIGN.md)、[ADR-0003](../adr/0003-phase1-local-runner-lifecycle.md)、[fixture规范](../judge/PHASE1_FIXTURES.md)、[Agent任务包](../judge/PHASE1_AGENT_TASKS.md)与 [P1-002评审任务](P1-002-design-review.md)。设计明确观察时点 /deadline顺序、完整cgroup /reaper清理、原始证据和verdict分离、预算 /I/O /错误映射、CLI /report和真实验收矩阵。

设计包待独立审阅，任务进入in_review，不冒称done或局部接口已经冻结。P1-010尚未锁定实际GCC /runtime及runner权限；没有创建执行器、CLI或C++fixtures，没有运行Phase1生命周期测试。原基础服务继续健康；源码摘要未因这些MD改变，检查见 [本轮交接](../handoffs/2026-10-01-phase1-design.md)。
