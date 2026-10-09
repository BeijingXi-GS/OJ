2026-10-02范围已实际实施，本文保留2026-10-01停止时的历史记录。当前使用、Agent职责和最新阻塞分别见 [README](README.md)、[VALIDATION](VALIDATION.md)、[最终交接](../handoffs/2026-10-02-file-judge-mvp.md) 与根PROGRESS。

# 文件测试与判题 MVP 范围记录

- 日期：2026-10-01；作者root /Architect；版本judge-mvp/0.1.0-draft，**提案 /未审 /未实施**。
- 用户明确要求：在现有产品增加读取文件、运行例子和AC/WA/TLE/MLE/CE/RE等结果；沙箱参考 [Anolix](https://github.com/Sh1roganeMiyuki/Anolix)。用户已选择服务器读取每题目录中的 `.in/.out` 文件。
- 当前停止原因：[B-JUDGE-001](../issues/2026-10-01-anolix-source-access.md)。此记录说明恢复后要完成的具体工作，不是已实现接口或独立评审结论。
- 原只编译契约保留为旧功能；新运行路径采用独立契约与worker。编译 /运行不放进API进程，完整Phase1 /生产OJ不自动视为完成。

## 文件用例与运行流程

拟采用 `testdata/problems/a-plus-b/manifest.json` 显式列出case ID、input `.in`、expected `.out`、比较器及时间 /内存 /输出限额，不依据目录枚举顺序推断绑定。至少提供样例、零值、负数和大整数四个case。文件由维护者管理，网页不接受任意服务器路径。

文件加载应拒绝缺失配对、重复ID、越界路径、符号链接逃逸、超限文件和无效限额；读取后固定输入 /标准答案快照与SHA256，不在执行中重新读取可变文件。标准答案和其他题目文件不放入选手容器；仅当前输入作为stdin。缺失或损坏用例是SYSTEM_ERROR，不能判WA或跳过后宣布AC。

一次请求固定source快照 → 独立受限编译 → 绑定本次编译产物 → 每case使用全新隔离环境 → 输入文件stdin → stdout /stderr分开有界收集 → 完整停止与资源证据读取 → checker比较 → 逐例 /汇总响应。CE不继续运行；基础设施或清理失败停止后续case。其他正常业务失败可串行继续，以展示每个case结果；所有case AC才汇总AC，若混合业务失败汇总采用manifest顺序第一项非AC，必须显示逐例结果，不把它称为错误严重度排序。

checker's默认提案为token比较（忽略分隔空白，数字仍按原token文本比较），可支持exact字节比较；manifest显式选择并固定版本。不能前后端各自trim或忽略部分输出。此比较策略与原Phase1可信fixture的exact提案分开，需独立审阅后采用。

## verdict证据与界限

| 结果         | 必要证据与处理                                           | 禁止推断                                |
| ------------ | -------------------------------------------------------- | --------------------------------------- |
| AC           | 编译成功、case正常退出、完整输出与标准答案匹配、清理成功 | 仅exit0或编译成功不能当AC               |
| WA           | 正常退出、完整stdout与checker不匹配                      | 运行异常 /截断 /缺失标准答案不能判WA    |
| TLE          | 监督器观测到已定义CPU累计或wall预算超限并终止            | CPU配额不是CPU累计时间限制；取消不当TLE |
| MLE          | 该case cgroup内存超限 /OOM事件等可信证据                 | 单凭exit137 /SIGKILL不能判MLE           |
| CE           | 编译器正常启动后编译 /链接失败，返回真实诊断             | 编译器不存在 /启动失败不是CE            |
| RE           | 运行退出非零或异常信号，且无更优先的资源 /基础设施失败   | 监督器主动kill不是自动RE                |
| OLE          | 原始stdout /stderr实际字节达到停止线，有界停止与清理     | 只截断展示后不能继续对不完整输出判AC    |
| SYSTEM_ERROR | 沙箱不可用、文件错误、证据不可信、启动 /I/O /清理失败    | 不用无隔离subprocess兜底，不伪造0资源值 |

同一case发生多个原因时拟采用：基础设施 /清理失败优先；其次可信MLE、TLE、OLE，再RE，最后checker；独立审阅必须明确观察时点和竞态规则，不把事后资源事件覆盖为一律RE。

时间单位ms、内存 /输入 /输出 /临时文件单位bytes，PID整数。编译预算与运行预算分开；CPU ms若不可信就明确null /未知，不能把wall耗时或DockerCPU配额当它。峰值内存不以各case相加，不把单次RSS抽样伪称cgroup峰值。具体默认限额、可靠资源读取和artifact边界须在预检 /设计审阅后确定，现有64MiB预检不是已冻结题目限额。

## Anolix参考核对和隔离要求

已读 [仓库README](https://github.com/Sh1roganeMiyuki/Anolix) 描述OCI/runc、namespace、seccomp、capabilities、NoNewPrivileges与容器超时清理；已读 [go.mod](https://github.com/Sh1roganeMiyuki/Anolix/blob/main/go.mod) 声明Go1.27.1。核心实现未取得，尚未锁commit，不登记为源码审阅。

采用前需要核对namespace、rootfs、seccomp名单、取消 /runc强制delete、后代 /FD清理与真实退出证据，并补齐OJ所需内存 /时间 /输出预算、答案checker和结果协议。不根据README推断已经可靠支持MLE或每组资源峰值。

真实Docker预检已观察Linux /cgroup v2、内存64MiB /swap0 /PIDs16 /CPU配额1核、NoNewPrivs1 /Seccomp2 /能力集全0。参考 [Docker资源文档](https://docs.docker.com/engine/containers/resource_constraints/) 中memory-swap同memory可禁用swap及CPU配额含义，和 [Docker seccomp文档](https://docs.docker.com/engine/security/seccomp/) 的默认profile说明。预检不等于已采用Docker替代Anolix，也不等于完整沙箱安全通过。

运行阶段必须无网络、无宿主敏感目录 /Docker socket、只读rootfs、必要的独占限额临时空间、最小权限、受限PID /输出 /内存、可信监督器控制与完整tree清理。编译也在受限环境执行；旧compile-worker不是可直接复用的运行沙箱。缺项拒绝启用，不改主机全局安全设置或开启公网。

## 恢复后Agent具体分工

| 任务      | Owner角色 /实际执行者                               | 修改范围                                                               | 交付与验收                                                                                                                                                                                            |
| --------- | --------------------------------------------------- | ---------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| JUDGE-001 | root /Architect，独立Architect +Security Review审阅 | docs/judge、ADR、task /问题 /profile说明                               | 取得并锁定Anolix参考，明确采用方案、输入 /输出 /错误 /预算 /清理契约，独立审阅后局部冻结                                                                                                              |
| JUDGE-002 | root /Backend Builder                               | 独立judge-worker、internal判题适配、testdata、deploy /scripts、API转发 | 真实受限编译与逐例执行、文件checker、verdict映射、健康与能力预检、各种失败和清理证据                                                                                                                  |
| JUDGE-003 | 已授权Frontend Builder，恢复后再派发                | apps/web，按新DTO实现                                                  | 题面显示运行限额 /case数；新增“运行测试”按钮与独立isJudging状态；保留原编译按钮；逐例表格case ID /verdict /耗时 /可信内存值 /诊断；总结果与CE报告；请求期间防重复 /源码变化标旧结果 /错误重试 /响应式 |
| JUDGE-004 | 已授权独立reviewer，未参与实现                      | docs/reviews及必要issues                                               | 判题 /沙箱 /API /前端实现审阅及真实矩阵复核；未测如实标明，不用自检替代                                                                                                                               |

前端继续采用App.tsx列表 /路由、EditorPage.tsx题面和Monaco、lib/api.ts传输DTO、styles.css响应式的现有分层。新judge API DTO在JUDGE-001审阅后才定；不能先硬编码随机verdict。可提案 `POST /api/v1/judgements` 携带problem_id /language /source，返回run ID、总verdict、compile诊断和case数组；服务错误用现有error envelope，基础设施判题结果与HTTP输入错误须区分。API仍不持有执行权，完整Submission /DB /队列不在此增量范围。

最少实际验收：正确A+B→AC，算错→WA，无限循环→TLE，实际内存页增长→MLE，语法 /链接错误→CE，崩溃 /非零退出→RE，持续输出→OLE；另测缺失case文件 /错误manifest /沙箱不可用→SYSTEM_ERROR，程序后代 /网络 /可写目录 /标准答案不可见 /取消 /清理和无旧产物复用。所有结果用真实环境验证，不使用前端固定状态或mock替代资源超限证据。浏览器旧阻塞B-MVP-001仍单独保留，不能冒称已解除。
