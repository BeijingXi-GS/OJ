# P1-002：Mini Judge 独立设计评审

- 实时状态：见PROGRESS.md；Owner：独立Architect；实际执行者：未认领。
- Review owner：授权集成执行者；风险R2。
- 依赖：P1-001设计包交付，Phase0验收完成。
- 输入：[DESIGN](../judge/PHASE1_DESIGN.md)、[ADR-0003](../adr/0003-phase1-local-runner-lifecycle.md)、[fixtures](../judge/PHASE1_FIXTURES.md)、[任务包](../judge/PHASE1_AGENT_TASKS.md)及基础不变量。
- 允许修改：独立REVIEW、明确指出的设计修订、局部冻结记录及任务输入。
- 禁止：冒称Phase0评审覆盖本设计、代写整阶段、修改未受审核心源码、执行外部源码、自动实施前端。

## 必须裁定

观察时点和timeout /exit /cancel顺序是否一致；stop candidate与最终verdict分离、CPU最终读取、开发guard归类是否可测试；launcher /单reaper协议、cgroup起始归属、完整清理、输出drain、工作根安全与profile预检是否可实现；默认预算、CLI /report边界和真实Linux验收矩阵是否明确。

输出实际独立性、版本 /摘要、通过或需修改及排除范围。通过后仅冻结实际采用的Phase1局部契约；compiler实际版本和环境能力仍由P1-010取得真实证据。未通过不进入实现，不以作者自检登记独立审阅。
