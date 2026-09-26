# 领域模型与数据不变量

状态：proposed / 0.1.0-draft；参考 internal/domain，业务数据库 schema 尚未冻结。

## 身份与版本

| 对象                                          | 身份 / 生命周期          | 约束                                                       |
| --------------------------------------------- | ------------------------ | ---------------------------------------------------------- |
| User                                          | 用户稳定 ID              | 授权不依赖前端传来的 owner                                 |
| Problem                                       | 题目稳定 ID              | 内容通过 ProblemVersion 发布                               |
| ProblemVersion                                | 不可变版本 ID            | 声明 JudgeType、Role 集合及资源规则                        |
| TestdataVersion / Testcase                    | 不可变数据版本及 case ID | 存储 key + 内容 hash；不覆盖历史对象                       |
| Submission                                    | 提交稳定 ID              | 绑定用户、ProblemVersion 和完整源码集合                    |
| SubmissionSourceUnit                          | 源码单元 ID              | Submission 内 Role 唯一，源码 key/hash/语言/请求版本不可变 |
| JudgeRun                                      | 一次业务判题 ID          | N:1 Submission；rejudge 创建新 ID                          |
| JudgeJob                                      | 投递任务 ID              | 关联确定的 Run；attempt/fence 属于执行恢复                 |
| CompiledArtifact                              | 产物 ID                  | 关联 Run + SourceUnit + 实际 compiler/runtime + 内容 hash  |
| TestcaseResult                                | Run + case               | 同一 Run 最终 case 结果唯一，重复投递不重复写入            |
| Language / RuntimeVersion                     | 语言与不可变运行版本     | 编译器/运行镜像或 digest 可追溯                            |
| Contest / ContestProblem / ContestParticipant | 比赛范围身份             | 时间、权限、榜单冻结由 Phase 6 设计                        |
| ScoreboardEntry                               | 比赛投影                 | 可依据确定 Run 结果重建                                    |
| JudgeNode                                     | 节点身份和能力           | Phase 9 绑定 lease、资源和 feature                         |

## 提交、编译和重判

Submission 保存不可变 SourceUnit 快照；编译状态、诊断和产物放 Run 级 Compilation / CompiledArtifact。避免将原文“SourceUnit 记录编译状态”理解为重判时修改源码快照。

JudgeRun 解析并固定 problem/testdata/checker 版本；逐 SourceUnit 固定实际 compiler/runtime。通常沿用原版本；主动升级环境的重判必须记录新版本和原因。Run 的 SourceUnit 集合必须来自同一 Submission 快照，不能在运行途中重新读“最新”源码。

数据库拟设唯一约束：(submission_id, role)、(run_id, source_unit_id, artifact_kind)、(run_id, testcase_id)。快照变更需新身份；Run 终态不能原地重新排队。具体 UUID/ID 格式、FK、on delete、索引与事务隔离等待 P0-004 / Phase 2 审阅。

## 资源模型

ResourceLimit 单位明确为 ms、bytes、PID 数；任何运行参数都必须是正且有限的值，缺省策略由冻结设计给出，不允许 0 被隐式解释成无限。ResourcePolicy 区分 per-process、aggregate、compile 和 node overhead。

Job 聚合内存直接取同一 Job cgroup 的实际占用/峰值，不能累加不同时间的独立 peak。Scheduler 要考虑编译与运行是否重叠；至少预留阶段峰值与系统开销。Go 草案只声明字段，目前没有验证器、计量器或资源限制实现。

## SourceBinding 与输出

每个 ProcessSpec 显式引用 SourceUnitID 与 ArtifactID。所有必需角色编译完成、hash 和 binding 通过检查后才运行。失败结构保留 category、reason code、scope、role、observed、limit。用户视图不能泄露宿主路径、对象存储凭据、隐藏测试输入或无限量编译输出。
