# 基础不变量契约

- 版本：foundation-invariants/1.0.0。
- 状态：frozen（仅下列七项原则），2026-10-01 采用。
- 范围：下列七项原则；不冻结整个 `0.1.0-draft` 文档集合或现有 Go structs。
- 整理 / 采用执行者：根 Agent / Architect，依据用户再次授权修复并推进。
- 独立审阅依据：`/root/phase0_review`，2026-10-01，[正式报告](../reviews/2026-10-01-phase0-review.md)第 3–4 节。

## 冻结内容

1. API 不编译、不运行选手代码，也不持有执行权限；Phase 2 同步 bridge 保持独立判题进程 / 服务边界。
2. Submission / SourceUnit 源码快照不可变；编译诊断和产物由 Run 持有；rejudge 新建 Run，不覆盖终态历史。选定的源码与 problem / testdata / checker / compiler / runtime 版本可追溯。
3. SourceUnit、Artifact、Process 以显式身份绑定；不依据数组顺序推断来源。Artifact 必须属于该 Run 使用的同一 Submission 快照。
4. lifecycle 与 verdict 分离：非终态无 verdict，CANCELLED 对应取消 verdict，FINISHED 有合法终态 verdict；终态唯一且不回退。具体状态 / verdict 组合、取消与完成竞争算法由相应阶段设计，不因方向 helper 存在就宣称实现。
5. 编译、逐进程、Job 聚合预算与节点开销分别表达；时间 ms、容量 bytes、进程 PID 数。限制为正且有限，不隐式无限。聚合峰值取同一 Job 的实际观测，不累加独立进程峰值。
6. Runtime 返回原始监督证据并负责完整清理，JudgeEngine 映射 verdict；失去可信计量或清理能力归类基础设施失败。隔离缺失时没有普通 subprocess 的隐式降级；Phase 1 仅使用维护者提供的可信 fixture，Phase 4 安全验收前不开通外部任意源码执行。
7. Queue 至少一次投递；业务结果事务成功后才 ACK；旧持有者通过 fence 拒绝写回；提交、版本、attempt 与 rejudge 历史可追溯。此项是 Phase 2/3/9 的实现约束，当前没有业务事务、outbox、lease 或 fencing 的运行验收。

以上是后续实现必须遵守的原则，不是现有工程已经具备这些能力的声明。

## 明确排除

完整 Graph / SourceBinding schema、管理进程信任与通道模型、ResourcePolicy 默认值与强制执行语义、Runtime / Engine 的返回值与错误语义、DB / API schema、身份授权、outbox / fencing SQL / lease / 恢复策略、Linux 安全隔离及 Phase 7–9 细节均保持 proposed。实现前完成各阶段 DESIGN 和独立审阅。

现有 `internal/domain`、`internal/execution`、`internal/sandbox`、`internal/judge` 草案不能直接充当长期兼容 DTO 或已冻结数据库 schema。本文件没有替换它们，也没有改动代码字段。

## 变更与验收记录

原则变更必须提交 ADR，说明兼容、迁移、恢复和重判影响，并由独立 Architect 审阅；版本升级只覆盖明确受审范围。

2026-10-01 独立 `/root/phase0_recheck` 的 [独立复审](../reviews/2026-10-01-phase0-recheck.md) 已通过，确认七项语义与原审阅一致。根 Agent 根据用户修复并推进授权采用 `foundation-invariants/1.0.0`，仅登记状态与证据，没有扩大受审范围；完整核心草案仍 proposed。[阶段验收](../reviews/2026-10-01-phase0-acceptance.md) 对实现范围及未测项单独负责。
