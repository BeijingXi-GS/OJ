# 2026-10-04 特判恢复工作交接与停止点

当前blocked，用户要求的最终增量尚未完整交付。停止原因：[自动审批通信再次中断](../issues/2026-10-04-approval-resume-interruption.md)，动作明确未执行；遵守用户外部阻塞归档后停止规则，没有绕过审批。

## 已完成和保留

内置不相邻排列、教师C++20特判源码上传/独立编译/受限执行/反馈、学生说明精简和工程师证据已经实现且局部独立复读通过。23项真实矩阵、Linux单测、完整check和两smoke均通过相应版本；详见[验证](../judge/SPECIAL_JUDGE_VALIDATION_2026-10-04.md)。教师页面已成功发布新题并真实AC2/2，页面具体事实见[QA](../judge/BROWSER_SPECIAL_QA_2026-10-04.md)。

分支task/SPJ-001-student-construction-2026-10-03；当前200非MD文件SHA256 `3406da9bef0efd1719a405bd48c2cc7206609368f0329f8012c5b6280bcf8df3` 是停止快照。镜像 `sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74`。没有新依赖、commit/push、删卷或覆盖原工作区。

最新补充：supervisor-stop-test等待exit+close、保留首错、独立时间戳报告、严格查询/端口检查及可信本地导入自测；mvp-dev新增5条无敏感值启动阶段日志。最后真实停止只bootstrap、未达到preflight，新日志版本未实跑；完整check在最后两项补充之前，不冒称当前全部源码已check。

当前13题、6上传、34文件的重启前快照和全部旧/失败证据保留；不要重新创建相同题或清除失败题来改变基线。16:56Web/API/Judge不可访问，Compiler8081 HTTP200，job0；最近成功Docker查询16:48为空，审批失败后未新查。标签保留为handoff，无有效截图。

## 恢复后的执行职责与顺序

1. **用户**恢复Codex自动审批连接并通知继续。已有项目修复/并行/独立审阅授权有效，不需重问；无需Docker重置或改安全设置。Agent先经正常审批渠道核实Docker Linux/cgroup2、锁定image/profile、owner残留、端口和job事实，失败仍如实停止。
2. **root / lifecycle_fix Builder**继续已知启动诊断。运行 `npm run mvp:stop-test`，查看最新独立 `spj-supervisor-stop-<时间>.json` 与mvp阶段日志，区分compiler health、旧judge探测、Go build、child启动及实际预检。可信本地Unicode/空格相对路径/file URL动态import和继承stdio均正常，不能断言它们是原因。不得仅增45秒预算、去掉阶段或清理断言、强杀Judge或手删job来制造PASS。
3. 三阶段必须实际确认preflight正在运行且未ready、idle已ready、running有case容器；每次Node exit0/stdio close0、helper确认、Docker查询成功为空/jobs0、端口ECONNREFUSED。新报告不覆盖 `.local/evidence/spj-supervisor-stop-2026-10-04T08-48-01-624Z.json` 和10-03失败档。
4. **root Backend Builder**严格重启 `npm run mvp`；执行已有 `.local/spj-persistence-20261004.mjs after`，与before快照比对全部6题题面/34文件，再逐题AC及清理。脚本只读保存的题目与数据，不替换旧源码；不重写before。实际提交使用X-OJ-Mode header和language/source字段。
5. root对本次自建特判题补单项完整工程师返回，明确输出故障message为checker_output_limit；持续输出fixture可能同时超时，须分清64KiB/16KiB展示截断与1MiB终止线，不得只看output_truncated。修改仅限本次拥有的生成题checker.cpp并finally恢复canonical，复验摘要/AC/0残留。不得触碰历史3上传题。
6. **root / Frontend Builder**使用原IAB标签或同一浏览器的新标签实际QA：常驻链接下载与当前源码一致；传统textarea真实输入/编译/提交；教师WA与反馈、学生CE及隐藏case/工程信息、工程师特判编译/hash/运行证据、移动布局和真实截图。当前静态代码修正并非下载/编辑验收完成。旧学生sort和工程师说明迁移、教师上传AC可参考记录，不重复冒称所有页面通过。
7. **root**完成最终 `npm run check`；核心Go若再变再做Linux专项，其他已通过检查仅在新变化要求时重跑。冻结最终实际source-digest，同步PROGRESS、四任务卡、使用指南、问题关闭证据及本机在线事实。保留Monaco体积等非阻塞历史项。
8. **spj_acceptance独立Architect/Security reviewer**只读复审变化代码及所有真实证据，绑定最终源码/镜像，明确待修0后才签产品PASS。当前[报告](../reviews/2026-10-04-special-judge-acceptance.md)pending；root自检不能代替它。**docs_curator**仅维护问题索引和特判用户指南；其他文件由root整理，恢复时先登记所有权。

## 可供使用的文档与例子

[详细教师/学生/工程师指南](../judge/SPECIAL_JUDGE_USER_GUIDE.md)、[判题契约](../judge/SPECIAL_JUDGE_CONTRACT.md)、[canonical checker.cpp](../../testdata/special-judge-example/checker.cpp)、[sample.in](../../testdata/special-judge-example/sample.in)、[sample.out](../../testdata/special-judge-example/sample.out)、[无解输入](../../testdata/special-judge-example/impossible.in)。问题索引见[issues README](../issues/README.md)。

产品仍为本机手动模式MVP：无真实登录、持久提交历史、任意交互器或完整生产Phase1，不能以当前增量扩大这些范围。
