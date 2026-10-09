# Docker恢复与五题MVP交接

日期2026-10-02；root集成；独立/root/judge_reviewer审阅。实时任务状态只在 [PROGRESS](../../PROGRESS.md)；本交接对应用户“尝试解决并继续推进任务进度，增加一些题目”。

最终结论：本机五题MVP验收通过，JUDGE-005/006/007完成；独立review正式PASS、必修0，最后完整check/Linux单测及109份MD本地链接检查通过。最终19:17:43独立数组和5例AC、四服务200、无判题容器/job残留。两项已知阻塞已关闭，暂无新的外部阻塞。

## 产品和运行

五题：A+B、数组和、最大值、回文、整数排序；完整中文题面、C++20演示模板、编译和运行双按钮、逐例/整体判定，共24组文件用例。本机 [题库](http://127.0.0.1:5173)，API8080、原compiler8081、judge8082。最终supervisor保留运行，关闭时应正常等待judge清理，不强杀执行中的judge。

本轮当前Docker Linux管道原不存在，启动已有Docker Desktop后恢复linux amd64 /cgroup2；没有删除数据卷、重置主机、调整隔离限额或打印.env。上轮11:54查询的确切原因仍未知，已改进脱敏诊断区分CLI失败与成功输出非2。

最终源摘要151文件/a07de43e447cea2b5af370225eafa13018a20cfee4d6c0bf011461f88026761f；分支task/JUDGE-005-recovery-problems-2026-10-02；镜像sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74。没有提交/推送或依赖升级，不覆盖历史未提交内容。源码最终修正只收束启动IO脱敏，执行profile保持。

## 实际分工与交付

| 执行者            | 任务与产物                                                                                        | 验证来源                                                          |
| ----------------- | ------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| root              | JUDGE-005：Docker恢复、固定诊断、全部题目worker/preflight、真实API/Docker/UI、统一文档            | 实际10次五题smoke、17项矩阵、工程/旧编译/Linux/浏览器/最终重启    |
| problem_catalog   | JUDGE-006：统一internal/problems、API/loader/client、跨题绑定、四题fixture及oracle、problem-smoke | 亲自执行Go tests/答案oracle；真实C++由root执行                    |
| frontend_followup | JUDGE-007：移除固定题号、补artifact DTO、复核切题草稿/结果/取消隔离                               | 亲自执行TS/lint/build/format；页面由root实测                      |
| judge_reviewer    | 独立Architect/Security设计、代码/失败路径、fake tests、答案oracle与实测证据审计                   | 亲自fake测试/Node oracle；只读root的Docker/UI原始证据，未冒充实测 |

## 检查、异常与范围

五题正确解AC/错误解WA全部通过，原17项矩阵包括AC/WA/TLE/MLE/CE/RE/OLE/SYSTEM_ERROR、真OOM与非OOM137、截断/UTF8、隔离/后代、busy/取消、文件故障及动态metadata/checker通过。完整check、smoke、mvp:smoke、Linux可信单测通过。实际页面覆盖五题、编译/AC/CE、双过期提示、恢复清空、在途切题/缓存及手机布局，证据见 [VALIDATION](../judge/VALIDATION.md) /[浏览器QA](../judge/BROWSER_QA_2026-10-02.md)。最后修正后受控关闭/0残留、19:15严格重启和最终AC复验另记验证末节。

B-JUDGE-002与B-MVP-001恢复条件已满足。开发期控制缓冲方法提升缺陷、Linux可信测试缺fixture及漏导入的修正分别在docs/issues记录前因后果。自动化fill/fullPage/confirm适配如实写入浏览器QA；没有新增未解决外部阻塞。已知Monaco大包警告继续非阻塞。

旧编译按钮仍只编译、不执行；严格OCI证据覆盖新judge内部编译/运行。CPU累计与可信内存峰值null，wall含启动开销。没有数据库题库编辑、账号、提交历史、持久化队列或公网部署；不声称完整Phase1/生产安全。后续新未知阻塞继续一事一档交用户，阻碍进度则停止。

## 下一步Agent任务（待认领）

1. Frontend Builder按既有 [详细框架任务包](../frontend/AGENT_TASKS.md) 开始题库检索/难度筛选：使用当前list DTO、先本机过滤，定义加载/空态/错误/重试和路由参数，完成桌面/手机及键盘操作，不自行引入账号/历史假数据。下一独立卡再做编辑器键盘焦点、确认取消和更细可访问性。
2. Catalog Builder新增题：先设计输入/输出/约束及Definition，再登记统一allowlist；每题4..8文件用例覆盖样例/边界/常见错解，独立oracle验证答案，扩展正确AC/错误WA smoke，检查全部题目预检及metadata绑定；不能仅复制目录就开放题目。
3. Backend Builder维护判题矩阵与启动/正常清理。数据库与提交历史另按P1受审设计实现不可变提交快照/异步Job/幂等结果，不将当前同步HTTP流程悄然替换成未审队列；资源峰值/CPU采集需可信来源和独立审阅。
4. 独立Architect/Security继续审契约与输入/输出/路径/生命周期边界，绑定源摘要/镜像/实测日期；不得以同一个执行者自检冒充独立评审。新需求先拆卡，再认领范围。
