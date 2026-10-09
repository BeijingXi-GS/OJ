# MODES-004 页面与独立验收

Frontend Builder只改apps/web；mode选择/缓存与在途隔离、学生聚合反馈/教师详情/工程师debug、教师上传标准/固定交互题与数据、题型标识、错误/进度/刷新/响应式。无新依赖/真实登录假承诺；mode更改不泄露旧结果。root实际UI，独立roles_review仅docs/reviews/2026-10-02-modes-interactive-review.md及只读可信fake/oracle。设计先独立通过，输出范围/证据/未测项明确；当前用户持续并行与独立review授权有效。

## 实际交付

modes_frontend只改apps/web九个文件：App、EditorPage、JudgeFeedback、UploadPage、DebugPage、lib/api、lib/mode、styles及verify-modes脚本。三模式默认student；同步取消请求、结果清空、mode查询key/工作区绑定和响应过期拒绝，保留同题源码。上传表单有题面/样例/资源/文件校验、普通配对与固定交互题模板，成功跳转题面；工程页只读debug。

实际Prettier/ESLint/TypeScript/Vite和8项可信模式检查通过；root真实浏览器上传、AC/WA分级结果、在途切换、工程证据/debug和手机布局通过，重启后页面9题/默认学生/无console warn-error，见[浏览器验收](../judge/BROWSER_QA_2026-10-03.md)。独立roles_review审读实际证据，root不冒称独立review。风险R2（前端隔离）/R3关联执行边界，最新状态[PROGRESS](../../PROGRESS.md)。

## 前端接续步骤

沿现有React/TS/Vite、Router、TanStack Query与Monaco，避免重复建框架。扩页先约定API DTO，再增加api.ts请求、查询key和取消边界；页面具备loading/empty/error/retry/禁用/过期状态。学生不可从缓存或教师结果恢复隐藏详情，服务端也必须按mode投影。搜索/筛选/分页/表单草稿/提交记录分任务实施，逐项键盘/手机验收；完整目录和验收清单见[框架指南](../frontend/IMPLEMENTATION_GUIDE.md)和[详细任务包](../frontend/AGENT_TASKS.md)。
