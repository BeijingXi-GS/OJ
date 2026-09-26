# 阶段路线图与验收基线

范围基于原计划 Phase 0–10。**本文件不记录实时完成状态**，状态统一见 PROGRESS.md。每阶段启动前都要具体化 DESIGN / 任务卡，完成上阶段依赖和适用审阅。

| Phase           | 目标与主要产物                                                       | 主导职责                                              | 验收证据                                                                             | 暂不引入                        |
| --------------- | -------------------------------------------------------------------- | ----------------------------------------------------- | ------------------------------------------------------------------------------------ | ------------------------------- |
| 0 Foundation    | Monorepo、环境、工作流、核心契约草案及冻结评审                       | Architect 设计；Builder 骨架；Worker 工程             | API smoke、前端构建、基础服务健康、package/状态机/身份模型独立评审                   | 真实判题、业务 CRUD             |
| 1 Mini Judge    | Linux 开发环境下 C++ 编译/运行、输入输出、Exact Checker、CLI         | Architect 生命周期；Builder pipeline；Worker fixtures | AC/WA/CE/RE/TLE；超时退出竞态；子进程完整清理；编译超时                              | 外部不可信提交、并发队列        |
| 2 Web OJ        | Problem / Submission / JudgeRun、pgx/sqlc、迁移、Monaco 提交页、REST | Architect DB/授权审阅；Builder 服务；Worker UI        | 打开题目→提交→持久化→展示判题；越权失败；事务回滚；独立判题进程                      | API 内执行代码、开放部署        |
| 3 Async OJ      | Redis Streams、独立 Worker 消费、Rejudge、outbox/恢复                | Architect delivery/ACK/fencing；Builder 队列          | Worker 失联可恢复；DB commit 后 ACK；ACK 前崩溃不双写；新 Run 重判                   | 分布式复杂调度                  |
| 4 Secure OJ     | namespace/cgroup v2/seccomp、只读 root、无网络、编译隔离             | Architect 主导及 Security Review                      | CPU/Memory/PID/输出/文件/网络/子进程逃逸测试；TLE/MLE/OLE/RE；清理证据               | 未经安全验收的生产执行          |
| 5 Standard OJ   | S3、不可变 testdata、版本化语言/runtime、缓存、token/float/SPJ       | Architect 可复现和 Checker 边界；Builder 实现         | source/problem/testdata/checker/compiler/runtime 全链追溯；缓存失效；SPJ 超限        | 未冻结信任边界的 Checker        |
| 6 Contest       | 比赛/榜单、优先级、限流、准入、SSE、背压                             | Architect 一致性和调度；Builder 服务；Worker UI/负载  | Rejudge 不饿死比赛；冻结榜单一致；队列上限；事先定义负载和延迟达标                   | 按 API QPS 代替判题吞吐评估     |
| 7 Interactive   | Contestant↔Interactor、监督器、idle/EOF/pipe、transcript             | Architect 主导；Builder 实现                          | AC/协议 WA/RE/interactor failure/TLE/idle/deadlock；先退出/写阻塞                    | 多角色通信扩展                  |
| 8 Communication | 多源码、逐份产物、ExecutionGraph、通道计量、聚合资源                 | Architect 主导；Builder executor；Worker 测试         | Role 完整性；显式绑定；单进程/聚合 MLE；真实 Job peak；编译预算；failure propagation | 用峰值之和冒充聚合峰值          |
| 9 Cluster       | Node 注册/heartbeat、lease/fencing、容量预留、调度恢复               | Architect 调度；Builder node/scheduler                | 节点失效不丢任务；容量恢复；不接受过期结果；重复投递受控                             | 1 submission = 1 slot           |
| 10 Production   | K8s、专用 Judge 池、可观测性、备份/恢复、审计、扩缩容                | Architect 生产安全；Builder 集成；Worker 运维         | 恢复演练/RPO/RTO；负载和容量报告；安全审阅；告警与回退演练                           | 无真实需求的多区域/Service Mesh |

## 里程碑映射

M1=Phase 1 Mini Judge；M2=Phase 2 Web；M3=Phase 3 Async；M4=Phase 4 Secure；M5=Phase 5 Standard；M6=Phase 6 Contest；M7=Phase 7 Interactive；M8=Phase 8 Communication；M9=Phase 9 Cluster。Phase 10 为生产准入门禁，补齐原计划未单列的最终里程碑 M10。

## 保留决策

初期采用 Go 模块化单体 API + 独立 Judge Worker + PostgreSQL + Redis Streams + S3。Prometheus / Grafana / OTel 随实际服务与指标接入；后期再决定 containerd / gVisor / K8s。技术栈选择固定，不表示每项工具必须在 Phase 0 部署。
