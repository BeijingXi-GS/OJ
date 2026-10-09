# 2026-10-06：恢复回归、Docker 与审批中断记录

## 背景与停止决定

用户要求继续解决文档中的问题，补齐通信题环境和题面，最后上传 BeijingXi-GS/OJ。此前已允许本地开发先行、保留真实 Docker 验收待办，并授权并行开发与独立审阅。

本轮完整回归出现一个 Windows 测试收尾失败；随后读取相关测试源码时自动审批服务连接中断，命令未执行。用户随后明确选择 **“先保存问题记录并停止”**。当前仅保存记录，不再修复、测试、启动环境或上传。此决定优先于此前继续工作和上传指令。

## 已保存的工作

- 题库发布改为受限目录内的 pending 标记协议：完整写入并关闭后才提交可见性；回滚仅处理本次创建的目录，删除失败保留标记。设计见 [Windows 兼容方案](../judge/WINDOWS_FS_COMPATIBILITY.md)。
- Windows 判题清理增加四种已知产物的只读属性恢复；保留目录身份校验、taint 拒绝和失败上报。此前专项结果见 [独立报告](../reviews/2026-10-05-windows-filesystem-review.md)，不能据此声称本轮完整回归通过。
- 通信 A+B 的 [题面与教师包](../../examples/communication-a-plus-b/README.md)、发送端/接收端源码、四对数据和 descriptor 已存在；输入范围 ±10¹²，通信上限 32 字节。
- [环境指南](../judge/COMMUNICATION_ENVIRONMENT.md) 和 `npm run judge:communication:test` 已保存。脚本用于已正常启动的真实环境，十二项预期覆盖 AC、WA、CE、RE、TLE、MLE、OLE；本轮没有执行真实矩阵。

## 实际操作与结果

1. 阅读已有工作区、任务卡、环境包和设计；保留全部已有改动、教师题库和数据卷。
2. 显式查询 Docker Desktop Linux 管道，返回 `The system cannot find the file specified`。进程查询没有列出 Docker 进程；安装路径存在。
3. 通过正常工具取得本轮 network grant，用于本地 HTTP 测试及已授权远程核验。隐藏启动现有 Docker Desktop 的命令 exit 0；后续管道查询仍返回 not found，仍未列出 Docker 进程。启动命令成功不代表引擎启动成功。先前的 permission denied 仅是历史观察，不能覆盖本次结果。
4. 已有环境包、示例 Markdown/JSON、验收脚本、设计和旧评审报告的格式整理 exit 0。没有修改选手 C++ 或执行它们。
5. 使用项目 `.local/tmp` 作为 TEMP/TMP、离线 Go 依赖执行完整 `npm run check`，实际 **exit 1**。日志保存在 `.local/evidence/windows-recovery-20261006-check.log`。
6. 格式检查、Go vet、前端 lint 通过。Go 阶段 problemstore 通过（1.259s）；filejudge 失败。后续 Node 监督器测试、SSR 检查及 build 因串行检查提前退出而 **未执行**，不能沿用旧结果。
7. 为定位剩余失败，发起读取相关测试代码及前端说明的命令。自动审批返回 `stream disconnected before completion` / `error decoding response body`，明确动作未执行。没有改用其他工具或让 Agent 代做该被拒读取。
8. 已说明故障并请求用户指导，用户选择保存记录后停止。本轮没有创建提交、推送、远程分支或 PR。

## 剩余回归失败：事实与待验证方案

唯一已见失败为 `TestSpecialJudgeLifecycleAndPriority/checker_cleanup`：Go 的 `testing.go:1369` 在 TempDir RemoveAll 收尾时，删除本次临时目录下 `jobs/run-…/checker-build/program` 返回 `Access is denied`。其余通过结果不代表 filejudge 整包通过。

根据已有只读产物复现和此次错误位置，推测故意模拟 checker 清理失败的测试留下了只读假产物，触发 Windows 临时目录收尾错误。**相关测试源码的本次读取未执行，因此该解释尚未完成源码核实，不把推测登记为已定位或已修复。**

后续方案：恢复后先正常读取该测试，核实是否故意保留假产物；如属测试夹具问题，只在测试拥有的临时目录中恢复已知假产物属性，让测试框架完成收尾。必须保留生产逻辑的 taint、失败证据和停止规则，不以删除不明目录或忽略清理错误使检查通过。修复后先运行受影响用例，再重跑完整检查并保留新日志，不覆盖本次失败证据。

## Docker 与自动审批的未确认原因

Docker：目前只能确认该 CLI 所见 Linux 管道不存在、启动后未观察到进程。尚不能确认是 Desktop 启动退出、会话/访问限制、WSL 后端故障或其他原因。用户可先在正常桌面检查 Docker Desktop 启动状态及其诊断；恢复后先确认 Linux API、平台/cgroup、镜像和预检，再进行通信矩阵，不能直接判定沙箱已可用。

审批：这是审批服务的传输/解码故障，**不是操作被判定为不安全**。network grant 已取得也未避免本次审批故障。用户需先恢复正常工具审批连接并通知继续；无需重新授权既有功能范围、并行审阅或 GitHub 上传，但本次停止要求在用户通知继续前持续有效。

## 接续顺序和责任

1. 用户恢复正常工具连接并通知继续；root 核对现场后完成上述测试源码核实和受限修复。
2. root 完整回归、示例数据静态核验、当前源码摘要及文档链接检查；独立 reviewer 复核当前证据，不能用 root 自检替代独立审阅。
3. 题面仍需删除末尾“模板是完整演示解法”等学生无需了解的说明；本轮只提出，尚未修改。环境指南中的 10-05 状态为历史，以本档和 PROGRESS 的 10-06 状态为准。
4. Docker 恢复后，环境负责人运行真实通信矩阵、教师上传与重启持久化、三模式页面、取消/清理以及旧题型回归。未知残留不自动删除，真实验收未完成前保持 pending。
5. 当前版本检查与独立审阅完成后，root 按用户要求交付 GitHub 开发分支并清楚列出真实环境待办。发布前重新核验远程状态，保留远程历史；成功后再记录实际提交和 PR。当前没有上传成功证据。

本次没有生成新源码 manifest，旧摘要不作为当前快照。既有源码、数据和失败日志均保留；`.local` 证据为本机文件，不随 GitHub 源码上传。
