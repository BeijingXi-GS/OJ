# 三模式、教师上传与交互判题验证

日期2026-10-03，Windows/PowerShell，Docker Linux amd64/cgroup2。root实际执行，roles_review独立审计代码和证据；各Builder可信单测与root真实资源验证分开登记。本轮局部设计、实现和本机产品已获独立PASS，实时状态见[PROGRESS](../../PROGRESS.md)。

## 版本和功能范围

分支task/MODES-001-interactive-2026-10-02；非MD源码176文件，SHA256 `8785b5dd63074e25633de5f6d78e52ddbaf4b94c6b6a4df08e3349a4fbb60b1c`。该摘要包含已有未提交基础工程与本轮增量，不代表已提交Git版本。无新增依赖/提交/推送。

固定运行镜像 `sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74`，GCC14.2.0/C++20；固定交互器source SHA256 `bd867a6c212c0b77d0792aa70296bed2c024fb95987a63e80901a89f5333c37f`。Anolix仅作已锁本地源码参考，commit4502d2a798f3130d3c540c5c31d45251ca103623，不执行其CLI。

内置5道普通题/24例与1道交互题/3秘密；另有3道本轮教师上传题（普通2例、交互各2例）。模式是本机手动演示，没有真实账号权限。学生服务端响应不含逐例秘密/错误；教师含教学详情；工程师另含退出/OOM/hash证据和只读debug。交互仅固定猜数协议，不能上传任意交互器。

## 实际检查

| 检查             | 实际结果与原始证据                                                                                                                                                                                                                                                                                     |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 完整工程         | 首轮npm run check exit0：format、Go vet、ESLint、全Go tests、Go/TS/Vite build；`.local/evidence/modes-check.log`。16MiB修正后的最终完整check也exit0，证据modes-final-check.log，含最大转义与超16MiB拒绝                                                                                                |
| 基础API/旧编译   | npm run smoke与mvp:smoke exit0；真实正常C++/语法错误/链接失败/仅编译循环源码通过                                                                                                                                                                                                                       |
| Linux可信单测    | judge:test:linux exit0；文件与Store单测均在受限Linux容器实际执行，TestCaseSymlinkRejected与TestStoreRejectsSymlinkRootsAndRegistration均PASS；modes-linux-unit.log，不把Windows跳过当通过                                                                                                              |
| 前端可信检查     | verify-modes8项通过，frontend lint/TS/Vite通过；独立reviewer另执行同类可信检查                                                                                                                                                                                                                         |
| 旧标准17项矩阵   | 09:50:19–09:53:12全部PASS，judge-matrix.json completed:true。含AC/WA/CE语法与链接/RE/非OOM137/TLE/真OOM MLE/OLE/展示截断/无效UTF8/隔离/后代/busy429与在途取消/文件故障/metadata/exact与tokens；故障文件finally原字节恢复                                                                               |
| 五题独立解       | 09:56:47–09:59:20，problem-smoke.json10次全部PASS，覆盖24例正确解AC与常见错误WA（共48case），每次无残留；不复用公开starter作为答案oracle                                                                                                                                                               |
| 真实交互13判定   | 09:31:55–09:34:20，interactive-matrix.json13提交/36case全部PASS：二分AC、越界最终答案WA、等待后监督停止WA、非法消息/超询问/EOF WA、CE、崩溃RE、137非OOM RE、不flush TLE、真OOM MLE、stderr OLE、后代清理WA。报告第14项只证明两容器存在后取消清理，不能当双方Running证据                                |
| 交互生命周期补测 | 10:07:09–10:07:34，interactive-lifecycle-smoke.json2项PASS：严格source/run绑定、双方State.Running=true后abort并0残留；仅本次interactor TERM，最终exit143非OOM、选手清理exit137非OOM，SYSTEM_ERROR停止后续例且双方/job无残留。首次观察器失败已另保留并记录，不删除历史                                  |
| 教师上传与三投影 | 10:00:26–10:01:33，teacher-problem-smoke.json6实际提交全部PASS：上传普通与交互题，缺省学生AC/教师WA/工程师AC；隐藏字段严格检查，教学输入/期望和双向日志吻合，工程字段存在；学生上传403、任意interactor拒绝且题库未增、teacher debug403/engineer200、0残留。交互教师!1是合法范围错误最终答案，秘密7/900 |
| 重启持久保留     | upload-persistence.json completed:true，3题所有发布文件SHA256与详情重启前后完全一致；10:09:13前新实例逐题独立源码AC2/2、testdata SHA256相同、0残留                                                                                                                                                     |
| JSON最大转义     | Client普通/交互8例最坏展示wire>6MiB接受，超16MiB拒绝单测通过；10:10:57前真实8组60000 NUL，student/engineer均AC8/8，响应290/8642720字节，fixture原字节恢复且自建文件清理；escaped-response-smoke.json completed:true                                                                                    |
| 真实浏览器       | [浏览器验收](BROWSER_QA_2026-10-03.md)：默认学生、教师实际上传/失败校验、三模式运行/分级详情/在途切换/代码保留/debug/手机布局/无warn-error通过；最终重启后刷新默认学生且显示9题，无warn/error，实际截图与AX已保存                                                                                      |

## 上传持久化对象

| ID                                      | 来源/题型                 | 结果                                                                       |
| --------------------------------------- | ------------------------- | -------------------------------------------------------------------------- |
| upload-396b76945f98c85be30bea7de065aee8 | CLI教师两数求和/普通题    | 重启前后AC2/2，文件与testdata hash不变                                     |
| upload-e31e9a631d3b4f7c885c89f70136da95 | CLI教师猜数字/交互题      | 重启前后AC2/2，文件与testdata hash不变                                     |
| upload-5996de397727f9c0404617993a76c448 | 浏览器教师猜数练习/交互题 | 重启前工程师WA与重启后独立解AC都绑定相同testdata hash，文件/元数据完全不变 |

10:06:32 judge严格真实预检ready，10:06:36 API启用judge/compiler并启动Vite。第一次停止supervisor未获得shutdown响应，后续只读确认全部拥有端口停止、0容器/job后才启动；没有冒称握手成功，详见[非阻塞退出提示](../issues/2026-10-03-supervisor-stop-warning.md)。Docker恢复、执行边界审阅修正、响应膨胀和观察器修正分别有[问题档案](../issues/2026-10-03-modes-development-corrections.md)。

## 尚未纳入与最终收束

本机同步单Worker，提交源码/结果不持久保存；上传题目/数据会保留。无真实登录、任意交互器、全局交错通信时间线、CPU累计/内存峰值、分布式队列和公网生产验收。本机CGO race和远程CI未运行。既有Monaco体积警告仍非阻塞。

最终工程检查、转义边界和重启浏览器均通过。10:13:32 modes-final-readiness.json completed:true：四服务200、9题、A+B恢复tokens4例、0容器/0job，源码176文件摘要绑定一致。独立[产品审阅](../reviews/2026-10-02-modes-interactive-review.md)已正式签署PASS，MODES-001..004完成；本机产品通过，生产/完整Phase1范围未包含。

最后MD收尾再次中断，实际服务/引擎停止；原签署证据原样保留。已按现有授权恢复Desktop，确认linux/cgroup2及0判题残留，新judge17:41:23严格ready、API17:41:25开启编译/判题。独立恢复报告modes-resume-readiness.json四服务200/9题/原A+B4例/0残留/同176源码摘要；modes-resume-judgement.json新实例对已上传交互题再次AC2/2、同testdata hash、0残留。源码/镜像无变更，不将上午实证冒称下午一直在线；[完整经过](../issues/2026-10-03-final-interruption-services.md)。

最终MD收束：format:check exit0，125份Markdown/456个本地链接目标全部有效，git diff --check通过。证据modes-doc-check.json completed:true，入口/四任务卡/交接/实时进度和问题档案已同步。完整工程不因仅MD收尾重复执行，源码摘要不变。
