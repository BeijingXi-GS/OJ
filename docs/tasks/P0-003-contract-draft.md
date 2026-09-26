# P0-003：核心类型与设计草案

- 实时状态：见 PROGRESS.md
- Owner / 执行者：Architect + Builder / 本次初始化 Agent
- Review owner：P0-004 的 Architect reviewer（待指派）
- 风险：R2
- 允许修改：internal/domain、judge、execution、sandbox、docs/architecture、docs/judge、docs/sandbox
- 禁止：宣称已冻结，扩展完整 Judge/Sandbox 实现，直接上线
- 依赖：原始计划及 ADR-0001

## 交付

领域快照、SourceUnit/Artifact/Run 分离、资源预算、ExecutionGraph/Engine/Runtime 草案、状态机和错误模型、事务/仓储边界。Sandbox 返回原始执行数据，由 JudgeEngine 映射 verdict，避免把执行与评分合并。

## 验收

类型编译；状态边方向与终态不可变测试；人工核对源版本、角色绑定、编译/运行/聚合预算。此项交付只证明草案存在且编译通过；语义完整性由 P0-004 审阅，尚无 persistence/fencing/资源强制实现。
