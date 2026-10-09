# 恶意代码边界与安全缺口独立审阅

日期：2026-10-07 开始，2026-10-08 收尾。任务：SEC-001 / PLAN-002。实际审阅者：`/root/security_gap_review`，独立 Architect / Security reviewer。范围：当前执行、编译、模式投影、教师上传与通信边界的只读审阅；只写本报告，不修改实现。

## 结论与证据范围

**当前适合作为受信任操作者使用的本机演示原型，不能据此签署多用户或公网恶意提交安全验收。** 主要缺口是旧的“仅编译”入口尚未采用逐任务受限容器、模式切换没有真实身份鉴权，以及容器配置核验和服务级资源配额尚不完整。下列 P0 表示“扩大到不可信多用户之前的门禁”，不表示已经发生入侵，也不要求停止本轮纯文档工作。

已读 AGENTS、PROGRESS、DEVELOPMENT 及判题、特判、通信局部契约，独立读取本报告所链接源码与相应测试。未运行 Docker、GitHub 操作、攻击载荷或任何 C++；未重新执行单元测试、真实容器测试或完整检查。本轮没有漏洞利用成功、宿主逃逸成功或运行时安全 PASS 的证据。

[10-07 恢复复审](2026-10-07-recovery-review.md)记录的本地工程检查及 250 文件摘要属于此前版本证据，本 reviewer 读取该报告，不冒称自己重新计算全部摘要或重新执行检查。真实 Docker 及产品验收仍按 [PROGRESS](../../PROGRESS.md)保持 pending。安全整改设计、模拟传输测试、真实 Linux 证据必须分别登记。

## 已有防护及适用范围

| 边界                    | 静态代码事实                                                                                                                                                     | 不能据此推出的结论                                                         |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| API 与执行分离          | API 调用 compiler/judge 服务，不直接启动选手程序；API 默认监听 loopback                                                                                          | loopback 不是账号权限，也不能阻止同机进程直接调用 worker                   |
| 新 filejudge 编译和运行 | 创建 UID 10002、只读 rootfs、无网络、cap-drop ALL、no-new-privileges、私有 IPC/cgroup namespace、PID/CPU/memory/swap/临时文件/输出限制的容器；固定入口及镜像摘要 | 这套约束没有自动覆盖旧 `compilerworker`；容器也不是内核零漏洞保证          |
| 启动预检                | 检查 Linux amd64、cgroup v2、无声明卷和本项目遗留；真实 probe 要求能力、NoNewPrivs、Seccomp 模式、cgroup 值、网络和挂载隔离符合预期，失败不监听服务              | 当前环境尚未重跑；`Seccomp: 2` 不是完整 syscall 规则版本或兼容性证明       |
| 结果和清理              | 启动失败、控制证据不可信、未知 I/O 错误进入系统错误；独立清理预算，不能确认清理就 taint 并拒绝新任务                                                             | 只有 fake 清理通过不足以证明真实后代进程、容器和宿主工作目录均已收回       |
| 学生投影                | `projectResult` 用服务端允许字段列表剔除隐藏用例、答案、checker 和通信 trace                                                                                     | 客户端可改 `X-OJ-Mode`，所以是演示模式投影，不是防学生越权                 |
| 上传                    | multipart 请求 1 MiB；文件 64 KiB；固定文件数/后缀/格式；拒绝额外路径字段，生成题目 ID，目录内受限读取，普通文件与打开后的身份核对                               | 单次限制不等于总存储配额；上传成功不等于 checker 正确或已运行验证          |
| 页面输出                | 已读题面和结果组件通过 React 文本节点 / `pre` 展示，无 `dangerouslySetInnerHTML` 路径                                                                            | 将来增加 Markdown/富文本/终端解释器时需要重新审阅；本次没有浏览器 XSS 实测 |

代码入口：[容器配置与清理](../../internal/filejudge/docker.go)、[预检与生命周期](../../internal/filejudge/engine.go)、[模式投影](../../internal/httpapi/modes.go)、[上传解析](../../internal/problemstore/upload.go)、[题库文件操作](../../internal/problemstore/store.go)、[输出组件](../../apps/web/src/JudgeFeedback.tsx)。

## G-01：仅编译路径的执行边界低于新判题路径

**优先级：P0，多用户前必须完成。事实成立，未尝试利用。**

证据：[compilerworker/worker.go](../../internal/compilerworker/worker.go)的 `Compile` 在常驻服务进程所在环境调用固定参数 `g++`，有单并发、15 秒 deadline、有界诊断和临时目录。[Linux 取消](../../internal/compilerworker/process_linux.go)通过进程组 SIGKILL 与 WaitDelay 结束编译；`defer os.RemoveAll(dir)` 未检查删除结果。[compiler compose](../../deploy/compiler/compose.yaml)只设 loopback 端口、init、重启策略，没有逐任务 memory、PID、CPU、只读 rootfs、禁网、cap-drop 或 no-new-privileges 设置。[镜像](../../deploy/compiler/Dockerfile)设非 root UID，这是一项保护，但不是逐任务资源和文件隔离。

恶意模板实例化、预处理、巨大编译产物及编译器缺陷不需要运行产物即可消耗资源或影响编译环境。现有 deadline 不能代替内存/PID/文件上限；进程组取消不是独立容器销毁证明。这里没有发现“API 在宿主启动 g++”，也不能把常驻编译服务的限制误写为宿主完全无约束执行。

整改：保留旧 HTTP 响应兼容适配，将仅编译请求路由到受审的逐任务 compile stage；复用 source 快照、固定镜像/参数、编译预算、产物导出及清理验证。不得在 Docker 不可用时回退到宿主或旧无同等限制路径。旧服务退出默认启动链，或只作为显式本机兼容功能并禁止承担不可信提交。

验收：编译成功/CE 的旧 UI 兼容；编译内存、时间、产物、诊断超限及取消均有明确类别；cleanup 失败拒绝后续请求；模拟断言证明每条编译入口都经过统一调度。真实 Linux 专项绑定镜像和源码摘要，确认失败后无本 Job 容器、进程及工作目录残留。测试载荷只能进入专用隔离环境，不能用当前旧服务做破坏性压力验证。

## G-02：模式头和内部 worker 入口没有身份边界

**优先级：P0，在本机演示之外开放前必须完成。当前是明确授权的产品限制。**

证据：[modes.go](../../internal/httpapi/modes.go)直接采用 `X-OJ-Mode`；教师上传只检查该模式和 Origin 允许列表，允许空 Origin；工程师调试也依赖该头。客户端可以合法切换，符合用户“仅本机演示”选择。[engine.Handler](../../internal/filejudge/engine.go)的 `/judge`、`/internal/communication-judge` 直接返回完整 Result，没有调用方凭据检查。[judge 服务](../../services/judge-mvp/main.go)固定监听 `127.0.0.1:8082`，只有 `/shutdown` 使用单独控制 token。`/shutdown` token 不保护判题接口。

整改：上线方案必须把 UI 模式和服务端身份分离；会话/角色从服务端取得，读题目、读提交、教师发布、工程师日志均做对象级授权。worker 不成为学生可达公共入口；采用受控内部连接及独立调用凭据或等价的本机访问控制。教师只能管理自己授权范围内的题目；工程师数据与密钥分离。保留本机模式时必须显式标注 demo，不能把角色头接到生产权限分支。

验收：修改模式头、缺少/过期身份、跨用户 submission ID、跨教师 problem ID 都不能扩大数据范围或写权限；直连 worker 被网络层或服务端身份边界拒绝；上传 CSRF/跨站来源与会话策略一致。日志不能记录 token。学生投影测试继续存在，但不能单独作为授权验收。

## G-03：容器创建配置与检查闭环不完整

**优先级：P0，完成沙箱正式验收前补齐。是核验缺口，未证实用户可绕过 Docker 创建参数。**

证据：[docker.go](../../internal/filejudge/docker.go)的 `createArgs` 设置 tmpfs、fsize/nofile/core ulimit 和 shm 大小；`verifyContainer` 核对 UID、镜像、主要 namespace、memory/swap/PID/CPU、cap、nnp 等，但 bind 只核对 destination、只读和数量，不比对实际 source；所有 `Type=tmpfs` 条目均可通过，未核对路径、选项、大小。结构中已有 `Tmpfs` 字段但校验未使用；没有核验 ulimit/shm 等实际返回值。启动 probe 验证若干内核值，是有用的第二层证据，但不是每个 stage 完整配置核对。

整改：建立一个版本化有效 profile，明确创建参数、inspect 归一化规则和内核探针分别验证什么。核对完整 bind 集合及源文件身份映射、tmpfs 集合/容量/执行属性、shm、ulimit、entrypoint/args 及必要 stage 身份。Docker Desktop 可能将 Windows bind source 映射为 Linux 路径，不能简单字符串比较后退化为忽略 source；须明确平台映射证明，无法证明时拒绝启动。默认 seccomp 和 LSM 需要记录可用性/规则来源，不能由 `Seccomp: 2` 推出任意规则都符合要求。

验收：分别篡改 source、额外 tmpfs、无容量 tmpfs、错误 noexec、shm/ulimit、入口和参数的模拟 inspect 返回，证明在启动选手程序前拒绝，并保留清理/taint 行为；正向兼容 Windows Desktop 与目标 Linux 的路径表示。真实 probe 除无网/只读/cgroup 外验证有效挂载及资源值，缺关键能力不能静默降级。

## G-04：单任务限额尚未形成服务级配额

**优先级：P1；开放持续多用户提交前提升为门禁。**

证据：[Store.Create](../../internal/problemstore/store.go)在每个合法上传后持久新增目录，未见总存储、教师题数、速率或磁盘低水位检查。单次上传上限不限制累计增长。[Engine](../../internal/filejudge/engine.go)和旧 compiler 分别以单并发拒绝忙请求；两个服务的容量池独立，API 解析、存储和其他控制开销不被选手容器 cgroup 自动覆盖。通信每角色内存限制与结构上的两倍总上界不等于已采集的整个 JudgeRun 实际峰值。

整改：基于统一 admission control 明确节点总预算、编译与运行配额、用户/课程配额、上传总量、并发请求体数、队列长度、磁盘与 inode 低水位、超载拒绝及恢复条件。配额预留和发布必须原子化，失败释放预留；不能以自动删除未知教师数据腾空间。记录控制面自身资源预算，作业聚合指标不能相加各进程在不同时刻的峰值。

验收：有限规模的并发模拟及限额注入证明请求可控拒绝、不会超售、不会部分发布；低磁盘模拟而非填满宿主；重启后恢复配额账本；容量恢复后有界重试，不造成重试风暴。真实节点容量测试单独规划与审批，不能复用正常办公机器做耗尽实验。

## G-05：自定义 checker 的执行安全不等于判题正确

**优先级：P1；正式教师发布流程必需。**

已有实现：[special.go](../../internal/filejudge/special.go)给 checker 独立受限 stage（64 MiB、32 PID、2 秒），只读挂入 input、answer、output 与自身产物；exit 0/1 对应 AC/WA，其他退出、OOM、超时、控制失败进入 SYSTEM_ERROR，避免将判题器失败计入学生错误。教师反馈 1024 字节，学生无 checker 输出。[上传规范](../../internal/problemstore/manifest.go)限制固定 checker 文件与协议；当前交互题只接受内置 `guess-number-v1`，并不支持上传任意 interactor，不能把未来能力当既有入口。

缺口与整改：隔离无法发现“永远通过”的 checker，也不能保证规则覆盖完整。上传后形成草稿版本，绑定 checker/testdata 哈希；在独立容器内用标准解、多个合法构造、已知错误解、格式边界样例验证后才允许发布。教师发布权纳入对象授权与审计。已知错误解集通过不代表数学完备，只是有限回归证据。将来允许上传 interactor/generator 时同样视为不可信计算，并追加私密输入流向和协议预算评审。

验收：错误 checker 的 CE/超时/OOM/异常退出不扣学生成绩；永远 AC 的 checker 在已知错误解门禁被拦截；两种合法构造都能通过；发布版本和重判版本可追溯；学生接口和页面看不到 checker 源码、私密反馈和隐藏用例。

## G-06：真实隔离与生命周期证据尚未齐备

**优先级：P0，正式运行验收门禁；不是新出现的漏洞。**

已有 [communication.go](../../internal/filejudge/communication.go) 的 relay 在固定总字节预算内缓存/转发、EOF 关闭、共享截止时间、输出停止线、两个角色清理、未知传输错误与 taint 方面有明确代码处理。receiver 没有 sender 输入或答案挂载，sender 只拿自身输入。编译、运行、checker 和交互支持失败路径模拟测试，但当前主机 Docker 连接阻塞，没有新版本真实通信矩阵及恶意行为边界证据。

整改与验收：恢复正常授权环境后，在专用 Linux 环境分层验证资源拒绝、双角色 EOF/提前退出/不读输入、共享超时、通信超额、输出超额、OOM、取消、启动不确定、daemon 断开与清理不确定。所有用例都要求资源证据、判定优先级、后代/容器/目录清理和 worker 是否接受下一任务的联合断言。禁止仅用 exit 137 推断 OOM；checker OOM 仍是系统错误；监督器终止 peer 不能伪装成自然 RE。缺少清理证据必须保持故障和 taint，而不是把“端口已断开”登记为零残留。

## 顺序与文档门禁

先完成 G-01 / G-03 的受审设计与实现，然后绑定当前源码进行真实执行验收；G-02 / G-04 在多用户化前完成，G-05 与教师质检流程一起实施。此顺序允许当前本地原型保留人工模式，不要求本轮擅自加真实登录。

威胁方案应明确覆盖编译耗尽、CPU/睡眠、进程/线程、内存/swap、stdout/stderr、文件/FD/inode、越权读取与 symlink、网络外联、内核/运行时缺陷、checker、通信死锁及预算、上传累计增长、API 越权、前端渲染、供应链与审计。采取哪种隔离后端都不能写成覆盖“所有可能的恶意代码”；需要声明可信计算基、残余风险、补丁与应急撤销路径。

本报告只完成独立静态缺口识别与整改/验收建议。源码整改、真实负面用例、正式安全签署均未完成；下节主方案的文档审阅结论不与实现或运行验收混淆。

## 2026-10-08 主方案与工作包准确性复核

已独立读取 [恶意代码安全方案](../sandbox/MALICIOUS_CODE_SECURITY_PLAN.md)全文及 [后续开发工作包](../planning/2026-10-07-implementation-backlog.md)全文。安全方案将 22 类威胁与 SEC-002..006、12 组待执行验收对应；工作包将持久模型、后端、前端、异步恢复、教师质检、公开通信回放拆分并列出依赖和验收。

**实质内容可作为后续受审设计与实施的工作输入。** 两份文档明确区分当前本机演示、已有源码保护、计划整改和真实环境待验；正确保留旧 compile-only、模式头及直连 worker、profile 校验、总配额、checker 正确性等缺口。未把 gVisor、真实账号、父 cgroup 计量、任意 interactor 或异步恢复写成已经实现。前端任务沿用现有工程，公开 DTO、隐藏数据、双角色绑定、历史版本和缓存清理责任明确。

收尾提出两项精确措辞调整，root 已处理；reviewer 随后只读确认修订落盘，两项均已关闭：

1. 工作包末段已同步 10-08 授权正常重试记录：Git 连接 reset、连接器写入 403 后停止，没有远程提交/分支/PR 交付；Docker 本轮未重试。reviewer 仅核对该记录措辞，未代执行或独立复核网络操作。
2. 安全矩阵 T09 已明确 `transferred <= byte_limit`，`observed` 可以超过上限以触发超限；显示/缓冲另行有界。公开回放若使用 `accepted` 新名称，须映射当前 `transferred` 字段，并明确不代表对端程序实际消费。

文档准确性复核完成，未发现其余阻止本次规划文档交付的问题；这不是新安全契约冻结，也不构成源码整改或正式运行安全签署。

本次未重跑 root 的格式、链接或源码摘要检查；root 报告其检查通过和源码未改属于交付证据，由 root 保留。没有新增源码修复、安全用例运行或真实 Docker PASS。此前 GitHub 写入问题也不因本次文档审阅而关闭。
