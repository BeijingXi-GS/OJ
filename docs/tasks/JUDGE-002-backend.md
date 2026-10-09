# JUDGE-002 文件判题后端与运行适配

- Owner：root /Backend Builder；Review：独立Architect +Security reviewer。
- 修改范围：internal/filejudge、services/judge-mvp、services/judge-compile-helper、internal/httpapi、apps/api、deploy/judge、scripts、testdata及根说明；不修改Frontend Builder的apps/web。
- 依赖：JUDGE-001独立设计审阅通过，采用judge-mvp/0.1.0局部契约。
- 交付：[契约](../judge/JUDGE_MVP_CONTRACT.md) 的文件加载 /快照 /checker、Docker受限编译 /逐例、证据映射 /清理taint、HTTP client /worker /API及可启动环境。
- 检查：单测输入 /错误 /checker /快照 /生命周期，真实Docker verdict与隔离 /清理 /取消矩阵，API smoke和完整check。
- 禁止：无隔离fallback、API持有Docker权限、容器答案 /socket挂载、用户flags、任意主机删除 /改设置、基础数据卷修改或伪造计量。
- 2026-10-02交付上述实现与17项真实矩阵、完整check /回归 /Linux单测；独立实现代码审阅通过。最终重启受B-JUDGE-002阻塞，已停止；实际运行与未测范围见 [验证](../judge/VALIDATION.md) / [交接](../handoffs/2026-10-02-file-judge-mvp.md)，实时状态仅PROGRESS。

以上是上轮停止记录。本轮用户重新授权后，JUDGE-005/006已恢复该依赖并扩展五题，严格重启/最终AC/原矩阵回归通过；本轮事实见 [新交接](../handoffs/2026-10-02-judge-recovery-problems.md)。
