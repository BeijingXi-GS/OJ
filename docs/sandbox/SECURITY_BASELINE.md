# Sandbox 安全基线

状态：设计约束已整理，Linux 实现和 Security Review 未完成。

用户源码、编译输入、Checker 和 Interactor 都需要明确的信任边界。Phase 1 的本地 runner 只处理仓库维护者提供的可信 fixture，在专用 Linux 开发环境执行；在 Phase 4 安全门禁通过之前，不接受互联网来源任意代码。

## Phase 4 必须落实

- PID / mount / user / network / IPC namespace；默认无网络。
- cgroups v2：Job 父层聚合限制、角色子层限制、编译独立预算。
- seccomp、capability dropping、只读根、受限 tmpfs、白名单文件映射。
- CPU、wall time、memory、PID、stdout/stderr、磁盘临时空间都有上限。
- signal/exit/timeout 竞态下回收完整进程树；不残留 child / zombie。
- 计量证据可信，OOM 明确归属到 Role 或 Aggregate；不可信量测返回基础设施错误。
- 编译器也隔离；不把编译阶段视作可信主机操作。
- stdout/stderr 和 transcript 截断，不暴露隐藏测试、宿主路径和密钥。
- Runtime 接口无法保证隔离时拒绝执行，没有普通 subprocess 的隐式降级。

## 验证矩阵

必须验证无限循环、无限内存、无限输出、fork bomb、文件膨胀、网络访问、宿主敏感文件读取、子进程逃逸、异常系统调用、父先退出、超时与退出竞争，以及多角色聚合 OOM。每例记录输入、环境/内核、期望、观测、资源值和清理结果。

这些用例由 Architect 设计，Builder 完成真实实现与集成测试，Worker 扩展 fixture。当前仓库不自动运行上述攻击型测试，也不将 Compose 当作已验收的选手 Sandbox。
