# 判题扩展预检与停止交接

- 日期2026-10-01；执行者root /Architect；分支task/JUDGE-001-file-cases-sandbox-2026-10-01，保留旧改动，未提交 /推送。
- 最新授权：原编译产品扩展真实运行与AC/WA/TLE/MLE/CE/RE等判定，参考Anolix；用户确认服务器题目目录.in/.out，不是网页上传。
- 上轮只编译范围由新要求扩展；B-MVP-001旧浏览器验收仍未关闭。此次仅进行运行扩展需求核对与能力预检，未修复或绕过浏览器。

## 已观察

Docker29.8.1 /linux amd64 /cgroup v2 /cgroupfs；独立临时容器实际memory.max64MiB /swap.max0 /pids.max16 /cpu.max1核配额 /CapEff全0 /NoNewPrivs1 /Seccomp2。临时容器自动移除，旧服务及卷保留。这不是运行隔离、OOM映射、CPU计量或真实判题安全已通过。

Anolix仓库首页和go.mod网页内容已读；核心源码网页cache miss，本机git ls-remote GitHub连接被重置，commit未取得。Anolix声明Go1.27.1，本机Go1.25.4；最终参考适配与工具链策略未裁定，未改源码或升级工具链。

## 停止及交付

[B-JUDGE-001](../issues/2026-10-01-anolix-source-access.md) 一事一档记录命令、返回、事实 /推测、影响与用户恢复条件。按用户要求停止，不再下载 /换方案 /写判题代码 /启动新Agent。[范围记录](../judge/MVP_JUDGE_SCOPE.md) 描述恢复后的文件流程、verdict证据、隔离要求和详细前后端 /review任务；[任务卡](../tasks/JUDGE-001-file-cases-sandbox.md) 与PROGRESS同步。

本轮未修改应用源代码、npm /Go依赖、Docker部署配置或.env；未运行任何新增判题源码，无AC/WA/TLE/MLE/CE/RE/OLE证据，没有独立review。旧compile MVP行为不变，judge_enabled仍false；不重复执行旧检查后将其当新判题通过。

用户恢复GitHub访问或放入完整Anolix源码并说明版本后通知继续，先复验 /锁定 /独立审阅，再按JUDGE-002→003→004实现和验收；新的阻塞继续单独记录并停止。浏览器最终验收仍依赖B-MVP-001解除或用户明确调整验收方式。
