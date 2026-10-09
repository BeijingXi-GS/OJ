# JUDGE-006 多题目录与可信文件用例

- Owner /root/problem_catalog Builder；Review /root/judge_reviewer Architect/Builder。
- 允许修改：新internal/problems、internal/httpapi/problems.go /router.go及相关tests、internal/filejudge/contract.go /cases.go及相关tests、testdata新增题目目录、scripts/problem-smoke.mjs。禁止改engine.go /docker.go /apps/web /根MD /锁文件。
- 输入：[多题扩展裁定](../judge/PROBLEM_CATALOG_EXTENSION.md)，先等独立设计通过再修改核心接线；普通fixtures可先规划。
- 在A+B基础上增加array-sum /maximum-element /palindrome /sort-numbers，中文题面、约束、样例、C++20模板；每题至少4例覆盖样例 /边界 /负数或重复 /特殊输入，manifest仍用受审限额。
- 统一服务端compiled allowlist：API、compile请求、judge请求和Load只接受目录中的已登记ID。禁止用户任意目录 /paths；每题仅自己的快照和用例。
- MetadataProvider改内部Metadata(ctx,problemID)，client GET /problem/{id}，API使用当前请求ID。root负责worker对应路由和全部题目preflight。
- 验证：Go输入 /catalog /API /loader单测、新题独立正确解和错误解真实smoke；不以同一starter字符串回放来代替题目答案检查。无新增依赖。
- 实时状态仅PROGRESS；所有新未知阻塞报告root并停止，不改主机设置或自行绕过。

## 本轮交付证据

Catalog Builder已交付唯一编译期目录、API/loader/client所有入口allowlist、Metadata(ctx,id)及响应ID绑定和四题各5例，A+B仍4例，总24组。自己执行Go输入/目录/API/client/答案oracle tests通过；独立reviewer另执行fake tests/Node oracle通过。root通过problem-smoke.json十次独立正确AC/常见错误WA和清理；全部新题预检/详情/真实页面通过。无新依赖。详细 [验证](../judge/VALIDATION.md) 与 [交接](../handoffs/2026-10-02-judge-recovery-problems.md)。
