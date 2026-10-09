# ADR-0003：可信 Mini Judge 的观察时点、清理和错误语义

- 日期：2026-10-01；作者：根 Agent / Architect；关联 P1-001 / P1-002。
- 状态：proposed / mini-judge/0.1.0-draft；独立审阅未执行，不允许据此直接实现。
- 完整边界与输入：[Phase 1 DESIGN](../judge/PHASE1_DESIGN.md)。

## 背景

已有 Go 草案没有可用的运行实现，不能仅按 `CommandContext` error 或主进程 exit 判 TLE / RE。正常退出和 timeout 可能同时到达，父进程也可能退出而后代仍占用资源 / 持有输出 pipe。输出 EOF、主进程 wait 和全树清理是三个不同的事实。

## 拟采用的决策

1. 每个编译 / case stage 一个独立监督器和唯一 arbiter；monotonic 时间、观察事件 seq 与原始证据分开。wall 从 launch 前开始，成功返回不能重置预算。初始化 profile / 镜像不计入选手 wall；启动失败归 INFRA。
2. EXIT 只有在 arbiter 观察时点严格早于 deadline 才能成为正常停止候选；deadline相等或以后观察即 wall超时。当前策略保守，不推断尚未被监督器观察的内核退出时刻。
3. 取消只在尚未登记 stop 时获得 CONTROL_CANCEL；若观察时已经到 deadline，wall超时优先。已经登记 EXIT 的迟到 timer / cancel 无效。CPU / wall同次快照均超限时保留两个观测，主 reason=CPU_TIMEOUT。
4. `stop_reason` 只登记一次，用于解释为何停止；它是候选执行证据，不直接等于 verdict。最终要读完整有界输出和资源证据并完成清理后，才发布唯一终态。
5. 子进程起始即进入专属 cgroup；监督器在组外。Linux stage supervisor 为唯一 reaper，负责直接 child 与 subreaper 收养后代；不能一处 `Cmd.Wait`、另一处 `wait4(-1)` 争抢同一 child。独立进程组用于辅助控制，不单凭 PGID 证明全树结束。
6. 正常 root exit 也验证后代：compiler遗留活跃后代是 INFRA，fixture遗留后代为 RE / CHILD_REMAINS。所有停止路径在2秒内完成专属组清理、wait、I/O worker和FD回收；kill控制信号不当成自发用户signal。
7. 清理失败、失去计量、损坏产物或 I/O / Checker failure优先 SYSTEM_ERROR。清理超时将 runner隔离，保留标识、禁止下一任务，由用户处理；不能把结果先写AC再后台处理残留。
8. 可信fixture环境与生产Sandbox接口分开，放 `internal/devrunner`；不向 API / Web / Worker开放，不把普通本地执行实现成缺少隔离时的生产回退。

## 最终归类的顺序

以下在 report 发布前执行，report发布后不可改变：

1. 证据 / 清理 / 计量不可信 → SYSTEM_ERROR；即使之前停止原因是取消或超时也不声称处理已完成。
2. CONTROL_CANCEL 是已登记的停止原因且清理成功 → CANCELLED。后来的取消不参与归类。
3. 若编译阶段：可信输出 / 资源预算触发或编译非零退出 → CE；compiler / 启动 / 产物校验异常 → SYSTEM_ERROR。
4. 若运行阶段：开发保护 memory / output / PID / temp 触发 → SYSTEM_ERROR / DEV_RESOURCE_GUARD；这里只拒绝未验收的资源归因，不实现 MLE/OLE。
5. 可信累计 CPU 达到预算或 WALL_TIMEOUT → TLE；最终CPU读取必须参与，正常EXIT候选不能掩盖已经超过CPU预算的事实。
6. 自发非零exit / signal或 CHILD_REMAINS → RE；监督器kill造成的退出不参加该条。
7. 退出0、证据完整、无前述停止 / 异常、清理成功 → ExactChecker；匹配AC，不匹配WA。

同次事件快照保留所有 limit / observed / scope；最终 reason由该固定顺序选择，不以 goroutine调度随机决定。不以单一 exit code（例如137）断言 OOM或TLE。

## 可比较的备选与代价

按 select分支先后决定用户verdict难复现；等到所有pipe EOF再判断exit会被后代持有pipe拖住；只kill主PID / PGID不能证明完整清理。拟采用观察时点和显式清理，可以审计并用可控事件测试，但仍不是内核精确退出时钟，临界样例可能保守判超时。

本阶段增加 cgroup权限、launcher / reaper和受控profile前置；缺少它们时不启动。不会通过降低限制或无隔离执行来换取“演示成功”。M1资源guard的保守SYSTEM_ERROR不代表最终OJ的用户错误策略，Phase4必须另行设计并真实验证。

## 兼容、回退与验收

无DB迁移，无外部客户端。现有 domain / Runtime / Graph仍draft；本ADR不改其字段。回退为停止CLI并停用新增devrunner，保留诊断，不改基础服务或卷。

独立 P1-002 必须明确批准观察时点、取消顺序、最终归类、完整reaper协议和预算；P1-010须实际锁定Linux / compiler并验证前置。验收至少包括 deadline前 / 相等 / 后、事件顺序排列、真实重复退出竞争、编译timeout、父先退出 / setsid后代及全部异常路径清理，具体 [fixture规范](../judge/PHASE1_FIXTURES.md)。当前全部为未执行的验收要求。
