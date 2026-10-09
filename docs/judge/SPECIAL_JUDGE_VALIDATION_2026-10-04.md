# 学生信息与特殊判题验证记录

日期2026-10-04；执行root及已授权独立Agent；当前结论：实现与部分真实验收通过，最终产品验收blocked。不能把本记录登记为完整交付。

## 版本和环境

- 工作区 `outputs/oj`，分支 `task/SPJ-001-student-construction-2026-10-03`；无本轮commit/push、无依赖升级、无数据卷删除，保留原工作区全部未提交改动。
- 局部契约special-judge/0.1.0，先独立设计审阅后实现；旧judge/modes契约兼容。
- 真实Docker Linux/x86_64/cgroup2已在本轮实际连接；镜像锁定 `sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74`，GCC14.2.0、C++20。Anolix沿用已取得本地源码及既有锁定版本。
- 最终停止快照200非MD文件SHA256 `3406da9bef0efd1719a405bd48c2cc7206609368f0329f8012c5b6280bcf8df3`，不是已通过冻结。最新停止自测扩展与启动日志变更发生在完整check之后。

## 实际通过的检查

| 命令/检查                                           | 实际结果                                                                       | 证据和边界                                                                                                                                     |
| --------------------------------------------------- | ------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `npm run check`                                     | exit0，格式/lint/vet、Go tests、4个shutdown helper、8组前端模式检查及build通过 | `.local/evidence/spj-20261004-full-check-final.log`；含下载链接/textarea修正，后续停止脚本补测及mvp阶段日志不在此次完整检查版本内              |
| `npm run smoke`                                     | exit0                                                                          | `.local/evidence/spj-20261004-scaffold-smoke.log`：API启动/健康/元数据、无数据库fail-closed、缺省不启用判题                                    |
| `npm run mvp:smoke`                                 | exit0                                                                          | `.local/evidence/spj-20261004-compile-smoke.log`：真实合法编译、语法/链接失败、死循环仅编译、API到Compiler桥接                                 |
| `node scripts/judge-unit-linux.mjs`                 | exit0                                                                          | `.local/evidence/spj-20261004-linux-unit.log`：锁定镜像中可信单测，含真实Linux符号链接拒绝测试；不把Windows SKIP当PASS                         |
| `node scripts/supervisor-stop-test.mjs --self-test` | 生命周期Agent及独立reviewer实际通过首4类；Agent后续本地文件扩展自测也通过      | 动态导入/信号、提前失败、exit与stdio排空、首错保留；最新增加Unicode/空格相对路径和file URL导入及后代stdio，不能代替真实三阶段                  |
| 独立代码复读                                        | 核心契约/API/client/checker/安全边界、下载/编辑窄修无必修项                    | [10-03报告](../reviews/2026-10-03-special-judge-review.md)、[本次报告](../reviews/2026-10-04-special-judge-acceptance.md)；最终产品签署pending |

初次完整check只在format失败，记录 `.local/evidence/spj-20261004-full-check.log`；修正格式后完整重跑才exit0。构建仍有已归档Monaco体积非阻塞告警，没有靠调高阈值关闭它。

## 23项真实特判矩阵

`node scripts/special-judge-test.mjs` 修正后实际exit0，16:32:38结束，completed:true / checker_source_restored:true。完整报告 `.local/evidence/spj-20261004-special-matrix-passed.json`、日志 `.local/evidence/spj-20261004-special-matrix-fixed.log`。每项判题后要求Docker查询本身成功、owned容器为空、job目录为空，均记录cleanup_confirmed:true。

| 范围         | 已执行断言                                                                                                            |
| ------------ | --------------------------------------------------------------------------------------------------------------------- |
| 内置构造题   | 两个不同正确排列各AC6/6；相邻差1、重复数字、错误无解输出均WA                                                          |
| 教师上传特判 | 两种正确排列各AC3/3；错误答案WA及反馈；超过展示上限但未超过执行输出限的原始答案仍AC；三路径读取与可选空标准答案AC     |
| 选手失败优先 | CE零例且不执行特判；选手RE/TLE保持，不误算特判失败                                                                    |
| 特判故障     | 特判CE、exit2、exit137但无OOM、超时、真实OOM、大量输出分别SYSTEM_ERROR，不能误判选手CE/RE/TLE/MLE                     |
| 恢复与投影   | 修改的自建checker恢复后AC3/3；teacher/student实际都AC，teacher有bounded反馈且无内部证据、student无cases/特判反馈/hash |
| 旧题回归     | A+B AC4/4；交互猜数字AC3/3                                                                                            |

首次矩阵的TypeError不是Judge错误：[可选反馈脚本问题](../issues/2026-10-04-special-feedback-optional.md)已复验关闭；失败报告和首次生成题保留。规范checker2368字节，SHA256 `58aaf8c9b4d5c4d07eb6ec2f2fd4514c827545b9d14bbdbebd2e30f8ac17b765`，下载示例及后端例子一致。

限制：checker-output-limit报告确认SYSTEM_ERROR与output_truncated，但保存的摘要未包含engineer具体message；该fixture持续输出可能同时触发时间限制，需补单项完整返回、明确checker_output_limit（并区分16KiB展示截断与1MiB停止线）。独立reviewer要求更具体证据，仍待执行；不把现有摘要当这项精确归因证明。

## 上传、页面与持久化

目前实际目录13题：7内置（5普通/1交互/1构造）+6上传。3道历史上传保留，新增矩阵首败题/通过题、浏览器教师特判题。上传前后没有删除其他题目。

浏览器真实选择checker.cpp与sample.in/sample.out/impossible.in，发布 `upload-29db51251c6398020722c527cfd55470`，AC2/2并展示无解特判反馈。学生说明精简、工程师说明迁移实际可见；详见[页面部分验收](BROWSER_SPECIAL_QA_2026-10-04.md)。下载/编辑修正已代码复审，真实复验和截图未完成。

重启前保存6上传/34文件SHA256及公开题面元数据：`.local/evidence/spj-20261004-persistence-before.json`。重启后对比和各题重判未执行；不能宣称持久化新增量已验收。

## 停止测试与本次阻塞

实际共享终端Ctrl+C已出现helper“Owned judge stopped; container and job cleanup confirmed.”，外层PowerShell会话退出1；不把外层返回码冒称Node成功退出证明。真正Node退出事实由三阶段脚本独立观测。

新版真实测试 `.local/evidence/spj-supervisor-stop-2026-10-04T08-48-01-624Z.json` 仍completed:false：preflight未达到，仅bootstrap，45s阶段等待超时；Node exit0/stdio close0、实检清理0/端口关闭不能替代缺失阶段/helper标记。idle/running未运行。启动阶段诊断日志已加入，但申请重跑时审批通信再次中断，动作未执行。

[审批复发](../issues/2026-10-04-approval-resume-interruption.md)是当前外部阻塞，[测试观测档案](../issues/2026-10-04-supervisor-observation.md)是另一个已知待修点，不能混作Docker运行失败。16:56只读HTTP：5173/8080/8082不可访问、8081 HTTP200、jobs0；最后成功Docker查询来自16:48清理阶段，审批失败后未再查询。

恢复后按[具体交接](../handoffs/2026-10-04-special-judge-paused.md)完成阶段诊断/三阶段实测、重启持久化、页面、输出限精确证据及最终全量check/独立签署。当前[PROGRESS](../../PROGRESS.md)为blocked。
