# Windows 文件系统兼容性独立审阅

日期：2026-10-05。任务：COMM-001/002/004。实际审阅者：`/root/windows_review_resume`，独立 Architect / Security reviewer；实现者 root 与后端 Builder 不计为独立审阅者。

## 结论与范围

**局部设计 PASS；Windows 清理实现窄复审 PASS。题库发布实现复审 pending；工程全量回归和真实 Docker / 产品验收未签署。**

本次只读核对 [兼容设计](../judge/WINDOWS_FS_COMPATIBILITY.md)、现有 `internal/problemstore/store.go`、`internal/filejudge/engine.go`、通信清理路径及本机 Go 1.25.4 标准库源码。只允许修改本报告，没有执行测试、环境修复、权限申请、Docker、Git、服务或宿主 C++。

读取 `AGENTS.md`、`PROGRESS.md`、`DEVELOPMENT.md`、[COMM-004](../tasks/COMM-004-review.md) 与 [判题 MVP 契约](../judge/JUDGE_MVP_CONTRACT.md)。本报告允许的修改范围由 root 委派确认；实时任务状态继续由 root 在 PROGRESS 维护。

## 设计审阅依据

- 题目发布以同一个受限 `os.Root` 创建空 `.staging-<ID>` 标记，再独占创建最终题目目录。只有最终目录全部数据与注册文件写入、逐文件 Sync / Close、子 Root Close 完成后，删除空标记才是提交点。不存在无约束绝对路径 rename 回退。
- 读取入口、打开最终目录之后且任何数据读取之前、所有最终成功或失败出口均检查标记。第二次检查覆盖“第一次无标记，随后发布者创建标记与目录”的交错；末次检查防止写入中的注册错误污染整个列表。标记存在时隐藏；标记查询异常时不可用。
- final 目录独占创建失败不得取得其回滚所有权。仅删除本次确实创建的 final；必须确认其清理成功后才删除本次标记。final 清理失败保留标记，重启后仍隐藏；不回收来源不明的旧 pending。旧的无标记 version 1 注册继续走原严格校验。
- Windows 清理只恢复 `program`、`checker-build/program`、`sender-build/program`、`receiver-build/program` 四个固定产物。容器完整清理、Engine 未 taint 和直属 run 路径校验均是前置条件；运行期产物仍为 0555，容器仍只读挂载。
- 从 jobsRoot 的 Root 打开直属 Job，各目录先 Lstat 拒绝非目录与链接，再核对打开对象身份；固定子目录采取同样校验。产物先确认普通文件，持有已打开文件并核对 SameFile，在该固定 Root 中仅对单名 `program` 调用 Chmod；修改后再次 Lstat 与已持有句柄核对身份。关闭后、删除前再次核对直属 Job 身份。
- 缺失尚未编译的固定路径允许跳过；非普通文件、异常查询、身份不符、公开 Close API 返回错误、属性恢复或删除失败均保留原 SYSTEM_ERROR / taint 路径。不得全树 chmod、忽略清理失败或在 taint 后修改现场。Linux 不采用 Windows 属性恢复。

## Go 1.25.4 与已有复现的独立核对

本机 `D:/GO/VERSION` 为 `go1.25.4`。已读 `D:/GO/src/os/root.go`、`root_openat.go`、`root_windows.go`：

- `root_windows.go` 的 `chmodat` 直接采用 `O_OPEN_REPARSE | O_WRITE_ATTRS`。在已锚定子 Root 中对固定单组件 `program` 调用时，不跟随末级 reparse link；它读写目标对象的 Windows readonly 属性。这支持设计中恢复可删除状态的方式，不能外推为任意多组件路径或所有平台行为。
- Root 的目录操作以持有句柄定位；公开契约仍不承诺阻止硬链接、文件系统边界或挂载边界。正常 Judge 仅由可信控制端创建产物，不向选手容器暴露可写 Job 目录，恢复属性前所有容器必须已清理。
- `root_openat.go` 中 Root.Close 内部忽略底层 syscall.Close 返回值。因此实现应调用并检查公开 API 的返回值，但本报告不宣称可据此证明所有底层句柄关闭绝对成功；File.Close 可报告的错误仍应传播。
- root 提供的 `.local/evidence/windows-filesystem-probe-20261005.log` 实际记载 Root.Rename invalid handle、Root.Remove 空目录成功、0555 产物导致递归删除 AccessDenied、File.Chmod 失败、Root.Chmod 后删除成功。本 reviewer 读取了该日志，没有重新执行 probe，也不把日志解释为主机 ACL / 安全软件根因已经查明。

Root.Chmod 前后身份核对不能消除拥有宿主目录写权限者主动替换为硬链接的全部竞态。本局部 PASS 依赖现有可信宿主管理与容器无可写 Job 挂载边界，不提供对抗宿主同权限攻击者的通用 TOCTOU 保证。题库标记方案承诺进程并发和中断时的可见性，不新增断电持久性保证。

## 本次读取版本

以下为实现开始前读取文件的 SHA256；仅用于明确设计评审基线，后续实现须重新复读。

| 文件                                     | SHA256                                                             |
| ---------------------------------------- | ------------------------------------------------------------------ |
| `docs/judge/WINDOWS_FS_COMPATIBILITY.md` | `3b52f2c1d7340415a405456a22cd097383c22d044856da7e44931b62d92a0c47` |
| `internal/problemstore/store.go`         | `f83b8194a073dede366ec639d527d7c5c9a201b684318f98082147e4ef760182` |
| `internal/filejudge/engine.go`           | `43112eee1df3e34722ee9ab1ee41a8a79320b3788f30c45dc689c5bcc438bb33` |
| Go `os/root_windows.go`                  | `8793f156f62950501778dff93c27e4f5c631d1ef034a9ccbf550da2737e7d785` |

## 实现复审与验证门禁

题库发布实现复审 pending。需要复读实际代码与测试证据，重点核对所有提前返回、回滚所有权、提交与回滚失败时隐藏、并发读取与旧 version 1 兼容。Windows 清理实现增量结论及实际测试边界如下。

本 reviewer 未执行任何测试。Windows 若无法创建符号链接，相应用例必须记 SKIP；静态阅读或 Linux 单独结果不能称为 Windows 链接实测通过。当前设计 PASS 不代表已修复工程回归，也不代表 Docker Linux、镜像、cgroup、真实双进程清理或最终产品通过。GitHub 开发分支 / draft PR 的交付状态由 root 另记。

## Windows 清理实现增量复审

**窄实现复审 PASS，无本范围未解决必修项。** 已独立复读新增 `internal/filejudge/cleanup.go`、`cleanup_test.go`、`engine.go` 的 removeJob 与调用处、`docker.go` 的 stage 清理及通信双角色清理路径。未由本 reviewer 执行测试。

- `removeJob` 在任何文件变更之前拒绝 taint 与非直属 `run-` 路径。Windows 路径从 jobsRoot 锚定 Root，依次验证目录、持有对象身份；只对四个固定产物恢复 readonly 属性，修改前后核对普通文件与 SameFile；删除前再次核对 Job 身份。Linux 不进入属性恢复分支。
- File / 子 Root / 父 Root 的公开 Close 返回错误通过返回值或 `errors.Join` 传播。产物不存在允许跳过，非普通产物拒绝；未知 readonly 路径没有 chmod 回退。Job 清理失败仍由 Judge / Preflight 原有调用处传播 SYSTEM_ERROR / taint。
- stage 在容器创建失败与正常执行后均确认清理；通信双角色清理等待双方完成，任一失败先 taint。Job 清理延期调用在 taint 时保留现场，符合恢复属性必须发生在容器清理确认后的约束。

已读取 root 实际执行的 `.local/evidence/windows-cleanup-20261005.log`：Go 文件清理专项与 fake daemon 通信流水线输出 `PASS`，filejudge 用时 1.336s。包含不属于本 Job / taint 拒绝、四种 readonly 产物、已不存在 Job、缺失角色、非普通产物，五种编译失败及十种通信流水线场景，其中含 A+B `pair_sum` / `pair_wrong`。独立检查对应测试代码确实断言 Job 目录空且未 taint；该日志证明 Windows 可信 Go fixture 与 fake daemon 回归，不证明真实容器执行。

新增 `.local/evidence/windows-cleanup-boundaries-20261005.log` 输出 `PASS`，filejudge 用时 0.579s：未知 readonly 路径测试实际 PASS；Job / build / artifact 三种符号链接子用例全部因 Windows 缺少创建链接权限 **SKIP**。父测试的 PASS 不计为三个链接边界实测通过。已复读测试代码，其预期断言为拒绝链接清理且外部文件内容和 readonly 属性保持；仍需具备权限的 Windows 环境实际验证。

| 增量文件 / 证据                                              | SHA256                                                             |
| ------------------------------------------------------------ | ------------------------------------------------------------------ |
| `internal/filejudge/cleanup.go`                              | `fb8cfc210711d45b102bc4cc1011bbeb0846021916d0a13609e875483987d5c1` |
| `internal/filejudge/cleanup_test.go`（含链接与未知路径用例） | `ccc636eb1e75e890b8bbcf1d14fbd9c25e8f9ccdd1a993b24f09fbf8ab0e7904` |
| `internal/filejudge/engine.go`                               | `ff858559d13d99808fbd6503f06fd8e59892368fd0bb8b3df44e78f053d5653e` |
| `windows-cleanup-20261005.log`                               | `12db41ef60bba7250fb37d4e859ea073a131e248d91b1763eee5d10f74905c41` |
| `windows-cleanup-boundaries-20261005.log`                    | `fa4150ad4b27cfd3b602c7cce53984bd798919140031c1ed9f5100b34b761a1a` |

没有将未执行的 Close 故障注入、身份替换竞态或链接用例描述为实测通过。本次静态实现签署维持前述可信宿主范围，不扩大为对抗宿主同权限攻击或最终产品安全验收。
