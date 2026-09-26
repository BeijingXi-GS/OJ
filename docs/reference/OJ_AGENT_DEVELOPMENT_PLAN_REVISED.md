# OJ Agent 开发计划与职责分工

> 目标：采用分层 Agent 协作方式，建设一套可持续演进至高并发、交互题、通信题及多 Judge 节点场景的 Online Judge（OJ）系统。
>
> 核心原则：**高推理能力 Agent 负责高影响、难以回滚的架构与安全决策；Builder Agent 负责需要完整上下文理解的模块实现；Worker Agent 负责边界清晰、可通过自动化验证的大规模工程任务。**

---

# 1. 项目目标

本项目的目标能力范围包括：

- 普通题（Standard Judge）
- Special Judge / Checker
- 多语言编译与运行
- 高并发提交
- Redis 队列与异步判题
- Rejudge
- Contest / Scoreboard
- Judge Worker 横向扩容
- Linux Sandbox
- CPU / Memory / PID / Output / Wall Time 限制
- Interactive Judge
- Communication Judge
- 多进程 Execution Graph
- 多 Judge Node 调度
- 可观测性
- 后续 Kubernetes / gVisor 部署

开发过程中**不得一次性引入全部目标能力**。

必须按照本文档定义的阶段逐步实施；每个阶段均应具备可独立运行、可验证、可回滚和可交付的明确边界。

---

# 2. 技术栈冻结

## 2.1 Frontend

- React
- TypeScript
- Vite
- Tailwind CSS
- shadcn/ui
- Monaco Editor
- TanStack Query
- React Router

## 2.2 Backend

- Go
- `net/http` + chi
- REST API
- OpenAPI
- PostgreSQL
- pgx
- sqlc

## 2.3 Infrastructure

- PostgreSQL
- Redis
- Redis Streams
- S3 Compatible Object Storage
- MinIO（开发环境）
- Docker Compose（开发环境）
- Kubernetes（后期生产环境）

## 2.4 Judge

- Go Judge Worker
- Linux
- namespaces
- cgroups v2
- seccomp
- capability dropping
- read-only filesystem
- tmpfs
- no-network by default

后期可增加：

- containerd
- gVisor / runsc

## 2.5 Observability

- Prometheus
- Grafana
- OpenTelemetry
- Structured JSON Logging

---

# 3. 三类 Agent

整个项目统一采用以下三类 Agent，并以职责边界而非代码难度作为分工依据。

---

## 3.1 Architect Agent

使用**最高推理能力的 Agent**，优先承担架构、安全、一致性及高风险并发设计。

### 负责

- 系统架构
- 核心抽象
- 安全模型
- 并发模型
- 状态机
- 数据一致性
- Queue 语义
- 幂等性
- Sandbox 设计
- Scheduler
- Interactive Judge
- Communication Judge
- ExecutionGraph
- Process 生命周期
- 协议设计
- 核心数据库 invariant
- 高风险 PR Review
- Threat Model
- 故障恢复策略

### 不应该主要负责

- CRUD
- 简单页面
- DTO
- 普通 migration
- 重复性 endpoint
- YAML 配置
- 基础测试 fixture
- 文档排版

### 判断标准

如果一个任务：

> 代码可能只有几十行，但设计错误会迫使整个系统重构，

则必须交给 Architect Agent。

---

## 3.2 Builder Agent

使用**中高能力、成本适中的 Agent**，负责在冻结契约下完成模块级实现与集成。

### 负责

- 根据 Architect Spec 实现完整模块
- Backend Service
- Judge Worker
- Queue Consumer
- Checker
- Contest 模块
- Scoreboard 模块
- Storage 模块
- Runtime 管理
- API 集成
- Repository / Service / Handler
- Integration Tests

Builder 不应自行重新定义 Architect 已冻结的核心接口。

如果发现接口存在结构性问题，应停止扩散实现并提交设计问题给 Architect。

---

## 3.3 Worker Agent

使用**成本较低、适合批量执行的 Agent**。

### 负责

适合机械执行、容易通过测试判断对错的工作：

- CRUD
- API DTO
- React 页面
- 表单
- 表格
- Pagination
- Filtering
- Migration
- Mock
- Test Fixture
- Seed Data
- 单元测试补齐
- Integration Case 扩充
- YAML
- Dockerfile
- CI
- README
- API Docs
- Grafana Dashboard
- lint 修复
- formatter
- 重命名
- 机械重构
- OpenAPI Client
- Boilerplate

### 不允许独立决定

- Sandbox 安全边界
- Judge 状态机
- Queue ACK 语义
- 幂等策略
- Scheduler
- Interactive Supervisor
- IPC
- ExecutionGraph
- 通信题计量协议
- 核心数据库一致性模型

---

# 4. Agent 分配的总原则

不要使用：

> 难代码 = 强 Agent  
> 简单代码 = 便宜 Agent

作为判断标准。

应该使用：

> **这个任务做错以后，影响范围有多大？**

例如：

一个 500 行 React 页面写错：

- UI 有 bug
- 修掉即可

适合 Worker Agent。

一个 20 行 Sandbox 接口设计错：

- Standard Judge 要改
- Interactive Judge 要改
- Communication Judge 要改
- Worker 要改
- Tests 要改

必须 Architect Agent 负责。

---

# 5. 不可违反的架构规则

以下规则在整个开发过程中视为项目“宪法”。

## 5.1 Backend 不得执行用户代码

```text
Backend != Judge Worker
```

API Backend 永远不能直接运行 contestant code。

---

## 5.2 Submission 与 JudgeRun 必须分离

```text
Submission
 ├── JudgeRun #1
 ├── JudgeRun #2
 └── JudgeRun #3
```

Rejudge 必须创建新的 JudgeRun。

---

## 5.3 Problem / Testdata 必须版本化

Submission / JudgeRun 应能够追溯：

- Problem Version
- Testdata Version
- Checker Version
- Runtime Version
- Compiler Version

---

## 5.4 Judge Job 至少一次执行

Queue 采用：

```text
at-least-once delivery
```

因此 Judge Job 必须：

```text
idempotent
```

---

## 5.5 ACK 只能发生在结果持久化之后

禁止：

```text
ACK
↓
写数据库
```

必须：

```text
完成判题
↓
数据库事务提交
↓
ACK
```

---

## 5.6 Judge Engine 不依赖具体容器实现

业务代码依赖：

```text
Sandbox Interface
```

而不是直接调用：

```text
docker run
containerd
runsc
```

---

## 5.7 Standard / Interactive / Communication 必须共享底层执行抽象

最终底层统一为：

```text
JudgeJob
   ↓
JudgeEngine
   ↓
ExecutionGraph
   ↓
Sandbox Runtime
   ↓
Process + Channel
```

---

## 5.8 Communication Submission 必须以多源代码集合建模

通信题的一次提交可能同时包含多个角色或进程的源代码，例如：

```text
Submission
  ├── SourceUnit(role=Alice)
  ├── SourceUnit(role=Bob)
  └── SourceUnit(role=Relay)
```

这些 SourceUnit **共同构成一个 Submission 的不可分割版本快照**，不得为每个角色分别创建互不关联的 Submission。

每个 SourceUnit 至少记录：

- Role / Process ID
- Language
- Source Location
- Source Hash
- Compiler / Runtime Version
- Compile Status
- Compiled Artifact Location
- Artifact Hash

要求：

- JudgeRun 必须绑定一组确定且不可变的 SourceUnit 版本。
- 各 SourceUnit 可以独立编译，但运行阶段只有在必需的编译任务全部成功后才能启动。
- 编译错误必须能够定位到具体 Role / SourceUnit。
- Rejudge 不修改原 Submission 源码快照，只创建新的 JudgeRun，并重新解析其对应的编译与运行产物。
- ExecutionGraph 的 ProcessSpec 必须显式引用对应 SourceUnit / Artifact，禁止依赖隐式顺序绑定。

---

## 5.9 Communication Judge 必须同时管理单进程与聚合资源

通信题存在多个用户进程并发运行，内存限制不得只按单进程峰值判断。资源模型至少区分：

```text
Per-Process Memory Limit
+
Judge Aggregate Memory Budget
+
Compile Memory Budget
+
Node Reservation / Runtime Overhead
```

必须满足：

- 每个 contestant process 可以具有独立的内存上限。
- 同一 JudgeRun 的全部 contestant process 必须受一个聚合内存预算约束。
- 编译阶段必须具有独立资源预算，避免恶意源码通过编译器耗尽 Judge Node 内存。
- 聚合峰值应以同一时间点的 Job / cgroup 层级实际内存占用为准，不得简单相加各进程在不同时刻出现的独立 peak。
- Scheduler 的资源申请与预留必须考虑多进程并发运行的总内存需求，而不能继续使用 `1 submission = 1 slot` 的简化模型。
- 触发单进程或聚合内存限制时，必须记录触发层级、相关 Role / Process、观测值和限制值，并按照既定 failure propagation 规则终止 ExecutionGraph。

---

# 6. 核心领域对象

从 Phase 0 开始冻结以下对象。

```text
User

Problem
ProblemVersion

TestdataVersion
Testcase

Submission
SubmissionSourceUnit
CompiledArtifact

JudgeJob
JudgeRun
TestcaseResult

Language
RuntimeVersion

Contest
ContestProblem
ContestParticipant
ScoreboardEntry

JudgeNode
```

---

# 7. 三个最关键的核心抽象

## 7.1 JudgeJob

表示：

> 系统需要执行一次什么判题任务。

应至少包括：

- Job ID
- Submission ID
- Judge Type
- Problem Version
- Testdata Version
- Language / Source Bundle
- SourceUnit List（Communication Judge 可为多份）
- Source Location / Source Hash
- Resource Policy
  - Per-Process Limit
  - Aggregate Judge Budget
  - Compile Budget
- Priority
- Attempt
- Runtime Version

---

## 7.2 ExecutionGraph

表示：

> 本次 Judge 中需要运行哪些进程，以及这些进程之间如何通信。

基本对象：

```text
Process
Channel
ResourcePolicy
SourceBinding
Evaluator
```

其中 `SourceBinding` 用于把 ProcessSpec 显式绑定到 SubmissionSourceUnit / CompiledArtifact；`ResourcePolicy` 同时表达单进程限制与整个 ExecutionGraph 的聚合预算。

### Standard

```text
Input
  ↓
Contestant
  ↓
Output
```

### Interactive

```text
Contestant
    ⇅
Interactor
```

### Communication

```text
Submission
  ├── Alice.cpp ──compile──> Artifact A ──> Process Alice
  └── Bob.cpp   ──compile──> Artifact B ──> Process Bob

Process Alice
      ↓
   Channel
      ↓
Process Bob
```

通信题中的多份代码属于同一个 Submission，但由不同 ProcessSpec 分别绑定和执行。以后所有 Judge 类型都应最终映射到 ExecutionGraph。

---

## 7.3 Sandbox

表示：

> 如何安全运行一个不可信进程。

Sandbox 负责：

- Process start
- PID isolation
- Mount isolation
- Network isolation
- CPU limit
- Memory limit
- PID limit
- Output limit
- Wall timeout
- CPU timeout
- Signal handling
- Process tree cleanup
- Exit status
- Peak memory collection

---

# 8. Verdict 统一定义

至少支持：

```text
PENDING
QUEUED
COMPILING
JUDGING

ACCEPTED
WRONG_ANSWER

COMPILE_ERROR
RUNTIME_ERROR

TIME_LIMIT_EXCEEDED
MEMORY_LIMIT_EXCEEDED
OUTPUT_LIMIT_EXCEEDED

IDLENESS_LIMIT_EXCEEDED

PARTIAL_ACCEPTED

SYSTEM_ERROR
CANCELLED
```

必须区分：

```text
User Error
```

与：

```text
Infrastructure Error
```

例如：

```text
SIGSEGV
→ Runtime Error

Sandbox Runtime Crash
→ System Error
```

---

# 9. 分阶段开发计划

---

# Phase 0：Architecture Foundation

## 目标

不追求产品功能。

目标是冻结不会轻易变化的核心结构。

## Architect Agent

必须负责：

- Architecture Document
- Domain Model
- Judge State Machine
- Submission / JudgeRun 关系
- JudgeJob
- ResourceLimit
- JudgeResult
- Sandbox Interface
- JudgeEngine Interface
- ExecutionGraph 初始模型
- Multi-Source Submission / SourceUnit 模型
- ResourcePolicy（单进程 / 聚合 / 编译预算）
- Error Model
- Transaction Boundary
- Repository Boundary
- Module Dependency Rules

输出：

```text
docs/architecture/
docs/judge/
docs/sandbox/

ARCHITECTURE.md
DOMAIN_MODEL.md
STATE_MACHINE.md
```

## Builder Agent

负责：

- 初始化 Go 项目
- 初始化 API project
- 初始化 Judge Worker project
- 定义 Architect Spec 对应的数据结构
- 建立基础 package

## Worker Agent

负责：

- Monorepo 初始化
- Vite
- React
- Docker Compose
- PostgreSQL
- Redis
- MinIO
- Makefile
- CI
- formatter
- lint
- README skeleton

## 强制 Architect Review

必须 Review：

- package dependency
- core interfaces
- Judge state machine
- database identity model

---

# Phase 1：Local Mini Judge

## 目标

实现：

```text
judge solution.cpp
```

能够：

```text
Compile
Run
Feed stdin
Capture stdout
Exact Compare
```

并返回：

```text
AC
WA
CE
RE
TLE
```

## Architect Agent

负责：

- Process lifecycle
- Timeout semantics
- Process group cleanup
- Signal semantics
- Judge result model
- Compile / Runtime boundary
- Error classification

重点检查：

```text
timeout 与 process exit 的 race
```

以及：

```text
父进程退出后子进程是否残留
```

## Builder Agent

负责：

- C++ Compiler
- Local runner
- stdout/stderr capture
- Exact Checker
- Judge pipeline
- CLI

## Worker Agent

负责：

- AC fixture
- WA fixture
- CE fixture
- RE fixture
- TLE fixture
- CLI test
- logs
- test scripts
- documentation

## 完成标准

必须通过：

```text
normal AC
wrong output
compile error
segfault
infinite loop
```

---

# Phase 2：Backend + Database + Frontend Skeleton

## 目标

实现最小 Web OJ。

支持：

- Problem
- Submission
- Submission Detail
- Code Editor
- Judge Result Display

## Architect Agent

只负责 Review：

- Database schema
- Submission / JudgeRun mapping
- transaction boundary
- authorization boundary
- API contract

## Builder Agent

负责：

- Problem module
- Submission module
- JudgeRun module
- PostgreSQL integration
- repository/service/handler
- API
- basic frontend integration

## Worker Agent

重点负责：

- CRUD
- migration
- DTO
- pagination
- filters
- Problem pages
- Submission list
- Monaco Editor
- forms
- tables
- OpenAPI client
- mock data

## 完成标准

用户可以：

```text
打开题目
↓
提交 C++
↓
看到 Submission
↓
看到 Judge Result
```

此阶段 Judge 可以暂时本地同步调用，仅用于开发。

---

# Phase 3：Redis Queue + Independent Judge Worker

## 目标

把：

```text
HTTP Request
```

和：

```text
Judge Execution
```

完全解耦。

架构：

```text
Backend
   ↓
Redis Streams
   ↓
Judge Worker
```

## Architect Agent

必须负责：

- at-least-once semantics
- JudgeJob idempotency
- XACK timing
- pending job reclaim
- Worker crash recovery
- duplicate execution handling
- JudgeRun state transition
- Rejudge semantics
- Database commit ordering

必须分析场景：

```text
Judge 完成
↓
DB COMMIT
↓
Worker 在 ACK 前崩溃
```

以及：

```text
Worker 拿到 Job 后直接失联
```

## Builder Agent

负责：

- Redis Streams consumer
- producer
- consumer group
- worker loop
- Judge result commit
- Rejudge pipeline
- queue metrics

## Worker Agent

负责：

- queue wrapper boilerplate
- admin rejudge endpoint
- test mocks
- job fixtures
- Redis integration test
- queue dashboard

## 完成标准

必须证明：

- Worker crash 后 Job 可重新执行
- 重复执行不会生成错误最终状态
- DB 成功后 ACK
- Rejudge 会创建新 JudgeRun

---

# Phase 4：Real Linux Sandbox

## 目标

将用户代码从普通 subprocess 执行升级成真正受限的 Linux Sandbox。

## Architect Agent

必须主导整个阶段。

负责：

- Threat Model
- namespace design
- user namespace
- PID namespace
- mount namespace
- network namespace
- IPC namespace
- cgroups v2
- seccomp
- capability dropping
- read-only root filesystem
- tmpfs
- process limit
- memory accounting
- CPU accounting
- wall timeout
- output limit
- OOM detection
- process tree kill
- filesystem access policy
- syscall policy
- host escape analysis

## Builder Agent

在 Architect 明确 Spec 后实现：

- LinuxSandbox
- cgroup manager
- namespace launcher
- seccomp policy loader
- filesystem builder
- process supervisor
- resource metrics collection

## Worker Agent

大量负责攻击测试：

- infinite loop
- fork bomb
- unlimited memory
- unlimited stdout
- file spam
- read `/etc/passwd`
- scan filesystem
- network access
- child process escape
- zombie processes
- syscall abuse

以及：

- benchmark scripts
- integration tests
- security test fixtures
- docs

## 强制 Review

Sandbox 代码合并前：

```text
Architect Security Review
```

是强制步骤。

## 完成标准

Contestant：

```text
不能访问网络
不能读取 Host 敏感文件
不能逃出 mount
不能无限 fork
不能无限占内存
不能无限占 CPU
不能无限输出
```

并可稳定返回：

```text
TLE
MLE
OLE
RE
```

---

# Phase 5：Production Standard Judge

## 目标

完善普通题系统。

加入：

- S3 / MinIO
- Problem Version
- Testdata Version
- 多语言
- Runtime Version
- Testdata Cache
- Token Checker
- Float Checker
- Special Judge

## Architect Agent

负责：

- immutable testdata model
- runtime reproducibility
- language abstraction
- checker trust model
- SPJ execution boundary
- artifact model
- runtime image/version semantics

必须回答：

```text
Checker 是可信代码还是不可信代码？
```

```text
Checker 能访问哪些文件？
```

```text
Submission 如何记录 compiler/runtime 版本？
```

## Builder Agent

负责：

- S3 integration
- Artifact Manager
- Testdata Cache
- Language loader
- runtime definitions
- checker framework
- SPJ implementation

## Worker Agent

适合并行完成：

- C++ config
- Python config
- Java config
- Go config
- Rust config
- Runtime Dockerfile
- language fixtures
- syntax-error fixtures
- compile test
- docs

## 完成标准

同一 JudgeRun 能追溯：

```text
source
problem version
testdata version
checker version
compiler/runtime version
```

---

# Phase 6：Contest + High Concurrency

## 目标

让系统开始具备比赛级负载能力。

加入：

- Contest
- Scoreboard
- Queue Priority
- Rate Limit
- Admission Control
- SSE
- Rejudge Isolation
- Cache
- Backpressure

## Architect Agent

必须负责：

- priority scheduling policy
- contest / normal / rejudge isolation
- scoreboard consistency
- contest freeze semantics
- submission rate policy
- pending-job limits
- admission control
- resource model
- queue starvation prevention
- rejudge storm protection
- backpressure

## Builder Agent

负责：

- Contest Service
- Scoreboard Projector
- Priority Queue integration
- Rate Limiter
- Redis Cache
- SSE submission status
- Contest submission pipeline

## Worker Agent

负责：

- Contest CRUD
- Contest pages
- scoreboard UI
- admin UI
- Redis middleware
- rate-limit middleware boilerplate
- Prometheus metrics
- load-test scripts

## 完成标准

系统必须保证：

```text
大量 Rejudge
```

不会完全阻塞：

```text
实时 Contest Submission
```

---

# Phase 7：Interactive Judge

## 目标

支持：

```text
Contestant ⇄ Interactor
```

实时交互。

## Architect Agent

必须主导：

- Interactive Supervisor
- IPC topology
- Pipe semantics
- EOF semantics
- buffering
- deadlock handling
- idle timeout
- wall timeout
- process lifecycle
- signal propagation
- transcript
- secret isolation
- Interactor trust boundary

必须定义以下情况：

```text
Interactor 先退出
```

```text
Contestant 先退出
```

```text
双方都阻塞
```

```text
一方停止读取导致另一方写阻塞
```

```text
Interactor crash
```

分别得到什么结果。

## Builder Agent

负责：

- Supervisor
- duplex pipe/channel
- interactor runner
- idle detector
- transcript recorder
- InteractiveJudge implementation

## Worker Agent

负责：

- sample interactor
- transcript viewer
- interactive fixtures
- idle timeout tests
- malformed protocol tests
- admin debug UI
- documentation

## 完成标准

至少覆盖：

```text
normal AC
protocol WA
contestant RE
interactor failure
TLE
idle timeout
deadlock
```

---

# Phase 8：Communication Judge + ExecutionGraph

## 目标

在既有 Standard / Interactive 执行能力之上，将底层执行系统正式推广为：

```text
N SourceUnits
      ↓ compile
N Artifacts
      ↓ bind
N Processes + M Channels
```

通信题的一次 Submission 允许提交多份代码；这些代码以角色（Role）或逻辑进程为单位独立编译，并在同一 JudgeRun 中组成一个 ExecutionGraph 运行。

## 提交与编译模型

建议采用：

```text
Submission
  ├── SubmissionSourceUnit(role=A, language=C++)
  ├── SubmissionSourceUnit(role=B, language=Rust)
  └── SubmissionSourceUnit(role=C, language=Python)

JudgeRun
  ├── CompiledArtifact(A)
  ├── CompiledArtifact(B)
  └── CompiledArtifact(C)
```

规则：

- ProblemVersion 必须声明允许或要求的 Role 集合。
- Submission 必须在创建时形成完整 SourceUnit 快照，并进行 Role 唯一性与完整性校验。
- SourceUnit 可使用不同语言；语言与 runtime version 必须逐份记录。
- 编译可以串行或受控并行，但必须受 `CompileBudget` 限制。
- 任一必需 SourceUnit 编译失败时，不得启动不完整的运行图；结果中必须保留失败 Role、编译器输出摘要及内部 reason code。
- CompiledArtifact 必须内容寻址或至少具有稳定 hash，以保证 JudgeRun 可追溯。

## Architect Agent

必须负责：

- ExecutionGraph final design
- SubmissionSourceUnit / SourceBinding
- Role / Process identity model
- ProcessSpec
- ChannelSpec
- graph validation
- multi-source compile orchestration semantics
- multi-process lifecycle
- failure propagation
- communication metering
- byte/message accounting
- channel quotas
- scoring model
- security boundary
- per-process memory limit
- aggregate Judge memory budget
- compile memory budget
- memory accounting / OOM semantics
- scheduler resource request projection

必须明确以下不变量：

```text
一个 Communication Submission = 一组不可变 SourceUnit

一个 ProcessSpec = 一个明确 SourceBinding

所有运行进程的聚合资源 <= JudgeRun ResourcePolicy
```

### 内存与资源计量要求

建议使用父级 Job cgroup + 子级进程/角色 cgroup 的层级模型：

```text
JudgeRun cgroup (aggregate limit)
├── role-A cgroup (per-process / per-role limit)
├── role-B cgroup (per-process / per-role limit)
└── role-C cgroup (per-process / per-role limit)
```

计量至少记录：

- per-role peak memory
- aggregate current memory
- aggregate peak memory
- OOM / limit trigger scope
- compiler peak memory（按 SourceUnit 或 compile group）

不得通过“各进程 peak memory 直接求和”计算聚合峰值，因为各进程峰值可能发生在不同时间。聚合峰值应从 JudgeRun 级资源容器直接观测。

失败语义至少覆盖：

- 某 Role 超过自身限制：标记该 Role 资源超限，并按 ExecutionGraph 规则终止相关进程。
- 整体超过 Aggregate Memory Budget：JudgeRun 以聚合 MLE 原因终止，并清理全部进程。
- 编译阶段超出 CompileBudget：记录具体 SourceUnit 与 compile resource failure；对外 verdict 映射由统一 Error Model 决定。
- 资源计量或 cgroup 状态不可可信：不得伪装成用户 MLE，应进入 Infrastructure / System Error 路径。

示例：

```text
Alice.cpp ──> Process Alice
                  ↓
               Channel
                  ↓
Bob.rs   ──> Process Bob
```

或者：

```text
A ⇄ B ⇄ C
```

## Builder Agent

负责：

- multi-source submission loader
- per-SourceUnit compiler pipeline
- artifact binding
- ExecutionGraph executor
- Channel proxy
- communication metrics
- multi-process supervisor
- aggregate resource supervisor
- CommunicationJudge
- evaluator

## Worker Agent

负责：

- graph fixtures
- multi-source submission fixtures
- mixed-language communication fixtures
- communication tests
- metric tests
- per-process memory-limit tests
- aggregate memory-limit tests
- compile-memory tests
- sample communication problems
- visualization/debug UI
- documentation

## 完成标准

必须证明：

- 一次 Submission 可以可靠保存、展示并重判多份 SourceUnit。
- 每份 SourceUnit 的语言、源码 hash、编译产物和 runtime version 均可追溯。
- SourceUnit 与 ProcessSpec 的映射不存在基于数组顺序等隐式约定。
- 单进程 MLE 与聚合 MLE 均能稳定复现并正确清理全部进程树。
- 并发进程的聚合内存峰值由 Job 级计量得到，而非简单累加独立 peak。
- Scheduler 能根据 Communication Job 的总资源请求进行容量预留。
- 任一角色异常退出、阻塞、超限或 channel failure 时，failure propagation 行为符合冻结规范。

JudgeResult 可包含：

```text
status
score
cpu time
per-role memory peak
aggregate memory peak
communication bytes
message count
failure role / failure scope
```

---

# Phase 9：Multi-Node Judge Cluster

## 目标

让 Judge 从：

```text
一个 Worker
```

扩展为：

```text
Judge Cluster
```

加入：

- Judge Node Registry
- Heartbeat
- Scheduler
- Node Capacity
- Job Placement
- Failure Recovery
- Node Labels
- Node Features

## Architect Agent

必须负责：

- Scheduler model
- resource request
- resource reservation
- heartbeat
- node failure
- lease
- stale node detection
- capacity accounting
- placement algorithm
- job retry
- duplicate placement
- split-brain prevention

资源不能只按：

```text
1 submission = 1 slot
```

必须支持：

```text
CPU
Aggregate Memory
Per-Process Memory Constraints
Compile Memory / Compile Slots
PIDs
Disk / Temporary Storage
Judge Type
Process Count
Node Feature
```

对于 Communication Judge，Scheduler 必须依据 ExecutionGraph 与 ResourcePolicy 计算 Job 级资源请求。若编译阶段与运行阶段不重叠，可按阶段峰值进行容量规划；若允许流水化或重叠执行，则必须预留可能同时发生的资源总量。

## Builder Agent

负责：

- Judge Node Agent
- node registration
- heartbeat
- scheduler service
- resource reporting
- job dispatch
- failure recovery implementation

## Worker Agent

负责：

- Node API
- dashboard
- Grafana
- Kubernetes manifests
- Helm values
- Node test simulator
- load tests
- docs

## 完成标准

Node 故障后：

- Job 不永久丢失
- capacity 能恢复
- Job 可重调度
- 不产生不可控重复 Judge

---

# Phase 10：Production Hardening

## 目标

进入正式生产部署。

加入：

- Kubernetes
- dedicated Judge Node Pool
- gVisor optional/hardened runtime
- Autoscaling
- Full Observability
- Backup
- Disaster Recovery
- Security Audit
- Load Testing

## Architect Agent

负责：

- production threat model
- network policy
- secrets
- judge node isolation
- autoscaling semantics
- DR design
- security review

## Builder Agent

负责：

- deployment integration
- production configuration
- autoscaling hooks
- trace propagation
- alerting

## Worker Agent

负责：

- K8s YAML
- Helm
- Grafana dashboards
- alert rules
- operational docs
- runbooks
- load testing
- deployment scripts

---

# 10. 高推理任务清单

下面这些任务**禁止只交给便宜 Agent**。

```text
Sandbox security model

namespace / cgroup / seccomp

Judge state machine

JudgeJob idempotency

Redis ACK / retry / reclaim

Rejudge semantics

core DB invariants

resource scheduling

capacity accounting

Scheduler

Process lifecycle

signal handling

Interactive Supervisor

IPC

deadlock / idle handling

ExecutionGraph

Communication Judge

multi-source submission identity / artifact binding

aggregate memory accounting / OOM semantics

multi-process failure propagation

scoreboard consistency

authorization model

security boundary
```

---

# 11. 适合便宜 Agent 的任务

下面这些任务可以大量并行交给 Worker Agent：

```text
React pages

forms

tables

CRUD

DTO

migrations

repository boilerplate

OpenAPI client

YAML

Dockerfile

runtime definitions

fixtures

unit tests

integration cases

mock

seed data

README

API documentation

Grafana dashboards

CI YAML

lint

format

rename

mechanical refactor

simple middleware

admin pages
```

---

# 12. 每个 Phase 的标准 Agent 工作流

每个 Phase 均必须执行统一的阶段门禁（Stage Gate），不得将整个阶段作为一个无边界任务直接交给单一 Agent。

概要流程：

```text
1. Phase Intake / Scope Confirm
       ↓
2. Architect Design & Risk Classification
       ↓
3. Contract Freeze
       ↓
4. Work Breakdown & Ownership Assignment
       ↓
5. Builder Implementation
       ↓
6. Worker Parallel Tasks
       ↓
7. Automated Verification
       ↓
8. Builder Review / Architect Review
       ↓
9. Integration & Failure Injection
       ↓
10. Phase Acceptance
       ↓
11. Merge / Release / Archive
```

禁止采用：

```text
“Agent，把整个 Phase 写完”
```

此类任务描述缺少契约、责任边界、验证标准和故障处理要求，不得进入实施阶段。

## 12.1 完整分步协作与交付流程

### Step 0：Phase Intake 与范围确认

输入至少包括：

- Phase 目标与 Non-Goals
- 依赖的上一阶段能力
- 预期交付物
- 风险等级
- 验收指标

任何超出当前 Phase 的能力应记录到 Backlog，不得在实现过程中顺带扩张范围。

### Step 1：Architect 设计与风险分类

Architect 必须先识别：

- 哪些接口属于长期契约
- 哪些数据结构需要版本化
- 哪些操作涉及并发、一致性、安全或不可逆迁移
- 哪些失败模式需要主动注入测试
- 哪些文件或模块必须进入 Architect Review

输出 `DESIGN.md`、必要 ADR，以及可以直接用于实现的接口与不变量说明。

### Step 2：Contract Freeze

冻结内容包括但不限于：

```text
API / OpenAPI
DB invariants
Core interfaces
State machine
Error model
Queue semantics
ResourcePolicy
ExecutionGraph schema
Security boundary
```

冻结后，Builder / Worker 不得通过“为了方便实现”私自修改核心契约。确需变更时，必须提交 Design Change Request，由 Architect 评估影响后重新冻结版本。

### Step 3：任务拆分与所有权分配

将 Phase 拆分为可以独立验证的小任务，并为每个任务标记：

- Owner Agent 类型
- 允许修改的目录 / 文件
- 输入契约
- 输出契约
- 依赖任务
- 验收测试
- 是否需要 Architect Review

同一核心文件原则上不得由多个 Agent 同时进行结构性修改。需要并行时，应优先按模块、接口实现、测试、文档或 fixture 维度拆分，避免产生不可控冲突。

### Step 4：隔离实现

推荐每个任务使用独立 branch / worktree。

```text
main
 └── phase-N-integration
      ├── task/builder-xxx
      ├── task/worker-aaa
      ├── task/worker-bbb
      └── task/tests-ccc
```

Worker 对已冻结的核心 interface、schema 和安全策略默认只读。Builder 可以实现接口，但无权重新定义架构语义。

### Step 5：Builder 主实现

Builder 先完成贯穿核心路径的最小闭环：

```text
input
  ↓
core service / judge path
  ↓
persistence / result
  ↓
observable output
```

要求同时覆盖正常路径、主要失败路径、日志、指标和 integration tests。

### Step 6：Worker 并行工程任务

在核心路径稳定后，并行分配：

- CRUD / DTO / migration
- UI / admin page
- fixtures / regression cases
- runtime config
- docs
- monitoring / dashboard
- lint / mechanical refactor

Worker 任务必须包含明确的“禁止修改项”。如果任务需要突破边界，应停止并提交 Blocker，而不是自行扩大修改范围。

### Step 7：本地与自动化验证

每个任务合入 integration branch 前至少通过：

```text
format
lint
unit tests
module tests
integration tests（适用时）
race / security / failure tests（高风险模块适用时）
```

测试结果必须与代码版本绑定，禁止仅以 Agent 的文字声明作为完成依据。

### Step 8：分级 Review

- 普通 CRUD、页面、文档、fixture：Builder Review。
- 核心接口、状态机、queue、sandbox、scheduler、interactive、communication、ResourcePolicy：Architect Review。
- 安全边界与 Sandbox 修改：必须完成专门 Security Review。

Review 必须检查契约遵循情况，而不仅是代码风格。

### Step 9：集成、冲突处理与故障注入

合并顺序原则：

```text
contract / migration
      ↓
core implementation
      ↓
integration adapters
      ↓
parallel worker outputs
      ↓
UI / docs / dashboards
```

出现冲突时，以冻结契约为准解决，不允许通过临时兼容代码长期保留两套语义。高风险 Phase 必须执行 crash、timeout、duplicate、OOM、partial failure 等故障注入。

### Step 10：Phase Acceptance

验收必须同时满足：

- 功能验收
- Failure Path 验收
- Invariant 验收
- 安全验收
- 性能 / 资源验收（适用时）
- 可观测性验收
- 文档与运维说明完整

任何“功能可用但无法定位故障、无法回滚或无法证明资源边界”的实现，不视为阶段完成。

### Step 11：Merge、Release 与归档

阶段通过后：

- 合并至主分支。
- 记录 schema / protocol / runtime version。
- 生成 release notes 或 milestone summary。
- 归档关键 benchmark、测试报告及 ADR。
- 对需要迁移的数据提供 rollback / forward-fix 方案。

### Step 12：回归与变更控制

后续 Phase 若需要修改已冻结契约，必须说明：

```text
Why change
Affected modules
Backward compatibility
Migration path
Rollback path
Required rejudge / rebuild scope
```

不得以“后续功能需要”为理由直接破坏历史 JudgeRun 的可追溯性。

## 12.2 Communication Judge 的专项操作流程

通信题由于一次 Submission 包含多份代码，建议使用以下固定流水线：

```text
1. Validate Role Set
       ↓
2. Freeze SourceUnit Snapshot
       ↓
3. Resolve Language / Runtime per SourceUnit
       ↓
4. Compile SourceUnits under CompileBudget
       ↓
5. Persist Artifact Hashes
       ↓
6. Build SourceBinding + ExecutionGraph
       ↓
7. Validate Graph / Channel / ResourcePolicy
       ↓
8. Reserve Aggregate Node Resources
       ↓
9. Start Job-level Resource Container
       ↓
10. Start Role Processes
       ↓
11. Supervise Channels + Aggregate Memory
       ↓
12. Propagate Failure / Cleanup Process Tree
       ↓
13. Persist JudgeResult + Metrics
       ↓
14. Commit DB Transaction
       ↓
15. Queue ACK
```

其中第 8～12 步必须按照多进程总资源进行管理。特别是内存，应同时验证单 Role 限制和 JudgeRun 聚合限制，并保留可供复盘的计量数据。

---

# 13. Architect 应输出什么

每个复杂 Phase 开始前，Architect 至少输出：

```text
DESIGN.md

Goals

Non-Goals

Data Model

Interfaces

State Machine

Failure Cases

Concurrency Model

Security Boundary

Invariants

Test Plan

Migration Plan
```

Builder 应以冻结后的设计和契约为唯一实现基线；发现结构性问题时，必须先发起设计变更流程。

---

# 14. Builder 的完成标准

Builder 不以：

```text
代码写完
```

为完成标准。

而以：

```text
接口实现完成
+
integration tests passing
+
failure path covered
+
metrics available
+
no architecture contract violation
```

为完成标准。

---

# 15. Worker 的任务应尽量“小而清晰”

不要给 Worker：

```text
“帮我完成 Contest 系统”
```

应该给：

```text
实现 Contest List API Handler，
接口已经定义在 xxx.go，
不得修改 service interface，
补充成功、404、权限失败三个测试。
```

或者：

```text
为 cpp23/python3/java/rust/go
生成 runtime YAML 和对应基础编译测试，
不得修改 LanguageLoader。
```

Worker Agent 的交付质量高度依赖任务边界、禁止修改项与自动化验收条件是否明确。

---

# 16. Review 规则

以下文件/模块修改必须 Architect Review：

```text
sandbox/**

scheduler/**

execution/**

interactive/**

communication/**

queue semantics

JudgeJob

JudgeRun state machine

ResourceLimit

ExecutionGraph

core database constraints

authentication / authorization core

security policies
```

普通：

```text
frontend page
CRUD
DTO
docs
fixtures
```

可以 Builder Review。

---

# 17. 测试责任分工

## Architect

设计：

- Invariant Tests
- Security Tests
- Concurrency Tests
- Failure Injection
- Race Conditions
- Crash Recovery Cases
- Multi-process resource accounting tests
- Aggregate memory / OOM attribution tests

## Builder

实现：

- Integration Tests
- Module Tests
- End-to-End Core Flow
- Multi-source compile / artifact binding tests
- Communication failure propagation tests

## Worker

批量扩充：

- Fixtures
- Edge Cases
- Regression Cases
- Language Cases
- UI Tests
- API Cases
- Multi-source / mixed-language fixtures
- per-role / aggregate memory stress cases

---

# 18. 安全开发原则

Contestant Code 一律视为：

```text
完全恶意输入
```

永远不能假设用户代码：

- 会退出
- 不 fork
- 不攻击 kernel
- 不读系统文件
- 不访问网络
- 不耗尽内存
- 不无限输出
- 不制造子进程
- 不尝试利用 compiler
- 不尝试利用 checker
- 不尝试利用 runtime

Compiler 处理的 Source Code 同样属于不可信输入。

因此：

```text
Compile
```

也必须处于受控环境。

---

# 19. 高并发原则

OJ 的主要瓶颈不是：

```text
HTTP QPS
```

而是：

```text
Untrusted Compute Throughput
```

因此优化优先级：

```text
Judge throughput

Queue

Resource scheduling

Testdata cache

Runtime startup cost

Compile cost
```

对于 Interactive / Communication Judge，还必须优先关注单个 JudgeRun 内部的多进程并发与聚合资源占用。Node 容量评估应基于实际 ResourcePolicy，而非 Submission 数量。

高于过早进行：

```text
数据库分库分表
复杂微服务
Service Mesh
```

---

# 20. 不要过早引入

在真正产生需求之前禁止引入：

- Kafka
- 全量微服务
- Service Mesh
- Multi-region
- Event Sourcing 全家桶
- CQRS 全家桶
- 自研数据库
- 自研消息队列
- 自研 container runtime
- Kubernetes Operator

初期保持：

```text
Go Modular Monolith
+
Independent Judge Worker
+
PostgreSQL
+
Redis Streams
+
S3
+
Linux Sandbox
```

---

# 21. 推荐仓库结构

```text
oj/
│
├── apps/
│   ├── web/
│   └── api/
│
├── services/
│   ├── judge-worker/
│   └── scheduler/
│
├── internal/
│   ├── judge/
│   ├── sandbox/
│   ├── execution/
│   ├── communication/
│   ├── artifact/
│   ├── checker/
│   ├── queue/
│   ├── storage/
│   └── testdata/
│
├── packages/
│   ├── proto/
│   ├── judge-sdk/
│   └── language-spec/
│
├── runtimes/
│   ├── cpp/
│   ├── python/
│   ├── java/
│   ├── go/
│   └── rust/
│
├── migrations/
│
├── deploy/
│   ├── docker-compose/
│   ├── kubernetes/
│   └── monitoring/
│
├── docs/
│   ├── architecture/
│   ├── judge/
│   ├── sandbox/
│   ├── interactive/
│   └── communication/
│
├── scripts/
│
├── Makefile
└── README.md
```

---

# 22. 推荐 Agent Prompt 模板

## Architect Prompt

```text
你是该项目的 Architect Agent。

你的任务不是快速产出代码，而是保证系统核心设计正确。

请：
1. 明确目标和非目标。
2. 定义核心对象、接口、状态机和 invariant。
3. 分析并发、失败恢复、安全边界和边缘情况。
4. 给出 Builder 可以直接实现的 specification。
5. 不写不必要的 CRUD 和 UI。
6. 标出必须由 Architect Review 的代码区域。
7. 如果现有架构存在系统性问题，应优先指出，而不是绕过问题实现。
```

## Builder Prompt

```text
你是该项目的 Builder Agent。

Architect 已经冻结设计。

请严格按照 specification 实现模块。

要求：
1. 不擅自修改核心接口。
2. 完成正常路径和失败路径。
3. 添加 integration tests。
4. 添加必要 metrics / logs。
5. 如果发现 specification 无法正确实现，提交设计问题，不要私自重构架构。
```

## Worker Prompt

```text
你是该项目的 Worker Agent。

你的任务边界已经明确。

要求：
1. 只修改指定文件/模块。
2. 不改变核心 interface。
3. 不重新设计架构。
4. 补充测试。
5. 确保 lint / format / test 通过。
6. 如果需要修改任务范围之外的核心模块，停止并报告。
```

---

# 23. Agent 成本分配建议

粗略预算：

```text
Architect Agent
20% ~ 30%

Builder Agent
30% ~ 40%

Worker Agent
40% ~ 50%
```

不要因为 Architect 的代码量少就减少其预算。

Architect 的价值在于：

```text
用 20% 的 Agent 成本，
决定 80% 的系统方向。
```

---

# 24. 最值得投入 Architect Agent 的阶段

最高优先级：

```text
Phase 0
Architecture

Phase 4
Sandbox

Phase 7
Interactive Judge

Phase 8
Communication / ExecutionGraph

Phase 9
Distributed Scheduler
```

这些阶段应该优先使用最强推理模型。

---

# 25. 版本里程碑

## Milestone 1 — Mini Judge

```text
C++ only
Local
AC / WA / CE / RE / TLE
```

## Milestone 2 — Web OJ

```text
Problem
Submission
PostgreSQL
Frontend
```

## Milestone 3 — Async OJ

```text
Redis Streams
Independent Judge Worker
Rejudge
```

## Milestone 4 — Secure OJ

```text
Linux Sandbox
cgroup
namespace
seccomp
MLE / OLE
```

## Milestone 5 — Production Standard OJ

```text
Multi Language
S3
Versioned Testdata
SPJ
```

## Milestone 6 — Contest OJ

```text
Contest
Scoreboard
Priority
Rate Limit
High Concurrency
```

## Milestone 7 — Interactive OJ

```text
Interactor
Supervisor
Idle Timeout
Transcript
```

## Milestone 8 — Communication OJ

```text
Multi-source Submission
Per-SourceUnit Compile / Artifact Tracking
ExecutionGraph
Multi-process
Channel Metering
Per-Process + Aggregate Memory Control
```

## Milestone 9 — Distributed OJ

```text
Judge Cluster
Scheduler
Capacity Scheduling
Node Failure Recovery
```

---

# 26. 最终架构目标

```text
                     Control Plane
                          │
                Backend / PostgreSQL
                          │
                          ▼
                  Scheduling Plane
                          │
                 Queue / Scheduler
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
      Judge Node      Judge Node      Judge Node
          │               │               │
          ▼               ▼               ▼
       Sandbox          Sandbox          Sandbox
          │
          ▼
     ExecutionGraph
       /    |    \
      ▼     ▼     ▼
  Process Process Process
      \     |     /
        Channels
```

核心：

```text
JudgeJob
+
ExecutionGraph
+
Sandbox
```

这三个抽象正确，项目才能自然扩展到：

```text
Standard
SPJ
Interactive
Communication
Distributed
```

---

# 27. 一句话开发原则

> **Architect 负责高影响、难以回滚的决策；Builder 负责需要完整上下文理解的模块；Worker 负责边界清晰、可自动验证且返工成本较低的工程任务。**

每次安排 Agent 前先问：

> 如果该任务出现错误，其影响是局限于单个实现单元，还是会破坏一层架构契约？

前者优先交给 Worker Agent。

后者必须由 Architect Agent 主导或审核。
