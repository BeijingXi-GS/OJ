# COMM-003：通信题双编辑器、上传与反馈

2026-10-07 更新：当前前端 lint/tsc/build/SSR 随完整检查通过；通信题双编辑器和教师包已交付，学生题面删去演示解法说明。真实浏览器操作仍待环境恢复，按 [接续职责](../handoffs/2026-10-07-recovery-delivery.md) 验证，不能用 SSR 替代。

- 状态：以[PROGRESS](../../PROGRESS.md)为准；本地实现已交审，真实页面pending。
- Owner：comm_frontend Builder；Review owner：comm_review Builder/Architect；风险R2。
- 范围：仅apps/web，不改后端、契约、共有MD、依赖、Docker或Git。
- 输入：[受审契约](../judge/COMMUNICATION_CONTRACT.md)及后端固定DTO。
- 输出：CommunicationWorkspace双角色编辑器、lib/communication显式请求与示例、api DTO/独立端点、EditorPage双草稿与取消防旧响应、App通信标签、UploadPage固定图/byte_limit/双模板/配对.in.out、JudgeFeedback三模式投影与角色编译、styles和verify-modes；新增教师A+B说明与双模板按钮，SSR使用communication-a-plus-b真实组件fixture。
- 学生：公开题面与限制、整体判定/自己的实际角色CE；教师：逐例有界方向日志/通信量；工程师：退出/OOM/终止/hash，不能把工程说明放学生。
- 验证：typecheck/lint/build/verify-modes/前端format实际PASS；保留原Monaco体积告警，真实双编辑器/教师发布/模式取消由root环境恢复后验收。
- 开发修正与方案：[档案](../issues/2026-10-05-communication-development.md)，使用与验收：[指南](../judge/COMMUNICATION_USER_GUIDE.md)。
