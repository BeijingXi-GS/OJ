# Windows 恢复增量独立复审

日期：2026-10-07。任务：COMM-001/002/004、PUB-001。实际审阅者：`/root/recovery_final_review`，独立 Architect / Security reviewer。

## 当前结论

**本次恢复增量实现与本地工程回归 PASS，未发现本范围必修缺陷；真实 Docker 和产品最终验收 pending。** 已独立读取本轮完整回归日志，并逐项重算 250 份非 MD 文件及聚合摘要，全部与新 manifest 一致。本结论限于下述受审代码及本地工程检查，可支持开发分支 / draft PR 交付，不代表允许合并、生产发布或真实产品验收完成。

用户本轮明确恢复本地工作及独立审阅。本 reviewer 只读取代码、文档和既有证据，仅写本报告；没有执行测试、Docker、Git、宿主 C++、环境修复或资源删除。读取 root 实际执行的日志属于独立证据审阅，不冒称本 reviewer 亲自执行测试。

## 已读增量与判断

已读当前 `PROGRESS.md` 前部、适用职责约定、[兼容设计](../judge/WINDOWS_FS_COMPATIBILITY.md)、[上轮独立报告](2026-10-06-recovery-delivery-review.md)，并核对以下增量。

### 测试夹具的只读产物收尾

`internal/filejudge/special_test.go` 的 `checker_cleanup` 场景继续要求结果 SYSTEM_ERROR、Engine taint 且唯一 Job 留存，并检查 `program` 和 `checker-build/program` 均为普通只读文件。未用删除现场或放宽 taint 断言让测试通过。

随后执行的 `t.Cleanup` 只将本测试的两个可信假产物恢复 0600，供 Go testing 清理其 TempDir。`cppFixture` 在 `t.TempDir()` 下创建 jobs，测试使用的 Docker 传输是 fake；回调在断言后执行，并先于更早注册的 TempDir 清理执行。恢复失败通过 `t.Errorf` 报告。该修正没有调用生产清理、清除 Engine taint，或改变真实失败现场的保留语义。

独立读取 `.local/evidence/checker-cleanup-20261007.log`：`TestSpecialJudgeLifecycleAndPriority` 的 14 个子用例全部 PASS，主测试 2.89s、filejudge 包 3.470s。此日志支持可信 fixture / fake Docker transport 专项通过，不能证明真实容器清理、资源边界或最终产品已通过。

### 发布、生产清理与通信验收脚本

已重新读取 `store.go` 的 pending 查询、registered 三处检查和 publish 回滚路径。合法空 pending 隐藏未提交题目；畸形标记或查询失败返回 unavailable；最终目录独占创建成功才取得回滚所有权，清理失败保留标记；关闭最终子 Root 后才删除提交标记。设计新增文字明确这些行为，与既有实现一致。

已重新核对 Windows 清理入口，并重算 store、publication tests、cleanup、cleanup tests、engine 和通信验收脚本的 SHA256，均与上轮独立已读版本相同。因此保留上轮对并发发布、回滚所有权、固定四类产物恢复、taint 边界及通信脚本规范摘要/清理证据的局部结论。本次没有因测试夹具收尾修改扩大生产清理权限。

这些结论仍依赖可信宿主管理，不扩展为防御宿主同权限替换的通用竞态保证，也不增加断电持久性承诺。此前 Windows Job / build / artifact 符号链接子用例 SKIP 的证据边界继续保留；本轮未补跑这些用例。

### 学生题面

已独立复读 `examples/communication-a-plus-b/PROBLEM.md`，确认“模板是完整演示解法”等说明已删除。发送端/接收端分工、±10¹² 输入范围、32 字节、共享 5000ms、每角色 64MiB、公开样例和提交操作仍完整。

## 环境与未完成门禁

root 转述本轮正常 Docker 查询返回 Linux pipe permission denied；随后只读升级查询在自动审批阶段出现 stream disconnected，动作未执行。reviewer 没有复跑或代执行该命令，也没有独立诊断审批系统。传输故障不能被报告成安全审核否决，正常查询的权限拒绝也不能证明 Docker 引擎未启动。

真实通信矩阵、Linux/cgroup v2、双角色资源与清理、服务 ready、页面投影、教师上传与重启持久化、旧题型和监督器停止验收继续 pending。已有静态复审、fake 测试或 worker 版本字段均不替代这些证据。

## 当前局部摘要

以下局部摘要由 reviewer 独立读取；完整非 MD 版本绑定见末节。文档摘要记录本次读取时的字节，后续纯 MD 编辑不纳入非 MD 聚合摘要。

| 文件                                           | SHA256                                                             |
| ---------------------------------------------- | ------------------------------------------------------------------ |
| `internal/filejudge/special_test.go`           | `9a60504a4651a38b3111afd17d7530ebf8ab967082e3c670019b1d7f26afc8aa` |
| `internal/problemstore/store.go`               | `069a86a899ac87375c9cc9e7d362c9f9e2d9c5a6acb2ee340e1bf80a4dae1f34` |
| `internal/problemstore/publication_test.go`    | `9a905f1533e7764b3357d2619c726cafb86909f2e875bb05a52ea178651a3dad` |
| `internal/filejudge/cleanup.go`                | `fb8cfc210711d45b102bc4cc1011bbeb0846021916d0a13609e875483987d5c1` |
| `internal/filejudge/cleanup_test.go`           | `ccc636eb1e75e890b8bbcf1d14fbd9c25e8f9ccdd1a993b24f09fbf8ab0e7904` |
| `internal/filejudge/engine.go`                 | `ff858559d13d99808fbd6503f06fd8e59892368fd0bb8b3df44e78f053d5653e` |
| `scripts/communication-test.mjs`               | `5d32359acb04d21d8dd01b2ec8d7dd75acfb76a92d3701ff69c88b6c9e6bb6d6` |
| `docs/judge/WINDOWS_FS_COMPATIBILITY.md`       | `d7fb9f733330089c5b71ae236d6f3bcb58844766067e51d73af428fd99e948bd` |
| `examples/communication-a-plus-b/PROBLEM.md`   | `6a25fc858ec60e12ea4ce0d45788c63921e344b3e49cf2e424a06f305578e2dc` |
| `.local/evidence/checker-cleanup-20261007.log` | `a43523382a555331d358ef2d3674d3c32b51d6a8de7cd2ade5306950c2bb41b1` |

## 完整回归与版本绑定复核

独立读取 `.local/evidence/recovery-20261007-network-check.log` 全文。日志包含完整 `npm run check` 链路：格式检查、Go vet / ESLint、Go 各包测试、4 项 Node shutdown 测试、模式与通信页面 SSR 检查、Go build、TypeScript 和 Vite 构建，均完成且未见失败。filejudge 用时 7.002s，problemstore 0.614s；部分其他 Go 包明确标为 cached，不冒称每个包均重新执行。Node 4 pass / 0 fail，页面渲染脚本明确 PASS，Vite 构建完成。root 提供的命令返回值为 exit 0，与日志完成链一致；本 reviewer 未另行执行这些检查。

Vite 仍有既有大 chunk 警告，不是此次构建失败，也不证明浏览器性能已验收。SSR 检查和可信 fixture 不替代真实页面、服务及 Docker 容器执行。

另独立读取第一次 `.local/evidence/recovery-20261007-check.log` 的失败尾部，其中多个本机 HTTP 协议测试报告 `judge_unavailable`，整体 FAIL；历史失败未被删除或当作通过。root 记录该次尚无正常 network grant，取得正常授权后复跑完整链路通过。此先后事实支持区分测试连接权限与产品断言结果，不能单靠日志断言底层权限或审批传输故障的唯一根因。

已独立读取 `.local/evidence/recovery-20261007-snapshot.json` 及生成器规则，但未执行包含 Git 调用的生成器。规则为：root 使用 `git ls-files -co --exclude-standard -z` 去重并按 JavaScript `sort()` 排序，排除 `.md`；每项固定对象字段 `{path, sha256}`，文件 hash 基于原始字节；整体对无末尾换行的 `JSON.stringify(sources)` 计算 SHA256。

reviewer 仅使用文件读取和 SHA256 运算独立重算清单内全部记录，核对路径范围、重复路径、文件数量、顺序、逐项 hash 和聚合 hash。结果为 **250 / 250 一致，重复 0、差异 0**：

`5ffef55d8235aa1ef4f41e58be2e0a7adb4bef6505856a8cfda76cd2fe9a8a45`

该重算确认读取时清单中全部源文件与快照绑定；没有执行产品测试，也没有把文件摘要核对说成对 250 份文件逐行完成新的代码审阅。root 说明之后只编辑 MD；如非 MD 文件继续变化，本地回归与本结论需重新绑定。

| 证据                                                  | SHA256                                                             |
| ----------------------------------------------------- | ------------------------------------------------------------------ |
| `.local/evidence/recovery-20261007-network-check.log` | `8ad175f5f04f721b848a326e769bac8df1226106efcd5d1771d9bb54c5013f07` |
| `.local/evidence/recovery-20261007-snapshot.json`     | `bb2e0571686496d1163d5db438c7e60c64eb93319a76937a46b7eb21414ba682` |

开发分支 / draft PR 的实际远程创建与内容核验由 root 另行记录，本 reviewer 没有操作 GitHub 或确认远程交付成功。真实 Docker pending 继续作为单独门禁，不因本地恢复 PASS 关闭。
