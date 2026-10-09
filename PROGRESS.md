# OJ 开发进度与任务台账

**唯一实时进度入口** · 更新日期：2026-10-09 · 当前阶段：**源码与文档保留；已授权创建分支并合并，但连接器建分支实际403，远程交付blocked**。

## 2026-10-09 Git 分支、推送与合并接续

用户要求分析并解决GitHub上传问题，创建Git分支并合并；本次明确授权合并，覆盖此前仅draft PR的范围。root认领PUB-001，允许修改交付/问题文档及必要Git元数据、正常推送和合并；不强推、不弱化分支保护、不改变仓库可见性、不部署产品。已重新读取现状，正常权限请求只返回本轮network grant，未返回.git写入grant，因此尚未写本地Git元数据。具体诊断和交付结果随后在本节与问题档案登记。

实际结果：Git直连github.com:443失败/curl超时；Git已经HTTP/1.1，无已读进程/Git代理配置，系统代理/DNS诊断因审批传输失败未执行。GitHub只读正常，main仍7a952c6；用户选择连接器恢复路径后，建分支请求到达GitHub但明确403 `Resource not accessible by integration`。**未创建分支/commit/PR/merge**，未写本地.git；停止写操作，等待正常连接器仓库写授权恢复。账号admin/push不等于集成令牌Contents写权限，插件已启用也不证明写权限。[详细原因、官方依据与对策](docs/issues/2026-10-07-github-write-access.md)、[本轮交接](docs/handoffs/2026-10-09-github-blocked.md)。Docker/代码测试未重跑，旧包不含本轮诊断MD。

收尾：本轮6份MD的Prettier exit0，git diff --check exit0；181份MD/855个本地链接目标缺失0，250份非MD文件仍与已审快照逐项一致，常见秘密模式和排除文件命中0。证据 `.local/evidence/delivery-20261009-audit.json`；不是完整安全扫描或真实Docker验收。

## 2026-10-08 工作补齐、安全盘点与上传复验

root（Architect/Builder）完成 PLAN-002 / SEC-001 文档内容：[NEXT-001..006实施工作包](docs/planning/2026-10-07-implementation-backlog.md)包含持久提交、前后端接口/页面/异常状态、异步恢复、教师质检、公开通信回放；[单列恶意代码方案](docs/sandbox/MALICIOUS_CODE_SECURITY_PLAN.md)包含22类行为、当前防护与缺口、SEC-002..006整改和12组验收。安全整改仍待实施，当前交付不改变本机手动模式或冒称Phase4通过。独立security_gap_review只读并写[审阅报告](docs/reviews/2026-10-07-security-gap-review.md)；检查收尾与交付见[本轮交接](docs/handoffs/2026-10-08-security-plan-delivery.md)。

新增确认问题：[B-SEC-001旧编译入口隔离](docs/issues/2026-10-08-compile-isolation-gap.md)、[B-SEC-002配置验证未闭合](docs/issues/2026-10-08-sandbox-profile-verification-gap.md)。模式无真实身份、worker本地入口信任、累计配额缺失亦在安全方案中单列。均为静态缺口，不是已发生入侵。

PLAN-002 / SEC-001 **文档交付 done**：独立审阅完成六项差距盘点和两份主文档复核，两处措辞已修正；不签署真实安全PASS。格式与本地链接检查通过，250份非MD文件逐项匹配10-07已审快照。本轮源码包位置为 `../oj-delivery/OJ-MVP-2026-10-08-source.zip`，逐项核验结果以同目录 `OJ-MVP-2026-10-08-source.manifest.json` 的 verified 字段为准；旧10-07包保留。

用户再次要求GitHub推送，PUB-001恢复正常检查后仍blocked：普通Git查询失败，获准沙箱外查询后 connection reset；连接器读main成功，单文件create_tree明确403。main仍7a952c6；未创建commit/分支/PR，未写本地.git，停止写重试。[复验与恢复条件](docs/issues/2026-10-07-github-write-access.md)。源码及文档本地保留，Docker未重试；历史检查/旧包记录以下保留。

## 2026-10-07 恢复与所有权

本轮新增认领：**PLAN-002 / SEC-001**，按用户要求补齐后续工作包，并单列恶意代码安全风险与解决方案。root（Architect/Builder）仅修改规划、安全、任务、问题与交接 MD；security_gap_review 沿用已有独立审阅授权，只读执行/API/上传边界及写自己的审阅报告。依赖当前源码和已保存验证，不运行攻击载荷，不重试被阻塞的 Docker/GitHub 动作。实现任务与安全措施的实际落地状态另行标明，不把规划写成已完成保护。

后续方向评估已整理为 [开发进度与产品方向建议](docs/planning/2026-10-07-product-direction.md)：依据原 Phase 0–10 计划、当前代码/验证和官方项目文档及 2024/2026 研究，建议先完成当前真实验收与持久提交，再验证教师题目质检、公开通信回放和教师可控提示。此项仅文档提案，未改变功能范围或验收状态，未派发实施、重试 Docker 或上传；现有源码 ZIP 保持原 422 文件交付版本，不包含本次新增规划文档。

用户要求继续尝试完成，并总结审批失败可能原因和对策，撤销上轮停止要求。root 认领剩余测试夹具收尾修复、文档与证据、Docker 诊断和 GitHub 交付；沿用既有独立审阅授权，reviewer 仅只读当前改动与日志并写自己的报告。优先使用正常工作区工具；审批传输失败、真实权限拒绝和产品错误分别记录，不弱化权限/清理/判定断言。当前正常源码读取已成功，不能据此断言连接故障永久修复。下方 10-06 停止状态为历史。

Windows 测试夹具收尾已修复，14 项专项通过；正常取得本轮 network grant 后 `npm run check` exit 0，全 Go 测试、格式/lint、4 项 Node、SSR、Go build/tsc/Vite 通过。通信题包四对输入答案、32 字节、模板/descriptor/内置数据一致性检查通过。250 份非 MD 源码摘要 `5ffef55d8235aa1ef4f41e58be2e0a7adb4bef6505856a8cfda76cd2fe9a8a45`；旧摘要仅为历史。[完整验证](docs/judge/RECOVERY_VALIDATION_2026-10-07.md)、[问题与对策](docs/issues/2026-10-07-approval-diagnosis-and-recovery.md)、[接续职责](docs/handoffs/2026-10-07-recovery-delivery.md)。

Docker 正常查询仍 pipe permission denied，提权只读申请因审批传输中断未执行；停止该动作重试，未运行真实矩阵。COMM-001/002/003 本地实现与环境包已交付，真实产品/页面保持 in_review；COMM-004 的 recovery_final_review 已核对 250/250 文件和实际日志，[本轮代码与本地回归 PASS](docs/reviews/2026-10-07-recovery-review.md)，最终真实产品签署 pending。

PUB-001 blocked：GitHub 连接器大请求传输失败后只读恢复，首个小批写接口明确 403 `Resource not accessible by integration`；普通 Git 也 connection reset。未创建 commit、远程分支或 PR，未写本地 `.git`。已停止写入尝试，待用户恢复正常 GitHub 写权限/连接，详见 [上传问题与对策](docs/issues/2026-10-07-github-write-access.md)。本地源码包与文档完成后停止依赖环境的操作。

本地交付包为仓库旁 `../oj-delivery/OJ-MVP-2026-10-07-source.zip`，核验清单为同目录 `OJ-MVP-2026-10-07-source.manifest.json`。清单 `verified=true` 才表示逐项读回及已审源码绑定全部通过；包不含本机配置、教师上传存储、Git 历史或依赖产物。最后仅文档更新，不改变本轮 250 份源码摘要。

## 2026-10-06 当前停止位置

用户在自动审批再次中断后明确选择“先保存问题记录并停止”。本轮停止新增修复、测试、环境操作和上传；下方 10-05 及更早状态为历史。详见 [本轮问题、实际结果与接续方案](docs/issues/2026-10-06-recovery-check-and-approval-stop.md)。

root 接续 COMM-001/PUB-001，仅完成现场核对、正常 Docker 启动尝试、格式整理和完整回归。Windows 题库发布及清理修复已保存；problemstore 本轮通过，但完整 `npm run check` exit 1：`TestSpecialJudgeLifecycleAndPriority/checker_cleanup` 在 TempDir 收尾删除假产物时 Access is denied。格式、Go vet、前端 lint 通过；后续 Node 测试、SSR 和 build 未执行。剩余问题的源码读取被审批传输故障中止，不能登记已修复。

通信 A+B 的 [完整题面和示例包](examples/communication-a-plus-b/README.md)、[环境与真实验收指南](docs/judge/COMMUNICATION_ENVIRONMENT.md)、十二项真实验收脚本已保存；本轮真实矩阵未执行。Docker Linux 管道本次两次返回 not found，Desktop 隐藏启动命令 exit 0 后仍未观察到 Docker 进程；不能声称引擎可用或资源零残留。

独立 `recovery_final_review` 已保存 [本轮审阅报告](docs/reviews/2026-10-06-recovery-delivery-review.md)：marker/清理和通信脚本的已读静态范围无新增必修发现；本地完整回归 FAIL、真实 Docker pending，明确区分独立已读与 root 转述。停止后仅保存已有结论，未生成新完整源码摘要，不签最终 PASS。COMM-001..004 的真实产品验收与 PUB-001 上传均待接续；本轮没有新增 commit/push/分支/PR。

用户通知连接恢复并继续后，root 先核实并修复剩余测试收尾，再补当前完整回归、示例核验、文档检查和独立版本审阅；Docker 实测与 GitHub 交付按问题档案顺序继续。当前不再执行这些步骤。

## 2026-10-05当前恢复、修复与GitHub交付

用户最新授权“根据文档中的问题，尝试解决，并继续写入文档，最后上传GitHub”，中断后再次要求继续并强调通信题环境和题面。本轮root认领COMM-001/PUB-001的环境最小复现、Windows清理兼容、联调与MD/源码交付；中断后store_recovery负责受审题库发布兼容及测试，windows_review_resume独立审阅设计/实现与自己的报告；communication_env仅新增通信真实验收脚本、A+B例题包与环境指南。既有未提交内容、教师题库与卷保留。下节停止记录为本次恢复前状态。

已正常取得network grant并读取获准Docker日志；当前CLI仍Linuxpipe permission denied，日志近期有API活动，不断言引擎没启动。本地Git命令网络连接失败，.git写权限请求未授予，未写本地索引/引用。GitHub连接器已确认目标仓库push权限、远程main=本地HEAD7a952c6；按最新用户上传要求，拟交付开发分支并清楚记录真实验收待办，不代表生产发布或最终产品PASS。

可信Windows最小复现已区分普通path rename/单文件remove成功、Root.Rename invalidhandle、只读产物RemoveAll失败；Root.Remove空marker以及Root.Chmod清除本Job产物readonly后清理成功。具体受审修复和实际回归结果随后写入问题与验证记录。

最新中断补录：重新正常请求network权限得到grant后，一条工作区读取命令在自动审批阶段stream disconnected/response decode，工具明确未执行，不能登记新检查结果。没有让其他执行者重复该受拒命令或旁路权限。后续实际已执行的设计、实现、检查与上传需逐项记录。

## 2026-10-05通信题当前授权与职责

用户要求继续修复并增加通信题，选择发送端/接收端两份源码，并明确允许先完成不依赖Docker的本地开发和单元测试，保留真实验收待办。此答复调整旧阻塞停止约定，仅覆盖已知Docker访问/审批阻塞；不授权旁路审核、宿主执行选手代码或虚报沙箱验收。沿用已有并行开发与独立审阅授权，旧未提交内容/上传题/数据卷保留。

| 任务                                                  | 状态      | 执行者与允许修改范围                                  | 依赖/审阅                                                     |
| ----------------------------------------------------- | --------- | ----------------------------------------------------- | ------------------------------------------------------------- |
| [COMM-001 契约/API/集成](docs/tasks/COMM-001-api.md)  | blocked   | root：严格契约/绑定/API三投影/版本门禁/集成MD         | A+B前完整check历史PASS；当前五包及权限复验exit1，停止依赖验收 |
| [COMM-002 双角色后端](docs/tasks/COMM-002-backend.md) | in_review | comm_backend：filejudge其余实现、题库/上传/三题12用例 | 新A+B专项独立PASS；当前完整回归权限失败，真实Docker待办       |
| [COMM-003 前端](docs/tasks/COMM-003-frontend.md)      | in_review | comm_frontend：仅apps/web，双编辑器/上传/反馈/A+B示例 | 新增后type/lint/build/SSR/format实际PASS；真实页面待办        |
| [COMM-004 独立审阅](docs/tasks/COMM-004-review.md)    | in_review | comm_review：只读及自己的审阅MD                       | 局部设计/旧实现PASS；A+B窄审PASS，当前完整版本与产品未PASS    |

新增 `communication-a-plus-b`：sender仅传递原始a、b，receiver计算和；long long、±10¹²、32字节、四例覆盖普通/零/负数/大数。保留原sum/maximum，内置总数10。教师上传页提供“填入通信A+B说明与双模板”，学生仍只看整体判定和自身编译诊断。235份非MD源码摘要 `71e866f2e74ab2f8e9d107a8cacaf9240ed03ded1c87bf963821b8ba9a3cb634`；旧226份80a摘要及完整check仅为新增前历史证据。

root已取得正常network grant并改用项目TEMP/TMP，本地HTTP相关错误未再集中出现；Store重命名与fake产物unlink仍AccessDenied。再次申请该目录read/write并获grant后，专项复验仍exit1，已按用户约定停止依赖验收，详细[问题与恢复方案](docs/issues/2026-10-05-local-test-permissions.md)。Docker最新显式Linuxpipe查询exit1“permission denied”；权限工具不能表示该管道路径，没有Linux/cgroup2/镜像/零残留/服务ready证明。

已受审[通信契约](docs/judge/COMMUNICATION_CONTRACT.md)、[A+B使用指南](docs/judge/COMMUNICATION_USER_GUIDE.md)、[对应版本验证](docs/judge/COMMUNICATION_VALIDATION_2026-10-05.md)、[接续交接](docs/handoffs/2026-10-05-communication-local.md)；Docker与审批事实见[环境档案](docs/issues/2026-10-05-local-docker-access.md)。SPJ最终停止/重启/页面待办与PUB-001继续待真实环境恢复；本轮未commit/push。下方旧停止叙述仅为此前历史。

收尾文档检查：本轮修改MD的Prettier执行exit0，git diff --check exit0；160份Markdown、705个本地链接目标、缺失0，证据.local/evidence/communication-ab-20261005-doc-audit.json。独立reviewer逐条核对235文件与新摘要完全一致；收尾只改MD，非MD摘要保持不变。这些记录检查不替代失败的当前工程回归或真实验收。

## 2026-10-05此前停止与推送授权（历史）

用户要求继续完成、解决问题，最后推到 `https://github.com/BeijingXi-GS/OJ`；必要commit/push现已明确授权，优先于旧任务中的不推送范围，既有并行/独立审阅授权继续。root读取项目指令与现有实现，origin的fetch/push URL本地核实匹配，分支task/SPJ-001-student-construction-2026-10-03，HEAD7a952c6。未取得真实远程最新状态，不能把本地remote匹配当推送成功。

首次Docker诊断实际执行但Linux pipe不存在；发起现有Docker Desktop隐藏启动命令，没有ready证明。随后申请 `git fetch origin` 时自动审批通信再次中断、响应解码失败，工具明确动作未执行，不是安全否决或已确认GitHub拒绝。按用户约定停止新增实现/测试/commit/push，未派发新Agent或绕过审核。详见[本次问题](docs/issues/2026-10-05-github-approval-interruption.md)。

09:30只读后果记录：5173/8080/8081/8082均不可访问、jobs0；未观察到Docker Desktop/backend进程，原因未定位；今日没有成功Docker owner查询，不能登记容器0。200非MD文件摘要仍3406da9bef0efd1719a405bd48c2cc7206609368f0329f8012c5b6280bcf8df3，与旧停止快照相同，非通过冻结。没有新增commit/push、题目或数据卷删除。

| 任务                                                        | 当前状态  | 执行与接续职责                                                                 |
| ----------------------------------------------------------- | --------- | ------------------------------------------------------------------------------ |
| SPJ-001 最终环境/API/联调                                   | blocked   | root，恢复后先定位已有启动卡点与三阶段停止，再做重启/页面/当前版本完整检查     |
| SPJ-002 特判后端                                            | in_review | 原实现已交root，23项矩阵是历史对应版本PASS；6上传/34文件重启仍待实测           |
| SPJ-003 前端                                                | in_review | 原实现及下载/编辑窄修已交root，真实复验/截图尚未齐备                           |
| SPJ-004 独立审阅                                            | in_review | 旧独立局部PASS/最终产品pending保留；恢复后再登记reviewer，不将root自检代替独立 |
| [PUB-001](docs/tasks/PUB-001-github-delivery.md) GitHub交付 | blocked   | root必要Git集成/commit/push；依赖完整SPJ验收和正常审批/远程核实                |

用户需先恢复自动审批连接并通知继续；无需重复授权项目修复/并行/独立审阅/GitHub推送。具体[10-05接续与发布职责](docs/handoffs/2026-10-05-github-paused.md)，剩余产品验收步骤沿用[10-04详细交接](docs/handoffs/2026-10-04-special-judge-paused.md)。问题与方案持续收录在[索引](docs/issues/README.md)。以下全部10-04结果保留历史，不是本轮复跑、在线或最终交付证明。

## 2026-10-04结果与停止原因（历史）

用户要求继续完成、解决问题并持续整理。16:11审批曾实际恢复，Docker Linux/cgroup2启动并完成23项真实特判矩阵、Linux单测、完整check及两smoke；真实教师页面成功发布C++特判题并AC2/2。学生工程说明已精简并迁往工程师。当前13题：7内置（含构造不相邻排列）+6上传，原3道上传保留。

当前阻塞 **B-SPJ-001 再次发生审批通信中断**：准备重跑启动阶段诊断时，require_escalated自动审批stream disconnected/响应解码失败，工具明确动作未执行、不是安全判定。依用户要求停工，未绕过或再次重试；详见[独立问题](docs/issues/2026-10-04-approval-resume-interruption.md)。请用户恢复审批连接后通知继续，已有修复/并行/独立审阅授权有效，不需重新授权。

另一个已知待修点：[停止测试观测](docs/issues/2026-10-04-supervisor-observation.md)。最新真实脚本只bootstrap、未达到preflight，45秒阶段失败；exit0/stdio close0和实际清理0不能代替阶段通过，idle/running未执行。新增5条启动阶段日志尚未实际重跑；不放宽断言。下载/编辑兼容已窄修并代码复审，真实复验未完成；截图API失败另单开[档案](docs/issues/2026-10-04-browser-screenshot.md)。

| 任务                                      | 当前状态  | 实际职责与接收所有权                                                                                  | 审阅                                              |
| ----------------------------------------- | --------- | ----------------------------------------------------------------------------------------------------- | ------------------------------------------------- |
| [SPJ-001](docs/tasks/SPJ-001-api.md)      | blocked   | root Architect/Builder：契约、API/client、worker、环境/联调及当前MD；最新mvp阶段日志未完整check       | spj_review历史局部PASS；spj_acceptance最终pending |
| [SPJ-002](docs/tasks/SPJ-002-backend.md)  | in_review | special_backend实现已交root；store/checker/snapshot/fixture完成，23项矩阵PASS，重启仍待验收           | spj_acceptance                                    |
| [SPJ-003](docs/tasks/SPJ-003-frontend.md) | in_review | student_frontend实现交root；lifecycle_fix窄修UploadPage下载/EditorPage输入；教师上传AC2/2，后续QA未完 | spj_acceptance                                    |
| [SPJ-004](docs/tasks/SPJ-004-review.md)   | in_review | spj_acceptance独立Architect/Security仅只读与review MD；未签最终产品PASS                               | root核实真实证据                                  |

docs_curator只维护[问题索引](docs/issues/README.md)与[详细特判使用指南](docs/judge/SPECIAL_JUDGE_USER_GUIDE.md)，lifecycle_fix仅窄修停止脚本和已交付前端控件；所有Agent已停止新增工作。归档区与旧进度保持历史，不作为当前状态。

[完整实际验证记录](docs/judge/SPECIAL_JUDGE_VALIDATION_2026-10-04.md)、[页面部分验收](docs/judge/BROWSER_SPECIAL_QA_2026-10-04.md)、[独立当前报告](docs/reviews/2026-10-04-special-judge-acceptance.md)、[下一次精确接续与Agent职责](docs/handoffs/2026-10-04-special-judge-paused.md)。可选checker反馈验证脚本问题已重跑关闭，旧失败报告保留；三阶段/输出限精确归因/上传重启/页面复验/最终完整check及独立产品签署仍未完成，不登记done。

停止快照200非MD文件SHA256 **3406da9bef0efd1719a405bd48c2cc7206609368f0329f8012c5b6280bcf8df3**，分支task/SPJ-001-student-construction-2026-10-03；不是通过冻结。镜像仍sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74。完整check已通过对应版本，后续本地导入自测扩展和启动阶段日志尚未重跑全部检查；无新依赖/commit/push/删卷，所有旧改动/数据保留。

16:56只读记录：Web5173/API8080/Judge8082不可访问，Compiler8081 HTTP200，jobs0；最近成功Docker owner查询16:48为空，审批失败后未新查。6上传/34文件重启前SHA与公开元数据保存在.local/evidence/spj-20261004-persistence-before.json，尚无after对比。本机手动模式不是真实登录，完整Phase1/生产安全/持久提交仍未完成。

## 三模式已验收基线（历史）

用户明确要学生隐藏具体用例错误、教师查看/上传题目数据、工程师调试及交互题，并确认仅本机手动切换。四项任务全部交付，roles_review独立[最终签署](docs/reviews/2026-10-02-modes-interactive-review.md)局部设计、实现和本机产品PASS，未解决必修0。最终完整check、Linux单测（含真实symlink）、17项旧矩阵、五题10次、13项交互判定、双Running取消/交互器故障2项、教师三投影6次、三题重启后可判、8组最大JSON转义实测和[浏览器验收](docs/judge/BROWSER_QA_2026-10-03.md)均通过。详细[验证](docs/judge/MODES_VALIDATION_2026-10-03.md)、[使用与Agent职责](docs/judge/MODES_USER_GUIDE.md)、[本轮交接](docs/handoffs/2026-10-03-modes-interactive.md)。

| 任务                                                 | 状态 | 执行者/范围                                                                       |
| ---------------------------------------------------- | ---- | --------------------------------------------------------------------------------- |
| [MODES-001](docs/tasks/MODES-001-api.md)             | done | root：API模式/投影、上传路由、client DTO/校验、集成/MD                            |
| [MODES-002](docs/tasks/MODES-002-catalog.md)         | done | teacher_catalog：Store/上传解析/loader/题面与fixture；单测/vet通过                |
| [MODES-003](docs/tasks/MODES-003-interactive.md)     | done | interactive_runner：监督器/固定交互器已实现，13项判定及双Running取消/系统故障通过 |
| [MODES-004](docs/tasks/MODES-004-frontend-review.md) | done | modes_frontend仅apps/web已交付；root实际页面通过，roles_review最终独立PASS        |

前一五题MVP验收是历史已通过结果，不能替代本轮新增执行/上传/模式验收。分支task/MODES-001-interactive-2026-10-02，无新未知阻塞；所有旧改动保留。

本轮最终绑定：非MD源码176文件SHA256 `8785b5dd63074e25633de5f6d78e52ddbaf4b94c6b6a4df08e3349a4fbb60b1c`；镜像仍 `sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74`。内置6题/27例，3道教师上传题持久保留，当前9题。无新增依赖/提交/推送，保留数据卷和原工作区。

最终收尾再次中断后，Docker和服务未运行；按已授权范围恢复现有Docker Desktop，确认Linux/cgroup2且成功查询0本项目残留，再重新严格启动。17:41:23 judge ready、17:41:25 API开启compiler/judge，Web5173/其余8080/8081/8082均200，源码绑定与9题不变，见[恢复档案](docs/issues/2026-10-03-final-interruption-services.md)。原已审阅10:13:32证据保留，本次另写modes-resume-readiness.json与实际AC复验，不能把历史在线状态当中断后事实。

开放但非阻塞：Windows共享终端Ctrl+C时supervisor未获shutdown响应，已核实无残留并重启通过；提示本身未修改，单独[交用户处理](docs/issues/2026-10-03-supervisor-stop-warning.md)。手动模式不是身份认证，任意交互器/提交历史/完整Phase1与生产安全未实施。

收尾检查：源码冻结后仅同步MD，最终format:check exit0；125份Markdown、456个本地链接目标和git diff --check全部通过，modes-doc-check.json completed:true。新实例上传交互题AC2/2、同数据摘要、0残留；任务卡/交接/入口/问题档案同步完成。

核心契约：七项 [foundation-invariants/1.0.0](docs/architecture/FOUNDATION_INVARIANTS.md) 已 frozen；完整核心类型 / 接口仍 `0.1.0-draft / proposed`。原计划只是依据；本次局部文件判题已实现，完整判题核心与业务持久提交仍未完成。

此前 P0-005 修复和复验已完成，独立 `/root/phase0_recheck` 的 [正式复审](docs/reviews/2026-10-01-phase0-recheck.md) 通过；[Phase 0 验收](docs/reviews/2026-10-01-phase0-acceptance.md) 已登记。历史源码摘要56文件 / `469c5c38da2a199f64c6e1e04d16bce4a3e7493145af1e72f97e43154a86bc00`仅适用于MVP实施之前。origin已配置，远程CI未执行。

此前P1设计交付时71份MD链接检查通过；该结果不代替本次MVP记录。P1仅交付设计，未运行真实Judge测试，见 [历史交接](docs/handoffs/2026-10-01-phase1-design.md)。

## 本轮恢复与多题任务（2026-10-02）

用户明确授权修复并继续、增加题目。初始Docker Linux管道不存在；启动已有Docker Desktop后恢复linux amd64 /cgroup2，完整预检和实际页面已通过。当前事实不还原上轮11:54查询失败根因。新增数组和、最大值、回文、排序，连同A+B共5题/24组文件用例，五题正确AC/错误WA的10次实测与17项判定/故障矩阵全部通过。

| ID                                                   | 状态 | 实际执行者 /职责                                                                      | 允许修改范围                                                                                       |
| ---------------------------------------------------- | ---- | ------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| [JUDGE-005](docs/tasks/JUDGE-005-recovery.md)        | done | /root Backend Builder                                                                 | Docker诊断、filejudge/docker.go /engine.go及专项tests、启动 /联调scripts、根MD                     |
| [JUDGE-006](docs/tasks/JUDGE-006-problems.md)        | done | /root/problem_catalog Builder                                                         | internal/problems、httpapi、filejudge/contract.go /cases.go及相关tests、四题fixture、problem-smoke |
| [JUDGE-007](docs/tasks/JUDGE-007-frontend-review.md) | done | /root/frontend_followup Frontend Builder；/root/judge_reviewer 独立Architect/Security | frontend仅apps/web；reviewer仅docs/reviews和只读fake检查                                           |

目录扩展采用 [裁定](docs/judge/PROBLEM_CATALOG_EXTENSION.md)，先独立局部设计通过再实施核心接线。原实现 /检查 /停止记录保留下文，实时状态由本节更新；新未知阻塞仍归用户并停止。

本轮完整工程check /基础smoke /旧编译smoke /Linux单测通过；浏览器会话恢复并完成 [五题页面QA](docs/judge/BROWSER_QA_2026-10-02.md)，已关闭 [B-JUDGE-002](docs/issues/2026-10-02-judge-final-preflight.md) /[B-MVP-001](docs/issues/2026-10-01-mvp-browser-session.md)。最后启动IO脱敏修正后19:15:06 judge再次严格ready、19:15:12 API开启judge与compiler，19:17:43最终独立数组和AC5/5、四服务200、0容器/0job残留。最终完整check与含新诊断测试的Linux单测均exit0，109份MD本地链接/差异空白检查通过。独立 [review](docs/reviews/2026-10-02-judge-recovery-problems-review.md) 设计/实现与本机产品PASS，绑定最终源码/镜像；无遗留必修问题。

分支 task/JUDGE-005-recovery-problems-2026-10-02，非MD源码151文件 SHA256 a07de43e447cea2b5af370225eafa13018a20cfee4d6c0bf011461f88026761f；镜像仍sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74。无新增依赖/提交/推送，保留所有历史改动和基础数据卷。统一 [使用及Agent职责](docs/judge/README.md)、[验证](docs/judge/VALIDATION.md)、[本轮交接](docs/handoffs/2026-10-02-judge-recovery-problems.md)；本产品仍为本机同步MVP，完整Phase1/生产安全/持久化未完成。

## 上轮文件用例判题任务

Anolix源码已本地拉取并锁commit，Docker Linux此前恢复；局部judge-mvp/0.1.0设计独立通过，前后端实际实现完成，独立代码 /执行边界审阅通过。**最终重启11:54:40报cgroup_v2_required，按用户规则停止，B-JUDGE-002交用户处理；不是已确认cgroup v1。** API /Web最终未成功上线，浏览器验收未完成，不标产品done。

| ID                                                      | 工作                                               | 状态    | 实际执行者 /职责                                         | 交付 /停止位置                                                                                       |
| ------------------------------------------------------- | -------------------------------------------------- | ------- | -------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| [JUDGE-001](docs/tasks/JUDGE-001-file-cases-sandbox.md) | 参考源码、能力预检、增量设计                       | done    | /root Architect；/root/mvp_review 独立Architect/Security | 已锁Anolix4502d2a...；[设计通过](docs/reviews/2026-10-02-judge-design-review.md)                     |
| [JUDGE-002](docs/tasks/JUDGE-002-backend.md)            | 快照、checker、受限编译 /运行、API、脚本、真实矩阵 | blocked | /root Backend Builder                                    | 实现和17项矩阵 /check通过；最终新进程预检失败，B-JUDGE-002                                           |
| [JUDGE-003](docs/tasks/JUDGE-003-frontend.md)           | 双按钮 /独立结果 /逐例 /互斥 /过期 /取消           | blocked | /root/mvp_frontend Frontend Builder                      | apps/web七文件、静态check与源码review通过；等待真实浏览器验收                                        |
| [JUDGE-004](docs/tasks/JUDGE-004-review.md)             | 独立实现 /安全审阅与最终产品验收                   | blocked | /root/mvp_review Architect/Security                      | [实现代码审阅通过](docs/reviews/2026-10-02-judge-implementation-review.md)；最终启动 /浏览器证据未齐 |

当前非MD源码100文件 SHA256 `3ee10a98dbff54fa0e567a6a18161a3b38aa67bab8e6190f9643fddb42bb7357`，无依赖更新 /提交 /推送。严格沙箱只用于新judgements链路，旧compilations仍为原仅编译worker；完整Phase1 /生产隔离 /历史持久化未完成。

[使用与Agent职责](docs/judge/README.md) / [真实验证与未测范围](docs/judge/VALIDATION.md) / [当前交接](docs/handoffs/2026-10-02-file-judge-mvp.md)。17项修正版矩阵含全部verdict、真OOM及非OOM137、展示截断 /UTF8、后代清理 /取消 /busy、文件故障及动态metadata，全部通过；完整check /smoke /mvp:smoke /Linux单测与旧声明卷镜像拒绝通过。最后一次严格重启未ready，最终AC smoke ECONNREFUSED未通过，浏览器仅inventory无操作，不用此前证据冒称最终验收。

[新阻塞与用户恢复条件](docs/issues/2026-10-02-judge-final-preflight.md) / [源码获取已恢复](docs/issues/2026-10-01-anolix-source-access.md) / [开发期空占位检查修正](docs/issues/2026-10-02-judge-preflight-placeholder.md) / [Windows symlink跳过及Linux补验](docs/issues/2026-10-02-windows-symlink-unit.md) / [go doc非阻塞网络警告](docs/issues/2026-10-02-go-doc-metadata.md)。旧[B-MVP-001浏览器会话](docs/issues/2026-10-01-mvp-browser-session.md)未经真实复验保持未关闭；本轮未绕过。新的问题一事一档、阻塞交用户并停止。

## 上轮编译 MVP 结果与停止位置

实际前端 /API /独立G++14.2.0编译worker已实现，npm run check /smoke /compiler:test /mvp:smoke通过。四项真实源码验证通过，程序从不运行；无新增Go /npm依赖。上轮74文件源码摘要 b9d9ee3ed07ca11d96b99efb9eae472c189a1d8757dd970a7da39b334d508364。

**B-MVP-001：浏览器创建标签页被当前会话拒绝，inventory无标签页，浏览器QA未完成。按用户要求已停止后续工作；独立reviewer仅完成文档阅读，代码审阅未完成，不能标done。** [问题前因后果与用户处理](docs/issues/2026-10-01-mvp-browser-session.md) / [使用与Agent职责](docs/mvp/README.md) / [验证](docs/mvp/VALIDATION.md) / [交接](docs/handoffs/2026-10-01-compile-mvp.md)。当时本机5173 /8080 /8081保持运行；本轮最终关闭 /重启情况以当前文件判题任务为准，不能假设这些历史实例仍在线。

## Phase0交付与验证（历史，非本轮MVP状态）

| 类别                             | 当前状态                                     | 证据 / 说明                                                                                                                             |
| -------------------------------- | -------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| 正式开发流程、职责、模板、路线图 | 基础范围独立审阅和阶段验收通过               | 入口说明已修正；实际状态与职责分离                                                                                                      |
| Git / Monorepo / 依赖锁定        | 本轮修复及文档未提交 / 未推送                | 基线 7a952c6；分支 task/P1-001-mini-judge-design-2026-10-01；保留全部原改动；Go / npm 依赖锁文件未修改                                  |
| API / Worker / 核心类型          | 骨架可编译，基础测试通过                     | Worker 仅 idle；核心类型为草案                                                                                                          |
| Web / 编辑器                     | 本轮构建通过，源码未修改                     | [实施指南](docs/frontend/IMPLEMENTATION_GUIDE.md) / [任务包](docs/frontend/AGENT_TASKS.md) 已交付待实施；现有 Monaco 体积警告已单独记录 |
| pgx / sqlc                       | 依赖已安装；代码生成通过                     | 仅基础 SELECT 1，不存在业务表                                                                                                           |
| 格式 / lint / 单元测试 / 构建    | **2026-10-01 npm run check 通过**            | 格式、Go vet、ESLint、Go tests、Go / TypeScript / Vite build；另检健康探针 Go vet；未运行 race / 远程 CI                                |
| npm 依赖检查                     | 9 月 26 日报告 0 项                          | 本轮没有更新依赖或重新执行 audit                                                                                                        |
| Compose 配置                     | 2026-10-01 配置 / 完整 up 通过               | PG 宿主机 15432；MinIO 使用固定官方源码本地镜像；所有端口仍仅绑定本机                                                                   |
| Docker / PostgreSQL / Redis      | **2026-10-01 已恢复并验证**                  | Docker Server 29.8.1 / linux；PG / Redis healthy，原命名卷保留，真实数据库 API /readyz=200                                              |
| 完整基础服务 smoke / MinIO       | **2026-10-01 通过，B-ENV-002 关闭**          | 三项服务 healthy；MinIO UID 10001，首次创建项目命名卷，源码 / checksum 固定；[修复记录](docs/issues/2026-10-01-minio-source-rebuild.md) |
| 评审                             | 独立工程复审通过；基础原则窄范围采用         | [原报告](docs/reviews/2026-10-01-phase0-review.md) / [复审](docs/reviews/2026-10-01-phase0-recheck.md)；没有整体冻结核心                |
| 本机 Go race                     | **未能执行**                                 | 当前 CGO 不可用；CI 中已配置 Ubuntu race                                                                                                |
| 远程 CI / 独立评审               | Phase0独立评审完成；P1设计未审；远程CI未执行 | 不把Phase0 review自动扩展为Phase1设计 /执行安全审阅                                                                                     |

## 上轮编译 MVP 任务（历史状态）

用户已授权实际前后端并行开发及独立审阅，范围见 [接口契约](docs/mvp/CONTRACT.md) 与 [ADR-0004](docs/adr/0004-compile-only-mvp.md)。当前分支 task/MVP-001-compile-web-2026-10-01。旧 Phase1/P2 任务保留，但不代表完整判题已实施。

| ID                                        | 工作                              | 状态      | 实际执行者 / 角色                     | 修改范围                                                                      | Review owner  |
| ----------------------------------------- | --------------------------------- | --------- | ------------------------------------- | ----------------------------------------------------------------------------- | ------------- |
| [MVP-001](docs/tasks/MVP-001-backend.md)  | API、独立编译服务、环境及联调     | blocked   | /root / Builder                       | apps/api、internal、services/compile-worker、deploy/compiler、scripts、根文档 | 独立 reviewer |
| [MVP-002](docs/tasks/MVP-002-frontend.md) | 题目页面、Monaco 编辑器、编译反馈 | in_review | /root/mvp_frontend / Frontend Builder | apps/web                                                                      | 独立 reviewer |
| [MVP-003](docs/tasks/MVP-003-review.md)   | 独立实现审阅                      | blocked   | /root/mvp_review / 独立 Builder       | docs/reviews                                                                  | /root 验收    |

## 原路线任务（历史状态保留）

状态词的精确定义见 DEVELOPMENT.md。角色与实际执行者分开记录；以下“本次初始化 Agent”不冒充独立 reviewer。

| ID                                                                    | 工作                               | 状态      | Owner 角色 / 实际执行者                                | Review owner                                       | 依赖 / 下一步                                                    |
| --------------------------------------------------------------------- | ---------------------------------- | --------- | ------------------------------------------------------ | -------------------------------------------------- | ---------------------------------------------------------------- |
| [P0-001](docs/tasks/P0-001-foundation.md)                             | 工程骨架与依赖                     | done      | Builder + Worker / 本次初始化 Agent；本轮根 Agent 修复 | Builder reviewers phase0_review / phase0_recheck   | 原骨架、复现修复与阶段验收通过                                   |
| [P0-002](docs/tasks/P0-002-workflow.md)                               | 流程、职责及统一 MD                | done      | Architect + Worker / 本次初始化 Agent；根 Agent 同步   | Builder reviewers / 授权集成验收                   | 入口说明已修复；流程与阶段验收通过                               |
| [P0-003](docs/tasks/P0-003-contract-draft.md)                         | 核心契约草案                       | done      | Architect + Builder / 本次初始化 Agent                 | Architect reviewers phase0_review / phase0_recheck | 草案交付与原则审阅完成；完整类型仍 proposed                      |
| [P0-004](docs/tasks/P0-004-contract-review.md)                        | 独立 Review / 窄范围冻结           | done      | 独立 reviewers；根 Agent 登记采用                      | phase0_review / phase0_recheck                     | 仅七项基础原则 frozen；阶段验收完成                              |
| [P0-005](docs/tasks/P0-005-infra-validation.md)                       | 基础服务真实联通 / 复现            | done      | Builder / 根 Agent，再次授权后修复                     | 独立 /root/phase0_recheck                          | 连续两次构建、真实服务 / API、工程检查与复审通过                 |
| [P1-001](docs/tasks/P1-001-mini-judge-design.md)                      | Mini Judge生命周期设计             | in_review | Architect / 根Agent，2026-10-01                        | 独立Architect，P1-002                              | DESIGN / ADR / fixture规范 /任务包已交付；只写MD，尚未冻结或运行 |
| [P1-002](docs/tasks/P1-002-design-review.md)                          | 独立设计评审 /局部冻结             | ready     | 独立Architect /未认领                                  | 授权集成执行者                                     | P1-001设计已交付；不得以Phase0评审替代                           |
| [P1-010](docs/judge/PHASE1_AGENT_TASKS.md#p1-010)                     | Linux /compiler profile与preflight | planned   | Builder-Environment /未认领                            | Architect                                          | P1-002；真实权限、版本、能力报告通过后实现                       |
| [P1-011](docs/judge/PHASE1_AGENT_TASKS.md#p1-011)                     | 不可变快照与C++ Compiler           | planned   | Builder-Compiler /未认领                               | Builder /Architect                                 | P1-002/010；固定argv、有限诊断、显式产物                         |
| [P1-012](docs/judge/PHASE1_AGENT_TASKS.md#p1-012)                     | Linux监督器 /完整清理 /runner      | planned   | Builder-Runner /未认领                                 | Architect                                          | P1-002/010；cgroup、唯一reaper、超时竞争与raw证据                |
| [P1-013](docs/judge/PHASE1_AGENT_TASKS.md#p1-013)                     | Exact Checker与结果映射            | planned   | Builder-Pipeline /未认领                               | Builder /Architect                                 | P1-002；字节比较、可信证据与错误优先级                           |
| [P1-014](docs/judge/PHASE1_AGENT_TASKS.md#p1-014)                     | 可信fixtures与manifest             | planned   | Worker-Fixtures /未认领                                | Builder-Tests                                      | P1-002/010；五类样例、hash与生命周期输入                         |
| [P1-015](docs/judge/PHASE1_AGENT_TASKS.md#p1-015)                     | Pipeline /CLI /report集成          | planned   | Builder-Pipeline /未认领                               | Builder /Architect                                 | P1-011–014；串行case、单终态与有限报告                           |
| [P1-016](docs/judge/PHASE1_AGENT_TASKS.md#p1-016)                     | 真实Linux矩阵 /race /故障证据      | planned   | Builder-Tests + Worker-Tests /未认领                   | Architect                                          | P1-015；逐例raw证据和完整清理断言                                |
| [P1-017](docs/judge/PHASE1_AGENT_TASKS.md#p1-017)                     | Phase1独立复审 /验收               | planned   | 独立Architect + Builder /未认领                        | 授权集成执行者                                     | P1-016与适用检查；不冒称Phase4安全能力                           |
| [P2-000](docs/tasks/P2-000-frontend-plan.md)                          | 前端框架实施 MD 和任务包           | in_review | Architect（方案整理）/ 本轮 Agent                      | Frontend Builder / Architect，未指派               | 文档已交付；未修改前端，待评审                                   |
| [P2-100](docs/frontend/AGENT_TASKS.md#p2-100前端方案与接口边界评审)   | 前端方案与 API 边界评审            | ready     | Architect + Frontend Builder / 未认领                  | Architect / 项目负责人                             | 先批准 FE-A 范围，FE-B 业务契约另外冻结                          |
| [P2-101](docs/frontend/AGENT_TASKS.md#p2-101框架分层应用入口与路由)   | 目录、provider、router、layout     | planned   | Frontend Builder / 未认领                              | Builder                                            | P2-100 FE-A                                                      |
| [P2-102](docs/frontend/AGENT_TASKS.md#p2-102统一请求层类型生成与配置) | HTTP client、错误、类型与配置      | planned   | Frontend Builder / 未认领                              | Builder / Architect                                | P2-101                                                           |
| [P2-103](docs/frontend/AGENT_TASKS.md#p2-103通用-ui-与页面状态组件)   | 通用组件、状态和样式               | planned   | Worker-UI / 未认领                                     | Builder                                            | P2-101                                                           |
| [P2-104](docs/frontend/AGENT_TASKS.md#p2-104mock单元组件测试底座)     | mock 与组件测试底座                | planned   | Worker-Tests / 未认领                                  | Builder                                            | P2-102                                                           |
| [P2-105](docs/frontend/AGENT_TASKS.md#p2-105查询模型与轮询策略)       | 业务 hooks、view models、轮询      | planned   | Frontend Builder / 未认领                              | Builder                                            | P2-102/104；DTO 提案审阅                                         |
| [P2-106](docs/frontend/AGENT_TASKS.md#p2-106编辑器组件与源码输入)     | Monaco 与草稿组件                  | planned   | Worker-Editor / 未认领                                 | Builder                                            | P2-101/103/104                                                   |
| [P2-107](docs/frontend/AGENT_TASKS.md#p2-107题目列表与详情页面)       | 题目页面                           | planned   | Worker-UI / 未认领                                     | Builder                                            | P2-103/105/106                                                   |
| [P2-108](docs/frontend/AGENT_TASKS.md#p2-108提交列表详情和历史-run)   | 提交与历史 Run 页面                | planned   | Worker-UI / 未认领                                     | Builder                                            | P2-103/105                                                       |
| [P2-109](docs/frontend/AGENT_TASKS.md#p2-109真实提交闭环与后端联调)   | FE-B 真实 API 闭环                 | planned   | Frontend + Backend Builder / 未认领                    | Architect + Builder                                | 业务 API/身份/幂等冻结、页面就绪、真实后端可用                   |
| [P2-110](docs/frontend/AGENT_TASKS.md#p2-110e2e响应式与-ci)           | 浏览器回归与 CI                    | planned   | Worker-Tests / 未认领                                  | Builder                                            | P2-104/页面；真实项目另依赖 P2-109                               |
| [P2-111](docs/frontend/AGENT_TASKS.md#p2-111集成验收与交接)           | FE-A / FE-B 集成验收               | planned   | Frontend Builder / 未认领                              | 项目负责人                                         | 上游适用任务与审阅完成                                           |

## 阶段门禁

| Phase           | 状态        | 进入下一阶段前的条件                                                  |
| --------------- | ----------- | --------------------------------------------------------------------- |
| 0 Foundation    | done        | 独立复审、窄范围采用与本地工程阶段验收完成；完整核心仍 draft          |
| 1 Mini Judge    | in_progress | 设计包待独立审阅；P1-010须锁定真实Linux /compiler与能力，再实现和验证 |
| 2 Web OJ        | planned     | Phase 1 验收；DB/授权/API 评审                                        |
| 3 Async OJ      | planned     | Phase 2 验收；Queue/recovery 契约冻结                                 |
| 4 Secure OJ     | planned     | Threat Model 与 Linux 隔离设计、安全评审                              |
| 5 Standard OJ   | planned     | Phase 4 安全门禁                                                      |
| 6 Contest       | planned     | 普通题稳定、量化负载目标                                              |
| 7 Interactive   | planned     | IPC/生命周期 Spec                                                     |
| 8 Communication | planned     | 多源码/图/聚合资源 Spec                                               |
| 9 Cluster       | planned     | lease/fencing/容量调度 Spec                                           |
| 10 Production   | planned     | 安全、负载、恢复与运维门禁                                            |

## 阻塞与解除条件

**当前MVP无未解决阻塞。B-JUDGE-001已关闭**：已本地clone Anolix并锁commit/阅读源码，局部设计独立通过；[完整记录](docs/issues/2026-10-01-anolix-source-access.md)。

**B-MVP-001已关闭**：当前IAB成功创建/操作本机页面，实际QA与独立证据审阅完成；[完整档案](docs/issues/2026-10-01-mvp-browser-session.md)。**B-JUDGE-002已关闭**：Docker恢复、严格最终preflight/AC/0残留复验，原失败确切原因不追认；[完整档案](docs/issues/2026-10-02-judge-final-preflight.md)。

| ID           | 阻塞                                 | 当前处理                                                                            | 解除方式                                                                                                           |
| ------------ | ------------------------------------ | ----------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| B-ENV-001    | Docker Desktop Linux engine 连接问题 | **2026-10-01 启动 / 复验关闭**：启动前 Desktop 未运行，启动后 Server 29.8.1 / linux | 已完成有界启动；成功证据见 [交接](docs/handoffs/2026-10-01-environment-review-handoff.md)；不更改主机自启设置      |
| B-ENV-002    | MinIO 不可用，完整环境验收失败       | **2026-10-01 修复关闭**：固定源码本地镜像；完整 up / infra-smoke 通过               | [修复档案](docs/issues/2026-10-01-minio-source-rebuild.md)；原 registry 访问失败未宣称修复，生产存储需另行裁定     |
| B-ENV-003    | PostgreSQL 宿主机 5432 不可绑定      | **2026-09-30 修复关闭**：Windows 排除范围；改为 15432，宿主机 API readiness 200     | [原因、修复及关闭证据](docs/issues/2026-09-30-postgres-port.md)；已有副本同步 .env 两个端口值                      |
| B-TEST-001   | 历史Windows race需CGO工具链          | 未执行；不影响已界定的Phase0基础工程验收                                            | P1实现时须在锁定Linux环境实际执行适用race；CI配置不是运行证据                                                      |
| B-REVIEW-001 | 基础原则 / 工程阶段评审              | **独立审阅及阶段验收关闭**                                                          | 七项原则 frozen，完整核心仍 draft；[阶段验收](docs/reviews/2026-10-01-phase0-acceptance.md)                        |
| W-FE-001     | 现有 Monaco chunk 较大，非阻塞警告   | 最终Vite构建成功，EditorPage 2.76MB，gzip716.87kB                                   | 后续安排性能任务；本轮未调整分包或警告阈值；[详细记录](docs/issues/2026-10-01-frontend-bundle-size.md)             |
| B-BUILD-001  | MinIO 重复构建 LICENSE 只读          | **修复复验及独立复审关闭**                                                          | 输出可写、源只读、SHA一致，连续两次构建与服务 / API 通过；[问题档案](docs/issues/2026-10-01-minio-license-copy.md) |

## 首轮修复与评审结果（2026-10-01，历史）

- 用户重新授权尝试修复已知问题并推进工程；Docker 成功有界启动，已用固定官方源码构建并启动 MinIO，完整基础服务验收通过。
- 新增来源锁定、构建脚本、scratch 镜像、静态健康探针与 CI 构建步骤；源 / 二进制校验可追溯，Go / npm 依赖锁文件未改变，现有数据卷保留。
- 当前工作区 `npm run check` 通过；核心契约 / API / 前端源码未变，Judge 仍未实现；Monaco 非阻塞警告单独归档。
- 环境首建通过后曾进入 in_review，最终重复构建失败，P0-005 回到 blocked，已按规则单开 MD 并停止修复。用户随后明确授权独立 Agent，`/root/phase0_review` 已交付 [正式报告](docs/reviews/2026-10-01-phase0-review.md)：核心基础不变量通过，工程需修改；未自行冻结完整契约或进入 Phase 1。
- Reviewer 发现的 [README 远程仓库说明](docs/issues/2026-10-01-readme-remote-status.md) 和 [setup 启动提示](docs/issues/2026-10-01-setup-minio-instruction.md) 已分别归档，交用户处理。它们不单独阻塞现有服务，不能抵消构建阻塞。
- 详细证据、冷启动路径与当前停止位置见 [本轮交接](docs/handoffs/2026-10-01-environment-review-handoff.md)。

## 上轮授权范围与结果（2026-09-30，历史）

- 用户明确授权修复 `5432` 端口冲突并再次验证 Docker，随后继续操作；此前前端只编写 MD 的范围不变。
- 实际原因：`5432` 位于 Windows TCP 排除范围 `5384–5483`，绑定失败 10013，未发现普通监听进程；`15432` 绑定测试通过。
- 已修复 `.env.example` / Compose 默认端口 / 本机 `.env` 端口和连接串；PostgreSQL 按配置变化重建，原命名卷保留。Docker 连接、doctor、config、PG / Redis 与真实 `.env` 的 API readiness 均通过。
- 完整 infra-smoke 失败于 MinIO，已另建详细问题 MD。按用户规则停止后续开发，等待用户处理并通知继续；没有变更 MinIO 镜像、执行前端 / Mini Judge 实现或启动其他 Agent。
- 修复自检不替代独立审阅，未执行全套 check / race / 远程 CI。临时 API 已停止，PG / Redis 保持运行。
- 配置摘要和实际命令见 [本轮交接](docs/handoffs/2026-09-30-port-docker-validation.md)。此前未提交内容保留，本轮未提交 / 推送。

## 上轮授权范围与结果（2026-09-29，历史）

- 用户明确要求：① 继续连接 Docker Linux engine，失败则分析原因；② 只在 MD 编写详细前端框架实施说明，不实际搭建。
- Docker：重试失败；完成进程、context、服务、WSL 版本、设置与有限历史日志的只读诊断。用户处理根因，真实环境验证仍停止。
- 前端文档：按本次独立授权已交付 [实施指南](docs/frontend/IMPLEMENTATION_GUIDE.md)、[任务包](docs/frontend/AGENT_TASKS.md) 和 P2-000 任务卡，不因 Docker 阻塞停写。
- 本轮没有修改前端/后端源码、依赖、.env、Compose；没有自动启动其他 Agent；没有再次发起主机修复。
- P0-004 独立架构评审 / P1-001 Mini Judge 设计仍未完成；本次文档不替代它们，也不宣称 Phase 2 已实现。

## 原 Phase1 后续工作（MVP 后另行推进）

1. [P1-002](docs/tasks/P1-002-design-review.md) 独立Architect审阅 [DESIGN](docs/judge/PHASE1_DESIGN.md)、[ADR-0003](docs/adr/0003-phase1-local-runner-lifecycle.md)、fixture规范与任务包；本次没有独立P1设计结论，不能直接冻结 /实现。
2. 审阅通过后P1-010锁定实际Linux /GCC /runtime profile与权限、计量 /完整清理能力；当前只有Docker Linux基础信息，不等于runner前置通过。
3. 按 [Phase1 Agent任务包](docs/judge/PHASE1_AGENT_TASKS.md) 认领P1-011–017，保留原数据卷和用户.env；真实Judge矩阵 /race /远程CI尚未执行。
4. 旧前端规划保留；最新用户已授权编译 MVP 实际前端，完整 Web OJ /真实判题仍须按原门禁另行实施。
5. 最新设计交付与已测 /未测边界见 [P1-001交接](docs/handoffs/2026-10-01-phase1-design.md)。

## 变更日志

| 日期       | 记录                                                                                                                                                                                                      |
| ---------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2026-10-01 | 用户要求文件用例运行判题并确认服务器.in/.out，参考Anolix；Docker资源 /安全限制预检通过，参考核心源码网页cache miss及git连接重置，B-JUDGE-001阻塞，按规则一事一档并停止，未改应用代码或启用运行。          |
| 2026-10-01 | 用户授权编译MVP实际前后端并行与独立review；实现、真实GCC四项smoke、Linux生命周期专项、check及基础smoke通过。新浏览器会话阻塞，停止验收 /独立审阅，完整问题交用户处理。                                    |
| 2026-10-01 | 再次授权后修复LICENSE幂等性与入口说明，连续两次构建 /服务 /API /check /smoke及独立复审通过；Phase0验收、七原则窄范围冻结。P1-001设计 /ADR /fixtures规范 /任务包交付，待独立设计审阅，未实现runner或前端。 |
| 2026-10-01 | Docker / MinIO 首建、三项服务、真实 API readiness、check / smoke 通过；最终重复构建 LICENSE EPERM，P0-005 / Phase 0 blocked。独立评审已交付：基础原则通过、工程需修改；问题分别归档，停止后续开发。       |
| 2026-09-30 | Docker Linux Server 29.8.1 连接成功；核实 Windows 排除范围导致 5432 绑定失败，修复宿主机端口为 15432，原卷保留，API /readyz=200；完整 smoke 仍被 MinIO 阻塞，记录并停止后续开发。                         |
| 2026-09-29 | Docker 29.8.0 客户端可用但 Linux engine 仍不可连接；完成只读原因分析并独立交付前端框架指南与 P2-100–111 任务包，实际前端未搭建。                                                                          |
| 2026-09-27 | engine 再次不可连接，B-ENV-001 重新打开；依用户补充规则建立独立问题档案并停止后续工作；更新 Git 现状，保留未完成设计及前端分工目标。                                                                      |
| 2026-09-26 | 启动 Docker Desktop 解除 B-ENV-001；infra:config、doctor 通过，PostgreSQL / Redis 启动并健康，API /readyz=200；MinIO 镜像被官方移除，登记 B-ENV-002                                                       |
| 2026-09-26 | 归档原文，正式化 Phase 0–10 流程；建立工程和依赖；验证 API、构建、编辑器和 sqlc；记录 Docker/CGO/评审待办。                                                                                               |

## 2026-10-01 再次授权后的结果（历史）

已修复 LICENSE 幂等性及两项入口说明，连续构建 / 服务 / API / check / smoke 通过；独立工程复审通过，七项基础不变量窄范围 frozen，P0-001–005 和 Phase 0 在本地基础范围内完成。随后P1-001设计包与P1-002/010–017任务已交付；P1设计仍proposed /待独立审阅，真实执行未实现。旧失败日志保留，不冒称远程 CI、安全判题或产品前端已实现。
