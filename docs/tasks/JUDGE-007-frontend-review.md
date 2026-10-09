# JUDGE-007 多题前端回归与独立验收

- /root/frontend_followup Frontend Builder只改apps/web，读取现有DTO；验证多题列表、导航切题时草稿 /双结果 /过期 /请求取消的隔离；必要时小范围修正，避免题A的结果留在题B，无新依赖。
- /root/judge_reviewer 独立Architect/Security，先审扩展裁定，再只读实现 /测试 /runtime证据；仅写docs/reviews/2026-10-02-judge-recovery-problems-review.md，不参与实现，不自行执行不可信payload或修改环境。
- root运行实际UI和Docker集成，reviewer可独立运行fake Go单测；严格区分证据来源 /未测项。
- 依赖JUDGE-005/006及局部设计；完成后统一更新PROGRESS、交接与使用MD，不伪造浏览器完成。

## 本轮交付证据

Frontend Builder仅改EditorPage固定题号与lib/api artifact_sha256类型，无新依赖；亲自TS/lint/build/format通过。既有按ID查询/重建Workspace、AbortController与信号检查复核通过。root实际 [浏览器QA](../judge/BROWSER_QA_2026-10-02.md) 已覆盖五题、编译/AC/CE、过期/恢复/在途切题/缓存和手机布局，实际截图/完整AX存.local/evidence；工具操作适配如实记录。独立reviewer亲自fake tests/答案oracle、只读代码及root真实Docker/UI证据，正式结论见 [报告](../reviews/2026-10-02-judge-recovery-problems-review.md)。CPU/内存峰值未采集、未测分支保留。
