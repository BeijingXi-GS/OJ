# P0-004：独立架构评审与契约冻结

- 实时状态：见 PROGRESS.md
- Owner 角色：Architect reviewer
- 实际执行者：未指派
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
