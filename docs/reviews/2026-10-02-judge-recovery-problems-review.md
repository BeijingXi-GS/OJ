# 判题恢复与多题增量独立审阅

- 日期：2026-10-02；任务 JUDGE-007；独立审阅者：/root/judge_reviewer，Architect / Security。
- 允许修改范围仅本报告；未参与实现、不启动 Docker payload、不操作浏览器或修改主机环境。根 Agent 负责真实 Docker / API / UI 验证，problem_catalog 负责目录与题目数据，frontend_followup 负责前端。
- 审阅输入：AGENTS.md、PROGRESS.md、DEVELOPMENT.md、JUDGE-007 任务卡、PROBLEM_CATALOG_EXTENSION.md、judge-mvp/0.1.0 原局部契约。

**正式结论：本轮局部设计、实现 / 执行边界及本机五题产品验收通过，未解决必修发现为零。** 绑定非 MD 源码 151 文件，SHA256 `a07de43e447cea2b5af370225eafa13018a20cfee4d6c0bf011461f88026761f`；本审阅者独立执行 `node scripts/source-digest.mjs` 得到相同摘要。镜像仍为 `sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74`。真实 payload / HTTP / 浏览器由 root 执行，本审阅者只读实现与证据、执行可信 fake / 文件 oracle / 证据审计。

## 局部设计结论

**多题增量设计通过，允许按裁定实施。** 共享编译期 allowlist、指定 ID 的 metadata、所有已登记题目 fixture 的启动预检以及 Docker 查询错误分类不扩大执行权限，也不改变既有 DTO、资源控制或清理边界。此结论仅为实施前的局部门禁，不代替最终实现 / 产品验收。

实施约束已通知 root 与 problem_catalog：

- List / Get / Exists 是唯一题目 ID 来源，未知 ID 必须先拒绝再访问服务器文件；新增目录本身不能开启题目。
- Metadata 查询仅构造固定服务 URL 的编码路径，API 必须获取所请求 ID 的 metadata，严格校验响应。不得为新题目返回 A+B 的默认值。
- 启动时有界读取每个已登记题目的 manifest / 普通用例文件；任一失败拒绝 ready。运行中配置改变仍遵守原 SYSTEM_ERROR 语义。
- Docker info 查询失败与成功返回非 2 必须区分。失败只输出固定操作 / 分类，不泄露原始 Docker 输出、主机路径、环境值；查询失败仍拒绝执行，禁止放宽五秒控制预算或资源策略。
- 扩展裁定须明确覆盖原契约中“只支持 a-plus-b”的范围限制；其余原契约保持。

## 实现与验证

状态：多题接线 / 执行边界 / startup 诊断代码复读通过，最后窄修正后的严格重启 / AC 与最终源码摘要已复核。旧矩阵和旧代码审阅未被用来代替本轮实例复验。

已只读复审 internal/problems、filejudge 的 contract / cases / docker / engine、httpapi 的 router / problems 与新增 tests，以及 apps/web 的题库 / 编辑器 / DTO / 判定展示和 problem-smoke 脚本。

- 目录包含五题，List / Get 返回独立 Samples 副本，外部调用者无法修改内部 allowlist。请求 / API / 文件 loader 使用同一 Exists；未知 ID 在 filesystem / metadata / compiler / judge 调用前拒绝。
- Metadata Client 只访问配置服务的固定编码路径，严格 JSON / 大小 / 政策校验并核对 problem_id；API handler 本身也拒绝不匹配 ID。配置 judge 却无 MetadataProvider 时详情返回 503，避免默认限额冒充真实配置。旧 worker GET /problem 保持 A+B 兼容。
- Worker startup 对全部登记题目 load，先于通用编译 / 运行 probe。原普通文件 / symlink 拒绝、快照 hash、原始输出 checker、资源配置、容器身份与独立 cleanup / taint 边界保留。
- Docker 控制错误只保存固定操作 / 分类 / 数字退出码，不输出原始 argv、daemon 输出或环境值。cgroup 查询失败与成功非 2 分开，仍拒绝 ready。五秒控制预算、七十秒 job 和十五秒 cleanup 未放宽。
- 前端保留 encoded link、problemId queryKey、按 data.id key 重建 Workspace、卸载 abort、signal 检查、共享 controller 锁、双结果 / stale / reset / retry。修改仅去固定题号并补 artifact_sha256 类型字段，无新依赖。

### 开发期发现与修正

1. 独立复审要求 control 在命令返回 nil 时仍检查 context 取消 / 超期，避免超预算成功竞态。root 已加入 ctx.Err() 条件并补 deadline_success_nil fake 用例；本审阅者执行通过。
2. root 自检发现嵌入 bytes.Buffer 的 WriteString / ReadFrom 方法提升绕过 Write 限额，改成私有 data 和显式 Len / Bytes / Write。本审阅者复读确认 io.WriteString / io.Copy 都经过 Write；输出上限专项 fake 用例通过。
3. 提醒 root 新 metadata 新增 problem_id 后，原 judge-test 的动态 metadata deepEqual 预期需要同步。已复读确认预期包含 A+B ID，本轮真实矩阵适配后全部通过；此为测试脚本适配，没有误报环境阻塞。
4. root 的 Linux 单测脚本需读取新增 fixture；首轮仅挂可信测试二进制导致相对路径缺文件，随后添加测试专用 `/testdata` 只读挂载和 `/tests` workdir。只读复审确认该挂载仅供交叉编译的可信 Go 单测、无 daemon socket，fake Docker command；提交程序容器未获得答案目录。该适配可接受，不是执行边界绕过。
5. 后续复读发现 startup 的 filepath.Abs / MkdirAll / MkdirTemp 分支仍直接返回 OS error，可能经启动日志泄露主机路径。root 已收束这些分支为固定分类，并补 `TestStartupIOErrorDoesNotExposeHostPath`：jobs 路径是普通文件或其子目录均返回固定 jobs_root_create_failed，不包含临时主机路径。本审阅者复读并独立执行专项 fake 单测通过；该项已解决，当前未发生对应真实 IO 故障。

### 本审阅者亲自执行的证据

- Windows / 当前 Go 项目缓存，执行 `node scripts/go.mjs test -count=1 -v ./internal/problems/... ./internal/filejudge/... ./internal/httpapi/...`，exit 0。覆盖 catalog 所有权 / allowlist、五题 fixture / 公开样例答案、metadata 请求与响应绑定 / 协议拒绝、全部题目 startup 拒绝、脱敏诊断分类、超期 nil、既有生命周期 / 输出 / checker / client / HTTP 校验。Docker 调用使用 fake command，未启动真实 payload。`TestCaseSymlinkRejected` 因既有 Windows symlink 权限限制 skip，不称该项本机通过。
- 独立 Node BigInt / 字符串 oracle 只读计算全部五题 24 对 `.in/.out`：A+B 4 例，数组求和 / 最大值 / 回文 / 排序各 5 例。数组长度、数值范围、回文字符与长度及答案 token 均通过，未执行 C++。
- startup 最后窄修正后另执行 `node scripts/go.mjs test -count=1 -run 'TestStartupIOErrorDoesNotExposeHostPath|TestControlErrorClassifiedAndSanitized|TestCgroupQueryFailureDiffersFromUnsupportedVersion' ./internal/filejudge`，exit 0；只含可信构造 / fake Docker 单测，没有 payload。
- frontend_followup 报告的 typecheck / lint / build / prettier 成功属于该 Agent 执行；本审阅者只读代码，不能登记为自行执行 UI。

### root 真实证据

已只读 `.local/evidence/docker-recovery-probe.json`：root 原五秒预算下三次 info 成功返回 2，分别 651 / 423 / 444 ms，退出 0、stderr 0 字节。此为 root 的真实 Docker 只读恢复证据，不是本审阅者执行，也不能还原上轮误归类查询的根因。

已只读 `.local/evidence/problem-smoke.json`：本轮开始 2026-10-02 18:50:52、完成 18:53:26（Asia/Shanghai），completed=true，五题共十次 HTTP200、compile success、对应独立正确解 AC / 常见错误解 WA，全部 passed=true。A+B 每次四例、新题每次五例；错误解至少一例 WA，其余仅 AC 或 WA。审阅脚本明确在每次完成后成功查询 Docker 和 job 目录无残留，再登记 passed。本审阅者未亲自执行这些真实判题。

root 报告本轮 check / smoke / mvp:smoke 通过，Linux 可信单测适配后全部通过含 symlink / 24 对 oracle；这些属于 root 执行结果，不能登记成本审阅者亲自运行。

已只读 `.local/evidence/judge-matrix.json`：本轮 18:55:16 至 18:58:14（Asia/Shanghai），17 项 completed=true / 全部 passed。除了 root 脚本断言，本审阅者另执行只读 Node evidence audit，核对 MLE 各例 OOMKilled=true、非 OOM exit137 各例 RE、展示截断各例仍 AC、CPU / 峰值全部 null、metadata 的 A+B ID 及 exact WA / tokens AC 与 testdata hash 变化，exit 0。该审计仅核对已有证据文件，真实 Docker / C++ / HTTP 仍为 root 执行。

矩阵覆盖 AC / WA / CE 语法与链接 / RE 崩溃与非 OOM137 / TLE / MLE / OLE、展示截断 / 非法 UTF8、子进程清理、独立隔离检查、busy 429 与 API 取消、文件缺失与无效 manifest SYSTEM_ERROR、动态 metadata / checker。脚本每次判题核对成功 daemon 查询无容器和 job 目录为空，故障注入 finally 恢复原文件；对应测试均登记通过。

已阅读 [浏览器验收](../judge/BROWSER_QA_2026-10-02.md)、重存的 `browser-array-sum-ac.jpg` 和 AX 文本、`browser-mobile-controls.jpg` 及 `final-readiness.json`。截图确实展示 AC / 5/5 / 逐例区，AX 包含五例对应 AC；手机双按钮 / 双结果无遮挡，报告中 clientWidth=scrollWidth=375。root 真实操作记录覆盖五题导航 / 详情、编译成功、AC、编辑导致双结果过期、语法错误 CE、恢复模板清双结果、运行中导航取消、缓存返回不串题。报告准确区分等待 / fill / screenshot / confirm 工具适配和实际页面结果，没有声称未执行的取消确认、全面键盘可访问性或每个 verdict 浏览器遍历已通过。本审阅者只读证据，没有自行操作浏览器。

19:09:34 的 readiness 查询显示 Web / API / compile / judge 四服务 HTTP200，compile / judge 已启用，五个题目 ID 均列出，判题容器与 job 目录剩余 0；此为最后 startup 错误脱敏修正前的运行证据。

root 最后受控关闭旧服务并核对无残留，重新编译后 19:15:06 judge 严格 preflight ready、19:15:12 API judge_enabled=true。已只读更新后的 `final-readiness.json` 与 `.local/final-ac-smoke.mjs`：19:17:26 至 19:17:43（Asia/Shanghai）四服务200、五个正确 ID、独立数组和源码 HTTP200 / compile success / AC 5/5，源码 / testdata / artifact hash 均记录，固定镜像一致，成功 daemon 查询与 job 检查均 0 残留，completed=true。此为 root 执行最后版本的真实复验，不是本审阅者执行。

最后差异只收束 startup IO 错误分类，不改变已验证的资源 / mount / 执行 / checker / 清理 profile；专项 fake 检查和最终严格启动 / AC 直接覆盖该差异，无需再次运行全部 17 项矩阵或重做前端交互。整体工程 check / Linux 可信单测最终登记由 root 负责，原已通过检查和本审阅证据不冒称远程 CI。

## 结论边界

本轮支持原 A+B 与四道新增题，完整中文题面及 24 对服务器文件用例，多题 metadata、编译与逐例判题、取消 / 清理和浏览器切题隔离符合局部契约。已有 Docker / 浏览器阻塞具备本轮实际恢复证据；上轮 info 失败的历史根因仍未确认，不能把恢复误写为证实曾使用 cgroup v1。

这是本机文件判题 MVP 的局部验收，不冻结完整核心类型，也不覆盖生产隔离、公网服务、队列 / 持久提交、完整键盘可访问性、恢复确认的取消分支或浏览器逐项遍历所有 verdict。CPU 时间 / 内存峰值仍为 null，未采集值没有伪造。
