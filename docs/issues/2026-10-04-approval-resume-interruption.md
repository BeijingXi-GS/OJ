# 恢复验收时自动审批通信再次中断

日期：2026-10-04 16:54左右；当前阻塞 B-SPJ-001；处理人：用户。不是安全审核否决，也不是已确认 Docker 故障。

## 前因

用户要求继续完成、解决问题并持续整理。16:11同类require_escalated操作已经实际获执行，之后启动已有Docker Desktop，核实Linux/x86_64/cgroup2，启动现有Compiler和MVP，完成Linux单测、23项真实特判矩阵、真实教师上传及AC2/2。旧自动审批问题的恢复事实保留。

16:48三阶段停止脚本仍未通过：只bootstrap，未观测preflight，阶段等待45秒失败；子进程正常退出、stdio关闭、实际Docker查询为空/任务0/端口关闭。原错误保留，未放宽断言。针对已授权启动缺口，root加入5条无敏感信息的阶段日志，准备重跑定位具体等待点。

## 动作与实际返回

申请在仓库运行 `node scripts/supervisor-stop-test.mjs > .local/evidence/spj-20261004-supervisor-progress.log 2>&1`，通过require_escalated访问原已授权Docker Linux引擎。仅测试本项目owned实例预检、空闲、运行中停止，无数据卷删除、源码发布或主机设置变更。

工具实际返回：

> Automatic approval review failed: stream disconnected before completion: Transport error: network error: error decoding response body

工具明确“动作未执行，因为自动审批审核未能完成；这是审核流程失败，不是操作不安全的判定”。因此不存在这次重跑session、成功结果或三阶段证据。连接/响应解码故障是工具已报告原因；更具体的网络或服务端根因未确认。

## 后果和停止位置

- root及子Agent停止新增实现和测试，未换方式绕过审批、未重试Docker。只整理已完成结果、问题与交接。
- 16:56只读记录：Web5173/API8080/Judge8082不可访问，Compiler8081 HTTP200，judge-jobs目录0。最近成功Docker owner查询是16:48失败测试清理阶段，确实为空；审批失败后没有成功再查询，不能当作当前实时确认。
- 现有13题：7内置、6持久上传。6上传/34文件的重启前摘要已保存；未取得重启后对比或全部上传题重判结果。原3道上传题和数据卷均保留。
- 23项真实矩阵、Linux单测、完整check/两smoke已通过其对应版本；后来补充停止自测和5条启动日志，尚未重跑最终全量check。常驻下载链接和传统textarea已代码复审，真实下载/编辑提交仍待验证。
- 当前200非MD文件摘要 `3406da9bef0efd1719a405bd48c2cc7206609368f0329f8012c5b6280bcf8df3` 是停止快照，不是最终验收冻结。最终独立产品审阅保持pending。

## 用户处理与恢复条件

请恢复Codex自动审批服务连接，并在能继续审批时通知继续。不需要再授权已允许的并行开发/本项目修复，也不需要重置Docker、删卷或改安全设置。

恢复后先通过原审批渠道核实环境和残留，再按[本轮交接](../handoffs/2026-10-04-special-judge-paused.md)定位启动等待、运行三阶段停止、重启持久验证、页面复验及最终检查/独立审阅。实际审批恢复不代替产品通过；若再次出现外部阻塞，继续一事一档并停。当前事实看[PROGRESS](../../PROGRESS.md)。
