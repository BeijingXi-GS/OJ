# 文件判题验证记录

本档保留2026-10-02五题版本和此前停止/恢复历史；最新三模式、持久上传和交互增量状态由[PROGRESS](../../PROGRESS.md)维护，[使用说明](MODES_USER_GUIDE.md)与[浏览器记录](BROWSER_QA_2026-10-03.md)另行交付。下文“最新”指本档当日的五题验收，不代表当前产品范围。

日期2026-10-02，Windows /PowerShell，Node /Go使用现有锁定依赖，Docker29.8.1 linux amd64 /cgroup v2。此记录区分实现者实测与独立源码审阅；未声称完整Phase1 /生产安全或远程CI通过。

## 本轮恢复与五题验收（最新）

用户重新授权恢复已知阻塞并增加题目。初始Docker Linux管道不存在，启动已安装Docker Desktop后恢复linux amd64 /cgroup2；原5s预算3次info查询651/423/444ms通过。当前事实不能追认上轮11:54查询失败的具体原因。原始证据 `.local/evidence/docker-recovery-probe.json`。

18:50:06 judge严格ready、18:50:10 API judge_enabled=true并启动Vite。本轮统一internal/problems目录，全部入口/文件加载使用同一5题allowlist；metadata按题目ID查询且客户端拒绝跨题响应，preflight校验全部已登记题目。4道新题每题5例，A+B4例，总24组，模板为演示解法。

| 验证                                | 本轮实际结果                                                                                                                                                        |
| ----------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 五题完整详情、独立正确解/常见错误解 | 10次真实HTTP提交通过；每题正确解AC、错误解WA；每次daemon/job清理，problem-smoke.json completed=true                                                                 |
| 可信答案                            | catalog Agent Go oracle与独立reviewer Node BigInt/字符串oracle分别核对24组及样例，不执行选手代码                                                                    |
| 17项真实矩阵                        | 18:55:16→18:58:14，judge-matrix.json completed=true、全部passed；含全部verdict、真实OOM/非OOM137、输出/UTF8、隔离、后代、busy/取消、文件故障和metadata/exact/tokens |
| 完整工程                            | npm run check exit0，format/vet/ESLint/全Go tests/TS/Vite817modules；既有Monaco大包warning非阻塞                                                                    |
| API与旧编译                         | npm run smoke /mvp:smoke exit0，真实GCC正确/语法/链接/仅编译无限循环回归通过                                                                                        |
| Linux可信单测                       | 最初因新增oracle没有fixture挂载失败；只读测试输入适配后完整exit0，含symlink/24组oracle/诊断/metadata/全部题目预检；不是忽略失败或放宽提交权限                       |
| 实际浏览器                          | 同一IAB成功创建/操作5题；编译成功、AC5/5、逐例输出、CE、旧结果过期、恢复清空、在途切题和缓存隔离、手机布局通过；详见 [操作证据与未测分支](BROWSER_QA_2026-10-02.md) |
| 最终收束                            | 独立review要求启动IO错误统一脱敏；修正后受控supervisor关闭与0残留通过，再重启严格预检和最终AC，结果追加本节末尾                                                     |

开发期 [控制缓冲修正](../issues/2026-10-02-docker-control-buffer.md) 与 [Linux测试挂载适配](../issues/2026-10-02-linux-fixture-test-mount.md) 单独记录失败、原因、实际修正；不冒称首次即通过。B-JUDGE-002 /B-MVP-001已满足恢复条件；独立实现/证据结论见 [review](../reviews/2026-10-02-judge-recovery-problems-review.md)。下文记录上轮历史，不覆盖本节。

最终源码151文件 SHA256 `a07de43e447cea2b5af370225eafa13018a20cfee4d6c0bf011461f88026761f`，独立reviewer自己计算摘要一致并正式PASS、遗留必修0。[启动IO修正](../issues/2026-10-02-startup-io-diagnostics.md) 后19:15:06 judge严格ready、19:15:12 API开启判题，final-readiness.json在19:17:26→19:17:43实际验证四服务200/五题/独立数组和AC5例/零残留。最后npm run check及Linux可信单测均exit0（包含新的StartupIO专项）；109份MD本地链接目标与git diff --check通过。最终差异仅启动失败诊断分类，17项矩阵与页面证据来自相同执行路径的此前版本，未冒称全部重跑；最终严格preflight/AC直接验证最后实例。

本轮本机五题产品验收通过，服务保持运行。无新增依赖/提交/推送，数据卷保留；CPU/峰值null、未测crash recovery/race/公网生产/持久化等范围仍保留，下文历史“blocked”不代表当前状态。

## 来源、版本与运行证据

Anolix本地commit4502d2a798f3130d3c540c5c31d45251ca103623，副本保持原样。GCC14.2.0 /C++20，判题镜像实际ID：`sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74`。image.Config.Volumes为空；所有create使用该不可变ID。

修正版judge于11:41:13严格预检ready，API于11:41:15 judge_enabled=true；真实矩阵是在该版本重新编译 /启动后执行，最终API启动会再绑定最后协议修正。原始记录在 `.local/evidence/judge-matrix.json`：17项、completed=true、全部passed。以下为真实API → 独立judge → Docker限制容器结果。

| 实测                                                | 结果与检查                                                                                     |
| --------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| 正确A+B /错误42                                     | AC /WA，各4例                                                                                  |
| 语法错误 /无main链接失败                            | CE、compile_error、cases空数组                                                                 |
| abort /普通return137                                | RE；137的4例均exit137且OOMKilled=false                                                         |
| 无限循环                                            | TLE，各例超过自己wall后清理，非全局timeout                                                     |
| 分块new并memset实际增长                             | MLE，4例最终State.OOMKilled=true                                                               |
| 无界stdout写出                                      | OLE，原始停止线1MiB，共享stderr预算                                                            |
| 70k空格+正确答案                                    | AC且output_truncated=true；checker未用展示前缀                                                 |
| 非法UTF8字节                                        | WA，合法UTF8显示，没有API协议错误                                                              |
| fork子进程后父进程输出退出                          | AC；每请求查daemon与job目录确认无残留                                                          |
| 程序检查uid /NNP /socket /答案 /根写入 /外网connect | AC，所有预期隔离检查通过                                                                       |
| 实际进入case后第二请求 /API连接取消                 | 429 /传播取消并清理，无容器或job目录残留                                                       |
| 暂移sample.in /错误manifest                         | HTTP200 SYSTEM_ERROR、未运行case，finally恢复文件                                              |
| 改为1例 /3500ms /32MiB /exact-v1                    | 题目API返回真实新metadata；无换行输出WA；改tokens-v1后AC，testdata SHA改变，原manifest最终恢复 |

启动probe实际读取memory.max=67108864、memory.swap.max=0、pids.max=16、cpu.max=`100000 100000`、CapEff=0、NoNewPrivs=1、Seccomp=2；实际UID10002，仅lo网络接口，答案 /Docker socket不可见、根写入失败。完整编译 /产物导出 /readonly bind /exec走通，生成目录含空格和逗号，验证CSV mount参数。

另外 `node scripts/judge-refusal-test.mjs` 已真实通过：使用旧compile镜像的继承`VOLUME /data`，新进程exit1、reason=image_volumes_forbidden、未ready /监听，生成profile在finally恢复；结果在 `.local/evidence/judge-startup-refusal.json`。没有把“Docker容器能启动”当作隔离通过。

## 自动检查与独立审阅

- Go单测已覆盖：有界文件、类型 /重复ID /路径 /未知或追加JSON /策略拒绝、字节快照hash，checker /Unicode /exact差异，UTF8截断，共享输出预算，CSV空格 /逗号，响应必有字段 /矛盾AC /CE清理覆盖 /忙 /不可用 /请求取消，create响应丢失仍清理、attach不完整、清理期间cancel与deadline、清理失败taint、case timeout与OLE，API请求边界及预检缺项拒绝。
- Windows symlink测试SKIP，单独[记录](../issues/2026-10-02-windows-symlink-unit.md)；同一可信Go test在受限Linux容器实际全部PASS，symlink专项PASS，不挂Docker socket。
- 独立/root/mvp_review已只读设计与实现，并独立运行fake filejudge /httpapi单测通过。真实Docker矩阵由root执行，reviewer阅读原始证据，不冒充其亲自运行。
- Frontend Builder的typecheck /lint /build /格式检查已通过，只有[既有Monaco警告](../issues/2026-10-01-frontend-bundle-size.md)；最终完整check及browser结果在本节后续追加。

## 验收状态与实际限制

当前等待最终完整工程检查、浏览器复验和独立报告收束。B-MVP-001未经成功创建 /操作真实页面不能关闭。CPU累计 /可信内存峰值为null，不是0；wall包含启动开销。生命周期故障主要用可控fake单测覆盖，未通过故意切断主机Docker /制造真实磁盘损坏来测试；完整crash recovery /race /远程CI /公网生产安全未验收。

## 最终收束与停止（2026-10-02）

- 最后一次npm run check实际exit0：格式 /vet /ESLint /全Go单测 /TS /Vite build817modules通过，只有既有Monaco警告。最初check因刚编辑的本MD末尾空行格式失败，Prettier修正后重跑完整check通过，未当作首次通过。
- npm run smoke /mvp:smoke exit0，基础元数据 /readiness与四项旧真实GCC编译通过；测试隔离API未配置judge，不改变已启用产品的执行语义。
- 最终Linux可信单测exit0，含symlink与新增预检拒绝；独立reviewer自己的fake单测exit0，Windows symlink仍真实SKIP。
- 源码100文件，SHA256 3ee10a98dbff54fa0e567a6a18161a3b38aa67bab8e6190f9643fddb42bb7357（scripts/source-digest.mjs，Git非忽略非MD源文件排序hash）。17项真实矩阵运行在11:41修正版；其后仅增加client禁止SYSTEM_ERROR后继续cases及格式 /验证脚本，代码层检查通过，但最终运行实例未成功保留，不能说最终摘要版本已完整运行验收。
- 正常supervisor控制shutdown通过，judge退出且容器 /job目录无残留；最终重启11:54:40报cgroup_v2_required，judge退出1、supervisor未启动API /Web。最终AC smoke ECONNREFUSED8080，未成功执行。按[问题档案](../issues/2026-10-02-judge-final-preflight.md)停止，不重试、不弱化preflight；此分支也覆盖CLI失败，不能推断实际cgroup已变成v1。
- 浏览器仅读取接口文档与inventory（IAB id2、tabs空）；没有创建 /操作页面，B-MVP-001未关闭。独立代码审阅通过，最终产品 /浏览器验收blocked。
- 本轮严格OCI编译 /运行证据只覆盖新judge链路；旧编译按钮仍走原compile-worker、未升级为新完整profile。使用说明中“编译与运行共用隔离”均指新judge内部的两阶段，不扩大到旧API。
