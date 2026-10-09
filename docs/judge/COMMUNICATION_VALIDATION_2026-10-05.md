# 通信题本地验证记录

日期2026-10-05；Windows/PowerShell；root集成，comm_backend/comm_frontend实现，comm_review独立Architect/Security。用户允许先交付不依赖Docker的实现/可信测试，真实容器待办。

## 当前 A+B 绑定与结论

用户追加最简单通信题：sender读取并传递原始a、b，receiver读取消息并计算a+b。新增communication-a-plus-b与四例，保留旧sum/maximum；内置总数10。教师提供A+B题面/协议/双模板按钮，SSR改用新题。schema和固定通信图不变。

当前清单 `.local/evidence/communication-ab-20261005-local.json`：235个非MD文件，摘要 **71e866f2e74ab2f8e9d107a8cacaf9240ed03ded1c87bf963821b8ba9a3cb634**，使用下方相同JSON数组公式；工作分支与HEAD未变，未commit/push。源码冻结后仅收尾MD。

| 新增后检查                                   | 实际结果                                    | 证明边界                                                                                                                   |
| -------------------------------------------- | ------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| 前端typecheck/lint/build/verify-modes/format | comm_frontend实际exit0                      | 新A+B示例、双模板及SSR，既有Monaco告警；非真实浏览器                                                                       |
| 独立reviewer A+B专项与SSR                    | 实际exit0，增量必修0                        | 四组数据oracle/metadata/32字节，pair_sum AC与pair_wrong WA、原始消息/EOF/绑定；可信fake，不是C++执行                       |
| backend与reviewer新增后完整相关包            | exit1                                       | 默认临时目录rename拒绝、Store错误及HTTP连接失败，日志事实见各执行者记录                                                    |
| root network grant后项目TEMP/TMP五包-count=1 | exit1                                       | problems/judge-mvp PASS；problemstore/filejudge/httpapi失败，重命名或假产物清理AccessDenied；不登记整体PASS                |
| 精确.local/tmp读写grant后两个专项复验        | exit1，filejudge18.263s、problemstore0.189s | 逐例AC/WA已生成，但unlink假产物拒绝使最终SYS/taint，上传仍Store错误；日志communication-ab-20261005-permissions-recheck.log |
| 当前新源码完整npm run check                  | 未执行                                      | 上项失败后依用户阻塞规则停止，旧成功日志不沿用                                                                             |

**当前完整回归未通过；新增A+B窄审/此前可信专项PASS，真实产品pending。** 环境原因未定位，详见[新问题与解决策略](../issues/2026-10-05-local-test-permissions.md)。没有在宿主编译执行两份C++；没有降低清理可信性、绕过权限或伪造最终AC。

独立reviewer已对新清单235个文件重新计算原始字节摘要，唯一性、排序、逐项与总摘要完全一致，命令exit0；仅做只读证据绑定，不再测试环境。收尾MD格式执行exit0，git diff --check exit0；末次审计160MD/705本地链接/缺失0，新源码摘要不变，证据.local/evidence/communication-ab-20261005-doc-audit.json。

## A+B 增量前的绑定与结论（历史）

分支task/SPJ-001-student-construction-2026-10-03，HEAD7a952c6；包含原未提交工作，当时未commit/push。226份非MD非ignored文件摘要 **80a5377ba7141f550cdd2d8979a9a632e1cb12bab1ef678a170280be32a09936**。公式：git tracked+untracked去重排序，排除.md及ignored；各项{path,sha256(raw bytes)}组成JSON数组，再SHA256。历史清单与事实在.local/evidence/communication-20261005-local.json，历史完整check日志在.local/evidence/communication-20261005-check.log。新增A+B已改变非MD源码，此摘要不适用于当前版本。

独立[审阅](../reviews/2026-10-05-communication-review.md)已签局部设计、本地实现/可信测试PASS，五项必修0遗留。**最终产品pending**，没有真实Docker通信执行、真OOM/隔离/取消清理或真实浏览器结论。配置仍使用原锁定judge镜像sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74；本轮没有成功Docker inspect证明当前镜像存在/运行。

## A+B 增量前已实际执行（历史）

| 检查                                                           | 实际结果                  | 证明范围                                                                                                                   |
| -------------------------------------------------------------- | ------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| npm run check（GOPROXY/GOSUMDB off，已有项目缓存）             | exit0                     | 全仓format、Go vet/前端lint、Go可信单测、4项Node shutdown测试、verify-modes、Go build与tsc/Vite build                      |
| root API/filejudge专项-count=1                                 | PASS                      | 双源码最大转义、新旧端点边界、recursive duplicate/null/unknown/ASCII与Unicode alias、源/产物/数据/限额绑定、三投影         |
| backend三包tests/vet                                           | PASS                      | 题库/上传/loader与监督器；26生命周期、五编译失败、八完整pipeline、EOF/pipe/预算/取消/清理/观察证据/存储重载                |
| frontend typecheck/lint/build/verify-modes/format及文案窄修SSR | exit0                     | 双编辑器/教师上传实际组件渲染，UTF8源码/字节限额、配对数据、草稿/过期结果/取消、角色CE和三投影；非真实浏览器               |
| 独立reviewer五包test-count=1、同五包vet、SSR                   | exit0                     | filejudge/problemstore/problems/httpapi/judge-mvp；本人独立复读与可信验证，不是root自检                                    |
| node scripts/supervisor-stop-test.mjs --self-test              | 输出全部self-tests passed | 可信data/local Unicode import、继承stdout/stderr、signal、早退、exit/stdio close区分、保留首错；不证明旧三阶段真实停止通过 |
| git diff --check                                               | exit0                     | 已跟踪diff无空白错误；新文件另由工程format检查覆盖                                                                         |
| Markdown本地链接审计                                           | 首轮157MD/668目标/缺失0   | 最终文档补齐后另做末次format/链接检查，结果补入PROGRESS                                                                    |

保留原Monaco构建体积告警，不调高阈值掩盖；不新增/升级依赖。所有C++仅作为fixture/源码数据，未在宿主编译或执行。

## Docker本次实际失败与恢复

默认CLI配置/contexts读取Access denied，明确request_permissions返回授予后仍失败；默认pipe缺失。安装文件存在，隐藏启动Desktop命令exit0但没有ready证明。项目内空CLI配置显式查询dockerDesktopLinuxEngine曾返回pipe不存在；A+B增量期间最后查询实际exit1，改为Linuxpipe连接permission denied，仍没有Linux/cgroup2响应。请求该命名管道权限时工具返回permission path cannot be represented losslessly，未授予。没有成功owner容器查询，不能登记容器0、镜像存在或服务ready。错误变化不能证明引擎已恢复。

部分拟WSL/service/只读检查因自动审批通信中断未执行；未替换工具/执行者绕过。具体操作、已知事实与候选原因见[环境档案](../issues/2026-10-05-local-docker-access.md)，开发首次失败/解决策略见[开发档案](../issues/2026-10-05-communication-development.md)。

## 未执行

真实双角色AC/WA/CE/RE/TLE/MLE/OLE、通信越界、stdin/EOF/早退、同时取消与双清理/零残留、输入/答案挂载隔离；新题教师发布/重启/双编辑器实际页面；旧SPJ三阶段停止、六上传题持久快照对比和剩余页面验收。旧23项SPJ与其他历史真矩阵不能替代本版新通信执行。

环境恢复后按[使用指南](COMMUNICATION_USER_GUIDE.md)与[交接](../handoffs/2026-10-05-communication-local.md)验收。GitHub最终交付仍依赖当前真实产品验证及发布任务，不冒称已推送。
