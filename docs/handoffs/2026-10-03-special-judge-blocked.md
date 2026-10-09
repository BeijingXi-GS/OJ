# 学生页面与特殊判题：外部阻塞交接

日期2026-10-03，当前停止位置见[B-SPJ-001](../issues/2026-10-03-spj-approval-transport.md)。用户本轮确认内置与教师上传特判、示例/规则；此前并行与独立审阅授权持续。root已通知全部本轮Agent停止新增工作。

## 已交付实现

- root：工程师only题面说明、学生/教师总体结果简化、教师反馈/工程师checker分级投影、worker严格重复/unknown/null校验、可信快照checker源码/数据摘要和完整同顺序case绑定；停止启动竞态修复，私有pipe预检取消，真实退出与残留核实后确认。
- special_backend：内置nonconsecutive-permutation-v1与construct-permutation六组输入，教师cpp-special-v1固定checker.cpp/≤4例/可选.out、安全源码读取/持久化/快照摘要、10秒特判编译/2秒64MiB特判容器、原始stdout参与判定、checker失败SYSTEM_ERROR且选手错误优先，源码示例和真实验收脚本。
- student_frontend：学生页面工程信息清理、保留任务格式/时间内存/flush/自己的CE，工程师说明与checker证据；教师四种batch规则、可选参考答案、C++源码读取/编辑/示例/下载/上传、反馈。只有apps/web改动，无依赖变化。
- spj_review：已真实独立设计与局部实现PASS、现无代码必修；最终产品与冻结仍pending，报告[review](../reviews/2026-10-03-special-judge-review.md)。

示例：[checker.cpp](../../testdata/special-judge-example/checker.cpp)、[sample.in](../../testdata/special-judge-example/sample.in)、[sample.out](../../testdata/special-judge-example/sample.out)。协议与公开数学条件见[契约](../judge/SPECIAL_JUDGE_CONTRACT.md)：argv按输入、可选参考答案、选手输出，退出0=AC/1=WA/其他=SYSTEM_ERROR；内置构造n2/3只允许-1，其他为1..n排列且相邻差不为1。前端示例已和文件一致性测试绑定。

## 实际检查与未完成

可信Go tests、vet、4退出helper、privatepipe控制单测、前端verify-modes/typecheck/lint/build、局部独立review通过；rootAPI/协议两轮专项通过。新真实预检私有pipe取消一例exit0且helper确认清理。两个手动WindowsCtrl+C尝试：旧syncbuild版本未确认日志、async版本确认日志但外层PowerShellexit1，仍需脚本直接观察Node真实退出码。详细[开发修正档](../issues/2026-10-03-special-development-corrections.md)。

完整check第一轮format失败且未重跑；三阶段真实停止脚本第一轮assert失败，报告completed:false；重跑被审批通信失败阻止，动作未执行。special-judge-test.mjs未执行，Linux补测、真实teacher上传/重启、浏览器QA、最终源码冻结与独立产品PASS尚未做。不能拿旧MODES通过证据替代。

停止快照：200非MD文件 SHA256 `bb1ac048576337b8b8a67f29861b2c6dbe8dce43b50bac50de4f47aa45d34166`；锁定镜像仍 `sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74`。无新增依赖/提交/推送/删卷，保留全部原修改和3道旧上传题。没有本轮真实特判上传新题成功证据。停止时Web/API/Judge关闭、Compiler8081健康200，jobs0；最近成功Docker查询0owner，失败审批后未再次查询。

## 恢复后的Agent职责与顺序

1. root收到用户连接恢复/继续后核实审批和实际环境，保留数据，确认0残留。
2. root定位supervisor-stop-test日志观测缺口，预检/idle/Running分别实际验证并保存正确子进程退出码；共享Windows信号路径单独证据，不能忽略错误。
3. root完整check（Go+新增Node helper及前端SSR专项已加入npm test）、Linuxfilejudge/storesymlink；失败如实记录修正。
4. root正常启动并严格preflight，串行judge:special:test真实矩阵：两合法答案/错误约束、源码绑定、原始输出截断、选手CE/RE/TLE、checkerCE/exit2/137/timeout/OOM/OLE、三投影、旧batch/interactive；仅脚本自建题可以注入并finally恢复。
5. root验证上传题重启可判、全部文件摘要保留、0残留；使用真实页面验证学生无技术说明、工程师可见、教师选择/读.cpp/示例下载/上传/反馈、必要手机布局；再保持服务可访问。
6. student_frontend仅修复实际UI反馈；special_backend仅修复store/engine专项反馈。root统一issues/验证/用户指南/task/PROGRESS/交接，spj_review独立签最终实现/产品及新源码摘要。完成后才done。

新未知外部阻塞仍按用户规则一事一档并停止。此MD是当前可恢复交接，不是最终产品验收。
