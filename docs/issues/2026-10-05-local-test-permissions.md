# A+B 通信回归的本地文件权限阻塞

日期：2026-10-05；任务 COMM-001 / COMM-002 / COMM-004；记录人 root，处理责任交用户。当前未关闭，依赖完整回归与真实验收的后续工作停止。

## 前因与目标

用户要求继续开发、尝试连接 Docker，并追加最简单通信题：发送端给出 a、b，接收端计算 a+b。用户已允许不依赖 Docker 的本地实现和可信测试，保留真实验收待办。新增 `communication-a-plus-b`、四组文件数据、前端双模板与教师示例已经落盘；本次目标是复验新增后的完整源码，不能沿用新增前的工程检查。

此前 226 文件 / 80a5377b 摘要对应完整 check exit0 是历史证据。新增版本为 235 个非 MD 文件，源码摘要 `71e866f2e74ab2f8e9d107a8cacaf9240ed03ded1c87bf963821b8ba9a3cb634`，详见[验证记录](../judge/COMMUNICATION_VALIDATION_2026-10-05.md)。

## 操作顺序与实际结果

1. backend 与独立 reviewer 在默认临时目录执行 Go 回归，出现 `rename ... Access is denied`、公开错误 `problem_store_unavailable`，以及 HTTP Client 的 `judge_unavailable`。两个执行者都登记失败，没有改写为整体 PASS。reviewer 随后只执行不依赖其失败路径的三个 A+B 专项，实际 exit0；该事实保留，范围有限。
2. root 经正常 `request_permissions` 取得网络权限，在命令进程中将 TEMP / TMP 指向项目 `.local/tmp`，保持 GOPROXY / GOSUMDB off，仅使用已有缓存。未修改系统临时目录、主机 ACL 或项目上传数据。随后执行五包 `test -count=1`，实际 exit1：problems、judge-mvp PASS；problemstore / filejudge / httpapi 仍因存储操作或清理失败而 FAIL。本轮输出未再出现此前集中 HTTP 连接错误，不据此承诺所有网络权限一直有效。
3. 通信 fake 流程已产生正确的逐例 AC / WA，例如输入 `1 2\n`、receiver 输出 `3\n` 或 `5\n`。但收尾删除可信假产物 `sender-build/program` / `receiver-build/program` 返回 `unlinkat ... Access is denied`，Engine 按契约最终返回 SYSTEM_ERROR / taint。这不是一个可登记为完整 AC 的结果，不能降低清理要求让测试变绿。
4. root 再通过正常权限工具申请 `.local/tmp` 精确读写权限，工具返回 turn 范围授予。专项复验 `TestCommunicationUploadPersistsOwnedTemplatesAndConfig` 与 `TestCommunicationJudgeSuccessfulBindingsAndRuntimeFailures` 仍 exit1；filejudge 18.263 秒，上传公开错误与假产物清理拒绝仍在。完整日志保留于 `.local/evidence/communication-ab-20261005-permissions-recheck.log`。权限授予响应没有证明文件操作已恢复。
5. 不再重跑失败验收，不改安全策略、返回判定、主机权限或测试断言；没有让其他 Agent 重试同一被拒操作。完成问题、台账和交接记录后停止依赖环境的工作。

所有测试均为可信 Go / Node fixture 与 fake 控制工具；未在 Windows 宿主编译或执行任何选手、教师 C++。

## 已知事实与未确认原因

已确认：默认临时目录失败；项目临时目录和正常读写 grant 后仍失败；上传存储及假产物清理受影响。目录能够创建、文件能够生成，不等于重命名和删除权限可用。此权限问题与[Docker 管道访问](2026-10-05-local-docker-access.md)分别记录。

候选原因是执行环境对重命名 / 删除的权限实施、Windows 文件属性或文件句柄状态、主机安全软件限制。尚无证据确定是哪一项，也未证明新增数学题源码存在导致这些拒绝的回归。没有用推测替代失败事实。

## 解决策略与恢复条件

用户先恢复当前执行环境对项目测试临时目录的创建、目录重命名、文件删除与清理能力，核查权限控制或占用记录。不要删除 `.local/uploads`、数据卷或来源不明的容器；本轮失败留下的测试临时文件也未批量清除。Docker 另需让正常 CLI 能连接 Linux API。

恢复后先在同一正常执行环境中复验上述两个可信专项，确认没有 Store 错误、unlinkat 拒绝、taint 或临时目录清理失败，再执行五包回归与完整 `npm run check`。日志必须绑定新源码清单；随后由独立 reviewer 复核完整证据，再做真实 Docker A+B、其他判定、教师上传重启与浏览器验收。源码变动后重新生成摘要。

未取得上述实证前，本问题不关闭，当前完整回归不标 PASS，真实产品、GitHub 交付继续 pending。

# 2026-10-07 恢复补录

此前 rename/readonly 清理问题已按受审兼容方案修复，剩余特判测试夹具收尾也已处理。正常取得本轮网络权限后完整 `npm run check` 通过；实际日志、摘要和限制见 [本轮验证](../judge/RECOVERY_VALIDATION_2026-10-07.md)。本档原失败和停止信息保留为历史；Docker 与审批故障单独见 [原因与对策](2026-10-07-approval-diagnosis-and-recovery.md)，不因本地测试恢复而关闭真实环境问题。
