# Docker 控制命令有界输出实现修正

- 日期：2026-10-02；处理人：root，属于用户已授权开发范围内的实现修正；独立审阅：/root/judge_reviewer。
- 状态：已修正并通过专项单测；不涉及主机环境变更，不构成外部阻塞。

## 背景、原因与影响

为恢复 B-JUDGE-002，新增 Docker 控制命令脱敏分类测试，验证 stdout 64KiB / stderr 16KiB 上限。首次 `TestControlErrorClassifiedAndSanitized` 的 stdout_limit / stderr_limit 子测试失败：命令输出超过上限却没有返回错误。

原 cappedBuffer 匿名嵌入 bytes.Buffer，虽覆盖 Write，但提升的 WriteString / ReadFrom 方法可能绕过该 Write；测试中的 io.WriteString 实际走了嵌入类型的方法。这是控制命令输出防护的实现缺陷，不能仅修改测试让它通过。受影响的是 Docker 控制查询缓冲，不是另行实现的运行 stdout/stderr 共享停止预算。

## 实际修正与验证

1. 将 bytes.Buffer 改为私有 data 字段，仅提供 Len / Bytes / Write，所有写入经过限额判定；不提升可绕过的方法。
2. 控制查询在 ctx 已过期限时即使 command 返回 nil，也判为 timeout；成功查询且内容非 2 才是 cgroup_v2_required。
3. 返回固定操作名和错误分类，不输出原始 daemon stderr、argv、路径或凭据；原 5 秒预算保持。
4. root 执行 filejudge 专项单测通过；独立 reviewer 执行包括 deadline_success_nil / stdout_limit / stderr_limit 的 fake 单测通过，并只读复审修正。Linux 同一可信单测亦通过。

无需用户处理。本记录保留开发失败与修正的因果，不把首次失败登记成通过；完整真实执行证据见 [验证记录](../judge/VALIDATION.md)。
