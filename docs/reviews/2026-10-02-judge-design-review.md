# JUDGE-001 独立架构与安全设计审阅

- 日期：2026-10-02（Asia/Shanghai）。
- Reviewer：`/root/mvp_review`，未参与判题实现；职责为 Architect / Security Review。
- 范围：[judge-mvp/0.1.0](../judge/JUDGE_MVP_CONTRACT.md)、[ADR-0005](../adr/0005-file-judge-docker-oci.md)、本机同步文件判题的实施前约束。
- 当前结论：**设计审阅通过，按本报告及修订契约的硬性约束进入实现。** 已复读 root 保存的修订版，关键隔离 / 快照 / 输出 / 预算修改已采用；D2 普通文件与唯一 case ID、D4 cleanup 总预算、D5 监督触发导致 attach 取消的例外、D6 非空非 symlink artifact 作为本次采用的补充约束，必须落入实现并由 JUDGE-004 复核。设计通过不等于执行服务验收通过，不授权缺项时回退无隔离运行。
- 本报告不替代后续 JUDGE-004 实现审阅、真实隔离 / verdict 矩阵及浏览器验收。
- 本次复读的契约 SHA256：`7d78af83c69aa5c864b064075cb1d5127ddcd5e6bdfb52e28866a58e6915a568`；ADR-0005 SHA256：`85e714820d55c9d86df1f11e11b2e5956faf8bc194f7fb8643b021c3bf1096a6`。后续仅登记 reviewed 状态或将补充约束写入正文会改变 MD 摘要；不能据此扩大本次评审范围。
- 2026-10-02 后续复核：契约 SHA256 `523aaafd3717e7ce95fd34e41214fcff05df9a14d86dbef3567d5a812731c1b4` 已将末轮普通文件 / 非 symlink 非空 artifact、70s任务与15s最终清理、监督触发后的 attach 取消例外写入正文，状态为 reviewed；**正式设计通过结论保持有效，所列补充条件已纳入保存的契约。** 本次继续开展 JUDGE-004 实现审阅，仍不将设计通过视为实际执行验收通过。

## 实际阅读与验证

已阅读 AGENTS 最新用户授权、PROGRESS 当前任务、DEVELOPMENT 相关流程、JUDGE-001 / JUDGE-004 任务卡、契约与 ADR。参考副本为 `gen/inputs/Anolix`，独立执行 `git rev-parse HEAD`，结果为 `4502d2a798f3130d3c540c5c31d45251ca103623`。

已阅读参考的 launcher/spec.go、launcher/launcher.go、config/config.go、launcher/launcher_test.go、config/config_test.go 与 go.mod。参考实现确有私有 namespace、空 capabilities、NoNewPrivileges、可配置 seccomp 与 memory / pids / CPU 资源映射；rootfs 的 Readonly 为 false，memory swap 尚未映射，forceDeleteContainer 和自动 bundle 删除忽略错误。launcher 生命周期测试使用 fake runc，不能作为真实 cgroup、隔离或后代清理通过的证据。未运行 Anolix 测试或未知 payload，也未复制其代码 / 引入依赖 / 升级 Go。

独立只读执行：

```text
docker image inspect oj/compile-worker:gcc14.2.0-mvp --format '{{json .Config.Volumes}}'
=> {"/data":{}}
```

因此原 compiler 镜像的声明卷必须在 judge 镜像裁定中处理，不能将只读 rootfs 当作所有可写位置已受限。root 提供本轮严格 runtime probe 通过记录：memory.max=67108864、swap.max=0、pids.max=16、cpu.max=100000 100000、CapEff=0、NoNewPrivs=1、Seccomp=2。该 probe 是实施者证据，本 reviewer 未重复执行，也不将它扩大为完整安全验收。

## 采用版的硬性实施约束

### D1：编译与运行共享隔离底线，封闭继承的卷和日志

初稿只在运行段完整列举隔离参数，修订契约已明确编译同样使用 UID10002、init、network none、cap-drop ALL、NoNewPrivileges、默认 Docker seccomp、私有 PID / IPC / UTS / mount / cgroup namespace、只读 rootfs；编译允许的工作 tmpfs 与阶段资源值另设。编译本身属于不可信计算，不能使用宽松的控制服务环境。

采用 root 已确认的镜像方案：工具层建立固定挂载占位，最终 `FROM scratch` 复制工具 rootfs 与 helper，仅设必要元数据，避免继承 Redis `VOLUME /data`。启动拒绝 Config.Volumes 非空；每次 create 后核对实际 mount 无 volume 类型、无 Docker socket、无答案或敏感宿主目录。所有容器设 `--log-driver none`，禁止另外产生无界的 daemon 输出日志；删除使用 `rm -fv` 并确认实际容器不存在。Docker 的声明卷和日志配置需要分别处理，不能由 stdout 截断或 readonly rootfs 替代。[Dockerfile VOLUME](https://docs.docker.com/reference/dockerfile/#volume)、[日志驱动](https://docs.docker.com/engine/logging/configure/)、[删除容器与匿名卷](https://docs.docker.com/reference/cli/docker/container/rm/)。

镜像固定实际 image ID、显式固定入口与 argv，不继承旧 compile-worker 入口；不使用 privileged、host namespace、额外 capability 或 seccomp unconfined。本机非生产边界声明保持有效。

### D2：实际挂载必须来自本次不可变快照

只读挂载原 testdata 文件仍允许管理员改动影响本次运行，无法保证所报 hash 对应实际输入。source / manifest / case 文件限额读取一次后，使用本 run 的固定字节生成仅服务可写的 source 和 input 快照；容器仅挂各自的普通文件，不挂 testdata 或整个可写 job 目录，expected 只保存在可信 checker 一侧。

testdata 根、manifest、case 文件应明确拒绝 symlink 和非普通文件；manifest case ID 唯一，文件名拒绝路径分隔符、绝对路径、`.` / `..` 及 Windows 特殊路径语义。`os.OpenRoot` 限制路径逃逸，但会跟随仍留在根内的符号链接，也不会自动拒绝设备等特殊文件，因此仍需独立类型检查。可信维护者不能在一次快照读取过程中并发修改目录的约束应写明，不将 OpenRoot 描述为整个文件读取的原子快照。[Go os.Root](https://pkg.go.dev/os#Root)。

工作目录由服务以随机 run ID 唯一创建；挂载路径是生成并验证的绝对路径。root 已确认采用 `encoding/csv.Writer` 生成单独 argv 中的 `--mount` 参数，避免含逗号 / 空格的 Windows 路径被解析成额外 mount 配置；不拼 shell 命令、不自动创建用户提供的宿主路径。应真实预检 Docker Desktop 对该目录、只读文件及可执行 artifact 的支持，失败不能回退主机执行。[Docker bind mounts](https://docs.docker.com/engine/storage/bind-mounts/)。

### D3：checker 使用原始完整字节，展示限额独立

修订契约已区分原始输出丢失和展示截断。stdout 内部保留到共享原始 1MiB 停止线，checker 使用完整原始 bytes；HTTP stdout 仅展示64KiB、stderr16KiB，展示截断本身不改变 AC / WA。原始停止线必须共享原子计数，stdout / stderr 并发写入不能各自多保留一份无限数据。

修订契约已确定 `exact-v1` 为 bytes.Equal；`tokens-v1` 为 bytes.Fields 按 Unicode 空白切分、逐 token bytes.Equal 且 token 数量相等，忽略空白量但不解析数值。不要在 checker 前修复 UTF-8 或做文本归一化；此语义与 Go1.25 的 bytes.Fields 一致，不另改为 ASCII 空白规则。

修订契约采用 strings.ToValidUTF8 后按展示的 UTF-8 字节上限安全截断，不能把多字节字符切断；响应枚举和每项输出限额校验须与此一致。二进制或无效 UTF-8 输出不能因为 JSON 替换后变长而被 API 误判协议失败。

### D4：全局预算、控制调用及清理必须分别有界

采用 root 已确认的预算：job 总70s包含快照后控制阶段、compile 与所有 cases；create / inspect / kill / rm / 查询等每个 CLI 控制调用≤5s；cleanup 独立总预算至少15s，即使 caller 取消或 job 超时仍能运行；API client110s、API / worker WriteTimeout120s、shutdown120s。这些上限不表示每个 case 都可耗满预算：超过全局 deadline 必须停止后续 case。

仅 case wall deadline 触发归 TLE；job 全局 deadline 或 caller 取消归 SYSTEM_ERROR，不把基础设施开销耗尽总预算写成用户程序 TLE。case wall 包含 start --attach 发起后的启动 / 传输监督开销，必须在页面明确，不能当 CPU 时间。

create 前记录随机 name / label。create 超时或响应丢失并不证明容器没有创建，仍必须在独立 cleanup 中按该标识查询和清理。只有成功的 daemon 查询证明无匹配容器才可认定不存在，不能把连接失败、取消或任意 inspect 错误当“已删除”。控制子进程、stdout / stderr drain 及等待也必须在限额内收束。

### D5：完整输出、最终状态及清理是判定的证据门槛

Docker CLI 的退出码不能代表选手 ExitCode；可信 inspect 应确认容器确已启动、最终 not-running、无 State.Error，并读取 ExitCode 与 OOMKilled。**已记录的 case wall / OLE / caller cancel / job deadline 导致 CLI 或 attach 被监督器终止时，应按该触发原因与最终状态判定；不能因为主动终止造成 CLI ExitCode 与 container ExitCode 不同而覆盖为未知传输错误。** 无可信监督触发的 attach / drain / wait 基础设施错误或无法证明输出完整，先归 SYSTEM_ERROR；不能仅凭 inspect exit0 和一个匹配前缀判 AC。

记录本次监督原因，最终按 SYSTEM_ERROR → OOM MLE → case wall TLE → 原始输出 OLE → 非零 exit RE → checker AC / WA。保留 OOM 与 timeout / OLE 的原始事实，kill 后的137不覆盖先前原因；普通 `return 137` 且无 OOM 证据必须 RE。请求取消和全局预算耗尽优先 SYSTEM_ERROR。事件应一次性收束，不能因为清理时钟已超时重新制造 case TLE。

每个容器清理失败或 job 目录删除失败都 taint，拒绝后续请求。必要时保留 job 目录供用户接手；若仍可能有容器使用 mount，不应先删宿主快照。重启发现本项目标签残留拒绝 ready，不能静默复用或删除其他项目容器。shutdown 应先停止接收新请求，再等待 / 取消当前 job，并完成同一清理协议。

### D6：artifact 通道与 helper 失败不能混入 CE

helper 固定调用 g++，将其所有诊断送 stderr，stdout 专用于有界 artifact bytes；成功导出前确认本次生成的是普通非 symlink 文件、非空、≤32MiB。控制服务只写随机 run 目录的固定 artifact 文件，完整输出、最终容器 exit0 和清理成功后再封存并计算 SHA256，运行容器只读挂载该产物；root 已确认 artifact 权限0555和预检实际 exec。

helper 自身文件 / I/O / 导出 / 启动编译器失败及异常终止必须保留 SYSTEM_ERROR 路径，不能全部映射为“源码 CE”；正常 g++ 语法 / 链接失败才 CE。返回中或内部证据应明确记录 artifact SHA256 与所有 cases 的同一绑定。编译超时 / OOM / 原始诊断超线为编译阶段 SYSTEM_ERROR，且不运行任何 case。禁止复用前次成功 artifact 或在失败时继续导出旧文件。

## 保持的正确裁定

- API 仅固定地址 HTTP 桥接，不拥有执行 payload 的路径；独立 judge 控制服务只监听127.0.0.1，禁止用户指定路径 / Docker 参数 / 镜像 / 命令 / 限额。
- 独立编译域与每 case 运行域，固定镜像 ID、全部 capability 丢弃、NoNewPrivileges 与受检 seccomp / cgroup；memory-swap 与 memory 同值以禁 swap，不能写 `--memory-swap 0` 代替。[Docker 资源约束](https://docs.docker.com/engine/containers/resource_constraints/)。
- MLE 仅由本容器真实 OOMKilled 证据支持；CPU累计和可信内存峰值未采集时返回 null。
- CE 不运行 cases，SYSTEM_ERROR停止，其他业务失败继续；总结果按 manifest 顺序第一非 AC，SYSTEM_ERROR优先。
- 旧编译接口兼容；前端编译 / 运行锁共享，源码变化标过期，离页取消，未知资源显示未采集。旧浏览器问题仍需复验。

## 实施后必须提供的证据

JUDGE-004 必须核对最终源码与镜像 / profile 标识，实际 API AC / WA / TLE / MLE / CE语法与链接 / RE崩溃与非OOM137 / OLE，以及 token / exact 边界。专项测试还应覆盖原始输出正确但展示截断、无效 UTF-8、源文件与输入快照不随原文件变化、路径含空格 / 逗号、helper 部分导出、create 不确定失败、attach不完整、全局deadline与取消、cleanup失败taint、后代进程和容器 / 卷 / job目录无残留。

runtime probe 须核对实际 image ID、镜像无声明卷、host config隔离、mount allowlist、cgroup / NoNewPrivs / Seccomp / CapEff，并证明只读rootfs、无Docker socket / 答案、无网络及用户权限限制。失败必须拒绝服务 ready / 执行，禁止普通 exec 回退。

本次只完成设计审阅和所列只读命令；未运行判题 payload、真实 verdict 矩阵、取消 / 清理故障注入、浏览器、race或远程CI。修订是实施前设计反馈，不作为新环境阻塞；如后续出现实际新阻塞，仍按用户一事一档、停止规则处理。
