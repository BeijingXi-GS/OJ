# B-SPJ-001：真实验收自动审批通信失败

日期：2026-10-03 18:56前后（Asia/Shanghai）；处理人：用户。状态：外部阻塞，已停止新增工程操作。

## 背景、目标与操作

本轮内置/教师上传特判、学生页面精简、工程师说明与启动清理修复均已实现，局部独立设计/代码审阅与可信单测已通过。真实验收尚未完成，准备执行已授权的本项目停止验证脚本，核实预检/空闲/选手运行三种阶段的正常退出、Docker容器及任务目录清理。

第一轮脚本执行已结束，记录不是PASS：在验证helper确认日志处assert失败，报告tests为空且completed:false；有一条脚本子进程退出0但日志未被观测，具体原因尚未核实，不能断言实际三阶段已运行。已修改测试记录为在断言前保存实际子进程日志，准备重新执行以定位本轮测试脚本缺口。

申请执行：在本仓库运行 `node scripts/supervisor-stop-test.mjs` 并把输出存入 `.local/evidence/spj-supervisor-stop-rerun.log`，使用require_escalated使Docker查询在已有用户授权下可访问本机Linux引擎。工作目录为outputs/oj，无数据卷删除、无代码发布、无任意宿主程序执行。

## 实际返回

工具返回：

> Automatic approval review failed: stream disconnected before completion: Transport error: network error: error decoding response body

工具同时明确该动作**未执行**，这是审批流程无法完成，**不是对动作不安全的判定**。审批拒绝发生在进程创建之前，没有本轮重跑session或成功证据。错误原因是工具已说明的连接/响应解码失败；未确认具体网络链路或服务端根因，不把它写成Docker故障，也没有通过换执行方式绕过审批。

## 已知环境与影响

- 本轮已获权成功核实Docker为linux/x86_64/cgroup2，旧supervisor停止后owner查询成功为空、任务0，基础卷保留。
- 已单独真实执行预检期私有pipe取消：18:47:59 preflight started、18:48:05取消、helper确认无残留，进程exit0。这只证明该次预检取消，不等于三阶段脚本或产品验收通过。
- 停工记录时仅核实已发起操作后果：5173/8080/8082连接失败、8081健康200，judge-jobs计数0。最近成功Docker owner查询为空；审批失败后没有再次成功Docker查询，不把它冒称当前确认值。
- 完整check第一次只到格式检查失败（读取Agent仍编辑的前端/审阅MD）；前端与review MD随后已整理，但最终完整check尚未重跑。Linux symlink单测、真实special矩阵、上传重启、浏览器、最终产品签署均未完成。
- 源码当前200非MD文件 SHA256 `bb1ac048576337b8b8a67f29861b2c6dbe8dce43b50bac50de4f47aa45d34166`，分支task/SPJ-001-student-construction-2026-10-03；这只是停止快照，不是最终通过冻结。

## 用户接手与恢复条件

请恢复Codex与自动审批服务的连接，并在可继续执行审批后通知继续。Agent收到继续指令后先验证审批可完成和真实环境/残留，再定位测试脚本日志缺口、完成全部未完成验收、恢复本机页面服务与最终独立产品审阅。无需重置Docker数据或删卷。

遵循用户“出现问题详细写入单独MD、由用户处理、阻碍后续进展则停止”的规则，本轮不继续重试审批、不绕过检查、不把局部PASS标为MVP增量已完成。状态见[PROGRESS](../../PROGRESS.md)与[交接](../handoffs/2026-10-03-special-judge-blocked.md)。

## 2026-10-04恢复进展

用户最新明确继续完成、解决问题并持续整理。16:11本轮同类require_escalated操作已实际创建并执行，自动审批通道可用，未绕过审核。实际Docker Linux管道未运行，全部4服务连接失败、jobs0，正在按既有授权启动已安装Docker Desktop。审批通信故障可登记恢复，真实验收仍未完成；不能用本次工具可执行代替Docker或产品验收。后续实际结果继续补录。

## 2026-10-04通信故障复发

随后已完成真实Docker恢复、Linux单测/23项特判矩阵/对应版本完整check和教师上传AC2/2。准备再次运行三阶段停止诊断时，同一审批通信错误复发，动作未执行；不能将16:11恢复事实当作一直可用。已新开[本次前因后果与恢复条件](2026-10-04-approval-resume-interruption.md)，B-SPJ-001当前仍阻塞。全部Agent停止新增实现/测试，当前[交接](../handoffs/2026-10-04-special-judge-paused.md)替代旧停止位置，历史证据保留。
