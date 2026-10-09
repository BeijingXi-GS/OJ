# JUDGE-001 文件用例与沙箱判题扩展

- Owner：root /Architect；Review owner：独立Architect +Security Review（恢复后在用户已有独立审阅授权内安排，当前未委派）。
- 当前状态见PROGRESS；因B-JUDGE-001阻塞，未进入应用实施。
- 输入：用户2026-10-01运行判题要求及服务器.in/.out选择，Anolix参考仓库；[范围和后续任务](../judge/MVP_JUDGE_SCOPE.md)。
- 分支：task/JUDGE-001-file-cases-sandbox-2026-10-01；保留原未提交代码与数据卷。
- 允许范围：本轮必要任务 /范围 /问题 /交接MD；恢复并审阅后才由JUDGE-002/003认领代码目录。
- 当前完成：需求 /参考README与go.mod核对、Docker真实限制预检；未完成：参考核心源码、commit锁定、沙箱设计独立评审、实现及所有真实verdict运行证据。
- 计划交付：每题manifest显式绑定.in/.out，版本快照；独立受限编译 /运行 /可信资源证据 /完整清理；checker /逐例及总verdict；API桥接 /前端结果与真实矩阵。
- 非目标：公网生产、登录、提交持久化、Queue /集群、交互 /通信题；原完整Phase1设计不冒称验收。
- 风险R3：运行不可信源码的隔离、资源和清理必须独立审阅，缺项不启用执行，不能回退普通subprocess。
- 恢复条件：[Anolix访问问题](../issues/2026-10-01-anolix-source-access.md) 复验通过；新问题仍归用户，阻塞即停止。
- 本轮结果：[停止交接](../handoffs/2026-10-01-judge-preflight-blocked.md)。

## 2026-10-02恢复与设计交付

用户授权先拉取源码后继续；Anolix本地commit4502d2a798f3130d3c540c5c31d45251ca103623、Docker Linux真实probe、局部judge-mvp/0.1.0契约与ADR已交付。独立/root/mvp_review [正式设计审阅](../reviews/2026-10-02-judge-design-review.md)通过并采用；已派发JUDGE-002/003/004，不再停在历史源码获取阻塞。实时状态仅PROGRESS。
