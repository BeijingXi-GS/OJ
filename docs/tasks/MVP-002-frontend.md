# MVP-002 前端实现

- Owner：/root/mvp_frontend，Frontend Builder；Review owner：独立 reviewer（MVP-003）。
- 修改范围：apps/web；不得修改 API 契约、根台账、依赖锁文件或后端实现。
- 依赖：[契约](../mvp/CONTRACT.md)，用户授权实际前端和并行开发。
- 交付：首页题目卡片、题面 /样例、Monaco C++ 输入、真实 API 编译反馈与恢复路径、响应式布局。
- 验收：frontend typecheck /lint /build；root 执行真实浏览器与 API 联调；错误诊断按文本呈现。
- 确保：空 /UTF-8 超限源码校验，防重复提交，输入变动清理或标明过期结果，重置前确认，明确只编译不运行。
- 新问题及时报告 root，不绕过阻塞；现有 Monaco bundle 警告已单独记录。

2026-10-01交付：实际mvp_frontend仅修改apps/web五个文件，无依赖变更；typecheck /lint /build /Prettier /diff检查通过，最终root check覆盖最后修改。浏览器尚未验收，因 [B-MVP-001](../issues/2026-10-01-mvp-browser-session.md) 停止；保留in_review，不登记done。具体分层与继续指令见 [MVP使用说明](../mvp/README.md)。
