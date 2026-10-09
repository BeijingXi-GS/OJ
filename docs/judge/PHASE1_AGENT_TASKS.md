# Phase 1 Agent 实施任务包

- 版本：mini-judge/0.1.0-draft；状态：设计提案，尚未冻结 / 实施。
- 编写者：根 Agent / Architect，2026-10-01；依据 [DESIGN](PHASE1_DESIGN.md)、[fixture规范](PHASE1_FIXTURES.md)、[ADR-0003](../adr/0003-phase1-local-runner-lifecycle.md)。
- 实时认领与状态只在 [PROGRESS.md](../../PROGRESS.md)。本文件是任务输入，不自动启动Agent或授权它绕过门禁。

## 1. 依赖与角色

| ID     | 任务                                  | Owner角色                    | Review owner                     | 依赖 / 当前准备情况                      |
| ------ | ------------------------------------- | ---------------------------- | -------------------------------- | ---------------------------------------- |
| P1-001 | 生命周期、错误与验收设计包            | Architect / 根Agent          | 独立Architect，P1-002            | Phase0通过；本包是输出                   |
| P1-002 | 独立设计评审与局部契约冻结            | 独立Architect                | 授权集成执行者                   | P1-001设计已交付后ready，当前未认领      |
| P1-010 | Linux / compiler profile与能力预检    | Builder-Environment          | Architect                        | P1-002通过后planned                      |
| P1-011 | 源快照与C++ Compiler                  | Builder-Compiler             | Builder；资源 /错误边界Architect | P1-002/010                               |
| P1-012 | Linux stage supervisor / local runner | Builder-Runner               | Architect                        | P1-002/010；角色绑定与产物输入提案已冻结 |
| P1-013 | Exact Checker与结果映射               | Builder-Pipeline             | Builder；映射规则Architect       | P1-002；集成依赖P1-011/012               |
| P1-014 | 标准 /生命周期fixtures与manifest      | Worker-Fixtures              | Builder-Tests                    | P1-002；真实生成 /执行依赖P1-010         |
| P1-015 | Pipeline / CLI / report集成           | Builder-Pipeline             | Builder + Architect              | P1-011/012/013/014                       |
| P1-016 | Linux矩阵、race与故障证据             | Builder-Tests + Worker-Tests | Architect                        | P1-015与受审fixture                      |
| P1-017 | Phase1独立复审与验收                  | 独立Architect / Builder      | 授权集成执行者                   | P1-016与全部适用检查通过                 |

同目录并行时指定文件所有权；Builder-Runner拥有 supervisor生命周期 / cgroup / reaper文件，Builder-Pipeline拥有CLI /报告 /映射文件，Worker-Fixtures只修改fixture和测试输入。角色不绑定模型；未经认领不得覆盖其他人的文件。以上只是未来分工，没有委派实现Agent。

<a id="p1-002"></a>

## 2. P1-002：强Agent先审设计

读取Phase0七项基础不变量、P1-001全部输出和现有Graph / Runtime草案。必须回答：观察时点超时策略是否可实现、取消与exit顺序、资源最终读取与verdict优先级、subreaper /唯一reaper协议、atomic cgroup归属、output drain、cleanup失败隔离、可信fixture边界、profile前置、报告与CLI是否可测试。

输出独立REVIEW及明确修改项；通过后只冻结实际采用的Phase1局部DTO / budget /错误与生命周期规则，不整体冻结未来Graph / Runtime。可以修订DESIGN / ADR /任务卡，禁止代写整个runner或冒充Linux测试已通过。未审就保持proposed，编译器实际版本仍须P1-010锁定。

提示词：`你负责P1-002独立Architect评审。核对DESIGN、ADR-0003和fixtures的可实现性与全部竞争 / 清理边界，记录实际独立性、版本、通过或修改项以及冻结范围；不实现执行器，不把基础原则审阅当作本设计审阅。新阻塞写独立问题MD并停止。`

<a id="p1-010"></a>

## 3. P1-010：Builder先证明真实环境可用

允许：受控开发profile、`deploy/local-judge/`、`scripts/judge-preflight.*`及环境说明。输出profile.lock.json与真实preflight报告，锁Go / GCC / libc /runtime digest /kernel /arch /uid /根目录；验证DESIGN第2节的cgroup能力、原子归属、资源限制、清理 / reaping、tmp与输出保护。

不修改主机全局WSL / Docker设置或安全策略，不挂Docker socket给payload，不删卷；需要用户处理的权限 / 下载 /环境问题一事一档，阻塞则停止。preflight缺项返回非零且不启动payload。仅基础Compose healthy不能通过本任务。

验收：以最终runner用户运行预检，记录每项观察 /错误；锁文件能标识真实compiler和runtime，不使用latest代替digest；在受控临时工作根验证新建 /删除不会触及项目数据。单独测试不受支持profile的拒绝路径。

提示词：`你负责P1-010。先阅读冻结后的设计，再锁定实际Linux编译环境与能力报告。只证明前置，不实现判题；任何未知 /无权限能力应拒绝并详细记录，不能用主机提权或放宽限制绕过。`

<a id="p1-011"></a>

## 4. P1-011：Builder实现编译边界

允许：`internal/devrunner/compiler*`、snapshot /artifact实现及其测试。输入受审LocalRequest和profile；固定argv、环境与工作根，复制不可变源快照，调用已审supervisor编译，保留有限诊断。产物显式绑定Run /Source /版本 /hash；失败不能返回或复用旧artifact。

验收：真实AC源编译得到可运行产物；真实语法错误CE；compiler缺失 /无法启动是INFRA；受控编译timeout、输出超限与取消；源 /artifact篡改拒绝；编译结束无后代或FD残留。不要把编译资源超限当运行TLE。

禁止：改变SourceUnit快照、在API内编译、接受用户compiler flags、写业务表或改冻结状态机。提示词：`按已审Compiler契约完成快照、版本记录、产物校验与全部编译失败路径；不能用旧产物掩盖CE，必须提供真实Linux编译和清理证据。`

<a id="p1-012"></a>

## 5. P1-012：Builder实现监督器与runner

允许：`internal/devrunner/supervisor*`、`runner*`、Linux平台适配及专项测试。独占生命周期、arbiter、cgroup、launcher /reaper、输出和cleanup实现；不能把现有sandbox.Runtime草案冒充生产隔离实现。

验收：独立case组、atomic归属、原始wait /signal与stop cause分离、CPU /wall检查、输出有界且双流无死锁、取消、root先exit / setsid后代、cleanup2秒预算、真实失去能力时fail closed。由Architect重点review竞争和清理；无真实child /cgroup证据不验收。

提示词：`按冻结生命周期设计完成Linux监督，不自行改timeout /cancel /映射策略。唯一reaper拥有wait，所有成功与失败路径先证明全树结束再返回；不能只kill父PID或依赖CommandContext当完整清理。`

<a id="p1-013"></a>

## 6. P1-013：Builder实现Checker与映射

允许：`internal/devrunner/checker*`、`result*`及对应纯单元测试。Exact按完整原始bytes比较，禁止trim /换行归一化；实现固定错误优先级，输入不可信时拒绝比较。只消费执行证据，不启动进程。

验收：空输出、NUL、CRLF / LF、末尾换行 /空格、首差异offset、stdout截断拒绝；验证CE/RE/TLE不从通用error或137猜出；cleanup /计量失败覆盖成功候选；取消 /guard策略与ADR一致。不要扩充SPJ /浮点 /Interactor。

提示词：`按已审ExactChecker与映射表实现字节比较和用户 /控制 /基础设施归类。测试行为差异和不可信证据拒绝，禁止用镜像实现逻辑的无意义测试替代边界用例。`

<a id="p1-014"></a>

## 7. P1-014：Worker制作fixture

仅修改 `fixtures/phase1/` 与fixture文档 /受审测试helper输入。逐项按PHASE1_FIXTURES创建源 /stdin /expected /manifest，算真实SHA-256，明确每例目标reason与cleanup断言。基础五类使用真实GCC；compile hang helper身份单独标注，不冒称真实gcc超时。

验收：每份字节 /hash可复核，明确role和case顺序；CE在固定compiler下确实失败、RE确定SIGSEGV；expected newline严格；可信来源经Builder审阅。Worker不修改runner /cgroup /核心接口，也不运行未保护的无限循环或fork bomb。

提示词：`你只制作P1-014可信fixtures和锁定manifest，逐项记录目标verdict /reason与bytes/hash。不要改变执行器为fixture通过开后门，未有受审保护环境时不执行无限循环。`

<a id="p1-015"></a>

## 8. P1-015：Builder集成最小CLI

允许：`apps/judge-cli/`、pipeline、report与该模块脚本。加载manifest /profile →编译一次→case串行→首错短路→清理→单JSON终态；CLI stdout与日志分开，exit0/1/2/3/130符合DESIGN，report不覆盖、原子发布、有界保留。

验收：五类真实CLI闭环、两case与SKIPPED、Ctrl+C /报告writer失败、字段完整、源与版本追溯；HTTP route和import仍无执行器依赖。禁止实际前端提交、DB /Queue接入或任意源码模式。

提示词：`集成已通过模块完成单机可信fixture CLI，只增加本阶段命令与有限报告。不得在API /Web提供执行入口；一个Run只发布一份终态且必须带清理证据。`

<a id="p1-016"></a>
<a id="p1-017"></a>

## 9. P1-016 / P1-017：测试与验收

Builder-Tests管理真实Linux生命周期测试，Worker-Tests可扩展fixture参数 /脚本与结果汇总，Architect审阅竞争 /清理 /错误证据。按fixture规范完成五类各10次、真实临界退出100次、确定性事件排列各1000次及所有异常路径；执行适用Linuxrace与项目基础API回归。

测试结果绑定代码 /profile /fixture摘要，逐例保留期待、实际、raw evidence、资源和清理，不把mock或已有基础服务healthy当runner安全验收。发生真实残留 /新阻塞立即按规则记录并停止。

P1-017由未参与实现的reviewer审阅最终diff与真实证据，明确Phase1完成范围和未实现的生产Sandbox；不能由实现者改角色名称代替独立性。通过后登记Phase1验收、未测项与回退，才设计Phase2业务事务 /API /授权。

## 10. 前端Agent继续使用已有详细指令

本次没有实施前端。Frontend Builder /Worker的目录、provider、router、请求层、DTO、mock、Monaco、页面状态、轮询、业务联调与E2E任务仍在 [前端框架实施指南](../frontend/IMPLEMENTATION_GUIDE.md) / [P2-100–111任务包](../frontend/AGENT_TASKS.md)。先完成P2-100方案审阅，FE-A框架与mock范围、FE-B真实API闭环分别验收；不得用当前基础API或本地CLI伪造提交业务已可用。
