# P0-004：独立架构评审与契约冻结

- 实时状态：见 PROGRESS.md
- Owner 角色：Architect reviewer
- 实际执行者：独立 Agent `/root/phase0_review`，2026-10-01 用户明确授权后认领；本轮原作者不作为 reviewer
- 风险：R2
- 输入：P0-003、ADR-0001、ARCHITECTURE、DOMAIN_MODEL、STATE_MACHINE、SECURITY_BASELINE、核心 Go 草案
- 依赖：P0-003 草案及验证证据已交付
- 允许修改：审阅记录、设计修订、核心契约；修订后需重跑受影响检查
- 禁止：开展整个 Phase 1，跳过审阅直接冻结

## 必须回答

1. API 与 Judge 的依赖与运行权限是否隔离？Phase 2 同步 bridge 是否保持独立进程？
2. Submission、SourceUnit、Run、Compilation/Artifact 的身份和不可变关系能否支持重判？
3. lifecycle/terminal verdict 组合、取消竞争、传输重试与 rejudge 是否一致？
4. ProcessSpec 的角色、源码和产物绑定是否显式？Graph 的可信管理进程如何扩展？
5. 单进程、聚合、编译资源与节点开销的单位和计量边界是否清晰？
6. Sandbox 与 JudgeEngine 是否分离原始执行结果和用户 verdict？
7. 数据事务、outbox、fence 条件和历史可追溯性哪些需要 Phase 2/3 继续具体化？

## 完成条件

使用 REVIEW 模板提供实际审阅者、版本、结论和修改项；无阻塞后将受审契约从 proposed 变成 frozen 并编号。冻结的范围必须明确，不能将尚未设计的 Phase 7–9 细节一并视为冻结。

## 2026-10-01 输入已更新

环境阻塞已解除，工程回归与完整基础服务验证通过。作者已准备 [具体评审包](../reviews/P0-004-review-package.md)，包含七项核心输入、工作区摘要、工程复核、已测 / 未测和冻结范围。该文是评审输入，不能视作独立结论。

随后用户明确授权独立 Agent；`/root/phase0_review` 已交付 [正式 REVIEW](../reviews/2026-10-01-phase0-review.md)，七项基础不变量审阅通过，建议仅按报告第 4 节明确列项建立窄范围冻结。完整 Go 类型、Graph、Runtime、资源执行语义和事务接口仍是草案，不可整体冻结。本任务进入 in_review 等待结论确认与冻结记录，本轮没有改动契约版本。

环境首次验收通过，但最终重复构建出现新 [EPERM 问题](../issues/2026-10-01-minio-license-copy.md)，基础工程审阅结论为需修改；reviewer 未修复或绕过它，主工程停止进入下一阶段。解除构建问题并完成受影响范围复审 / 阶段验收前，不登记 P0-004 或 Phase 0 完成。

## 最终采用与验收（2026-10-01）

独立七项回答与本轮 [独立复审](../reviews/2026-10-01-phase0-recheck.md) 完成；根 Agent 依据用户推进授权只采用 [foundation-invariants/1.0.0](../architecture/FOUNDATION_INVARIANTS.md)，没有扩大冻结范围。[阶段验收](../reviews/2026-10-01-phase0-acceptance.md) 已登记，P0-004 在这个窄范围内 done；原工程阻塞已解除，完整核心仍 draft。
