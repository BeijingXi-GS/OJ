# 三阶段停止测试的日志观测与启动缺口

日期：2026-10-04；处理人：root / lifecycle_fix；范围：已记录启动停止问题的授权修复，常规测试脚本修正，尚在复验。旧审批通信问题不作为本轮失败原因。

## 前因与事实

10-03旧停止测试出现子进程exit0、没有日志，被finally中的清理提示断言覆盖；未取得实际三阶段通过证据。10-04用户要求继续解决，分别等待exit和stdio close、保留首错误、每次保存独立报告并增加bootstrap记录，未取消阶段/退出/清理断言。

可信进程自测证明：暂停读管道可在exit0时观测0字节，随后close前读到完整输出。因此exit不是日志排空证明，但不能认定旧实际失败一定由此造成。动态import的data URL夹具正常；两次初始夹具不可靠（Windows后代继承管道、在Promise返回后才测量），改为在实际exit事件记录字节数后四类自测通过。独立reviewer也实际执行通过。

## 实际重跑

16:48:01–16:48:47，真实脚本只有bootstrap 28字节，没有所需preflight标记，45秒阶段等待失败。stop后子进程exit0、stdio close0，Docker查询成功且为空、任务0、端口明确拒绝连接；未取得helper确认提示。本轮如实记录首阶段错误和后续日志确认错误，没有把清理成功当作测试通过。

证据 `.local/evidence/spj-supervisor-stop-2026-10-04T08-48-01-624Z.json` 和 `.local/evidence/spj-20261004-supervisor-real.log`，completed:false。当前在检查真实本地文件动态加载及Windows子进程启动路径；不扩大超时或删除必需断言。尚未确认原因。root暂停服务重启，待窄修后再次验收。最终结果补录于本档及[PROGRESS](../../PROGRESS.md)。

## 进一步事实和停止

随后新增可信本地模块夹具：项目Unicode/空格路径中，相对动态import及编码file URL import均进入模块并ready，继承stdout/stderr的Node后代日志可观测，stdin信号后exit/close0；完整自测实际通过。这证明加载机制本身可用，仍未定位真实mvp等待在哪一步，不能把夹具正常当三阶段通过。

root为mvp-dev加入5条无敏感值启动阶段日志，覆盖Compiler探测、Judge探测、Judge构建/启动和API构建。准备重跑时[审批通信再次中断](2026-10-04-approval-resume-interruption.md)，动作未执行。未添加wrapper fetch/spawn诊断，未再修改超时或运行测试。本档保持未关闭，接续见[本轮交接](../handoffs/2026-10-04-special-judge-paused.md)。
