# MVP-001 后端、编译环境与联调

- Owner：/root，Builder；Review owner：独立 reviewer（MVP-003）。
- 状态：实时状态仅见 ../../PROGRESS.md；当前实施。
- 范围：apps/api、internal/httpapi、internal/compilation、internal/compilerworker、services/compile-worker、deploy/compiler、scripts、环境说明及验收文档。
- 依赖：用户最新实际 MVP /并行开发授权；[契约](../mvp/CONTRACT.md)。
- 交付：内置题目读取、严格编译请求校验、API 转发、真实 g++ 编译链接、输出上限 /超时 /单请求并发 /临时文件清理；永不运行产物。
- 验收：Go handler /worker /client 测试，真实 Docker success /CE /链接失败 /无限循环源码编译，完整 check /smoke 与浏览器联调。
- 不包含：数据库题库、登录、答案判定、sandbox、真实 runner。
- 新阻塞：单开 docs/issues 文档并停止，由用户处理；保留既有卷与工作区。

2026-10-01交付：实现与真实编译 /Linux专项 /check /基础smoke通过；浏览器联调和最终独立review未完成，见 [验证](../mvp/VALIDATION.md)、[停止交接](../handoffs/2026-10-01-compile-mvp.md) 与 [B-MVP-001](../issues/2026-10-01-mvp-browser-session.md)。当前根Agent已停止，不登记done。
