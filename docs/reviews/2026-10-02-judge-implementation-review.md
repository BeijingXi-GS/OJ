# JUDGE-004 独立实现与安全审阅

- 日期：2026-10-02（Asia/Shanghai）。
- Reviewer：`/root/mvp_review`，未参与实现；职责 Architect / Security Review。
- 范围：本机同步文件判题 `judge-mvp/0.1.0`，不替代完整 Phase1 或公网生产安全验收。
- 当前结论：**已读实现的代码与本机执行边界审阅通过；最终产品验收 blocked。** 首轮 R1–R5 已修正并复读，独立 fake 单测通过，修正版17项真实矩阵证据已核对。最终重新启动出现新的 preflight 阻塞，按用户规则停止；不能将此前矩阵通过登记为当前服务启动成功或产品 done。
- 当前独立重算源码摘要：100文件，SHA256 `674d70a785de4093b7912cdf500bf3df10e2be1e888925ef47a8fa52f3dd36b2`。摘要不含MD /docs；后续新增验证脚本或实现改动须记录新摘要并核对影响范围。

## 已审阅范围与证据来源

已读取最新契约、JUDGE-004、PROGRESS，以及 internal/filejudge 的 contract / cases / docker / engine、services/judge-mvp、services/judge-compile-helper、deploy/judge/Dockerfile、apps/api/main.go、internal/httpapi/router.go / problems.go、共用 compilation.Decode、scripts/judge.mjs / mvp-dev.mjs、apps/web 的 API DTO / EditorPage / JudgeFeedback / App / utils。

root 报告真实预检已 ready：镜像无继承卷、实际受限编译导出、含逗号与空格路径的快照挂载 / exec、runtime kernel probe。该证据来自实施者；本 reviewer 按本次委派约束未运行 Docker payload、浏览器或修改应用实现。启动 probe 的空占位检查曾被修正，详细问题由 root 另档记录；本报告不将最初失败掩盖为首次通过。

root 确认修正后控制服务于11:41:13 ready、API于11:41:15 ready（Asia/Shanghai），用于下述修正版矩阵；最终启动仍会重新编译API。

独立执行以下已读过的 fake 单测，不调用Docker或执行payload，exit0：

```text
node scripts/go.mjs test -count=1 -run '^(TestCheckerAndDisplay|TestClientProtocol|TestVerdictEvidence|TestMountCSV|TestSharedOutputBudget|TestStageLifecycle|TestStartupRefusesInvalidProfile|TestCaseSnapshot|TestRejectedCaseFiles|TestCaseSymlinkRejected|TestJudgeAPIValidation)$' ./internal/filejudge/... ./internal/httpapi/...
ok oj.local/oj/internal/filejudge 0.380s
ok oj.local/oj/internal/httpapi 0.188s
```

另独立检查符号链接测试的细项：本机 Windows 不允许创建 symlink，TestCaseSymlinkRejected 实际 SKIP，不登记该项在Windows通过。root 使用 [Linux可信单测脚本](../../scripts/judge-unit-linux.mjs) 补验：交叉编译同一Go测试，以UID10002 /无网络 /无cap /NNP /只读容器运行，不挂Docker socket；root报告该项实际PASS且整包通过。已阅读脚本及 [权限与补验记录](../issues/2026-10-02-windows-symlink-unit.md)，未自行重复Docker执行。

独立读取 `.local/evidence/judge-matrix.json`：2026-10-02T03:42:44.085Z至03:44:41.744Z，completed=true，17项均passed=true，实际镜像 `sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74`。已核对AC /WA /CE语法与链接 /RE崩溃与137非OOM /TLE /MLE /OLE、展示截断AC、无效UTF-8 WA、fork后代清理、隔离事实、busy与API取消、缺失input与非法manifest SYSTEM_ERROR、动态元数据和exact WA /tokens AC。MLE四例均有OOM证据；137四例无OOM；CPU /peak为null。读取证据文件是独立核对实施者结果，不宣称本 reviewer亲自执行了这些payload。

## 首轮事实与 D1–D6 对照

| 约束               | 实际源码事实                                                                                                                                                                                | 首轮判断                                   |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------ |
| D1 编译 / 运行隔离 | createArgs 共用 UID10002、init、network none、cap-drop、NNP、私有 namespace、只读 rootfs、资源上限和 log-driver none；scratch 最终镜像避免继承 VOLUME；启动拒绝声明卷；实际 mount 禁 volume | 主路径符合设计；真实专项证据待齐全         |
| D2 不可变文件快照  | OpenRoot + Lstat / 打开后 Stat + SameFile；严格 manifest、唯一 case ID、有界读取；source / input 写本 run；expected 不挂载；CSV mount 参数；产物仅只读单文件挂载                            | 主路径符合设计；文件配置元数据一致性见 R4  |
| D3 完整输出与展示  | 原始 stdout 保留到停止线，stdout / stderr共享计数；checker bytes.Equal / bytes.Fields；展示 ToValidUTF8 后按合法边界截断                                                                    | 算法符合设计，仍需对应边界测试             |
| D4 有界监督 / 清理 | job70s、控制命令5s、独立 cleanup15s、HTTP client110s / Write120s；随机 name / label 在 create前确定；失败也成功 list 后清理；不把 inspect 错误当不存在                                      | 最后清理时的 deadline / cancel 判定存在 R1 |
| D5 证据与优先级    | 最终 inspect 状态判定，OOMKilled 支持 MLE，137本身不当 MLE；已知监督原因不被 attach 中止覆盖；清理失败 taint；重启拒绝残留；shutdown等待请求                                                | 主路径正确；响应语义验证存在 R2 / R3       |
| D6 本次 artifact   | helper固定g++ argv、stdout仅导出产物、stderr诊断，Lstat普通非空文件及SameFile；控制服务O_EXCL写固定随机目录文件、封存0555、hash绑定结果；异常helper退出125，正常g++exit1为CE                | 主路径符合；编译fsize与artifact上限见 R5   |

API 的生产接线使用 filejudge.NewClient，路由仅输入校验与HTTP委派，不调用 Engine、Docker CLI或生成产物。未来 idle judge-worker 保留；旧 submissions 未添加业务处理。main的关闭预算已扩展为120s。

前端使用 controller ref 同步防止重复提交，编译 / judge 共用同一请求锁；两类结果分开，记录提交时 source 并标过期；模板恢复先确认；离页 abort；nullable资源显示未采集；diagnostics / stdout / stderr以 React文本呈现。未发现该轮阅读中前端互斥或源码过期的确定性缺陷，但未执行交互验证。

## 开发期发现及修订复核

### R1 / P1：最后 cleanup 中的取消或总deadline仍可能返回成功

docker.go 的 stage 在 attach 后检查 job.Err，随后调用独立 cleanup；cleanup 成功后不再检查 job。最后 case 或最终 removeJob 期间，caller取消或总70s到期可能仍返回 AC，违反 job_timeout / request_cancelled 优先 SYSTEM_ERROR 的契约。

要求：独立清理继续完成；清理成功后与最终返回前重新核对 job ctx，总体归 SYSTEM_ERROR并保留已观察到的编译 / case事实。测试须使取消或deadline发生在cleanup阶段，而不是仅发生在attach。

复核：已关闭。stage清理后重新检查job；Judge最终defer在目录清理完成后核对ctx，总体SYSTEM_ERROR，taint失败仍优先。TestStageLifecycle涵盖cancel_in_cleanup /deadline_in_cleanup，本 reviewer独立执行通过。

### R2 / P1：响应校验允许自相矛盾的成功结果

contract.go 的 validResult 只核对部分枚举 / 大小 / hash / 聚合结果。首轮版本可接受 AC case 的 exit_code=null、非0或oom_killed=true；没有全面验证必有字段是否存在；compile.success 不能与总CE匹配等语义也需要检查。

要求：对必有字段和可空字段分别核对，AC / WA必须可信非OOM exit0，RE必须非零退出且无OOM，MLE需OOM证据，case不能CE；compile状态与cases /总verdict一致；禁止SYSTEM_ERROR后继续cases。纳入缺字段、非法类型、矛盾状态的协议测试，不能只覆盖有枚举名称的正常DTO。

复核：已关闭。completeWire逐层核对必有字段 /不可null项；validResult拒绝AC无退出码 /非零 /OOM、success+CE等矛盾组合，并拒绝SYSTEM_ERROR后继续cases。已复读相应修改及protocol测试；独立选定测试通过，最后停止case的补充由root报告测试通过。

### R3 / P1：目录清理覆盖CE后被client误转为协议错误

engine.go 的 removeJob defer失败将总verdict改SYSTEM_ERROR并taint，但原Compile.Status仍compile_error。首轮validResult强制compile_error只能总CE，因此真正“清理优先”的结果在API桥接被转为502，丢失已获得的编译阶段事实和清理原因。

要求：选择并写明保留compile_error事实、总SYSTEM_ERROR，或统一改compile.status为system_error；保证worker /client约定一致，覆盖目录清理失败覆盖CE的单测。

复核：已关闭。保留compile_error原始事实，允许总SYSTEM_ERROR且cases空 /artifact空；cleanup_overrides_CE协议测试已读且独立执行通过。

### R4 / P2：公开固定元数据与可接受manifest参数不一致

problem详情固定4例 /2000ms /64MiB /tokens-v1，但 Load 接受1..8例 /500..5000ms /32..256MiB /exact-v1。当前内置fixtures一致；合法编辑manifest会造成页面与实际运行限额不同。

要求：若本机MVP公开题目只支持固定配置，preflight明确核对该配置并文档说明；若宣称manifest可配置，API应取得可信worker元数据，不继续硬编码。checker专项测试可以独立验证exact-v1，不能把当前公开metadata硬编码当作测试数据永远不变的证据。

复核：已关闭。worker GET /problem读取合法manifest快照，API的实际Client实现MetadataProvider并转发可信元数据；5s调用 /4KiB响应及枚举 /取值范围校验，失败不捏造固定元数据。真实矩阵已读取1例 /3500ms /32MiB /exact配置和exact WA /tokens AC的证据；root确认finally恢复原fixtures。

### R5 / P2：编译fsize低于声明artifact上限

createArgs将编译与运行都设fsize16MiB，而helper /协议允许artifact32MiB。超出较低文件限额可能使合法源码链接失败并落CE，实际上是编译资源预算触发。

要求：按阶段设编译fsize32MiB以匹配artifact上限，或明确更低编译文件限额并正确区分资源失败与源码CE。仍保持tmpfs /memory /输出等独立限额。

复核：已关闭。createArgs按Compile阶段将fsize改32MiB，运行仍16MiB；helper仍在导出前检查普通 /非空 /≤32MiB，异常返回125。最终隔离矩阵使用修正后服务。

## 待最终结论的检查

root 修正已复读，源码摘要 /镜像ID /profile与17项矩阵已登记。源码行为和所述本机边界满足已采用的局部契约，未发现本轮阅读后仍未解决的阻塞代码缺陷。未知CPU /峰值、默认Docker seccomp与Anolix自定义名单的区别、编译镜像的供应链限制和本机服务边界仍保留。

本轮严格OCI隔离审阅覆盖新 /judgements 的编译与运行。旧 /compilations 继续走原 compile-worker；原Compose没有此次judge的完整memory /pids /只读 /网络profile。本次保留旧接口兼容，不把新judge的资源和隔离证据扩大为旧worker也已升级。

## 最终停止与产品验收状态

root 于收束阶段通知新的实际启动阻塞，按用户约定本 reviewer 已停止新增测试 /源码探索 /诊断，未重试或弱化preflight，也未尝试替代浏览器。

root提供的已完成结果：最终npm run check exit0、基础smoke及mvp:smoke exit0；Linux可信单测包含symlink与新增preflight拒绝用例，exit0；真实声明卷镜像被preflight拒绝，exit1 /reason=image_volumes_forbidden，随后profile已恢复；17项真实矩阵completed且均passed；旧supervisor通过token受控关闭，judge正常退出，容器及job目录无残留。最后声明卷拒绝 /关闭验证与最终check是实施者提供的证据，本 reviewer依停止要求不追加独立操作。

最终源摘要由root提供：100文件，SHA256 `3ee10a98dbff54fa0e567a6a18161a3b38aa67bab8e6190f9643fddb42bb7357`。本 reviewer独立重算的是上述较早的 `674d70a7...` 版本；停止后没有再次读取变动源码或重算最终摘要。对本轮已读实现及R1–R5修正的审阅结论保持有效，不冒称已独立核验该最终摘要下的每个文件。

最终重新编译启动于11:54:40（Asia/Shanghai）返回 `cgroup_v2_required`，judge exit1、supervisor失败，未启动新的API /Web。随后最终AC smoke连接8080得到ECONNREFUSED，**未执行本次最终AC验证**。目前未知是Docker控制命令5s失败 /超时，还是info实际返回非2；不能仅凭该通用错误断言主机已改为cgroup v1。详细前因后果见root独立归档的 [最终preflight问题](../issues/2026-10-02-judge-final-preflight.md)。

浏览器本轮只读inventory仍tabs=[]，尚未创建页面复验；旧B-MVP-001不能关闭。题面 /Monaco /运行按钮 /成功失败 /逐例 /重置 /过期 /移动布局的浏览器验收仍未完成，不以HTTP或构建通过替代。

当前隔离缺项即拒绝运行的行为符合设计，但产品可用性仍blocked。用户处理并通知继续后，需要复验该preflight阻塞，重新正常启动、完成最终真实AC与浏览器验收，再由root记录产品结论。若受影响执行代码改变，应补相应独立审阅。当前不宣称公网安全、完整Phase1、真实CPU /峰值计量、Windows symlink测试或远程CI已通过。
