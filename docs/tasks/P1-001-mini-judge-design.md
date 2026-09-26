# P1-001：Mini Judge 生命周期设计（下一阶段首项）

- 实时状态：见 PROGRESS.md
- Owner 角色：Architect
- 实际执行者 / Review owner：待指派
- 风险：R2
- 依赖：Phase 0 契约评审；可用 Linux 开发环境
- 允许修改：Phase 1 DESIGN、生命周期/失败语义 ADR、任务拆分和 fixture 规范
- 禁止：一次实现全阶段；API 执行代码；接收外部恶意程序

## 目标

把 C++ 编译→受控本地 runner→stdin/stdout→Exact Checker→CLI 结果定义成可实现的 Spec，明确编译/运行分界、timeout/exit race、进程组清理、信号、输出上限、错误归类和可信 fixture 边界。

## 输出与验收

DESIGN 包含 AC/WA/CE/RE/TLE 五类成功复现标准、子进程残留检测、编译超时、平台要求、回退方式，以及 Builder/Worker 后续小任务。实际运行环境和编译器版本明确后才能进入实现。
