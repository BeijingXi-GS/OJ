# B-SEC-002：容器配置验证尚未完整闭合

- 发现方式：2026-10-07 至 10-08 只读 `internal/filejudge/docker.go` 的 createArgs / verifyContainer 和 engine.go Preflight；独立 security_gap_review 交叉确认。
- 影响：隔离配置漂移/不符未被全部拒绝的风险；非“已证实Docker可被绕过”。
- 当前事实：创建参数已有约束，验证缺口待整改；无真实攻击或环境变更。

## 前因和已确认结果

当前实现通过固定 Docker argv 创建受限容器，并在启动前 inspect 验证多数属性，还使用实际探针检查关键cgroup、capability、NNP、seccomp、网络及只读根。静态核对发现：bind mount 只比对目的地/只读/数量，未比对 Source；tmpfs 类型一律接受，没有校验目标与选项/大小；ulimit 和 shm-size 已设置但未反查。

`Seccomp:2` 表示过滤模式，无法单独证明具体过滤规则版本和危险调用约束。当前不能将这些已有检查称为“没有保护”，也不能把创建参数的存在当成所有字段的实际执行证明。

## 可能后果与解决方案

当控制面、daemon配置或适配器产生非预期容器配置时，部分漂移可能未触发现有拒绝逻辑。本轮未构造或确认可利用路径。方案是 [SEC-003](../sandbox/MALICIOUS_CODE_SECURITY_PLAN.md)：单一profile声明、生成与验证共用字段、完整mount/tmpfs/ulimit/shm检查、版本化seccomp依赖、故障注入和实际限制探针。

Docker Desktop bind Source 表示可能与Windows客户端不同，应先定义可核对的映射并用只读canary验证，不通过宽松字符串比较解决；未知情形失败关闭。不为通过测试改变宿主ACL或减弱隔离。

Owner：Sandbox Builder；设计/Review：独立Architect/Security。关闭要求：各字段错配fake测试、Linux实际限制/挂载证据、旧题型与通信回归、无残留和独立审阅。Docker恢复之前真实部分维持待验。
