# 通信题开发问题与解决方案

日期2026-10-05；root负责汇总，comm_backend/comm_frontend实现，comm_review独立Architect/Security。此档记录受审范围内的开发修正；外部Docker/审批故障单独见[环境问题](2026-10-05-local-docker-access.md)。所有C++源码仅作为数据，可信fake测试没有在宿主运行选手/教师程序。

## 已处理的开发问题

| 问题与原因                                                                                        | 处理策略/最终方案                                                                                                  | 实际验证与边界                                                            |
| ------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------- |
| 第一轮设计没有明确部分编译产物、旧case映射、终止证据和早期SYS形状，审阅未PASS                     | 补齐只有两success才有聚合artifact；旧exit映射receiver、OOM任一角色；真实kill/daemon确认；早期SYS允许零真实编译记录 | 补充实际落盘后独立设计复审PASS；不是实现/产品PASS                         |
| 默认Go命令访问用户缓存被拒绝且尝试拉模块，测试尚未进入                                            | 采用已有scripts/go.mjs的项目内缓存，GOPROXY/GOSUMDB关闭联网，不更新依赖                                            | backend problems/problemstore实际PASS；保留首次环境失败，不算代码测试FAIL |
| root metadata绑定helper误用不存在的Manifest.ProblemID，Go build失败                               | 明确传入可信catalog题目ID，manifest仅持有类型/规则/资源/用例；API外层也核对ID                                      | 已修正，待本轮完整检查绑定最终版本                                        |
| 部分编译sender success后receiver目录失败时，strict状态条件同时要求success与system_error，无法满足 | 分开最后真实编译状态与汇总SYS，保留sender产物但空聚合产物                                                          | backend五种编译失败可信矩阵已实际PASS，含目录失败                         |
| 双方stderr各未截断但合并超过16KiB，case截断不能仅取角色flag OR                                    | 按角色展示stderr拼接再有界display，并验证合并文本/合并截断一致                                                     | backend完整pipeline含合并截断已实际PASS                                   |
| RE/OOM/natural137 fake误要求正常payload，在空EOF时伪造transport错                                 | 只修fixture的异常角色行为：对端可见空EOF并自然退出2，返回匹配exec退出证据；未知attach/IO仍SYS                      | 三个首次失败保留，修正后filejudge全包实际PASS；不放宽生产监督器           |
| 部分容器启动超时被当未知状态SYS                                                                   | 仅共享deadline已确认且最终daemon peer未启动/exit0/noerror/noOOM/nonrunning时接受TLE；保持未启动角色无虚构exit/kill | 新partial_start_timeout可信测试实际PASS；真实启动超时待Docker             |
| wire允许terminated_by_supervisor=true但exit0/2/null，可能伪造退出豁免                             | 独立审阅要求true只能对应实际kill确认的exit137；OOM独立优先                                                         | validator已修，专项wire拒绝测试加入；待最终检查                           |
| 旧API题库fixture不知道新通信题元数据且继续走单源码编译路由                                        | fixture提供真实通信manifest限制，并走独立双源码端点；旧类型保留原路由                                              | 已修，最终API回归待本轮检查                                               |
| 前端EngineerResult与JudgeSummary交集使role_compiles丢失hash类型                                   | 使用Omit移除summary角色字段再定义工程师角色结构                                                                    | 前端typecheck/lint/build/verify-modes实际PASS                             |
| 从构造题切通信题遗留checker状态可能漏检.out                                                       | construction规则只作用batch；通信总是配对.in/.out且最多4例                                                         | 前端扩展verify-modes实际PASS                                              |

## A+B 增量前的修正与检查（历史）

独立复读进一步发现：旧cleanup在matching为空时返回zeroState，不能证明真实inspect确认未启动。后端新增cleanupWithObservation保留state/killed/observed，旧helper行为不变；通信attach已启动后发现容器消失必须SYS，合法部分启动TLE须真实inspect的unstarted状态。新增partial_start_missing=>SYS与partial_start_timeout=>TLE对照并全filejudge回归PASS。

另发现Go JSON大小写及Unicode SimpleFold别名可能覆盖同字段（例如byte_limit/Byte_Limit、source/ſource），null加alias也可能隐藏非法配置。Store与filejudge固定协议key统一限制ASCII小写[a-z0-9_]，保留题面/源码值原内容；Store另明确拒communication/starter_codes显式null。multipart、manifest解析/持久重载与request/worker wire补上alias/null回归，旧类型继续回归。backend三包tests/vet及模板null专项PASS，root最终check待执行。

五种编译失败测试进一步核对实际compile阶段数量/role，不能只凭结果记录证明失败后未编译；专项PASS。root当前API/filejudge专项-count=1 PASS（0.754s/7.527s），含最大JSON转义与源/产物/数据/限额绑定、三投影。前端通信共享启动超时改为至少一个角色未完成启动，保留已运行sender证据；教师不露工程说明，旧交互语义保持，SSR/type/lint/format窄修PASS。

上述修正后，当时root完整check exit0、独立五包test/vet/SSR PASS，五项审阅必修全部闭环，对应226文件80a历史摘要。可信测试PASS只证明受测本地逻辑；后续新增源码不能沿用该摘要和完整检查。Docker真实执行、挂载/资源/清理、上传重启、页面及旧supervisor最终验收均不在fake证据内。

## 最简单 A+B 增量与当前阻塞

用户要求一个角色给出a、b，另一个求和。新增communication-a-plus-b，发送端模板只传原始两个数，接收端才做加法；四例与32字节限额加入真实目录/manifest和题库。可信pipeline加入pair_sum/pair_wrong，断言raw消息、EOF、通信字节和规范绑定。教师新增A+B示例按钮、双模板及职责说明，原整数序列题保留；内置总数10。

新增第十题后，catalog测试的合成墙钟`1000+i*500`使最后一项到5500ms，越过既有5000ms上限。按fixture范围调整为`1000+i*300`，不改产品时限；API普通题编译次数改为按题型统计ordinaryCount，避免硬减两道通信题。新A+B API fixture明确sender stdout为原始输入、receiver stdout为标准和。

前端新增后type/lint/build/SSR/format实际PASS，独立增量代码与窄可信专项PASS。完整相关包回归在当前环境失败：默认临时目录与网络访问受限；root正常network grant及项目TEMP/TMP后Store重命名/假产物清理仍拒绝，精确目录read/write grant复验亦exit1。逐例数学答案正确仍因清理失败最终SYS，不放宽逻辑。详见单开[权限问题前因后果、实际策略与恢复方案](2026-10-05-local-test-permissions.md)。当前完整回归不PASS，按用户约定停止依赖验收。

真实环境步骤见[Docker问题](2026-10-05-local-docker-access.md)与[当前交接](../handoffs/2026-10-05-communication-local.md)。没有GitHub推送，不删除旧题/卷/未提交内容。
