# Windows 恢复交付独立静态审阅

日期：2026-10-06。任务：COMM-001/002/004、PUB-001。实际审阅者：`/root/recovery_final_review`，独立 Architect / Security reviewer。

## 结论与停止边界

**已读范围未发现必须修复的实现缺陷；当前本地完整回归失败，真实 Docker 与产品验收 pending，不签署最终 PASS。**

用户最新选择“先保存问题记录并停止”。收到该指令后，本 reviewer 停止新增审阅、读取和测试，仅保存已经取得的结论。未执行测试、Docker、宿主 C++、环境修复、删除或 Git 写操作；唯一写入文件为本报告。实时任务状态由 root 维护。

本轮没有新的完整源码 manifest，不能沿用历史 235 文件摘要绑定本次完整回归，也不能把下列局部文件摘要解释为整个交付版本已经验证。root 转述的失败证据与本 reviewer 独立已读证据分别记录。

## 独立已读范围

已读 `AGENTS.md`、`PROGRESS.md` 前部、`DEVELOPMENT.md`、COMM-004 任务卡、[Windows 文件兼容设计](../judge/WINDOWS_FS_COMPATIBILITY.md)、[此前独立审阅](2026-10-05-windows-filesystem-review.md)和[通信契约](../judge/COMMUNICATION_CONTRACT.md)。

实现范围包括 `internal/problemstore/store.go` 与 `publication_test.go`；`internal/filejudge/cleanup.go`、`cleanup_test.go` 和 `engine.go` 的调用边界；新增 `scripts/communication-test.mjs`、`examples/communication-a-plus-b` 包及 [COMMUNICATION_ENVIRONMENT.md](../judge/COMMUNICATION_ENVIRONMENT.md)。为核对契约，还读取通信规范摘要、数据快照、执行结果和三模式投影的相关实现，以及现有启动脚本、页面路由、教师 A+B 填充逻辑和 `package.json`。

### 题库发布与回滚所有权

静态实现符合受审的 pending marker 方案。`publish` 先独占创建空 `.staging-<ID>`，再独占创建最终目录；只有后一步实际成功才取得最终目录的回滚所有权。文件写入执行 Sync / Close，最终子 Root 关闭后删除标记才提交。失败时仅清理本次创建的最终目录，清理无法确认则保留标记；既有最终目录与来源不明的 pending 不取得清理所有权。

`registered` 在初次读取、打开最终子 Root 后且读取注册内容前，以及所有成功或错误出口检查 pending。有效 pending 隐藏未提交题目；标记查询异常、非目录、链接或非空标记失败关闭。标记在 Lstat 与 Open 之间被正常提交删除的情况有再次查询处理。当前正常发布使用随机 ID、发布后目录不可变；结论不扩大为对抗宿主同权限目录替换或断电持久性的保证。

已读测试覆盖部分 JSON 与完整未提交内容的 Get / List 隐藏、重启可读、旧 version 1、提交前关闭、open / write / commit 故障、回滚失败保留、既有 final / pending 所有权、查询故障、关键交错和并发读者。未独立运行这些测试，也未将未注入的底层 Close 故障或宿主替换竞态列为实测通过。

### Windows 判题清理

已读实现维持此前窄范围审阅结论。taint 和非直属 `run-` 路径在变更前拒绝；Windows 从 jobsRoot 的受限 Root 锚定目录，并检查目录非链接及打开后的对象身份。只对 `program`、`checker-build/program`、`sender-build/program`、`receiver-build/program` 四个固定产物恢复 readonly 属性，检查普通文件、持有句柄和 SameFile，修改后及删除 Job 前再次核对身份。

缺失尚未生成的产物允许跳过；未知路径没有全树 chmod 回退；失败继续保留 SYSTEM_ERROR / taint 路径。Linux 不进入 Windows 属性恢复分支。结论仍依赖可信宿主管理和容器无可写 Job 挂载，不承诺防御宿主同权限攻击者的所有 TOCTOU / 硬链接竞态。

独立读取了既有 `.local/evidence/windows-cleanup-20261005.log` 与 `windows-cleanup-boundaries-20261005.log`：前者记录可信 Go fixture / fake daemon 专项 PASS；后者未知 readonly 路径 PASS，但 Job / build / artifact 三种 Windows 符号链接子用例因缺少创建链接权限全部 SKIP。父测试的 PASS 不能当作这些链接边界已实测通过。这两份历史专项记录不替代本轮失败的完整回归。

### 通信脚本、教学包与指南

新增真实验收脚本没有 fake 模式，不启动服务、上传题目、删除资源或运行宿主 C++。其十二项矩阵连接既有本机 API / worker，要求真实 Docker 查询成功，并检查 Linux / amd64、cgroup v2、镜像 ID、owner 容器和任务目录清理。查询失败不会被解释为零残留。

源码摘要采用固定 sender / receiver 顺序及 Go JSON 对 HTML 字符和 U+2028 / U+2029 的转义规则；请求数组倒序场景仍按角色校验。数据摘要与后端一样按长度前缀组合 manifest、case ID、输入和答案；产物聚合摘要使用固定角色顺序。CE 检查实际尝试角色前缀和零运行例；运行矩阵检查逐例结果、真实角色 OOM / 自然退出 / 受控终止证据及清理。报告初始与失败状态保持 `completed:false`，全部通过才置 true。

已读示例的双模板、descriptor 和内置 A+B 数据声明一致：sender 发送原始 a、b，receiver 计算和，范围 ±10¹²、32 字节、四例。教师 README 明确说明当前页面一键填充使用 ±10⁹ 与 `2 3`，使用本包需同步题面和样例；未把较小范围与 large 数据混作同一规格。页面链接与已读路由一致；学生和教师投影说明符合服务端白名单，教师通信字段不含角色退出 / OOM / 监督器终止证据。

初读 `PROBLEM.md` 时末尾仍有“模板是完整演示解法”说明。已告知 root；root 表示会按用户要求删除。停止前未复读最终文本，本报告不声称该文案修改已经独立复验。

该脚本只覆盖内置 A+B 的十二项判题路径。双 attach / 取消 / busy、无答案挂载、提前退出 / EOF、教师上传和重启持久化、三模式页面、旧题型回归与监督器停止阶段仍需真实验收。脚本静态正确不等于其真实矩阵已通过。

## root 转述的本轮结果，未独立读取日志内容

root 报告 `.local/evidence/windows-recovery-20261006-check.log` 的实际结果为：format、lint 通过，problemstore 通过（1.259s）；filejudge 的 `TestSpecialJudgeLifecycleAndPriority/checker_cleanup` 在 testing.go 的 TempDir RemoveAll 清理只读 `checker-build/program` 时失败，整体 exit 1。后续 Node SSR 检查及 build 未执行。**本地完整回归必须登记 FAIL。** 本 reviewer 未读取该失败测试、未判断失败根因、未提出或执行替代修复。

root 随后的源码读取在自动审批阶段发生 `stream disconnected / response decode`，工具明确未执行。root 要求不让本 reviewer 代读该被拒操作，本 reviewer 没有代读或旁路执行。

root 报告本轮 Docker 首次查询为 Linux pipe notfound，未观察到 Docker 进程；正常隐藏启动 Desktop 命令 exit 0 后，再查仍为 notfound 且无进程。此前 permission denied 仅为历史事实。启动命令 exit 0 不能证明引擎启动成功；当前没有真实通信矩阵、服务 ready 或资源零残留验收。上述本轮环境事实来自 root 转述，本 reviewer 没有操作 Docker 或独立读取该轮环境日志。

## 局部读取摘要

以下 SHA256 由本 reviewer 在停止指令前读取，用于明确局部审阅对象。它们不是完整源码 manifest，也不绑定尚在变化的文档或最终远程交付。

| 文件                                              | SHA256                                                             |
| ------------------------------------------------- | ------------------------------------------------------------------ |
| `internal/problemstore/store.go`                  | `069a86a899ac87375c9cc9e7d362c9f9e2d9c5a6acb2ee340e1bf80a4dae1f34` |
| `internal/problemstore/publication_test.go`       | `9a905f1533e7764b3357d2619c726cafb86909f2e875bb05a52ea178651a3dad` |
| `internal/filejudge/cleanup.go`                   | `fb8cfc210711d45b102bc4cc1011bbeb0846021916d0a13609e875483987d5c1` |
| `internal/filejudge/cleanup_test.go`              | `ccc636eb1e75e890b8bbcf1d14fbd9c25e8f9ccdd1a993b24f09fbf8ab0e7904` |
| `internal/filejudge/engine.go`                    | `ff858559d13d99808fbd6503f06fd8e59892368fd0bb8b3df44e78f053d5653e` |
| `scripts/communication-test.mjs`                  | `5d32359acb04d21d8dd01b2ec8d7dd75acfb76a92d3701ff69c88b6c9e6bb6d6` |
| `examples/communication-a-plus-b/descriptor.json` | `bc081942a17f6b5808bb0930fe163c61db75d7405567834275041219213ff601` |
| `package.json`                                    | `b9c106b7d3b095ecfa8145a62ef07e8f4510cab2f9b015e77cc1cfc628b1359c` |

后续恢复须重新取得当前版本的完整回归、源码 manifest 与真实环境证据，并由独立 reviewer 核对；本次只交付停止时的静态审阅记录。
