# ADR-0001：正式化原计划的阶段边界

日期：2026-09-26。状态：工程初始化范围已采用；核心语义仍 proposed，待 P0-004 审阅。

## 背景

原计划完整覆盖 Phase 0–10，但流程与当前状态分散；部分描述需要消除歧义后才能成为实现契约。

## 决策

1. 本次落实 Phase 0 工程起点，保留 Phase 1–10 的路线图，不一次实现所有功能。
2. PROGRESS.md 是唯一实时台账，AGENTS.md 管职责，DEVELOPMENT.md 管流程；原始文档原样归档，不作为运行时指令来源。
3. “Phase 2 可本地同步调用 Judge”解释为通过独立判题进程/服务同步获取结果。API 内执行用户代码始终禁止。
4. 将原 Verdict 混合枚举拆成 lifecycle state + terminal verdict；展示兼容名由 API 层组合。
5. 不可变 SourceUnit 存源码与请求版本；可变编译状态及产物按 Run 记录，以保证重判历史可追溯。
6. Phase 0 对未来对象提供设计目录与最小草案，不用空壳 CRUD、假数据或未审 migration 伪装业务完成。
7. 核心接口没有独立审阅记录前保持 draft。初始化者的自检不得登记为独立 Architect Review。
8. Windows 用于控制平面开发；真实 runner / Sandbox 需要 Linux。基础依赖仅绑定 localhost。

## 影响与迁移

当前无业务数据与外部客户端，暂无数据迁移。未来契约变更仍需记录兼容与重判范围；不允许改写历史快照。保留原文副本和 SHA-256，方便溯源。

来源 SHA-256：
`5271A56F0C1D2E76AD52BCBD7CF5A27547DCD7AF629497E583BC89E85DA80137`
