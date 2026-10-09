# 双角色通信题独立设计与实现审阅

日期：2026-10-05；独立审阅者：`/root/comm_review`，Architect / Security。依据 [AGENTS](../../AGENTS.md)、[开发流程](../../DEVELOPMENT.md)、[基础不变量](../architecture/FOUNDATION_INVARIANTS.md) 与 [通信题增量契约](../judge/COMMUNICATION_CONTRACT.md)。

用户选定发送端／接收端两份 C++ 源码，并明确允许先交付不依赖 Docker 的本地实现与可信测试，真实容器验收保留待办。该范围例外没有取消独立设计门禁，也不将历史 SPJ 矩阵转换为本次通信题的验收证据。

本审阅者只读代码、文档和后续可信本地测试，仅可写本报告；不修改契约或实现，不控制 Docker、真实服务、浏览器或 Git，不在宿主运行选手／教师源码。

**当前设计结论：PASS。** 2026-10-05 恢复后的独立复读确认，通信契约“规范化和验收门禁”已经实际补齐下述四项状态约束与教师 descriptor 上限说明；未解决设计必修为 0。允许按受审的固定双角色 MVP 范围实现。实现与产品结论分别登记如下，设计 PASS 不代表沙箱已经验收。

**上一版本本地实现与可信测试结论：PASS，未解决实现必修 0。** 下述五项实现发现均已独立复读修正并验证，root 完整工程检查日志及 226 文件源码清单已经独立核读／核对。用户随后新增最简单通信 A+B 题，源码已变化；旧 PASS 与摘要保留为历史绑定，不能直接证明新增题目／模板已通过。**新增 A+B 窄范围代码与本人已执行的可信专项测试 PASS，新增必修 0；新完整工程回归未通过，现因文件权限阻塞停止依赖验收。** 本次 reviewer 全包检查及 root 后续权限复验均 exit1，不能登记为新完整 PASS，详情见新增复审记录。真实 Docker／页面／持久化／旧停止待办仍待实际证据。最终产品 pending，没有授予产品 PASS。

## 已独立复读的设计

当前完整草案限定固定单向 sender → receiver 图，以显式角色身份绑定两份 C++20 SourceUnit、分别编译的 Artifact 及两个独立运行容器；请求数组顺序不决定角色。旧单源码题型拒绝多源码与通信证据，新端点拒绝普通题型，API 始终不执行代码。

发送端仅获得自己的产物和本例输入；接收端仅获得自己的产物。标准答案仅用于可信比较，不挂到接收端。代理对发送端 stdout 原始字节计量，仅转发通信限额内的前缀；观察字节数、成功写入受控管道的字节数和程序实际消费字节数明确区分。EOF 关闭接收输入，已确认的接收端提前退出与未知传输错误分开处理，不能无条件吞掉 IO 失败。

两个角色共用从创建／启动开始的单一墙钟预算，取消、通信越界和角色故障须关闭所有管道、收束两个 attach、独立确认两个容器清理。清理不可信须 taint / SYSTEM_ERROR；真实 OOM 不可被监督器终止例外掩盖。

逐角色内存限额与固定两倍的 aggregate_memory_limit_bytes 被明确称作并行限额上界；没有真实聚合 peak、父 Job cgroup 或完整 Phase 8 图实现声明。该受限 MVP 不因上述明确排除而被要求在本轮补齐完整 Phase 8，但 CPU / 内存 peak 必须保持未采集状态，不能相加独立进程峰值。

请求编码上限单独为 1 MiB，保留每角色 64 KiB 原始源码上限与旧端点 262144 字节上限。严格 JSON、角色集合、摘要规范化、可信题目快照／字节配置／用例顺序绑定、三种模式投影及新 worker 契约标记均有设计。学生只保留本人逐角色编译诊断，不获得隐藏用例／通信／运行日志；教师按有界白名单看方向输出和通信量，工程师再看退出／OOM／终止／摘要。

## 首轮设计必修及门禁（历史）

当前落盘草案仍有以下状态形状未完全明确；root 已说明拟补充其中第 1–3 项，但其补丁审批通信中断，工具明确动作未执行。审阅者没有代写或重试该失败动作。下面是独立审阅要求，不代表补丁已成功或核心实现已获准开工。

1. **聚合产物摘要只在两个角色均编译成功时非空。** sender 成功、receiver 编译 CE / SYS 时，sender 自己的 artifact_sha256 必须保留，但 top artifact_sha256 为空；失败角色的产物摘要为空。不能把部分成功的产物序列摘要伪装成可运行双角色产物。编译汇总 success 必须等价于两条角色记录全部 success；CE / 编译 SYS 零运行用例，运行阶段 SYS 可以保留已产生的用例前缀。
2. **明确旧 Case 字段与双角色字段的关系，并独立校验通信状态。** 旧 case.stdout / exit_code 映射 receiver；oom_killed 为任一角色真实 OOM；stderr 为有界组合，原 CPU / 内存 peak 为 null。sender RE 而 receiver 正常退出时仍可 RE，监督器确认停止的 receiver 137 也不能通过普通单进程断言误判为自然 RE。每角色运行对象才是对应退出的证据，不得从 receiver 的旧字段反推 sender 状态。
3. **terminated_by_supervisor 必须来自实际确认的控制终止证据。** context 取消、发出 kill 指令或 attach 返回取消本身均不够；最终 daemon 状态与监督器动作确认后才允许豁免相应退出。自然非零退出不能因同时出现取消而被改成正常结束；OOM 始终保持高优先级。AC 要求双方正常结束、EOF／传输状态可信且没有终止例外。
4. **允许且限制编译开始前的 SYSTEM_ERROR 形状。** 快照读取／前置基础设施错误可能发生于第一份编译前，应允许零条 role_compiles、compile.status=system_error、零 cases、空 top artifact；仍须保留请求的规范 SourceUnit 摘要。进入编译后的记录必须是按 sender、receiver 的真实已尝试前缀：一条只能为 sender，两条的角色固定且唯一；sender CE / SYS 时不得伪造 receiver 成功。仅早期 SYSTEM_ERROR 可有零条记录，CE 不允许零条记录。

另一个需在教师规则／指南中如实说明的上限组合：现有 descriptor 编码上限 64 KiB 包括题面和双模板。每份模板的原始上限 64 KiB 不意味着两份最大模板与最大题面可以同时发布；无需为这个说明扩大旧上传上限。

在上述必修实际成为受审设计的一部分之前，结论为 **设计需补充，未最终 PASS；核心实现门禁尚未通过**。本报告不以“条件通过”替代门禁。其余已复读设计没有新增阻断发现。

## 恢复后的独立设计复审

root 通知读取恢复、此前补充已实际写入后，本 reviewer 亲自重新读取完整通信契约和本报告，没有代写契约，也没有依赖“拟修改”消息签通过。契约新增落盘段落明确：

- top artifact_sha256 只在双角色成功时非空，编译记录是实际尝试的 sender 前缀；编译前 SYSTEM_ERROR 可为零记录／零用例／空 top artifact，但规范请求源码摘要仍保留，CE 不能为零记录。
- 兼容 case 的退出与输出来自 receiver，OOM 为任一角色真实 OOM；通信校验使用逐角色证据，CPU / peak 为 null，不套单进程推断。
- terminated_by_supervisor 需要控制终止和最终 daemon 状态的实际确认，context / kill 请求 / attach 取消均不够；真实 OOM 始终优先，提前关闭仍须可信角色状态与 EOF 交接。
- 教师 descriptor 沿用包含题面和双模板的 64 KiB 编码总量，不将每角色原始源码上限误写为组合上传保证。

四项首轮必修及上限说明均已解决，固定双角色图、角色编译绑定、单向原始字节／EOF、失败传播／清理／taint、严格请求与响应、API 分级投影的设计没有新增必修。**communication-mvp/0.1.0 局部设计 PASS，未解决设计必修 0。** 完整 Phase 8、父 Job cgroup / 实测聚合 peak、生产身份／持久提交不在此次批准范围；真实容器验收必须另行补齐。

## 实现与验收状态

设计批准前没有执行新增通信测试。设计批准后的实现复读与实际可信检查记录如下；这与真实 Docker 验收分开。

后续可信本地复审至少覆盖：反序提交的角色显式绑定、缺失／重复／未知／null／递归重复 JSON 字段、双角色最大转义请求、sender CE 与 receiver CE、部分编译产物保留、提前 SYS、EOF 与接收端提前退出、部分管道写入及未知 IO 失败、通信限额前缀与字节计量、共享输出限／墙钟／双方取消、真实终止证据分类、双清理失败及 taint、严格 DTO 状态组合、数据／源码／产物摘要错配和三模式投影，以及旧 batch / interactive / special 回归。

真实 Docker 通信矩阵、挂载隔离、真 OOM／RE、并发取消与无残留、教师上传重启、双编辑器真实页面，以及旧三阶段停止／持久化待办均为 **未执行 / pending**。最终产品 **pending，未授予 PASS**；历史 23 项 SPJ 真实矩阵不证明任何新增通信图的真实执行。

## 独立实现复读与可信检查

本人独立复读 `communication.go` / `communication_contract.go`、共享 Request / Result / Client / snapshot 校验、Engine 双角色分支和内部路由、Docker stdin／清理证据扩展、API 路由／元数据／三模式投影、Store／manifest／multipart 双模板与限额、两道内置题和数据，以及 CommunicationWorkspace / EditorPage / UploadPage / JudgeFeedback / 前端接口和规则。

当前代码明确绑定 canonical sender、receiver 源码与各自产物；首次编译失败停止下一角色，保留实际前缀与已有成功产物，只有两份成功形成总产物摘要。两个运行容器独立使用本角色产物；sender 才获得当前输入，receiver 没有输入／答案／sender 产物挂载。代理有界保留消息并统计原始字节，超限只转发限额内前缀；接收端提前正常退出必须有 daemon／attach 证据，未知 stdin 传输或 attach 故障仍 SYS。

共享时间从容器创建开始；双方输入关闭、attach 收束与独立 15 秒双容器清理采用同一生命周期。OOM、墙钟、共享输出、通信越界及自然 RE 按受审优先级映射，清理失败 taint 并保留现场。元数据的两倍内存限额只作为上界，CPU／真实峰值保持 null。请求与 worker JSON 拒绝重复字段、大小写别名、未知字段和非法 null，API 绑定真实本地数据／字节限额／用例顺序、各角色源摘要和规范产物摘要后再投影。

学生响应没有用例或方向日志／通信计数／摘要，仍提供本次实际尝试角色的编译诊断。教师仅使用约定的有界方向／计数白名单，工程师保留逐角色退出／OOM／监督终止／摘要。双编辑器按 role 标识与独立 Monaco path 保存，切模式取消请求并保留双草稿，任一源码变动使结果过期；通信上传要求双模板、正确限额及配对数据，原普通／交互／特判路径保持分叉。

本人实际执行的可信检查：

- `node apps/web/scripts/verify-modes.mjs`，exit0：已有模式／取消／普通与特判规则回归，通信角色／独立端点／UTF-8 与字节上限／配对数据／角色 CE／学生隐藏／教师与工程师投影／实际组件 SSR 通过。
- `node scripts/go.mjs test -count=1 ./internal/filejudge/... ./internal/problemstore/... ./internal/problems/... ./internal/httpapi/... ./services/judge-mvp/...`，最终 alias／null 修正后的独立重跑 exit0；filejudge 8.690 秒、problemstore 0.408 秒、problems 0.131 秒、httpapi 0.485 秒、judge-mvp 0.157 秒，全部 PASS。此前 7.551 秒重跑保留为对应旧检查事实。
- `node scripts/go.mjs vet` 对同五包执行，exit0。
- root 新增 strict wire／canonical binding／termination 三项测试本人另外执行，exit0；完整五包重跑包含这些测试及后端观察证据修正。

这些测试只使用可信 Go／Node fixture 和 fake Docker 控制工具，不编译或执行宿主 C++。Windows 符号链接跳过不当作 Linux 文件安全通过；组件 SSR 不当作真实浏览器交互通过。本 reviewer 没有亲自执行前端 typecheck / lint / build 或真实 Docker，其他执行者结果须单独按来源记录。

## 实现审阅发现与解决

1. **监督终止 shape 可伪造。** 初版 `validRoleRun` 只要求终止角色退出码非空，允许 true + exit0 / exit2。root 已改为 true 必须 exit137，并新增 true + 0 / 2 / null 拒绝、supervised137 TLE 和 natural137 RE 测试；本人复读并运行通过，此项关闭。
2. **共享预算中的部分启动超时误归 SYS。** 初版只要 launched 就要求两角色 StartedAt，可能把 sender 已启动／receiver 尚未启动时的真实墙钟截止误判 SYS。后端已允许有最终可信状态的 unstarted 角色为 TLE，保留 nil exit、不虚构 kill，增加 partial_start_timeout 可信路径；本人复读及完整测试通过，此项关闭。
3. **缺失容器与已确认未启动的零状态混淆。** 上项第一次修正还把 matching 空返回的零值当真实未启动。后端新增 cleanupWithObservation 与通信专用并发清理：仅最终独立 inspect／身份确认取得状态时 observed=true；启动后 observed=false 归 SYS。partial_start_missing 与合法 partial_start_timeout 分别断言 SYS / TLE；本人复读及完整测试通过，此项关闭。旧清理签名由包装层保留，没有放宽旧交互行为。
4. **部分启动页面说明需对应角色事实。** JudgeFeedback 原“程序尚未开始运行”不能准确描述 sender 已运行、receiver 未启动的通信 TLE。前端已按通信题写至少一个角色未完成启动，并保留逐角色证据；本人复读新文案与 SSR 断言并实际重跑 verify-modes exit0，教师不露工程句、旧交互句保持，此项关闭。
5. **上传／Store 的 JSON 大小写别名可覆盖配置。** 经 root 要求追加检查，初版 `problemstore.strictJSON` token walk 只拒精确重复键，不能拦 `encoding/json` 的大小写字段匹配。通信配置 `{byte_limit:32, Byte_Limit:1}` 可能被后值覆盖，顶层 `communication:null` 与 `Communication:{...}` 也可形成非空配置；这不符合严格未知／重复／null 约束。后端已将固定 schema／role 键限定为规范小写 ASCII，拒绝 communication / starter_codes 显式 null，新增 multipart、manifest Parse／Store.Get reload 与旧题型拒绝检查；合法 canonical 恢复后仍可读取。root 对 filejudge 同步改成 ASCII canonical key，而非仅 ToLower，补 long-s `ſource`／`ſender` 折叠别名拒绝。本人独立复读这些代码与测试，并对最终五包 test／vet 实际重跑通过，此项关闭。题面和源码字符串值不受 ASCII 字段名限制；本人没有代改实现。

五项实现发现全部闭环，当前受审范围没有遗留必修；**通信增量本地实现审阅与可信测试 PASS**。真实产品结论始终 pending，与本地实现审阅分别登记。最终工程检查／源码清单绑定见下节；恢复后的实际镜像／真实矩阵／浏览器／持久化须另外验证，不得将本报告局部 PASS 登记为最终产品验收或完整 Phase 8 完成。

## A+B 增量前的本地版本与工程检查绑定（历史）

root 实际执行完整 `npm run check`，exit0；本 reviewer 独立读取 `.local/evidence/communication-20261005-check.log` 的 format／Go format、Go vet／前端 lint、全部 Go 包测试／4 项停止 helper／verify-modes、Go build／TypeScript／Vite build 日志，全部阶段完成。既有 Monaco chunk 大小提示保留为非阻塞警告。执行者是 root，reviewer 只复核其日志，不登记为本人再次执行完整 check。

本 reviewer 亲自读取 `.local/evidence/communication-20261005-local.json` 的全部 226 个非 MD 源码条目，逐个重新计算当前原始文件字节 SHA256，并独立核对排序／唯一性／文件数及总摘要。结果：226 条唯一，所有列出文件与记录逐一相同，mismatches 为空，复核命令 exit0。没有执行 Git 命令；tracked／untracked／非 ignored 的清单生成范围由 root 负责，本次复核明确针对其清单及当前文件，不冒称重新生成 Git 索引范围。

本轮增量前的历史绑定：226 个非 MD 文件，SHA256 **`80a5377ba7141f550cdd2d8979a9a632e1cb12bab1ef678a170280be32a09936`**；计算公式为 `SHA256(JSON.stringify(sorted [{path, sha256(raw file bytes)}]))`。此公式与历史 `scripts/source-digest.mjs` 的拼接文本摘要不同，不混用。分支 `task/SPJ-001-student-construction-2026-10-03`、HEAD `7a952c68414c1aeccef9af2475ec3e43124e292b` 是 root 提供的版本记录，本 reviewer 未操作 Git。当时冻结后只改 MD；随后新增 A+B 的非 MD 源码已经使该摘要成为历史记录，须另行核对新的完整检查与源码绑定。

root 当时最后显式 Docker Linux pipe 诊断仍失败，没有当前通信容器矩阵／Linux 文件安全／浏览器／持久化／停止验收。本 reviewer 未操作 Docker 或服务，不能确认当前实际镜像或无残留。**该历史版本独立结论：局部设计、该绑定源码的本地实现／可信测试与已复核工程检查 PASS；真实产品 pending，未授予产品 PASS。** 新增 A+B 的当前结论另行登记。

## 最简单通信 A+B 增量独立复审

用户新增要求为发送端传递 a、b，接收端求和。该增量沿用已受审的 communication-v1／双角色 schema，不扩大通信图或引入新执行路径。本人重新读取 `internal/problems/catalog.go`／catalog tests、`internal/filejudge/catalog_test.go`／communication tests、新题 `testdata/problems/communication-a-plus-b` 的 manifest 与全部四组数据，以及前端 `communication.ts`／`UploadPage.tsx`／`verify-modes.mjs` 的增量。

内置题 `communication-a-plus-b` 的 sender 模板只读取并输出原始 a、空格、b、换行；receiver 独立读取消息中的两个数并输出 a+b。双方使用 long long，输入范围 ±10¹²，和在 ±2×10¹² 内。最坏两个负数消息为 30 个原始字节，低于 32 字节限制。旧 legacy starter_code 等于 sender 模板，双模板按角色存储并沿用深拷贝。manifest 为 communication／communication-v1、32 字节、5000 ms、每角色 64 MiB，四组同名 .in／.out 覆盖普通、零、负数和大整数：1+2=3、0+0=0、−7−11=−18、10¹²+10¹²=2×10¹²。独立 Go 数据 oracle 按新题范围计算，公开样例与标准答案一致。

教师新增 A+B 示例自带 ±10⁹ 范围、2 3→5 样例、32 字节和分别传数／求和的双模板；该上传示例自己的范围与模板自洽，无需与内置题采用相同范围或样例。原整数序列求和示例保留。新增 SSR 使用该题 ID 和双模板，检查 sender.cpp／receiver.cpp 独立路径、题面／样例、双编辑器及三模式信息边界；上传页提供独立填充按钮和职责说明，源码核心契约没有变更。

本人实际执行的新增检查：

- `node apps/web/scripts/verify-modes.mjs`，exit0；新 A+B fixture 与此前模式／取消／编译反馈／信息投影断言通过。
- `node scripts/go.mjs test -count=1 ./internal/filejudge/... ./internal/problems/...`，本次 **exit1**：problems PASS 0.180 秒；filejudge 3.361 秒失败，日志包含临时目录 rename `Access is denied`，HTTP client 测试集中 `judge_unavailable`，以及依赖 Store 临时操作的 `problem_store_unavailable`。当前环境网络受限；sandbox 网络／临时目录权限是原因推断，未确认全部底层原因，不将这些失败改写为源码回归或完整 PASS。已立即通知 root，不调整权限／环境、不循环重跑完整包。
- 随后仅执行不依赖上述 HTTP 客户端／重命名路径的 `node scripts/go.mjs test -count=1 -v -run '^(TestRegisteredFixtureAnswers|TestCommunicationWorkerRoutesAndTrustedMetadata|TestCommunicationJudgeSuccessfulBindingsAndRuntimeFailures)$' ./internal/filejudge/...`，**exit0，0.993 秒**。新题四组数据／公开样例、题型／四例／32 字节／角色 metadata，以及 pair_sum AC、pair_wrong WA 均通过。可信 fake sender 从独有挂载读取并传递 `1 2\n`，fake receiver 确认原消息／EOF后求和；双方产物、严格 wire、请求／快照绑定、四个通信字节和清理结果均被断言。

上述判断由可信 Go fixture 模拟两个角色，不能证明 C++ 模板已经在真实 Linux 容器中运行；本人未在宿主编译或执行两份源码。新增题目／模板／前端示例的**窄范围独立代码审阅与本人已执行可信专项 PASS，新增必修 0**。226 文件历史摘要不沿用；新完整版本必须另行核对源码清单，真实产品验收 pending。

## root 后续权限复验与停止边界

root 反馈已正常取得 network grant，并将 TEMP／TMP 指向项目内 `.local/tmp`；HTTP 相关测试恢复，但 Store 目录 rename 和可信 fake program 的 unlink 仍报 AccessDenied。root 随后又通过 request_permissions 获得 `.local/tmp` 精确 read／write turn grant，专项复验仍 exit1，未将授权返回等同实际文件操作成功。

本人仅只读核对 `.local/evidence/communication-ab-20261005-permissions-recheck.log`：`TestCommunicationUploadPersistsOwnedTemplatesAndConfig` 仍 `problem_store_unavailable`（problemstore 0.189 秒 FAIL）；通信 pipeline 18.263 秒 FAIL，含 pair_sum、pair_wrong。两项已经生成正确的 case AC／WA 与输出 3／5，但 receiver-build/program 删除被拒，最终结果按清理失败优先级为 SYSTEM_ERROR。case 判定正确不能替代整个 job 成功；这些 root 后续失败不改写本人此前另一临时目录下 0.993 秒专项 exit0 的执行事实，也不支持声称当前环境完整回归或清理通过。

root 明确按用户阻塞后停止的约定，停止依赖验收，不再复验或调整权限。reviewer 同样停止测试与环境操作，仅补记录、核对新的源码清单。旧 `80a537…` 完整 check 是 A+B 增量前的历史证据；新完整回归**未通过／环境阻塞**，不能用历史 check 或新增窄范围 PASS 替代。

真实 Docker 的最新边界由 root 提供：Linux engine pipe 访问为权限拒绝，而该 pipe 无法作为可表示的普通路径授予文件权限。此处与历史 pipe missing 状态分开记录，不诊断为已修复或当前 Docker 可运行；本人没有操作 Docker／权限／服务，真实容器验收仍 pending。恢复后应先解决 root 已登记的权限问题，再实际完成 Store 发布／删除清理与完整回归，最后进行真实 Linux 通信矩阵及页面验收。

## 当前 A+B 源码清单绑定与最终边界

root 在停止依赖验收后生成 `.local/evidence/communication-ab-20261005-local.json`，其 source 为 count／digest／files；当前源码从该清单冻结，之后只收尾 MD。本人只读全部条目，并逐一重新计算 235 个当前非 MD 文件原始字节的 SHA256。独立复核结果：235 条、路径唯一且按字典顺序排列、mismatches 为空；重新计算 `SHA256(JSON.stringify(sorted [{path, sha256(raw file bytes)}]))` 与清单／root 指定摘要完全相同，复核命令 exit0。

**当前绑定：235 个非 MD 文件，SHA256 `71e866f2e74ab2f8e9d107a8cacaf9240ed03ded1c87bf963821b8ba9a3cb634`。** reviewer 没有执行 Git，清单的 tracked／untracked／非 ignored 范围由 root 生成；本次独立核对证明列出的当前文件与该清单逐一相同，不冒称重新生成 Git 范围。该摘要包含新增 A+B，与旧 226 文件／80a537… 历史完整检查明确区分。

清单明确记录当前 root 五包回归 exit1、精确权限复验 exit1，以及前端 Agent 的 typecheck／lint／build／SSR／format exit0。本 reviewer 已独立执行的 A+B 专项及 SSR 结果仍按前文实际事实登记；owner 的前端结果不改称 reviewer 全部亲自执行。root 因权限复验仍失败后遵守用户停止约定，**没有执行新增后的完整 npm run check**；源码摘要核对只证明版本一致，不能补成完整工程检查成功。

**最终独立结论：受限通信设计 PASS；该绑定源码新增 A+B 的窄范围代码审阅／本人已执行的可信专项 PASS，新增必修 0；当前完整工程回归未通过，文件权限阻塞，依赖验收已停止；真实 Docker／页面／持久化验收 pending，没有授予最终产品 PASS。** 本次收尾仅核对证据和更新本报告，没有重复测试、调整权限、操作 Docker／服务／Git 或运行宿主 C++。
