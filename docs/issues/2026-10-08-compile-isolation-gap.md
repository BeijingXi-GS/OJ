# B-SEC-001：仅编译入口未统一到逐任务沙箱

- 发现方式：2026-10-07 至 10-08 对当前代码静态盘点；root 与 security_gap_review 分别读取。
- 影响：接受非维护者任意源码前的安全门禁；不阻止文档整理或明确标记风险的源码交付。
- 当前事实：未实施整改、未运行利用程序、未证明实际入侵。实时任务状态见 PROGRESS。

## 前因与确认过程

工程先实现“只编译不执行产物”的本机 MVP，之后新增逐任务 Docker 判题。两个入口并存。用户本轮要求按恶意代码情形单列安全问题，盘点发现运行测试的保护不能自动覆盖旧编译按钮。

`internal/compilerworker/worker.go` 在常驻服务里通过固定参数 exec.CommandContext 调用 g++；已有 15 秒超时、单并发、诊断截断和 Linux 进程组取消。`deploy/compiler/compose.yaml` 只设置 loopback 端口、init、停止等待和restart，没有显式 memory/PID/CPU、只读根、无网络、cap-drop/NNP 等限制。与 `internal/filejudge/engine.go` / `docker.go` 的逐阶段预算不同。

## 原因、后果与边界

原因是早期范围的编译适配仍被保留，未在新增判题沙箱后合并安全策略。编译器也处理不可信程序文本，可能被资源消耗或编译器漏洞影响；不执行产物并不能排除这些风险。当前服务仍有非root用户和loopback边界，不能据此声称已获得主机root权限或发生逃逸。

## 解决方案

按 [SEC-002](../sandbox/MALICIOUS_CODE_SECURITY_PLAN.md) 先审接口兼容，复用 filejudge 受限编译阶段，逐次新容器/有界产物/清理确认；禁止隔离失败后回落到旧进程。所有编译入口覆盖相同压力、取消与清理测试。完成前不接入任意外部源码；不用源码关键词过滤代替隔离。

Owner：Backend Builder；设计与独立安全审阅：Architect。恢复/关闭要求：受审实现、本地故障注入、正常Linux真实编译矩阵、前端编译按钮联调和无残留证据全部绑定同一源码/镜像。当前真实环境受既有 Docker 问题影响，不能提前关闭本项。
