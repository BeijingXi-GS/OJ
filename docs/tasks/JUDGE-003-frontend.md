# JUDGE-003 前端文件测试和结果

- Owner：/root/mvp_frontend /Frontend Builder，已有用户并行授权；Review：独立reviewer。
- 修改范围仅apps/web，禁止修改根台账、API契约、后端或锁文件，不新增依赖。
- 依赖：JUDGE-001受审新DTO；实现可与后端按契约并行，最终必须真实接口联调。
- 交付：保留编译、增加运行测试按钮；共享请求互斥锁；题目运行限额 /case数量、逐例与总verdict、CE /SYSTEM_ERROR /服务异常、源码过期结果 /重置 /取消 /重试 /响应式。
- DTO：见 [judge-mvp/0.1.0](../judge/JUDGE_MVP_CONTRACT.md)，CPU /峰值null显示未采集，不编造0。前端诊断只按文本渲染，不暴露标准答案。
- 静态验收：typecheck /lint /build /格式；实际浏览器root联调及独立review。
- 2026-10-02已在设计通过后实际派发并交付apps/web七文件，无新依赖；typecheck /lint /build /格式及独立源码审阅通过。浏览器实测因最终启动与历史会话问题尚未完成，见 [交接](../handoffs/2026-10-02-file-judge-mvp.md)。所有新阻塞及时报告root，不自行绕过。

以上是上轮记录。本轮JUDGE-005/007恢复浏览器会话并完成五题/实际编译和运行/结果过期及清空/切题隔离/手机布局，见 [真实QA](../judge/BROWSER_QA_2026-10-02.md) 与本轮PROGRESS。
