# 本机模式、题目上传与交互判题独立审阅

- 设计日期：2026-10-02；最终复审：2026-10-03（上海时区）。独立审阅者：`/root/roles_review`，Architect / Security。
- 用户明确选择本机演示、手动切换模式，不实施真实账号登录。该选择不被改写为真实身份或权限保证。
- 本审阅者只读设计、实现与既有运行证据，允许执行可信 fake / 文件 oracle；只修改本报告。不运行 Docker payload、不修改环境、不操作浏览器、不参与实现。
- 审阅依据：AGENTS.md、PROGRESS.md、DEVELOPMENT.md、judge-mvp/0.1.0、PROBLEM_CATALOG_EXTENSION.md、[本轮契约](../judge/MODES_INTERACTIVE_CONTRACT.md)、MODES-001 / 002 / 003 / 004 任务卡，以及 root 关于 participant_stopped / Origin / DTO 的补充。

**最终结论：本轮局部设计、实现与本机产品验收PASS。用户要求的三种手动模式、教师上传与持久题库、固定猜数协议的真实交互判题已交付；本轮发现的必修问题全部闭合。结论绑定文末最终源码与运行镜像，不能扩展为真实身份权限、任意交互器或公网生产验收。**

下文保留各阶段发现、修正与证据追加顺序；历史“待补 / 尚未审结”描述仅代表当时状态，实时结论以本段与最后一节为准。

## 正式设计结论与补充裁定

固定 `guess-number-v1` 的真实双向交互、独立双方容器、来源明确的分别编译产物、共享输出预算与独立清理，符合现有 API 不执行 / 计算隔离 / 快照和清理不变量。动态题库采用服务器生成 ID、严格注册和完整验证后发布，避免通过目录枚举开放任意路径。手动模式采用 API 默认学生和分级投影，准确服从用户选择的本机演示范围。

实施时以下补充是审阅采用的约束，须同步进契约；不改变预算、隔离或用户范围：

- Case.Interactor 的 `participant_stopped` 表示固定交互器合法 exit 1 后，协调器为了收束 WA 对选手实施监督终止。仅此情形可允许 WA 加非零选手退出码；必须同时 interactor exit 1、无 OOM 且有监督终止依据。自然非零 / 非 OOM exit 137 保持 RE；任何 AC 都要求双方自然 exit 0、`participant_stopped=false`、完整输出 / attach 与确认清理。客户端严拒绝矛盾。
- 共享原始 1 MiB 停止线必须保存触发方归因：选手 stdout / stderr 超线归 OLE，交互器超线归 SYSTEM_ERROR。固定交互器 exit 1 只表达合法协议拒绝；配置 / 秘密解析 / 内部故障必须用非 0 / 1 的退出码，避免把基础设施错误解释成选手 WA。
- 教师看到的 input / expected 来自本次 worker 已读取的快照，不在 API 中从可变目录补读。显示文本 UTF-8 转换和 byte 截断仍有界，不改变原始秘密、checker 字节或快照 hash。两个方向 stdout 是分别保存的方向记录，不能伪装成带全序的完整 transcript。
- root 补充上传 Origin 检查可接受：只允许本机前端 / API 的固定同源，拒绝跨源；不开放 CORS，CLI 无 Origin 可使用。这是本机请求约束，不提供真实身份认证。页面和文档应继续明确可手动切换。
- Metadata 的 problem_type 在新 wire 中必填，并与 checker / interactor 元数据匹配；客户端继续严格字段、nullable 值、枚举、预算、大小及状态矛盾校验。题目上传与现有 built-in 目录共同来自同一 Store，不依赖可变全局注册。

当前未发现需改变方案才能实施的设计阻塞。root / 各 Builder 根据上述契约和补充实现；本次结论仅打开局部设计门禁。

## 实施前约束建议

1. 模式是本机演示选项。API 默认 student，严格接受 student / teacher / engineer；API 对结果作服务端投影，学生结果不包含逐例输出、错误、退出码、OOM 证据、隐藏 case 标识、交互器消息或对话。仅 UI 隐藏不满足用户“不展示具体样例错误”的要求。模式选择本身不提供真实权限，手动切换可见详情须明确标注。
2. 教师题目上传须有请求总量、文件数、每文件大小与内容类型限制；只接受受控题目 ID、普通文件 basename 与允许后缀。服务器生成目标目录，不接受宿主路径、命令、镜像、flags 或自定义资源上限。完整验证后原子发布；新 ID 避免覆盖在途题目。重启后可发现已发布题目，失败不暴露半成品。
3. 真实交互判题需要两个受限进程的双向通信，不能把交互题等同于普通 .in / .out 比较。优先采用固定内置协议与受控秘密数据格式，教师不需要上传任意交互器代码。秘密只进入交互器隔离容器，不进入选手容器。
4. 可信协调器须监督双方生命周期、总时长、双向通信字节 / 回合上限、EOF、无 flush、提前退出、异常退出、OOM、取消与独立清理。只有合法完成的协议及双方可信最终状态可判 AC；交互器异常归 SYSTEM_ERROR，不由任意容器退出码假冒 WA。选手异常仍按 RE / TLE / MLE / OLE 的可信证据分类。
5. 若允许教师上传交互器源码，则它也属于不可信计算，独立 OCI 编译和运行；API / 宿主不能直接执行。当前用户需求本身不强制开放该范围。
6. 验收须覆盖服务端学生投影、教师上传成功 / 拒绝 / 持久重启、标准题兼容、交互 AC / WA / CE / RE / TLE / MLE / OLE、非法协议 / 超查询 / 不 flush / EOF、取消 / 两方后代清理及系统故障。

## 审阅状态与证据

- 已只读当前目录、API、filejudge 请求 / 响应协议、文件读取、编译与普通用例执行、Docker 控制 / preflight 的相关实现，既有隔离、原始输出检查、独立 cleanup / taint 不应被本轮扩展弱化。
- 已向 root 发送上述约束和限定内置猜数协议的范围建议。
- 已完整只读本轮契约、四项任务卡、台账与最新授权。采用固定内置交互器，不开放任意交互源码上传。
- 本阶段未执行任何测试；最终实现、fake 检查、真实 Docker / API / 浏览器证据和源码摘要尚未齐备，后续补充后再给出实现 / 产品结论。

## 2026-10-03 实现首轮审阅

状态：当前代码 / 可信 fake 审阅通过；真实资源与产品验收等待本轮证据。已复读 Store / strict multipart / manifest / loader、engine / Docker 与双向监督、内部 DTO / Client、API 模式投影 / 上传 / debug 和 main 接线。不得将下文可信检查替代真实 Docker / API / 浏览器验收。

以下属于实现期审阅修正，已通知 root / runner；不是新外部环境阻塞。行号对应首轮复读版本，后续会更新处理状态。

| 发现                          | 首轮位置                                 | 触发与后果                                                                                                                                                                                                                             | 状态                                  |
| ----------------------------- | ---------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------- |
| casewall 在创建阶段映射不一致 | interactive.go:160–178                   | 双 create / inspect 耗尽已含创建的 casewall 时返回 interaction_start_timeout / container_create_failed，即 SYSTEM_ERROR；契约规定整体 casewall 超限 TLE。必须收束映射与未启动 exit_code=null 的可信证据或明确受审设计修订。            | 已通知，待修正 / 裁定                 |
| stdin 复制错误丢失            | docker.go:45                             | 自定义 io.Copy(stdin,input) 的非预期转发错误被忽略，stdout attach / 退出码完整不等于输入输送完整；不能据此 AC。须区分监督关闭 / 正常 peer EOF 与非预期 I/O。                                                                           | 已通知，待修正                        |
| 结果未与请求题型绑定          | contract.go:330、httpapi/router.go:85–87 | interactive 请求若收到形状合法但完全缺少交互证据的 batch AC，可被按 presence 推断的 Client 接受，API 再标成 interactive；固定 interactor source hash 也仅检验格式。须核对请求题型与固定版本 hash，编译前 CE / SYS 保留合法未开始语义。 | root 已修正并独立复读 / fake 检查通过 |
| 输出停止源存在竞态            | interactive.go:65–71                     | total.Add 与 owner.CAS 分离；第一超线 writer 被调度暂停时，第二 writer 可抢先登记 owner，错误区分选手 OLE 与交互器 SYSTEM_ERROR。须使预算累加 / 保留量 / 首次归因同一临界区，转发与 cancel 在锁外。                                    | 已通知 runner，待修正                 |
| WA 监督终止依据不足           | interactive.go 原 WA 分支                | 旧 inspect 的 Running 只说明当时运行；选手可在取消 attach 后、cleanup 前自然 exit137，使 cleanup 没有 kill，却仍 participant_stopped=true / WA。标记须由实际监督 kill 的执行与确认事实支持，自然非零不能假借旧 Running。               | 已通知 runner / root，待修正          |
| 取消后遗漏另一方自然系统异常  | interactive.go 原 event / verdict 分支   | 选手 RE / TLE / OLE 触发取消后另一方 event 被跳过；若交互器自然 exit2 而非监督 kill，最终可能提前返回选手判定，漏掉 SYSTEM_ERROR 优先。需要双方真实监督 kill 归因与最终状态收束。                                                      | 已通知 runner / root，待修正          |

本审阅者亲自执行 `node scripts/go.mjs test -count=1 -v ./internal/problemstore/... ./internal/filejudge/... ./internal/httpapi/...`，exit 0。仅可信 fake / 文件测试，没有 Docker payload 或环境修改。覆盖严格 multipart / 注册持久重载 / 无半公开题 / 内容边界、学生服务端投影 / 上传 Origin 与模式限制、交互隔离参数 / AC / supervised WA / 自然 RE137 / OOM / TLE / 双向输出 / 交互器错误 / attach / 不确定 create / 清理失败 / 取消等。Window symlink 两项因既有权限限制 skip，未称通过；需 root 的 Linux 可信单测补证。

root 已知的旧内置 allowlist 拒绝上传题问题随后修正；独立复读和实际 fake 测试 `TestEngineAcceptsRegisteredUploadBeforeCompile` 的 batch / interactive 均通过。本报告不将 root 自检重复登记为独立发现。

已确认没有新增任意 interactor 上传 / 宿主执行；教师详情来自 worker Snapshot 字节的有界显示；student 使用显式 allowlist map，结果没有 cases / stdout / stderr / 隐藏 ID / 哈希 / 镜像 / 交互编译，SYSTEM_ERROR 消息通用化。teacher 与 engineer 分别构造教学详情与内部证据。真实运行与浏览器尚待 root 提供本轮证据。

### 同轮窄补充裁定

同意 root 提出的 `start_timed_out` 可选 bool。仅交互题、选手及交互器编译均成功、双 create / inspect 阶段 casewall 超期而尚未启动任何 attach，可返回 TLE + 双方 exit_code=null；无 OOM、无 participant_stopped，清理必须确认。job 超期 / 取消 / 基础设施失败保持 SYSTEM_ERROR。Client 须拒绝 null / 非 bool、batch 或其他 verdict 配 true、已有退出事实配 true；启动后正常 TLE 保持原可信退出事实约束。该例外不允许 cleanup 的正常耗时在事后将已完整结束的 AC 改判 TLE；cleanup 仍受整体 job / 独立清理预算。

API 按所请求题型和本人 source SHA256 核对结果、Client 核对固定交互器版本 source digest 的方案可采用。编译前 CE / SYSTEM_ERROR 可以尚无交互证据；已运行的 interactive 正常结果必须有成功 interactor compile / 固定 source digest 和每例事实；batch 不接受交互证据。

本审阅者独立执行 Node 文件 oracle：guess-number 的 minimum / middle / maximum 三份秘密输入格式与范围正确；枚举所有 1..1000，二分均能在至多 10 次查询内找到，最大 10 次。未运行 C++。独立从固定 Go raw string 得到交互器源码 SHA256 `bd867a6c212c0b77d0792aa70296bed2c024fb95987a63e80901a89f5333c37f`；该字符串 CR 数为 0，按 Go raw string 规则规范化后摘要相同。

### root 所有权修正与前端可信检查

独立复读确认 Client 实际核对本人请求的 `source_sha256`，固定交互器 source digest 精确匹配，API 在投影前用可信登记题型执行 MatchesProblemType。batch 拒绝交互证据；interactive 正常已运行结果要求成功交互编译及每例交互事实，编译前 CE / SYSTEM_ERROR 保留没有交互证据的合法状态。

本审阅者执行 `TestInteractiveStartupDeadlineWire` / `TestInteractiveClientEvidenceContradictions`、HTTP 的错题型 worker / 编码请求总量 413 / 默认学生投影专项 fake，exit 0。随后 root 补本人 source hash 的 wrong_source 拒绝测试；本审阅者已复读该实现，最终完整可信检查将在 runner 收束后执行。

本审阅者亲自执行 `node apps/web/scripts/verify-modes.mjs`，八项可信 HTTP mock / React 静态渲染检查通过：mode 请求头、模式切换取消、迟到响应拒绝、调用者取消、学生 / 教师 / 工程师展示、过期模式层拒绝。未打开浏览器，不将 SSR/mock 称作真实 UI 验收。

前端只读确认 queryKey 包含 mode，Workspace key 为 id:mode 并卸载 abort；global request 登记在切换时取消，发起和读取响应后核对 getMode，反馈组件还核对 result.mode。教师 / 工程师的旧详细结果不能通过正常模式切换残留；同题源码保留但结果重建，页面如实说明。本机演示标签、仅教师 / 工程师上传、仅工程师 debug 路由与不同层次展示符合方案；真实上传与在途 UI 切换仍需 root 本轮浏览器证据。

### 当前修正复审结论

上表六项首轮发现现均已解决，代码复读及独立可信检查通过；表内“待修正”描述保留为首轮历史，实时处理状态以本节为准。

- startup create / inspect 的 casewall 超期使用 StartTimedOut + TLE，仍未启动双方，不编造退出码。job / 取消 / 真正控制错误保持 SYSTEM_ERROR，cleanup 确认后才能保留 marker，Client 严格限制。
- 真正 `runDuplexCommand` 保存 io.Copy 错误，非预期错误收束为固定 duplexInputError，case 以 SYSTEM_ERROR 处理；明确本次关闭 / 取消产生的允许 pipe 错误不会错判。可信 self helper 仅读 stdin 后 exit0，测试实际 StdinPipe / io.Copy 路径，非意图执行的题目程序。
- count / retention reservation / first owner 在同一短 mutex 内完成；存储、转发与 cancel 在锁外。并发测试暂停首次超线方的 retention，另一方无法抢先归因。
- cleanup 内部保存双方向真实 kill 证据：本次 Docker kill 控制成功，最终停止且 exit137，才记录 SupervisorKilled。WA 的 ParticipantStopped 来自该证据；旧 Running、kill 失败、已自然退出137均不冒称监督事实。新 fake 均保持 RE。
- 取消后的最终 interactor 自然非0/1（未经本次确认 kill）收束 SYSTEM_ERROR，保留系统故障优先；RE / TLE / OLE 先触发也不能漏掉迟到的自然交互器 exit2。双方清理仍共享独立15秒，失败 taint。
- 本人 source hash / 请求题型 / 固定交互器 digest / 编译次序 / 必有快照字段，均实际严格验证。新增健康 modes_contract_version，supervisor 检查后才可复用服务，旧仅健康实例不冒充新双向 probe 已通过。

本审阅者在修正全部落盘后再次亲自执行 `node scripts/go.mjs test -count=1 -v ./internal/problemstore/... ./internal/filejudge/... ./internal/httpapi/...`，exit 0。包含首轮测试、原始 relay 四种完成路径、确定性 first-owner 并发、startup create / inspect TLE 与 job超期、WA 后自然137与 kill 失败、选手 RE / TLE / OLE 后迟到 interactor异常、所有 snapshot / strict wire / HTTP 模式投影和上传拒绝。既有两项 Windows symlink skip 不变。当前没有未解决必修代码问题；真实资源 / 产品证据未齐，最终源码与镜像绑定待验收收尾。

## 2026-10-03 首批真实证据审计

root 报告 `npm run check`、基础 smoke / 旧 compile smoke、Linux 可信单测与 verify-modes 八项通过。本审阅者已只读 `.local/evidence/modes-check.log` / `modes-linux-unit.log` 及对应脚本；check 日志完成 Go / TypeScript / Vite 构建，仍只有既有 Monaco 大包警告。Linux 日志明确 `TestCaseSymlinkRejected` 与 `TestStoreRejectsSymlinkRootsAndRegistration` PASS，覆盖 Windows 中不能创建 symlink 的两项；Linux 测试为可信 Go 单测二进制与 fake Docker 调用，受限容器没有 daemon socket，只有测试专用只读 fixture 挂载，不等同于选手获得答案。

本审阅者只读 `.local/evidence/interactive-matrix.json` 和 `scripts/interactive-test.mjs`。root 的真实判题于上海时间 2026-10-03 09:31:55–09:34:20 完成，`completed=true`：13 次 HTTP200 判题、36 个 case，均 passed。包括二分 AC、越界最终值 WA、合法拒绝后监督终止 WA、非法协议 WA、超过查询次数 WA、最终回答前 EOF WA、CE、崩溃 RE、自然137非OOM RE、不 flush TLE、真实 OOM MLE、stderr洪泛 OLE、交互拒绝后的后代清理。每次登记 passed 之前脚本要求成功 Docker 查询没有本项目容器及 job 目录为空。

本审阅者另执行只读 Node evidence audit，exit 0：所有 verdict 与预期相等，runtime image 与 profile 一致，CPU / 峰值 null；AC 双方 exit0 / 无监督终止，对每个秘密重新计算每次查询正确回复和最终值，至多10次查询，3例均吻合；MLE 的真实 OOM=true，RE无OOM；监督 WA 真实退出137 + interactor1 + stopped；固定 source digest 与本审阅者独立计算值完全一致。真实 Docker / C++ / HTTP 由 root 执行，本审阅者没有执行 payload。

镜像仍 `sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74`。本阶段本审阅者亲自执行 source-digest 得到非 MD 源码175文件 / `18b496fa892b5ad09780acabc896f4827c56a78df0f31ab7ec2d73957453f175`，仅为此批审计时版本，最终绑定待收尾。

发现真实验收证据仍需补充，已通知 root：

- 现 paired cancellation 仅用 `docker ps -aq` 数量≥2，不核对双方 Running；当前可证明双 runtime 创建后取消和清理，不能冒称双方已在运行时取消。增强条件及保存实际 Running / label 证据后再登记该项。
- 当前真实矩阵没有交互器系统故障，fake 已覆盖 OOM / 自然错误退出 / attach 等，但契约列明实际验收；建议确认本次双运行后向 interactor 定向 TERM，期望 SYSTEM_ERROR / 非OOM退出事实和双方清理。未提供此项前不得称已真实覆盖。
- “WA-final”用的是 `! 0`，实际是越界拒绝，不能冒称已测试合法范围的错误最终值；可补 `! 42`。现有 source 审阅可确认 wrong-final 比较实现，本报告将运行范围如实记为越界 WA。

root 正在串行补其余标准题矩阵、题库、上传 / 重启和浏览器证据。当前仍没有产品总体 PASS，本批实证不替代这些项目。

## 2026-10-03 响应大小边界补充审阅

新增必修：Client 的完整 worker 响应仍限制为 6 MiB。新 input / expected 快照与原 stdout 一起进入内部 wire，JSON 对控制字符最多扩展为六倍。合法上传 8 组每对 60000 个 NUL 的 batch 数据总文件量 960000 bytes，可落在现有 1 MiB 请求预算内；正确程序输出同样的 60000 个 NUL 时，input / expected / stdout 的完整 JSON 约为 8.64 MB。Client 会在投影前拒绝该正确结果，学生和工程师均可能返回 502。该问题属于本轮实现边界，已通知 root 并获接受，不是外部环境阻塞。

root 提议仅将内部 MaxResponse 改为 16 MiB。独立上界审阅通过：每例 input / expected / participant stdout 各64 KiB，participant stderr16 KiB，interactor stdout64 KiB / stderr16 KiB，总288 KiB；8例按六倍 JSON 转义为13.5 MiB。两个编译输出各64 KiB 同样按六倍计为0.75 MiB，合计14.25 MiB；固定长度的键、ID、摘要、枚举、整数和有界消息远小于余下1.75 MiB。16 MiB仍是明确的内存 / 解析停止线，不改变单字段、执行、用例文件、上传请求或 job 的预算。

本项代码修正已独立复审通过：contract.go 的 MaxResponse 现为16 MiB，读入仍采用 LimitReader(MaxResponse+1) 并拒绝超线；契约同步且其他预算未变化。本审阅者亲自执行 TestClientMaximumEscapedSnapshots（batch / interactive，各8例最大控制字符快照）与 TestClientProtocol（含16 MiB+1拒绝），exit0。两类完整有效响应均超过旧6 MiB、低于16 MiB，Client接受并保留第8例完整expected。root 随后将实际使用临时8组60000 NUL fixture并正确输出，验证student / engineer均AC；临时manifest和自建文件须在finally恢复，避免留下测试题。

本项目前状态：设计、代码、可信边界检查通过，实际NUL边界验收待补。前文“当前没有未解决必修”仅描述首轮六项收束时的状态；当前仍不能登记产品总体通过。

## 2026-10-03 标准题回归与模式界面证据审计

本审阅者已复读当前 scripts/judge-test.mjs / problem-smoke.mjs 的断言、受控fixture恢复和clean检查，以及本轮两份真实报告；另亲自运行只读Node audit，exit0。judge-matrix.json上海时间09:50:19–09:53:12，completed=true / 17项passed；13个主要判题含44例，覆盖原AC/WA/CE/RE/TLE/MLE/OLE、展示截断后原始checker仍AC、无效UTF8、后代清理与原有隔离、忙429/运行时取消、配置错误SYSTEM_ERROR、exact/tokens和动态metadata。MLE均OOM、自然137均无OOM，未采集CPU/峰值均null，runtime image仍与profile绑定。脚本在主要passed登记前确认本项目容器和job目录为空，受控文件变更在finally恢复。

problem-smoke.json本轮上海时间09:56:47–09:59:20，completed=true / 原5题共10次passed、48个case。不同于starter的独立正确方案全部AC，常见错误方案均成功编译且至少一例WA；checker、case数/ID和题面metadata按fixture核对，未知或路径型题号404及每次清理通过。此前旧17项脚本固定metadata预期因新增problem_type而适配，仅测试预期变化，不改变产品行为；这些是root执行的真实HTTP/Docker证据，本审阅者没有实际提交。

已只读本轮modes-ui-student-wa / teacher-wa / engineer-evidence / debug / switch-cleared / inflight-switch / mobile-upload的AX文本与student-WA、teacher、mobile-upload截图。学生WA仅总体判定/0比2/本人编译；教师展开有秘密输入和双方向输出、逐例错误、分别记录非全序说明；工程师另可展开run/hash/image和固定交互器digest；debug如实说明读取配置不等于实际运行通过。切回学生的AX不含旧逐例内容，在途切换后的新工程师工作区结果为空，模式手动演示标签明确。移动上传截图宽度内控件可见，无明显横向裁切，console记录为空数组。真实浏览器操作由root执行，本审阅者仅查看既有文本/图片；最终浏览器QA记录、上传重启、双运行取消/交互器故障及NUL边界证据仍待收尾，以上不替代整体验收。

## 2026-10-03 三模式上传、生命周期与重启补证

已只读teacher-problem-smoke.mjs和本轮teacher-problem-smoke.json；上海时间10:00:26–10:01:33完成、completed=true / 6次passed。真实发布batch / interactive两题，各默认学生AC、教师WA、工程师AC，均2例；默认学生上传403，任意interactor源码名称400且目录列表不增加，debug仅工程师可访问。本审阅者亲自执行只读Node证据oracle，exit0：学生实际结果恰为10个允许字段且AC2/2；教师实际每例input / expected等于发布数据、没有exit/OOM/CPU/峰值，顶层无run/hash/image/交互编译，互动层没有exit/OOM/监督停止；工程师双方自然exit0 / 无OOM / 无监督停止。interactive教师实际stdout `! 1`、秘密7与900均合法范围而不等1，交互器stderr `wrong final answer`，补齐此前`! 0`只能证明越界拒绝的缺口。

新增interactive-lifecycle-smoke.mjs只观察本次source / run绑定的case-a双方，在定向TERM前再次inspect核对ID / Name / owner / run / stage / Running。completed报告于上海时间10:07:09–10:07:34完成。独立oracle核对两项before记录双方均Running、正确Name与labels、无OOM/State.Error、已经Started；所有保存CLI查询均success / exit0。取消确实AbortError并0容器/0job。系统故障项向固定已核对的交互器TERM，真实退出143 / 非OOM、结果和case均SYSTEM_ERROR；选手被收束为137 / 非OOM，participant_stopped=false，不冒充合法WA停止；同样确认0容器/0job。本批实证闭合此前仅创建后的取消证据以及真实交互器系统故障缺口。

首次观察器失败证据仍保留interactive-lifecycle-first-attempt.json：completed=false、tests为空、docker_inspect_failed。首版观察范围包含短命compile容器，ps/inspect间可能发生正常清理；首版没有记录失败ID，因此只能认定存在观察器查询竞态可能，不能断言当次唯一原因。第二版限定source/run绑定的runtime并记录全部查询，真实inspect错误仍失败，没有把错误当不存在或泛化重试。产品核心未修改。[修正档案](../issues/2026-10-03-modes-development-corrections.md)保存经过；独立审阅采用第二次完整证据，不覆盖首版失败。

本轮正常Ctrl+C停止提示fetch failed并exit1，root未宣称带token shutdown握手成功。随后无5173/8080/8082监听、0容器/0job并重新严格ready，是root实际只读核实和新启动记录；源码支持信号同时送达的可能，但当次先后未捕获，不能作为已知根因。该非阻塞运维提示已单开[用户处理档案](../issues/2026-10-03-supervisor-stop-warning.md)，未修改退出实现、未以忽略错误绕过预检。新judge上海10:06:32严格ready、API10:06:36启用，后续实际判题证明新实例可用；最终ready与源码绑定仍待收尾。

upload-persistence.json已completed=true，上海10:02:30保存before、10:09:13完成after。3道已发布题（教师batch / interactive与真实浏览器上传interactive）重启后均HTTP200 / AC2/2 / 同testdata_sha256。独立只读Node oracle重新逐文件计算当前注册、manifest、input / answer的SHA256，与before所有snapshot完全相等；又按Go loader的uint64小端length-prefix次序独立计算原始manifest+每例ID/input/expected摘要，与before及after的testdata hash全部匹配。脚本实际核对重启前后完整题目详情相等并要求每次清理；该实证确认发布持久性和可判，不把只读Store单测当重启证据。

本节代码 / 实证复审通过，没有新必修问题。实际NUL边界、最终完整check / readiness / 浏览器QA记录及最终源码摘要尚待齐备；整体验收仍不提前登记PASS。

### 大量转义字符的真实边界收束

escaped-response-smoke.json于上海10:09:55–10:10:57完成，completed=true / fixture_restored=true。root临时将A+B受控manifest改为8组exact-v1，每份input / expected60000 NUL，独立C++直接输出60000 NUL；两mode真实HTTP200 / AC8/8。学生公开响应仅290 bytes，没有cases / 摘要；工程师公开完整响应8642720 bytes，超过旧6 MiB、低于新16 MiB，每例input / expected / stdout均60000 NUL、退出0 / 无OOM。脚本实际对这些完整字段逐一断言、确认每次无容器/无job，并在finally恢复原manifest、仅删除以wx新建的文件，不覆盖已有fixture。

本审阅者亲自执行只读oracle，核对上述报告大小、模式、AC8/8和完成标记；当前A+B恢复tokens-v1原4例，没有escaped-bound文件。另独立重建临时JSON manifest及8组原始NUL字节，按Go loader的length-prefix算法计算testdata SHA256 `986b56e3d378d7bbf00f01515095b7081433f7629c23c3fad83c6ceeffaabb7c`，与真实报告完全一致。此前6 MiB必修边界的设计 / 代码 / 可信测试 / 实际结果已全部闭合，可以关闭该项。

本批结束时独立source-digest为176文件 / `8785b5dd63074e25633de5f6d78e52ddbaf4b94c6b6a4df08e3349a4fbb60b1c`，与root当前报告一致。镜像及固定交互器来源摘要不变。等待最终完整check / readiness / 浏览器QA文档和MD台账收尾后，再登记本轮产品结论。

## 2026-10-03 最终独立签署

最终局部设计PASS、实现PASS、本机产品PASS，无遗留必修代码或验收问题。此结论依据独立源码复读、本人可信fake / 文件及证据oracle、root本轮真实Docker / HTTP / 浏览器实证；本人没有执行真实选手程序、修改运行环境、操作浏览器或参与产品实现。六项首轮发现、额外wire严格约束、6 MiB响应边界及真实观察器证据缺口均已按上文收束，不用自检冒充独立审阅。

已复读modes-final-check.log完整format / Go vet / ESLint / 全Go tests / Go与TS/Vite build成功，包含16 MiB最终修正；既有Monaco大包警告仍单独记录。已复读[本轮验证](../judge/MODES_VALIDATION_2026-10-03.md)与[浏览器验收](../judge/BROWSER_QA_2026-10-03.md)，文字和原始证据一致。最终home AX / 截图显示默认学生模式、全部9题、3道上传题保留，最后console为空数组；无新增页面必修问题。

最终modes-final-readiness.json于上海10:13:32完成：Web / API / compiler / judge均HTTP200，新judge健康含modes-interactive/0.1.0、固定profile和镜像；9道登记题（6内置+3上传），原A+B恢复tokens-v1 / 4例，成功Docker查询证明0本项目容器且0job目录。本审阅者亲自运行只读最终证据oracle并再次执行source-digest，均通过，源码摘要与root最终readiness完全一致。

最终绑定：

- 非MD源码176文件，SHA256 `8785b5dd63074e25633de5f6d78e52ddbaf4b94c6b6a4df08e3349a4fbb60b1c`，分支task/MODES-001-interactive-2026-10-02；包含历史未提交基础工程和本次增量，不冒充已提交Git版本。
- OCI镜像 `sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74`，judge-mvp/0.1.0，GCC14.2.0 / C++20。
- 固定guess-number-v1交互器源摘要 `bd867a6c212c0b77d0792aa70296bed2c024fb95987a63e80901a89f5333c37f`；Anolix本地参考commit `4502d2a798f3130d3c540c5c31d45251ca103623`，未执行其CLI。

用户可继续本机演示和后续开发。手动切换不提供身份认证；目前只有固定猜数交互协议、单同步worker，提交源码和结果不持久化（上传题目与数据已持久化），CPU累计 / 峰值仍null，没有全序通信日志；本机CGO race及远程CI未运行，完整Phase1与公网生产不在此次签署范围。Windows正常退出时未获得shutdown响应的非阻塞提示保留在专档交用户处理，当前已核实无残留并重启通过，不冒称该提示已修复或握手成功。最终MD台账与链接检查由root完成，不改变此源码 / 镜像绑定。
