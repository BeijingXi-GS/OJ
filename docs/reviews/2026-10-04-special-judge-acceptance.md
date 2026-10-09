# 学生信息与特殊判题：恢复后的独立验收审阅

日期：2026-10-04；独立审阅者：`/root/spj_acceptance`，Architect / Security；任务 SPJ-004。依据 [AGENTS](../../AGENTS.md)、[流程](../../DEVELOPMENT.md)、[契约 special-judge/0.1.0](../judge/SPECIAL_JUDGE_CONTRACT.md) 与 [前日审阅](2026-10-03-special-judge-review.md)。前日局部 PASS 是历史依据，不替代本轮源码和真实产品验收。

本审阅者只读实现，运行可信数学 / 文件快照 / fake 测试，仅写本报告；没有控制 Docker、真实服务、浏览器，也没有运行上传或选手 C++。主执行者的真实测试另以明确的证据来源审阅，不登记为本人执行。

## 本轮实现复读与本人检查

已独立复读 manifest / multipart / 源码限制、不可变文件读取与摘要、构造题规则、C++ 特判编译 / 执行、固定容器参数与清理、Client 严格 JSON / 状态 / 快照绑定、API 分级投影、EditorPage / JudgeFeedback、私有 stdin 取消通道与启动监督器。

- 教师特判仅 batch、固定 checker.cpp、最多 4 例；每例输入必需、同名 .out 可选，缺失答案生成空文件；不允许上传命令、依赖、二进制或任意路径。
- 特判源码进入本次快照摘要，由独立 Judge 在现有受限编译容器中编译，预算实际为 10 秒。执行只挂载本例输入、参考答案、选手完整原始输出及自身产物；无网络、只读、nonroot、capdrop、64 MiB / 32 pids / 2 秒。选手失败优先，特判失败保持 SYSTEM_ERROR。
- Client 递归拒绝重复字段、未知字段与非法 null，并核对可信本地 checker、源码 / 数据摘要及同顺序用例前缀。非 SYSTEM_ERROR 的成功执行须包含所有用例；普通 / 内置 / interactive 拒绝 C++ 特判证据。
- 学生没有逐例数据、checker 反馈 / 编译 / 摘要；学生与教师题面工程说明省略，工程师可看。页面保留输入输出 / 公开样例 / 时间内存限制 / flush 等解题信息与本人编译诊断。
- 启动监督器 stopPromise 唯一、异步 build 与 await 后停止检查；owned Judge 通过受限私有 pipe 在预检前可取消，无 token 实例不读取 stdin。停止成功要求 owned exit0、成功 Docker 查询无残留和 jobs 空，HTTP 失败或关闭监听本身不会变为成功。

本人本轮实际执行：`node scripts/go.mjs test -count=1 ./internal/problemstore/... ./internal/filejudge/... ./internal/httpapi/... ./services/judge-mvp/...`，4 包 exit 0；`node --test scripts/judge-shutdown.test.mjs`，4 项 PASS；`node apps/web/scripts/verify-modes.mjs`，PASS。均为可信本地 / fake 检查；Windows 符号链接权限 skip 不记为 Linux 检查通过。

当前独立复读未发现新增核心实现必修。真实矩阵、运行中 / 预检 / 空闲停止、上传持久、页面、完整 check、最终源码绑定仍待主执行者证据，本报告当前不签最终产品 PASS。

## 恢复验收中的脚本发现

独立读取主执行者首轮真实矩阵日志，发现 `custom-checker-reads-input-answer-output` 在合法无反馈的第一例调用可选 `checker_feedback.includes`，触发 TypeError，未进入后续坏特判矩阵。已通知 root 按可选字段处理。这是验收脚本问题，不是核心判题故障；首轮报告 `completed:false`、13 项已记录且源码已恢复，不能冒称完整矩阵 PASS。修正与真实重跑结论仍待证据。

独立复读 [使用指南](../judge/SPECIAL_JUDGE_USER_GUIDE.md)，内置数学条件、多个正确示例、四种 batch 判定 / 数据配对、C++ 固定路径参数 / 退出值 / 限额、三种模式和当前真实验收 pending 的表述与契约和实现一致，没有文档必修。

## 已复验的真实矩阵与停止观测修正

root 修正可选 feedback 断言，并补 teacher / student 判定均须 AC。本 reviewer 复读其固定版日志与 `.local/evidence/special-judge-matrix.json`：2026-10-04 16:28:08–16:32:38（Asia/Shanghai），exit0、completed:true、23 项记录、源码 finally 恢复为 true。每项成功查询清理；内置两种不同合法排列与 C++ 两种答案均 AC，约束错误 / 重复 / 错报无解 WA，70,000 个前导空格的原始输出在页面截断后仍被 C++ 特判接受。特判 CE、exit2、无 OOM 的 exit137、超时、真实 OOM、输出超限均 SYSTEM_ERROR；选手 CE / RE / TLE 优先，checker 不运行；三文件参数读取、数据 / 源码摘要变化与恢复、teacher/student 投影以及旧 batch / interactive 回归通过。真实执行者为 root，审阅者仅独立读取这些证据。

独立复读新版 `supervisor-stop-test.mjs`，exit 与 stdio close 分开等待，保留首先发生的阶段错误，最后分别记录退出、资源和端口事实；每轮新报告保留历史失败。资源查询必须成功且为空、jobs 空，端口必须真实 ECONNREFUSED；没有强杀 Judge 或手工删 jobs。本人亲自运行 `node scripts/supervisor-stop-test.mjs --self-test`，动态导入 / 信号、提前失败、exit 先于 pipe 读取、原错保留四类可信进程测试 PASS。这证明测试观测修正，不替代真实 preflight / idle / running 停止。

上述局部证据已复验。完整 check、真实三阶段停止、教师下载页面修正 / UI、上传重启与最终版本绑定尚未全部齐备，最终产品仍 pending。

另独立复读 root 分配给 lifecycle_fix 的 UploadPage 下载窄修：标准常驻 `a download="checker.cpp"`，按当前有效源码创建 UTF-8 Blob；layout effect 在 paint 前替换 href，源码变化 / 禁用 / 卸载时释放旧 URL。空 / NUL / 超 64 KiB、正在读文件 / 发布、非自定义题均不可下载；下载与发布继续使用同一 source。代码复审没有新增必修，真实浏览器下载待 root 实测。旧临时 anchor 的 IAB 10 / 30 秒事件等待超时是观测事实，未独立确认其根因，不能把代码修正推测写成已查明浏览器故障原因。

对另获 root 窄授权的 `EditorPage` 兼容修正，独立读取锁定 Monaco 包的公开类型与注册 / compute 实现：`editContext:false` 是支持的布尔配置，选择现有 textarea 输入路径，没有换依赖或修改其他编辑器选项。本人在两处前端修正后再次执行 `verify-modes.mjs`，PASS；真实输入和提交仍须 root 页面验证，不把 IAB 的旧编辑观测失败泛化为所有用户无法输入。

## 再次外部中断后的停止结论

root 通知再次发生自动审批 `stream disconnected / response decode` 通信错误，申请的真实停止重跑动作明确未执行。依用户规则，reviewer 已停止所有新增检查与实现，仅收束报告；这不是对动作安全性的否决。最终产品结论保留 **pending / 未通过**，不签最终 PASS，也不把本轮已通过的局部真实矩阵降格或冒称完整验收。

为准确记录已经发起的操作结果，本 reviewer 只读复核 `.local/evidence/spj-supervisor-stop-2026-10-04T08-48-01-624Z.json`：16:48:01–16:48:47（Asia/Shanghai）仅进入 preflight 观测一项，`stage_reached:false`，日志只有 bootstrap；阶段等待超时，另缺少 owned Judge 确认日志。owned Node exit 和 stdio close 均为 code0 / signal:null；独立 Docker / jobs 查询及端口关闭核实成功，但这不证明 Judge 真实预检、空闲与 Running 三阶段均达到或停止通过。报告 `completed:false`，idle / running 未执行。

另外只读复核 root 本轮完整 check 的最终日志（Go / 前端 / helper / SSR 检查与构建通过，既有 Monaco chunk 非阻塞警告保留）、Linux 单测日志（两个符号链接检查实际 PASS）及两项 smoke 日志（基础 API 和实际 GCC 四种编译状态通过）。这些检查与 23 项矩阵均发生在随后 mvp-dev 启动观测日志和最新生命周期测试脚本变更之前；不能把它们登记为最终停止源码的全套检查通过。root 新增启动阶段日志后，下一次真实重跑在执行前遭审批通信中断，没有新真实结果。

下载 / editContext 窄修代码复审通过，真实页面下载 / 编辑复验、完整三阶段停止、上传重启后的文件摘要与可判验证、最终在线就绪均未齐备。现在没有新增已确认核心代码必修；当前停止位置与未完成范围仍阻止产品验收。

停止期间仅复核已发起页面操作的后果：[部分页面 QA](../judge/BROWSER_SPECIAL_QA_2026-10-04.md)及保存 DOM 证明 teacher 真实发布题并得到 AC2/2、无解输入3 / 输出-1与 bounded 反馈。源码常驻下载 / 传统输入已显示不等于完成实际下载或编辑提交；[截图请求失败](../issues/2026-10-04-browser-screenshot.md)没有有效图片。缓存页面也不证明已停服务在线。root 16:56记录 Web / API / Judge 不可访问、Compiler HTTP200、jobs0；审批失败后没有新的成功 Docker 查询，本 reviewer 不把旧 owner 空查询写成现在的实时事实。

本报告最终停止绑定：root 实际计算 **200 个非 MD 文件，SHA256 `3406da9bef0efd1719a405bd48c2cc7206609368f0329f8012c5b6280bcf8df3`**；锁定镜像 `sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74`，参考 commit `4502d2a798f3130d3c540c5c31d45251ca103623`。收到停止命令后没有复跑摘要工具或其他测试；摘要是 root 提供的停止记录，**不是通过冻结或最终版本全量检查的证明**。

独立最终结论：**已复读范围的局部实现 / 可信测试与 23 项真实矩阵证据 PASS；最终产品 pending，未授予最终 PASS，SPJ 增量不能登记全部 done。** 具体外部原因见 [本轮审批问题](../issues/2026-10-04-approval-resume-interruption.md)；适用已测 / 未测边界见 [验证记录](../judge/SPECIAL_JUDGE_VALIDATION_2026-10-04.md)，恢复顺序见 [当前交接](../handoffs/2026-10-04-special-judge-paused.md)。恢复后保留本报告历史失败与实际通过，重新补齐当前源码对应的三阶段停止、真实下载 / 编辑 / 分模式页面、上传重启摘要和可判、完整 check 与在线就绪，再独立签署。
