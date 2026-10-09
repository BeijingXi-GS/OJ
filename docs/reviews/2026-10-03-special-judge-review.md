# 学生页面精简与特殊判题独立审阅

日期：2026-10-03；独立审阅者：`/root/spj_review`，Architect / Security。实际任务 SPJ-004；依据 AGENTS.md、PROGRESS.md、DEVELOPMENT.md、[本轮契约](../judge/SPECIAL_JUDGE_CONTRACT.md)、现有 Store / loader / engine / Docker / Client / API / supervisor 源码。

本审阅者仅进行只读审阅和可信数学 / fake 检查，仅修改本报告；不修改实现，不控制实际服务或 Docker，不运行上传或选手 C++，不把主执行者自检登记为本人测试。

## 设计首轮

结论：主体方案可接受，核心设计门禁尚待下列澄清写入契约后正式通过。学生保留解题条件和本人编译错误，工程说明仅工程师可见；内置构造题允许多个合法输出；教师上传 C++ checker 被视为不可信计算，在独立 Judge 内隔离编译与运行。该方向符合 API 不执行、不可变快照、固定路径、资源证据与确认清理的既有不变量。

| 审阅项             | 必要约束与原因                                                                                                                                                                                                                                                                             | 首轮状态                |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------- |
| 特判语义与题目绑定 | 仅凭 `problem_type=batch` 无法区分普通、内置与上传 checker。需依据受信 manifest / metadata 的 checker 校验 Result；上传 checker source digest 应与本次题目快照绑定，普通 tokens / exact / 内置以及 interactive 拒绝 C++ 编译、hash 和 runtime 元信息。编译前 CE / SYS 保留合法未开始状态。 | 已通知 root，待契约明确 |
| 总预算             | `20 + 10 + 4 × (5 + 2) = 58` 秒只覆盖 attach 阶段，create / inspect / cleanup 仍消耗总体时间。70 秒是任务总体上限，不保证最慢容器控制路径下全部最大规格例子完成；超过总体预算归 SYS 并确认清理。checker 编译必须实际使用 10 秒，而非现有 compileNamed 的 20 秒硬编码。                     | 已通知 root，待契约明确 |
| JSON 重复字段      | 当前 worker Client 的 `DisallowUnknownFields` 和 map 解码不会拒绝重复 key。契约中的严格未知 / 重复 / null 要求应覆盖 metadata、Result 和嵌套 checker，增加递归重复字段检查及原始字节回归。                                                                                                 | 已通知 root，待契约明确 |
| 输出限制           | 非编译 stage 现有预算是 stdout + stderr 合计 1 MiB；草案“原始合计 / 每流”需统一为合计停止线。64 KiB / 16 KiB 是展示上限，判题读取原始 stdout；特判反馈另限 1024 UTF-8 字节。                                                                                                               | 已通知 root，待文字明确 |

可采用的其他约束已经独立确认：固定 `checker.cpp`、源文件和请求大小、每次读取普通文件快照并纳入 testdata hash、最多 4 例、独立子目录与产物、API / 宿主机不执行、无网络 / nonroot / capdrop / 只读文件、固定 argv、仅本例的输入 / 标准答案 / 选手原始输出以及特判自身产物挂载，避免暴露选手二进制、其他用例或 jobs 根。teacher checker exit 0 / 1 表达 AC / WA，编译错误、其他退出、超时、OOM、输出超限和基础设施错误均 SYS；选手未成功运行时不执行 checker，保持 RE / TLE / MLE / OLE 的可信判定。

Windows supervisor 的旧 warning 不能通过忽略 HTTP 错误关闭。所拥有的 judge 进程正常退出、Docker owner 查询成功且为空、jobs 目录为空共同构成停止事实；查询失败、非零退出或未确认清理保留失败，不停止借用实例，不打印控制 token。真实正常退出和运行中退出证据尚待本轮执行。

## 本人已执行的可信检查

设计数学 oracle 在独立 JS 中枚举 `n=1..1000`。排除 2 / 3 后，998 个 n 的“偶数升序后奇数升序”和其逆序均覆盖 `1..n`、无重复、每对相邻差不为 1；其中 997 个 n 的两份答案不同。此检查证明示例构造与多答案验收选择成立，不是已实现 checker 或 C++ 运行证据。

本阶段未执行工程单测、真实 Docker / API / 浏览器验收或最终源码摘要检查；尚无实现 / 产品 PASS。后续独立复读及本人实际 fake 测试和主执行者绑定证据将追加在本报告，历史审阅不替代当前实现。

## 正式设计通过

**局部设计 PASS，允许 SPJ-001 / SPJ-002 / SPJ-003 按 `special-judge/0.1.0` 开始核心实现。实现与产品验收尚未通过。**

独立复读更新后的契约及 SPJ-001 / 002 / 003 / 004 任务卡，首轮四项澄清全部写入：API 提交前加载受信本地快照，并核对 checker、checker source SHA 与 testdata SHA；worker metadata / Result 递归拒绝重复 key 及非法 null；70 秒明确包括控制操作与清理，超预算统一 SYS；stdout + stderr 合计 1 MiB 停止线与显示限制明确分离。普通 / 内置 / interactive 不得携带上传特判证据。

同意在 API 的共享 Store 中读取不可变期望快照用于返回结果绑定，读取不是编译或执行。真实运行加载到不同源码 / 输入 / manifest 时宁可拒绝协议结果，不能把错版本的判断投影给学生。生产 API 与 worker 应指向同一注册目录和内置根；现有 fake 测试必须创建真实 fixture，不通过“无根目录时放行”弱化生产约束。

core 实现须保留本报告和契约约束，并提交真实 C++ checker 的正常 / 编译错误 / 非法退出 / 超时 / OOM / 输出限、选手失败优先、原始 stdout 大于显示上限、上传重启绑定、mode 投影、旧 batch / interactive 和运行中关闭清理证据。测试未执行的路径必须如实登记，最终结论独立复读后另给。

## 实现首轮：root 适配与生命周期范围

此时实现仍开发中，未冻结。本次重点独立读取 checker_contract / 内部 DTO / Client、API 投影与快照绑定、judge-shutdown helper / mvp-dev 和服务 main；顺带读取新 store / engine / special 执行路径。未给最终实现或产品 PASS。

| 必修发现                 | 触发与实际后果                                                                                                                                                                                                                                                                                                                   | 状态                       |
| ------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------- |
| 启动中 stop 未被完整等待 | `stop` 遇到 `stopping` 直接返回。SIGINT 打断预检 loop 后，第二次 `await stop(1)` 可立刻结束，接着 `process.exit(0)`，首个 stop 仍在等待 owned judge 和清理事实。另 await existingJudge / judgeState 时收到信号，后续启动路径未检查 stopping，可能继续创建子进程。应共享一个 stopPromise 并等待，同步 guards 阻止停止后继续启动。 | 已通知 root，待修正 / 复验 |
| preflight 信号尚未注册   | judge main 在完整 Preflight 后才注册 Go signal context。预检中共享 Ctrl+C 可使 judge 默认立即退出，尚在执行的容器无法由该进程清理。应提前注册 signal，向 preflight 传入可取消 context，并完成独立 cleanup 后退出；真实预检中停止仍需证据。                                                                                       | 已通知 root，待修正 / 复验 |
| 结果例子未与快照列表绑定 | 首轮 MatchesSnapshot 只核对 type / checker / hash。返回少量 AC 例、填入完整快照摘要仍可被接受；custom wire 也仍允许 5..8 例。应验证返回例子为快照同顺序前缀，非 SYS 的成功运行结果必须完整，custom 至多 4 例。                                                                                                                   | 已通知 root，待修正 / 复验 |

当前已核实的正确路径：工程说明仅 engineer detail 包含；teacher / student TLE / MLE / WA / SYS 整体文案不含内部分类原因；student 显式 allowlist 不包含 checker 或反馈；teacher 只有 checker_feedback，不含特判 hash / artifact / 退出证据；API 提交前实际 LoadWithStore，并对返回结果调用 MatchesSnapshot。worker 新增递归 uniqueJSON，覆盖元数据和内部 Result；optional checker / compile / feedback 的非法 null 与形状校验存在。teacher source 在 worker load 时读入普通文件快照，先绑定 hash，独立子目录编译且确实传 10 秒预算；每例仅在 executionVerdict 为 AC 后运行 checker，使用四个固定只读挂载、64 MiB / 32 pids / 2 秒与固定 argv；原始 stdout 用于文件，反馈在 UTF-8 安全显示后限 1024 字节。

本人亲自执行 `node --test scripts/judge-shutdown.test.mjs`，4 个可信生命周期 helper 测试全部通过；执行 `node scripts/go.mjs test -count=1 ./internal/filejudge -run 'TestCheckerWireStateValidation|TestCheckerSnapshotBinding|TestMetadataNewCheckerLimitsAndDuplicateJSON'`，exit 0。只运行可信本地 fake，未启动任何 Docker payload；上述通过不覆盖 startup 竞态或真实清理，也不代替开发中的完整检查。

## 首轮修正复读与私有取消通道设计

已复读 root 修正：stop 返回同一个 stopPromise；异步构建和 await 后 guards 使停止后不继续启动，结束仅 await 而不直接 process.exit；Go signal context 在 Preflight 前注册；MatchesSnapshot 校验用例是同顺序前缀，成功编译且非 SYS 运行必须完整；custom wire 最多 4 例，pure-null 测试不再夹入 unknown 字段。前三项实现必修按这次复读已修正，真实生命周期结果仍待 root 的新证据。

**同意私有 stdin 取消通道的窄设计。** 仅本 supervisor 拥有的 judge 子进程使用 pipe；只有非空临时控制 token 的 Go 实例启用 reader，无 token 的独立实例不读取或取消。只识别短且固定的 shutdown 行，明确短长度限制；EOF / pipe 失效也触发生命周期取消，防止拥有者已经消失而 worker 继续运行。该通道不增加网络端口，不执行输入，不记录输入或 token；发出命令不是确认清理，仍须拥有的进程正常退出、Docker owner 查询成功为空、jobs 为空。parent 需处理 EPIPE 并保持写操作有界。root 保持 HTTP 受控 shutdown 作为正常阶段补充。

已读 control.go / control_test.go：Scanner 上限 16 字节、固定 shutdown 行、EOF / oversize 取消、unknown 行在未 EOF 时不取消、无 token reader 不读取均符合设计。取消 Preflight 的日志与退出语义不得在 cleanup 失败时宣称“已确认清理”；已另提醒 root 收束，真实预检 / ready / 运行中停止仍必须执行。

## 全范围实现首轮与本人可信检查

独立读取后端新的 manifest / multipart / 源码大小与格式限制、注册重载、checker 源快照摘要、内置 n 校验、构造规则、Docker 固定目的地与执行判定拆分、engine 编译顺序 / 10 秒预算 / 原始输出、bounded feedback、示例源码与真实矩阵脚本。该实现没有引入 API / 宿主执行；缺省标准答案创建为空普通文件；checker 的 CE / 非 0 非 1 / 超时 / OOM / 输出限 / attach / cleanup 失效保留 SYS；选手失败不运行 checker，原始输出超过 64 KiB 仍完整进入特判文件。

独立读取 EditorPage / JudgeFeedback / UploadPage / specialJudge / API mode handling：学生与教师均不展示工程阶段耗时、比较器 ID / 容器 / 墙钟说明，student 的工程字段即便被 mock 响应夹入也有 UI guard；本人编译诊断保留；teacher / engineer 可见 bounded 特判反馈；运行证据、特判编译及 hash 仅 engineer 可见。C++ 示例、编辑、下载与发布使用同一源文本，UI 说明发布尚未验证源码、先试正确 / 错误答案，并给出自包含 C++20 的路径参数、退出值、数据配对和配置错误规则。

本轮另外发现前端文件名预检初版允许 `_case.in` / `-case.in`、Windows 保留 stem / COM / LPT 前缀，而后端拒绝，导致教师本地预检成功后收到 generic 400。已通知 student_frontend 按现有后端策略修正，未要求扩大后端范围；其已回复完成，待最终复读与测试。并非未知外部阻塞，按本轮授权作为普通审阅修正记录。

本人实际执行 `node scripts/go.mjs test -count=1 ./internal/problemstore/... ./internal/filejudge/... ./internal/httpapi/... ./services/judge-mvp/...`，4 包 exit 0；执行 `node apps/web/scripts/verify-modes.mjs`，实际 PASS。范围包含 trusted fake / 文件快照 / 序列判定 / 协议 / API 投影 / SSR 模式与配对 / canonical checker 示例一致性，未运行 Docker payload、未操作 UI 或服务。Windows 中既有 symlink 权限 skip 不能登记为已通过 Linux 符号链接证据。

当前没有新增未闭合的核心判题实现必修；私有取消通道最终接线、前端文件名修正复验、真实矩阵 / 上传重启 / UI / 退出清理 / 最终源码绑定尚未齐备，因此不签最终产品 PASS。

## 实现修正复审通过（未作最终冻结）

**当前局部实现及可信测试 PASS，未闭合的代码必修为 0。产品验收 / 最终源码绑定仍待补齐。**

已再次独立复读完整取消接线：owned 子进程 stdin pipe 发送固定 shutdown 行，父进程处理 EPIPE，HTTP 仅补充；每 await 后检查 stopping，stopPromise 唯一，异步 build 不再通过同步 helper 直接退出。Go reader 仅在有 token 时启用，16 字节上限、EOF / 失效取消、提前注册 signal context 均落实。Engine 新增只读 CleanupFailed 返回真实 taint；main 在预检取消后的 cleanup failure 非零退出，正常取消日志要求 supervisor 另行确认，不将 false taint 或退出码当作实际无残留证明。

已复读 special Engine fake 的全部 14 路径都检查 strict validResult / completeWire / MatchesSnapshot；新双用例测试覆盖 AC / RE 全列表、checker SYS 只返回实际前缀；API 依据受信 snapshot 的 checker/source/data 和例子顺序拒绝混入或不完整 AC。前端文件名修正现在与后端首字符、长度、Windows 保留名称 / 前缀严格一致，并补 interactive 前导零输入与 server secretRE 的一致性；canonical 示例一致性保持。

修正后本人再次亲自执行 4 包 trusted Go tests、4 项 Node shutdown helper tests 与前端 verify-modes，全部 exit 0 / PASS。该重复仅针对新增修正。未执行真实 Docker / UI / 服务操作；root 真实三阶段停止、矩阵、上传持久、最终完整检查与源码绑定将作为下一次独立证据审阅输入。
