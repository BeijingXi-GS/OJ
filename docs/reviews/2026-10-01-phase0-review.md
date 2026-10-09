# P0-004 / Phase 0 独立审阅

- 被审版本：基线 `7a952c6`，分支 `task/P0-005-minio-source-2026-10-01` 的当前工作区。
- 独立复核的代码 / 配置摘要：56 文件，SHA-256 `95e401d13d7692a1d1c2f5872299ee7ada68226a583ad434233ad9dcb95d584f`。MD、被忽略的缓存 / 凭据不在摘要内；文档依据本次实际读取内容。
- 审阅类型：Architect（P0-003 / P0-004 核心契约）+ Builder（P0-001 / P0-002 / P0-005 基础工程与流程）。不是 Linux Sandbox Security Review，也不是项目负责人的最终阶段验收。
- 审阅者：独立 Agent `phase0_review`。
- 独立于实现者：**是**。本 Agent 在用户明确授权独立评审后启动，未参与被审代码、配置、原设计或作者自检的编写；自行读取实际代码、配置、设计文档并进行下述只读复核。
- 日期：2026-10-01（Asia/Shanghai）。
- 核心结论：**通过，限基础不变量的独立审阅；完整 Go 类型 / 接口及后续执行语义仍不可整体冻结**。冻结建议见第 4 节，本报告不自行修改版本或冻结记录。
- 基础工程结论：**需修改**。已运行服务健康，但 MinIO 重复构建存在已证实的阻塞，P0-005 和 Phase 0 尚不能验收通过。

## 1. 结论与执行边界

本次架构草案保留了必要的服务边界、不可变快照、Run 身份、终态原则和原始执行证据与 verdict 分离。没有发现必须立即修订源码才能保留这些基础原则的问题；尚未实现的能力已明确标为 draft，不能把“未实现”直接登记为缺陷。

但 `npm run infra:build-minio` 在已有输出目录下复制 LICENSE 失败，说明初始化路径尚不能稳定重复执行。运行中的容器来自先前成功构建，不能以其 healthy 抵消重建失败。按用户问题处理约定，应保持停止修复和下一阶段推进，等待用户处理 / 明确授权；本次只读评审是另行明确授权的工作。

本报告仅新增审阅文档。未修复文件属性、改动源码 / 根台账 / 核心版本，未构建镜像、启停容器或重复执行全套检查。

## 2. REVIEW 核验表

| 检查               | 证据与结果                                                                                                                                                                   |
| ------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 范围与契约一致     | API 仅三个基础 GET 路由；Worker 明确 idle；没有提交执行端点。架构的目标拓扑和当前能力分开描述。Phase 0 基础范围符合 ADR-0001；MinIO 变更限定本机开发。                       |
| 正常路径、失败路径 | 阅读 HTTP readiness 单元测试、API smoke 脚本及本轮作者的成功证据。独立核对三个容器 healthy。重复 MinIO 构建失败，不能验收为正常 / 缓存复现路径均通过。                       |
| 数据与状态不变量   | Submission / SourceUnit 快照与 Run 产物分离；rejudge 新 Run；终态不能回退。只有状态边方向 helper 和测试，没有业务持久化 / 组合校验器。                                       |
| 并发、重试、幂等   | 文档明确至少一次投递、提交后 ACK、attempt / rejudge 分离及 fence；无队列实现，本阶段不验收运行语义。工程构建幂等性有阻塞。                                                   |
| 安全及资源边界     | API 无 judge / sandbox / execution / os/exec 依赖；所有基础服务仅发布 localhost。MinIO 运行 UID/GID 10001。资源单位和预算分类明确；没有 Linux 隔离、资源强制或安全测试证据。 |
| 可观测性           | API / Worker JSON 日志及健康接口足以支持骨架；错误 readiness 不泄露连接细节。未来 job / attempt / fence / 清理事件与资源证据尚待各阶段设计。                                 |
| 兼容、迁移、回退   | 原 PG / Redis 命名卷保持；MinIO 本次首次创建项目命名卷。down 不删卷。ADR-0002 明确不能盲目降级已写入的新存储格式；旧 MinIO 数据格式迁移未验证。                              |
| 文档、交接和进度   | 角色、任务、状态、审阅与用户接管规则清楚；原始文档是资料，不自动授权。发现 README 远程仓库信息和 setup 末尾提示过时，见第 5 节；根 Agent 负责同步最终台账。                  |

## 3. P0-004 七项独立回答

### 3.1 API 与 Judge 权限分离

**通过基础边界审阅；部署权限隔离尚未验证。** `apps/api/main.go:3–15` 只依赖 HTTP 与 pgx；`internal/httpapi/router.go:16–38` 只注册健康、readiness、system；`services/judge-worker/main.go:12–18` 等待退出信号，不接受任务。实际 API / HTTP 包没有执行器或 `os/exec` 引入。

`docs/architecture/ARCHITECTURE.md:27–30` 禁止 HTTP 导入执行实现及 API 持有执行权限；`docs/adr/0001-foundation-boundaries.md` 第 3 项要求 Phase 2 同步 bridge 仍通过独立判题进程 / 服务。这是可冻结的不变量。当前没有 bridge、生产部署账户或权限清单，不能声明这些实际权限已隔离；Phase 2 设计须列明 bridge 进程、IPC、凭据、认证、超时和调用失败语义。

### 3.2 Submission / SourceUnit / Run / Artifact 身份与重判

**通过身份关系原则审阅；完整持久化形状继续 draft。** `internal/domain/model.go:13–39,70–76` 将源码快照与 Run 产物分离，Artifact 显式含 RunID、SourceUnitID、内容 hash 与实际 compiler/runtime；`docs/architecture/DOMAIN_MODEL.md:13–19,26–30` 定义 Role 唯一、Run N:1 Submission、重判新身份及版本固定。这些关系可以支持保留重判历史。

尚需 Phase 2 明确：ID 生成与外键、Artifact 的 source 确属同一 Submission 的约束、artifact_kind、角色全集、对象存储 key 与 hash 验证、Compilation 失败诊断及实际编译版本保存方式、删除 / 保留规则、允许升级哪些 Problem/testdata/checker/runtime 版本。Go 草案没有独立 Compilation 记录；这不阻塞 Phase 0 草案交付，却禁止把现有 structs 视为完整数据库 schema 或长期兼容 wire DTO。

### 3.3 生命周期、终态、取消与重投

**通过逻辑生命周期原则审阅；竞争与组合校验不得冻结为已实现能力。** `internal/domain/state.go:30–44` 只检查有向边；`state_test.go:5–23` 验证顺序、终态不可变及若干错误边。`docs/judge/STATE_MACHINE.md:25–31` 另外定义 state / verdict 组合，明确 helper 不校验 verdict / fence / 事务竞争，未隐藏限制。

因此，`CanTransition(Pending, Finished)` 返回 true 并不证明 ACCEPTED 可从 PENDING 直接提交；存储提交函数仍须校验来源状态与 verdict、终态唯一和持有者条件。Queue 重领不回退逻辑状态，rejudge 新 Run；后续必须设计恢复如何越过已持久化 COMPILING / JUDGING、取消与完成谁能成功提交、清理完成后何时持久化 CANCELLED、旧 attempt 写回如何拒绝。Phase 1 固化 timeout / exit / cleanup 的局部竞争；Phase 3 再固化 lease / reclaim / heartbeat / retry / fencing SQL。现有边测试不证明这些条件已通过。

### 3.4 ProcessSpec / SourceBinding / Graph

**显式绑定原则通过；完整 Graph / ProcessSpec 形状不可冻结。** `internal/execution/graph.go:6–15` 显式携带 SourceUnitID 与 ArtifactID，不靠数组顺序关联；Graph 也携带资源策略与 evaluator version（25–30 行）。

现有 ProcessSpec 只有 ID、Role、Binding，尚无可信管理进程的来源 / 身份 / 信任等级 / 运行版本，也没有可执行文件、参数、cwd、环境、输入输出映射或启动依赖。当前所有 Binding 都是同一形状，尚不能表示“不来自选手 SourceUnit 的 Checker / Interactor”。Phase 1 先定义单选手进程与可信 Exact Checker；Phase 7–8 再定义管理角色来源、通道方向、EOF、背压、拓扑验证、信任与计量边界。不应现在用空 SourceUnitID 暗约定管理进程，也不应冻结这个最小 struct 后要求未来实现绕过它。

### 3.5 资源单位与计量

**单位与分类原则通过；字段含义、默认值及执行结果模型待设计。** `internal/domain/model.go:41–55` 分为 per-process、aggregate、compile、node overhead；字段名标注 ms / bytes / PID 数。`docs/architecture/DOMAIN_MODEL.md:34–36` 要求正且有限、禁止隐式无限，并要求同一 Job cgroup 的聚合占用 / 峰值，不能相加异步出现的各进程峰值。

下一步必须明确：CPU 是逐进程还是合计的累计时间，wall 的起止点，多个相同 Role 进程如何受限，stdout / stderr / 通道输出如何计数，PID 包括哪些后代，temp bytes 如何可靠执行，compile 对多 SourceUnit 是逐份还是总预算，per-role 与 aggregate 同时触发时的归因，node overhead 是否参与判题计量或仅容量预留。`internal/sandbox/sandbox.go:10–24` 的结果目前没有 wall / output / PID / temp 观测字段，不能把 ResourceLimit 的存在视作测量完整。Phase 1 仅冻结实际采用的受控可信 runner 指标；Phase 4 / 8 / 9 分别补齐强制隔离、聚合计量和容量语义。

### 3.6 Sandbox 原始结果与 Judge verdict

**分层原则通过；Runtime / Engine 完整错误契约继续 draft。** `internal/sandbox/sandbox.go:10–29` 返回进程退出、signal、CPU、内存、输出对象与 Failure；`internal/judge/engine.go:8–10` 另返回 JudgeResult / Verdict。Runtime 注释要求 graph-wide cleanup 且子进程全部退出后返回；安全基线禁止普通 subprocess 降级。

后续须定义 `(result, error)` 同时存在时的优先级、context 取消 / 超时是否返回可采信的部分证据、无法清理或失去计量如何表示、编译 Runtime 调用与运行 Runtime 的边界。用户 RE/TLE/MLE 必须由可信监督证据映射，不能仅凭 `error != nil` 或任意 exit code。Phase 1 的可信 fixture runner 不意味着 Phase 4 隔离已完成，也不授权公开执行任意源码。

### 3.7 事务 / outbox / fence / 可追溯

**原则通过；事务接口、SQL 和恢复策略尚不可冻结。** `docs/architecture/ARCHITECTURE.md:36–40` 将提交快照 / 首 Run 放在一事务，Phase 3 同事务增加 outbox，结果带 run / attempt / fence 条件提交后才 ACK；`docs/judge/STATE_MACHINE.md:35–39,53` 将重投与重判分开。版本与 Artifact hash 字段支持未来追溯。

Phase 2 必须冻结快照创建事务、唯一 / 外键约束、存储接口、输入幂等键与回滚行为；Phase 3 必须冻结 outbox 发布 / 去重、fence 产生及所有状态 / 产物写入的条件、取消竞争、lease 与恢复、重试耗尽 / 死信规则，以及对象存储孤儿清理。当前没有业务表、outbox、仓储提交或 lease 实现，不能把上述文字升级为已保证事务一致性。

## 4. 可冻结范围与排除范围

建议项目负责人只为以下已审原则建立一份明确列项的基础不变量版本及冻结记录，引用本报告；**不把整个 `0.1.0-draft` 文档集合和现有 Go 类型自动改成 frozen**：

1. API 不编译 / 运行选手代码；同步 bridge 维持独立执行进程边界。
2. Submission / SourceUnit 快照不可变；编译诊断和产物由 Run 持有；rejudge 新建 Run，不覆盖终态历史。
3. SourceUnit / Artifact / Process 使用显式身份绑定，不依赖集合次序。
4. 生命周期与 verdict 分离；非终态无 verdict、CANCELLED 对应取消 verdict、终态唯一且不可回退。具体竞争算法继续设计。
5. 编译、逐进程、Job 聚合与节点开销分别表达；ms / bytes / PID 数明确；不采用隐式无限；聚合峰值不能相加各进程独立峰值。
6. Runtime 返回原始监督证据并负责完整清理，JudgeEngine 负责判题映射；失去可信计量归类基础设施失败；没有不安全执行降级。
7. 至少一次投递；业务结果事务成功后 ACK；旧持有者必须被 fence 拒绝；提交、版本和重判历史可追溯。这是实现必须遵守的目标，当前未验收实际实现。

完整 Graph / Binding schema、ResourcePolicy 可执行语义、Sandbox / Engine 接口错误语义、DB / API schema、授权、outbox / fencing / lease、Linux 隔离及 Phase 7–9 细节仍是 proposed。实施这些内容前必须完成对应 DESIGN 和独立审阅；本次基础原则审阅不能替代它们。

## 5. 发现与处理要求

### R1：MinIO 重复构建幂等性失败——阻塞

- 文件 / 位置：`scripts/build-minio.mjs:97`，源 LICENSE → `.local/minio-build/LICENSE` 的 `copyFileSync`。
- 证据：作者的最终重复构建 exit 1 / EPERM；独立只读核对两份文件均 `IsReadOnly=true`、34523 bytes，SHA-256 都为 `0D96A4FF68AD6D4B6F1F30F713B18D5184912BA8DD389F86AA7710DB079ABCB0`。
- 前因与后果：首次复制 / 镜像构建已成功，目标保留只读属性；重复执行尝试覆盖已有目标而失败。只读属性解释与证据一致，未进行修复后的排除性试验，不能冒充已复验根因。失败在 Docker build 前，不等于 Docker engine 或现有容器失效。
- 影响：`npm run infra:build-minio` 是 P0-005 验收和 README 完整初始化的显式步骤。已有缓存的正常重建无法完成，因此基础工程结论需修改，Phase 0 不能最终验收 / 进入 Phase 1。首次 Linux CI 的潜在成功不能覆盖 Windows 缓存路径失败。
- 处理人：用户；根 Agent 负责台账与独立问题档案。详见 [问题档案](../issues/2026-10-01-minio-license-copy.md)。本 reviewer 未修改属性或执行修复。
- 解除条件：最终构建脚本在同一已有输出目录连续两次 exit 0，LICENSE 内容 hash 保持，源码 / 镜像来源仍正确；复验 Compose / 服务 smoke / 真实数据库 readiness，保留既有卷。仅一次清除属性不是脚本修复的充分证据。

### R2：两个入口说明过时——非阻塞

- `README.md:70` 写“当前未配置远程仓库”，实际只读 `git config --get remote.origin.url` 返回已配置的 OJ GitHub origin，也与 `docs/development/ENVIRONMENT.md:96` 冲突。影响下一位 Agent 对远程 CI / 交接条件的判断；不能据此伪造远程 CI 通过。
- `scripts/setup.mjs:30` 最后仍提示 doctor 后直接 `infra:up`，遗漏当前 `pull_policy: never` 的 MinIO 镜像必须先 `infra:build-minio`。新机器按该提示直接启动会缺少本地镜像；README / ENVIRONMENT 的完整命令已正确，所以此提示不单独阻塞现有服务。
- 处理人：用户；建议根 Agent 按一事一档规则另建文档，暂不默默修订。验收时同步入口文字，再核对新副本初始化顺序。上述提示差异不影响已通过的 API / Web / Go 基础检查。

## 6. 基础工程 / MinIO / CI 的具体审阅

**工程骨架与流程：基础范围通过。** npm workspace 和精确依赖 / lock 保留；Go domain 不引入基础设施；setup 保留已有 `.env` 并随机生成新环境凭据；Go wrapper 限定项目包，避免遍历 node_modules 的 Go 源码；sqlc 只有 SELECT 1，无冒充业务 schema。Vite 仅 localhost 开发访问，Monaco 配置本地 worker，页面声明不能提交或保存源码。本轮没有实际搭建前端；前端实施 MD 的内容不在本报告的实现验收范围内。

流程文档明确任务状态、实际执行者、路径、风险和 Review owner；原资料不自动获得授权，自检不能登记独立评审，异常单独记录并按用户要求停止。项目负责人仍须据当前失败更新台账，不能保留“所有门禁通过”的阶段结论。

**MinIO 来源与镜像：来源方案可接受用于 localhost 开发，构建脚本需修复 R1。** `deploy/minio/source-lock.json:2–6` 固定完整官方提交与 CA 来源 digest；构建器以固定 commit 下载上游 module、以 `-mod=readonly` / `GOTOOLCHAIN=local` / `CGO_ENABLED=0` 构建 linux/amd64，没有修改 OJ go.mod。独立读取本机 `go env GOSUMDB` 为 `sum.golang.org`，构建元数据包含 module / go.mod sums；本次未重新下载校验。脚本继承主机 Go 环境，未强制拒绝使用者的 `GOSUMDB=off` 或其他跳过校验配置，所以“校验保持启用”是此次实际环境事实，不是对任意主机配置的硬保证；生产供应链策略不在本轮验收内。

Dockerfile 最终层为 scratch，只含二进制、静态探针、CA、LICENSE 与数据目录；`deploy/minio/Dockerfile:9–12` 保留 LICENSE、数据目录属主与 UID/GID 10001。独立 Docker inspect 确认实际运行用户 `10001:10001`、revision label 完整匹配 source-lock、挂载 `oj-dev_minio-data:/data`；镜像 ID 为 `sha256:544a55f20b45c95c8704253cf1f91074fae5184934e181e720a26aea6d47495f`。当前本地二进制 SHA-256 与构建元数据匹配：`ab90d4d794fc5826df7766ce54f02c099f75b76550ec26847e4a76b8b7a741cf`。没有把开发源码镜像冒称原预编译 release，也没有把 Docker Compose 当选手 Sandbox。

MinIO 本次使用新命名卷，健康证明该新卷可用；不能据此证明任意已有 root 属主 / 旧版本数据卷可直接迁移。对象存储业务持久化、受限账户、bucket policy、备份 / 恢复及生产维护另行审阅。ADR-0002 已记录上游归档、冷机器网络依赖、限定 localhost 和不可盲目数据降级；这些限制应保留。

**CI：静态配置顺序合理，远程执行无法验证。** `.github/workflows/ci.yml:24–36` 包含 setup、check、Linux race、API smoke、sqlc 再生成差异、Compose config、固定 CA 镜像拉取、MinIO 构建、完整 up / smoke 和 always down；退出不删卷。CI 从新 checkout 构建，不能证明已有 Windows LICENSE 的重复覆盖路径。`scripts/go.mjs:10–12` 的递归包集合不含 `scripts/minio-health`，普通 check / race 的 Go vet / tests 并未自动覆盖探针；本轮作者另跑探针 vet，构建脚本及容器 healthcheck 实际覆盖其编译 / 执行。若后续扩充探针逻辑，应将相应检查明确纳入 CI；这不是当前阶段的独立新增阻塞。

## 7. 已测、独立复核与未测

| 项目                                                         | 本次证据归属 / 结果                                                                                                                                                                                              |
| ------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 实际代码、设计、流程、配置、CI                               | 本 reviewer 独立静态阅读，结论见以上逐项。                                                                                                                                                                       |
| 代码 / 配置摘要                                              | 本 reviewer 执行 `node scripts/source-digest.mjs`，与评审输入 56 文件 / SHA-256 一致。                                                                                                                           |
| Docker 引擎、现有服务、用户 / 来源 / 卷                      | 本 reviewer 只读 `docker info` / `docker ps` / `docker inspect` / `docker image inspect`：29.8.1 / linux / x86_64，PG / Redis / MinIO healthy，端口均 localhost，MinIO UID / revision / 命名卷正确。未启停服务。 |
| LICENSE 属性 / 内容、MinIO 二进制元数据                      | 本 reviewer 只读 Get-Item / Get-FileHash 与元数据核对，结果见 R1 / 第 6 节。                                                                                                                                     |
| 首次构建 / Compose up / 完整 infra-smoke / 真实 DB readiness | 阅读作者 [交接](../handoffs/2026-10-01-environment-review-handoff.md) 与 [修复记录](../issues/2026-10-01-minio-source-rebuild.md) 的通过证据。未再次执行这些测试。                                               |
| npm run check / npm run smoke / 探针 Go vet                  | 作者本轮通过；reviewer 阅读测试实现并核对范围，未重复全套测试，不冒充独立重跑。smoke 覆盖未配置数据库 503 和提交路由 404。                                                                                       |
| 最终重复 infra:build-minio                                   | 作者失败，EPERM；reviewer核对失败点和文件属性，未重试 / 修复。                                                                                                                                                   |
| Windows race / 远程 CI / 冷机器首次初始化                    | 未执行 / 未核实，不能登记通过。                                                                                                                                                                                  |
| S3 业务持久化、旧 MinIO 数据卷兼容 / 回退                    | 未执行，后续阶段 / 存储变更任务验收。                                                                                                                                                                            |
| 真实 Judge、Queue / fence / outbox / 取消竞争、Linux 隔离    | 无实现或未执行，明确排除。                                                                                                                                                                                       |

## 8. 交给根 Agent / 项目负责人的裁定

- 独立评审授权已使用并产出实际 REVIEW，不再以“未授权委派 / 没有 reviewer”作为本轮停止理由；审阅结论不能自动变成 Phase 0 通过。
- 可以登记基础不变量的独立审阅通过，并按第 4 节决定是否建立窄范围冻结记录；完整核心契约保持 draft，未来阶段设计仍须审阅。
- P0-001 的原骨架和 P0-002 流程可记录本报告的基础范围结论；P0-005 因 R1 仍 blocked，Phase 0 尚未最终验收。R2 单独归档，由用户处理。
- 用户处理 / 授权修复 R1 并获得具体复验证据后，复审受影响构建脚本及初始化说明，再进行 Phase 0 最终验收。之后才恢复 P1-001，按第 3 节未冻结项设计编译、运行、清理、错误映射和可信 fixture 边界。
- 不继续实现前端；其详细 MD 与任务包保持既有交付，真正实施仍需明确授权及 P2-100 的方案 / 接口审阅。
