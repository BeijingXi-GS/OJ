# Phase 1 fixture 与验收规范

- 状态：proposed；作者：根 Agent / Architect，2026-10-01；关联 P1-001 / P1-014 / P1-016。
- 这是给 Worker / Builder 的制作要求；当前没有创建这些 C++ 文件、fixture manifest或运行测试。
- 输入契约：[DESIGN](PHASE1_DESIGN.md) / [ADR-0003](../adr/0003-phase1-local-runner-lifecycle.md)，须先经 P1-002 独立审阅。

## 1. 文件、manifest与可信来源

后续 Worker在 `fixtures/phase1/` 建立源码、inputs、expected和suite manifest。manifest声明 suite_id / version、case顺序、源相对路径、role=contestant、language=cpp、每份内容SHA-256、输入 / expected路径及hash、目标verdict、reason和重复次数。路径只指向该fixture根；拒绝重复case、越界路径 / symlink及hash不一致。

标准基准输入为UTF-8字节 `1 2\n`，expected为 `3\n`；expected bytes恰为 `33 0a`（十六进制）。Worker生成hash后必须复核实际bytes，不能只用编辑器显示。manifest锁定测试数据，profile另锁实际compiler / runtime / flags /预算，不伪造尚未测过的版本。

fixture均由维护者审阅后入allowlist。环境 / cleanup / 编译hang的test helpers同样版本化并锁hash，不能把任意用户路径或shell command塞进测试profile。

## 2. 必须实现的五类fixture

| ID        | 源码行为                                              | 必须断言                                                                    | 执行次数 |
| --------- | ----------------------------------------------------- | --------------------------------------------------------------------------- | -------- |
| F-AC      | 读取两个小范围整数并输出和及 LF                       | AC，exit=0，stdout字节完全匹配，未截断，版本与artifact绑定正确，cleanup完成 | 10       |
| F-WA      | 相同输入，输出错误值 `4\n`                            | WA，exit=0，mismatch offset可解释，不能因stderr正常就AC                     | 10       |
| F-CE      | 确定语法错误，例如缺失表达式的 `int main(){return ;}` | CE，真实gcc非零，诊断有界，不存在可用artifact，run stage未启动              | 10       |
| F-RE      | 使用标准 `raise(SIGSEGV)` 明确发出信号                | RE，可信raw signal=SIGSEGV，不能用不可预测的UB / null优化代替确定触发       | 10       |
| F-TLE-CPU | 可观测副作用的无限unsigned循环，避免优化掉整个循环    | TLE，CPU或wall监督cause及limit / observed存在，不单断言exit137；cleanup完成 | 10       |

CE示例最终须由Worker在固定compiler下确认；若compiler容许某语法，换用确定的非法token，不把warning当CE。无限循环只在已经验证保护和清理的可信Linux环境运行。

## 3. 真实生命周期与故障矩阵

| ID                 | 输入 / 注入方式                                                          | 期望及证据                                                                                                      |
| ------------------ | ------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------- |
| L-WALL             | 已登记fixture只sleep，远超过wall预算                                     | TLE / WALL_TIMEOUT，CPU低，返回有界                                                                             |
| L-COMPILE-TIMEOUT  | 内部test profile使用受审的真实Linux helper，按握手阻塞至compile deadline | CE / COMPILE_RESOURCE_LIMIT，无artifact，整个stage清理；记录实际helper身份，不能冒称gcc本身编译此源超时         |
| L-CANCEL-COMPILE   | 在可控compile helper运行后发送取消                                       | CANCELLED且完整清理，或已先发生的终态cause；不得启动run                                                         |
| L-CANCEL-RUN       | 在长运行fixture启动握手后取消                                            | CONTROL_CANCEL，完整清理后CANCELLED；后发取消不得改变已完成结果                                                 |
| L-CHILD            | root退出0，子进程sleep并持有stdout                                       | RE / CHILD_REMAINS；不能因为root正常就AC或无限等EOF                                                             |
| L-SETSID           | 注册fixture创建改变session / group的后代并sleep                          | 仍可完整清理，不能只依赖原PGID；group空 / wait回收证据完整                                                      |
| L-DUAL-STREAM      | stdout和stderr交替写入，超过一次pipe容量但低于限额                       | 不死锁，流不混合，字节计数正确                                                                                  |
| L-OUTPUT           | 受控持续stdout超过停止线；另测stderr超过其线                             | SYSTEM_ERROR / DEV_RESOURCE_GUARD，scope / observed / limit，prefix有界 / truncated=true，无残留；不猜已完成OLE |
| L-EMPTY / L-BYTES  | 空stdout、末尾LF差异、CRLF、NUL、前后空格                                | 按Exact bytes得到明确AC或WA，不trim / 统一换行                                                                  |
| L-TWO-CASE         | 同一artifact两case；另测首case WA                                        | 编译一次，逐case独立预算，次case真实执行；首错时剩余SKIPPED                                                     |
| L-ARTIFACT         | 内部test hook篡改artifact bytes或绑定                                    | SYSTEM_ERROR / ARTIFACT_INVALID，run未启动；不重新读取 / 复用旧产物                                             |
| L-MISSING-COMPILER | test profile指定不存在的受控compiler标识                                 | preflight拒绝、CLI exit2，无payload / verdict，不伪装CE                                                         |
| L-MISSING-CGROUP   | 隔离test环境中取消所需能力 / 委派                                        | preflight拒绝，无不受控回退；不得改宿主全局权限                                                                 |
| L-IO-FAIL          | 受控内部hook使输出writer失败                                             | SYSTEM_ERROR / IO_FAILURE，停止 / cleanup，不比较残缺stdout                                                     |
| L-CLEANUP-FAIL     | 测试adapter在隔离环境模拟cleanup失败；外部watchdog仍收尾                 | SYSTEM_ERROR / CLEANUP_INCOMPLETE，runner隔离、下一case不启动；不得制造真实宿主残留                             |
| L-REPORT-FAIL      | 不覆盖已有report、受控writer失败                                         | exit2（参数预检）或exit3（运行后写失败）；不声称报告已持久化                                                    |
| L-GUARDS           | 受控验证memory / PID / temp保护                                          | 明确保护前置 / DEV_RESOURCE_GUARD、清理；不能作为Phase4恶意逃逸测试通过证据                                     |

故障hook是内部测试依赖，不能通过公开CLI参数切换到任意命令或关闭限制。测试使用真实受控进程验证启动和清理；纯mock不替代真实child / cgroup / I/O证据。

## 4. 超时竞争的确定性测试

先对arbiter / reducer提供fake monotonic clock与可控事件：EXIT观察在D−1、D、D+1；timer先唤醒 / exit先唤醒；取消前后；CPU与wall同快照；exit候选后发现真实资源超限；cleanup失败覆盖取消；重复timer / cancel。每组合运行1000次，断言stop只登记一次、final只发布一次、字段一致。

事件中的时间必须按ADR的观察 / 提交规则生成，不把消息排队时间偷换成payload的实际退出时间。迟来的EXIT或取消不能改变已发布结果。

真实Linux运行临界exit fixture100次：允许调度边缘按既定规则产生AC或TLE；必须每次在预算和cleanup范围内结束、唯一终态、无残留。另有明确早于 / 晚于deadline的样例，时间裕量至少5个采样周期，稳定覆盖正常退出与TLE；不要以不稳定sleep精度作唯一正确性证明。

## 5. 每次运行的清理断言和证据文件

每个启动过payload的样例都必须记录：专属cgroup ID / empty观测、等待回收记录、`reaper_pending=0`、`owned_open_fds=0`、`io_copy_workers=0`、无活跃stage；正常路径也验证这些字段。group populated=0不单独证明无zombie。隔离stage supervisor自身退出并被CLI回收；工作目录按保留策略处理，不删除仍被使用目录。

资源证据写清CPU来源 / 累计值、wall观察区间、memory与output的观测scope、limits及未知值原因。不得以各process peak之和冒充Job真实peak；不得以杀死主PID冒充子树回收。

测试报告至少有 environment/profile digest、commit或代码摘要、fixture /数据hash、compiler / runtime版本、case、repeat序号、期待与实际、raw exit / signal、stop reason / seq、verdict、cleanup、截断与耗时。stdout JSON单对象、stderr有限、退出码与结果一致。汇总文件保留passed / failed / skipped，而非只写“全绿”。

拟运行入口由P1-016提供，例如 `npm run judge:test:linux`，**当前尚不存在**。Linux实现阶段须运行普通Go测试、适用race、上述真实矩阵以及API回归，确认HTTP仍没有执行依赖。首次真实残留 / 环境失败按项目规则单独记录并停止；当前这些测试全部未执行。
