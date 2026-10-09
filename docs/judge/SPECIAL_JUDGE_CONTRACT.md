# 学生页面与特殊判题增量契约

版本：special-judge/0.1.0 · 2026-10-03 · 状态：reviewed，[独立设计PASS](../reviews/2026-10-03-special-judge-review.md)后实施；实现/产品验收另行绑定证据。

用户本轮授权继续解决已有问题、精简学生信息、将实现说明放工程师模式，并明确要求“一部分内置，并且支持教师上传并提供示例和规则”。沿用此前并行开发与独立审阅授权。旧模式和交互协议保留；本机手动模式不是身份认证。

## 显示与兼容

- 教育提示保留在题目 notes；内置题默认解法、token/字节比较、容器、墙钟与启动开销等改到 engineering_notes，仅工程师题目响应包含此字段。教师上传的 notes 是公开解题提示，不自动按字符串删改。学生/教师整体结果文案避免容器/OOM等内部原因；教师可看具体例反馈，工程师可看实现证据。
- 时间/内存限制与交互协议中选手必须遵守的 flush、查询/回答格式是解题信息，保留。学生不展示比较器 ID、编译器版本、阶段耗时、运行环境/hash、判题详细用例。自己的编译错误保留。
- 老 manifest/上传题/判题 DTO 兼容，新字段均可选。Store保留strict JSON；worker metadata/result补充递归重复key检测，并拒绝未知/非法null字段。metadata 增加已允许的 batch checker。worker health 新增 special_judge_contract_version；启动器不复用缺少该标记的旧实例。API在提交前从受信本地目录加载快照，按该checker、checker源码SHA及testdata SHA绑定返回结果；普通tokens/exact/内置不得夹checker运行/编译摘要，交互也不得携带特判字段，不能只按problem_type=batch放行。

## 内置构造题

新 batch checker：nonconsecutive-permutation-v1。输入只允许一个十进制整数 n（1..1000），ASCII空白可存在；输出需是 1..n 的排列且相邻差不为 1。n=2/3 唯一合法输出为 -1，n=1 输出1。其他情况不能输出 -1。拒绝额外 token、非整数、重复、越界、数量错误。输入错误是题目配置错误，绝不归为学生 WA。

增加 construct-permutation 题及 .in 用例，不需要 .out。允许多个合法输出，规则ID/输入参与原快照摘要。教师可选该内置规则；原 tokens-v1/exact-v1 仍必须配对 .in/.out。

## 教师 C++ 特判

- 新 checker cpp-special-v1，仅 batch。manifest/上传 descriptor 的 checker_source 固定为 checker.cpp；multipart files 额外包含这一份源文件。每文件≤64KiB、总请求≤1MiB、源文件UTF8非空无NUL；只允许固定名称，不接受路径、脚本、编译参数、外部依赖或二进制。最多4例；每例必须 .in，可选同组 .out，缺省标准答案为空文件。普通/内置仍最多8例。
- 每次 JudgeRun 将 checker 源码读入不可变快照并计入 testdata SHA256。只由独立 Judge 编译它；API/宿主机永不执行。编译使用现有只读/无网络/nonroot/capdrop/seccomp/pid/memory沙箱，独立子目录与产物，限10秒/512MiB；选手编译20秒不变。特判源码编译失败或资源失败 → SYSTEM_ERROR，不是选手 CE。
- 每例先执行选手，RE/TLE/MLE/OLE/SYSTEM_ERROR 优先，只有成功退出且未超限才执行特判。把**原始未展示截断的**选手stdout快照为普通文件；独立受限容器运行特判，argv 固定为 `checker /data/input.txt /data/answer.txt /data/output.txt`，只读挂载这三份文件与自身产物，不暴露选手产物、其他用例、jobs目录或宿主配置。checker stdin为空，网络禁用，64MiB/32pids/2秒墙钟，输出stdout+stderr合计停止线仍1MiB，展示 stderr≤16KiB/stdout≤64KiB；容器和目录清理/taint策略沿用，任务总预算70秒不变。预算还包含容器create/inspect/cleanup等控制耗时，不保证所有最大规格都在70秒内完成；任何总预算耗尽均SYSTEM_ERROR并确认清理。
- checker exit0=AC，exit1=WA；其他退出、信号、超时/OOM/输出超限/基础设施失败均 SYSTEM_ERROR，不能误判选手 RE/TLE/MLE。反馈取 bounded stderr（最多1024 UTF8字节），教师可看；工程师额外看 checker运行退出/时间/输出/oom与源码/产物摘要、编译记录。学生响应不包含 checker或反馈，且 SYSTEM_ERROR 为普通重试提示。
- Result 可选 checker_compile/checker_source_sha256/checker_artifact_sha256，Case 可选 checker（结构包含exit_code、duration_ms、oom_killed、stdout、stderr、output_truncated）与 checker_feedback。Client严格校验长度、可空退出、状态组合及摘要；普通/交互题不得携带特判元信息。教师不返回源码/产物摘要、runtime或exitcode。源码上传成功不表示特判已经编译验证，教师页面明确说明，应先试运行验证。
- 提供不依赖testlib的 C++20示例、完整 .in/.out可选规则、退出值规范与坏特判系统错误说明；前端能读取checker.cpp，预填示例并下载/上传同一程序，选择规则时正确切换文件配对。内置例与可上传例的公开数学条件一致。

## 生命周期修复与验收

修复已有Windows共享终端退出提示；HTTP shutdown失败时只能在所拥有子进程正常退出、成功查询无本项目容器/任务残留之后登记已停，查询失败/非零退出/未确认清理保留失败，不能只因健康页断开而假设清理完成。不终止借用的外部实例，不读/输出token。

经独立追加设计批准：owned Judge 使用私有stdin pipe，父进程发送固定shutdown行，子进程仅非空control token时以16字节行上限启用读取，收到该行或EOF/pipe错误取消早注册的生命周期ctx，在预检/ready阶段均等待独立清理。无需新端口、路径或命令；无token standalone不读stdin；发送成功不作为退出确认，最终仍检查owned exit0、实际无残留。启动器构建异步，停止Promise唯一，每阶段await后检查stopping，不因同步构建退出跳过清理。

必须：适用单测/vet/typecheck/lint/build、完整check；真实内置两种不同答案AC与错误约束WA；真实教师上传特判 AC/WA、特判CE/非0非1/超时/输出限→SYSTEM_ERROR、选手CE/RE/TLE保持；上传重启可判/摘要绑定/0残留；旧batch/interactive回归；真实学生无工程细节、工程师说明可見、教师特判上传/反馈；正常与运行中退出确认。独立Architect/Security审阅设计与实现，最终绑定新源码摘要、实际证据，未执行不得登记通过。

新未知外部阻塞一事一档由用户处理并停止；本增量常规实现/审阅修正由已授权执行者完成并如实记录。无新依赖、提交/推送或数据卷删除。
