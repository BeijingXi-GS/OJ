# 交接：P1-001 Mini Judge 设计包

- 日期：2026-10-01（Asia/Shanghai）；实际执行者：根Agent / Architect。
- 授权：用户再次要求修复问题并推进工程；此前前端仅交付MD的约定不变。
- 依赖：[Phase0验收](../reviews/2026-10-01-phase0-acceptance.md) 已通过；[七项基础不变量](../architecture/FOUNDATION_INVARIANTS.md)已窄范围frozen。
- 基线：`7a952c6`；分支 `task/P1-001-mini-judge-design-2026-10-01`，保留原未提交内容，未提交 /推送。
- 代码 /配置摘要：56文件 / `469c5c38da2a199f64c6e1e04d16bce4a3e7493145af1e72f97e43154a86bc00`；Phase1本轮只写MD，不修改核心 / API /前端源码或依赖。
- 状态：P1-001设计产物交付，in_review；P1-002独立审阅ready，尚未执行；实现任务planned。

## 已交付与为什么

[DESIGN](../judge/PHASE1_DESIGN.md) 将编译、执行原始证据、ExactChecker、Pipeline和CLI分开；定义monotonic观察时点的timeout /exit /取消规则、完整后代 /FD清理、输出drain与有限报告，防止只看到主进程exit就报告成功。

[ADR-0003](../adr/0003-phase1-local-runner-lifecycle.md) 集中记录竞争与错误归类的拟定决策，供独立Architect裁定；[fixture规范](../judge/PHASE1_FIXTURES.md)给出五类基础样例、compile timeout、取消、父先退出 /setsid、guard、输出 /产物 /报告故障和可控竞争测试。所有测试要求均未执行，不以文档存在当运行证据。

[Agent任务包](../judge/PHASE1_AGENT_TASKS.md) 拆分P1-002和P1-010–017，逐项给出Owner /Review owner、依赖、允许路径、禁止内容、输入输出、具体验收和可复制提示词。前端仍链接已交付的框架实施指南和P2-100–111任务，不把本地CLI当业务提交接口。

## 本轮验证与范围

此前实际修复和验证结果见 [LICENSE修复交接](2026-10-01-license-fix-validation.md) 与 [独立工程复审](../reviews/2026-10-01-phase0-recheck.md)：连续两次构建、doctor /config /up、infra-smoke、真实数据库readiness、check /smoke /probe vet通过，原卷保留。已记录的问题关闭；现有Monaco非阻塞警告保留。

Phase1只验证MD格式、本地链接、任务 /职责 /依赖与代码摘要一致性；没有新增运行源码，因此不重复整个Go /Web构建。最后文档检查结果记录在PROGRESS。设计参考Go1.25.4官方源码和Linux文档，项目策略与默认平台行为明确区分。

最终实际检查：`npm run format:check` exit0，`git diff --check` exit0；71份项目MD的本地文件链接核查无缺失，原始参考资料 /缓存未纳入该链接检查，外网与全部标题锚点未自动验证。代码 /配置摘要仍为上列56文件 /SHA-256；Go /npm锁文件、前端和完整核心源码均未在本轮改变。

未执行：P1设计独立评审、Linux runner preflight、GCC实际锁定、五类真实Judge、timeout /cleanup测试、Phase1race、安全测试、远程CI。当前Docker kernel /cgroup信息不等于runner已获权限或cgroup控制测试通过。

## 下一位的第一步

P1-002独立Architect读取完整设计包，裁定timeout观察策略、唯一reaper /launcher协议、guard归类、默认预算和CLI报告；通过后仅冻结Phase1实际采用的局部规则。然后P1-010以真实runner用户锁定Linux /compiler /profile并验证全部前置，才可认领编译器、runner和CLI实现。

新问题继续由用户接管；若阻塞则单开详实MD并停止。当前没有新增阻塞问题，待审设计与待验证环境是已明确的实现前置，不冒称已完成能力。
