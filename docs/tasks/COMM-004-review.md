# COMM-004：通信题独立审阅

2026-10-07 增量由 recovery_final_review 只读审阅测试收尾、设计澄清、题面与当前日志/摘要，仅写 [本轮报告](../reviews/2026-10-07-recovery-review.md)。当前完整本地检查通过，真实 Docker 与最终产品结论仍 pending；不得沿用旧冻结摘要。

- 状态：以[PROGRESS](../../PROGRESS.md)与[实际独立报告](../reviews/2026-10-05-communication-review.md)为准。
- 执行者：comm_review独立Architect/Security；不得把root或实现者自检登记为独立审阅。
- 允许范围：只读契约/源/证据与可信测试，仅写自己的审阅报告；不修源/契约，不操作Docker/真实服务/浏览器/Git或宿主C++。
- 设计：首轮四项状态必修未PASS保留历史，实际补充后局部设计PASS；实现门禁按新发现逐项复审，真实产品pending。
- 必审：source/artifact/process显式绑定、双编译失败前缀、strict JSON与新旧端点、通信原始字节/EOF/提前退出/未知IO、共享wall/并行资源上界、真实kill/inspect/双清理与taint、上传快照、学生白名单/教师界限、取消旧响应。
- 真实Docker证据当前未执行，不以历史SPJ、health版本标记、可信fixture或局部设计PASS替代产品签署。
