# P0-003：核心类型与设计草案

- 实时状态：见 PROGRESS.md
- Owner / 执行者：Architect + Builder / 本次初始化 Agent
- Review owner：P0-004 的独立 Architect reviewer `/root/phase0_review`（2026-10-01）
- 风险：R2
- 允许修改：internal/domain、judge、execution、sandbox、docs/architecture、docs/judge、docs/sandbox
- 禁止：宣称已冻结，扩展完整 Judge/Sandbox 实现，直接上线
- 依赖：原始计划及 ADR-0001

## 交付

领域快照、SourceUnit/Artifact/Run 分离、资源预算、ExecutionGraph/Engine/Runtime 草案、状态机和错误模型、事务/仓储边界。Sandbox 返回原始执行数据，由 JudgeEngine 映射 verdict，避免把执行与评分合并。

## 验收

类型编译；状态边方向与终态不可变测试；人工核对源版本、角色绑定、编译/运行/聚合预算。此项交付只证明草案存在且编译通过；语义完整性由 P0-004 审阅，尚无 persistence/fencing/资源强制实现。

## 2026-10-01 独立审阅

[正式报告](../reviews/2026-10-01-phase0-review.md) 的七项回答认可基础不变量，建议仅明确列项窄范围冻结；完整 Graph、Runtime、资源执行语义、DB / API 与事务接口继续 proposed。没有更改当前 `0.1.0-draft` 或冻结完整类型。P0-004 等待结论确认 / 冻结记录，Phase 0 又被构建复现阻塞，本任务不登记 done 或实施后续执行能力。

## 最终验收（2026-10-01）

草案交付与七项基础原则独立审阅完成，[阶段验收](../reviews/2026-10-01-phase0-acceptance.md) 通过，本任务的草案产物范围 done。仅 [基础不变量](../architecture/FOUNDATION_INVARIANTS.md) 已 frozen；完整 structs / Graph / Runtime / schema / 执行语义继续 proposed。
