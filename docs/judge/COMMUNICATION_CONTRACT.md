# 双角色通信题增量契约

2026-10-05，communication-mvp/0.1.0，reviewed。用户明确采用发送端/接收端两份C++源码，并允许先进行不依赖Docker的本地开发与单元测试，真实容器验收留待环境恢复。

固定双角色已本地实现，comm_review独立局部设计/实现PASS。新增最简单A+B沿用此契约，sender传a、b，receiver求和；增量窄审与可信专项PASS。新增前完整工程检查为历史PASS；当前完整回归因本地文件权限失败，真实容器与产品验收pending。详见[独立报告](../reviews/2026-10-05-communication-review.md)、[本轮验证](COMMUNICATION_VALIDATION_2026-10-05.md)与[权限问题](../issues/2026-10-05-local-test-permissions.md)。问题、策略和验证统一记录，不能把未执行或fake测试当真实沙箱通过。

## 固定双角色图

本轮只支持sender和receiver两份自包含C++20源码，各最多64KiB，按显式role绑定。数组顺序不决定角色，缺失、重复或未知角色均拒绝；普通和交互题不接受多源码。

逐角色独立编译，各20秒/512MiB，保留role/source hash/artifact hash/诊断。两份全部成功才启动case，任一CE保留失败role且零运行用例；未编译角色不伪造成功。

每例并发两个独立受限容器：sender只读自身产物和本例输入；receiver只读自身产物，没有输入、标准答案或sender产物挂载。可信代理单向转发sender stdout原始字节到receiver stdin，EOF后关闭receiver输入，无反向通道；receiver stdout按token比较.out。

通信byte_limit为1..65536，每例超限WA并终止双方；bytes_observed记录读取字节，bytes_transferred记录实际转发字节且不超过限额。不能把展示字符数、分块次数或write调用次数当通信量。sender stderr与receiver stdout/stderr共享1MiB执行输出停止线，超限OLE。

双方共享一个500..5000ms wall截止时间，包含启动/等待，不能在sender完成后重置receiver预算；死锁TLE。取消/角色故障/超限必须关闭管道，等待两个attach并独立确认两个容器清理，清理不可信则taint/SYSTEM_ERROR。

每角色32..256MiB/32pids、无网络、非root、现有capdrop/seccomp/只读挂载。aggregate_memory_limit_bytes固定为2倍角色强制限额，是总受限内存上界，未采集真实聚合峰值或父Job cgroup。完整Phase8任意图、多语言、父级资源计量和异步持久提交不在此MVP。

任务总预算70秒，包括两编译及所有case；最多4例，独立cleanup预算保留。失败优先为基础设施/清理不可信SYS、真实OOM MLE、共享wall TLE、共享输出OLE、通道越界WA、未被监督器终止的自然非零RE，双方正常才比较AC/WA。主动终止的peer退出不能误判原始RE。

bytes_transferred精确定义为代理成功写入receiver受控输入管道的字节数，不等于用户程序真正消费字节；超限只转发剩余额度内prefix。receiver提前退出导致的已确认关闭管道可按其自然退出/最终答案判断，其余未知传输错误SYS；不能无条件吞掉IO错。AC需要双方正常结束、EOF交接、attach与daemon状态可信。

## 数据、模板和上传

Manifest type=communication、checker=communication-v1，新增communication对象仅byte_limit；每例必须同组.in/.out，禁止interactor/checker_source/任意命令图或文件路径。旧类型禁止communication配置。

Definition和Teacher Descriptor新增starter_codes可选map，固定键sender和receiver，通信题必须恰好两键，每份UTF8且非空无NUL、≤64KiB。starter_code保留sender作为兼容字段，须与starter_codes.sender一致；旧类型不接受starter_codes。Store深拷贝map。输入/答案/图ID/byte_limit/资源限额加入不可变数据摘要。

新增communication-sum（32字节）和communication-maximum（16字节），各4组负数/零/大数/单元素数据和正确双角色模板。教师可填任意公开数学题面、配对.in/.out、双模板和byte_limit，固定单向图不接受任意interactor或C++特判混合。所有源码只在Judge沙箱编译/运行，API不执行。

## 请求、结果和绑定

通信采用独立公开POST /api/v1/communication-judgements，内部POST /internal/communication-judge；旧单源码端点/262144字节请求上限保持。新端点总JSON≤1MiB，允许两份64KiB源码被JSON控制字符转义；每份原始源码上限不变。Request为problem_id加source_units数组，每项role/language/source（language=cpp），顶层language/source空或省略。必须恰好sender、receiver，无重复/未知/null/未知字段或重复JSON键。旧端点不接受通信题或source_units，新端点不接受普通题。

Result新增role_compiles数组，每项role/source_sha256/artifact_sha256（成功时）/compile。顺序固定sender、receiver，与请求数组顺序无关；成功两项，CE/SYS允许一项sender或两项真实已编译记录。top source_sha256为规范sender/receiver顺序SourceUnit JSON的SHA256；top artifact_sha256为固定role与成功产物hash序列JSON的SHA256。编译汇总success仅当两份成功，CE为compile_error且注明失败role，系统故障system_error。未编译角色没有记录。

Case新增communication对象：bytes_observed、bytes_transferred、byte_limit、limit_exceeded、sender、receiver。两角色运行对象含exit_code（可空）、duration_ms、oom_killed、stdout、stderr、output_truncated、terminated_by_supervisor。case stdout为receiver答案，case stderr为有界双方诊断；原CPU和内存peak均未采集。旧题禁止带这些证据，通信禁止interactor/checker证据。

Metadata新增communication可选对象：roles固定sender/receiver、byte_limit、aggregate_memory_limit_bytes；health新增communication_contract_version接口版本标记，启动器不复用缺标记的旧worker。仍必须通过现有严格沙箱预检才能开启服务，当前版本真实通信集成未执行须在验收MD明确，不以health契约版本冒称通信矩阵通过。

Client继续严格拒绝递归duplicate keys、未知/null/非法长度或数值及状态组合。API提交前加载可信snapshot，绑定类型、图配置/byte_limit、caseID和顺序，验证每role请求源hash、真实编译状态、源与成功产物聚合hash。旧普通/交互/特判回归不放宽。

## 三模式页面与投影

学生和教师的role_compiles仅保留role与compile（本次提交源码的编译诊断）；不展示源码/产物hash。学生仍不获得cases、通信记录、隐藏输入/答案或运行日志。教师case.communication仅投影bytes_observed、bytes_transferred、byte_limit、limit_exceeded、sender_stdout、sender_stderr、receiver_stderr、sender_output_truncated、receiver_output_truncated；stdout展示最多64KiB，stderr最多16KiB。receiver stdout沿用case.stdout。工程师获得完整运行、退出、OOM、终止原因及hash记录。教师投影不含exit/duration/OOM/terminated等角色工程信息。

通信页有两个明确标注角色的编辑器和一个运行测试按钮，先分别编译再运行整组用例。任一源码修改使旧结果失效；切题/切模式取消请求且防止旧响应覆盖。切模式保留本地双源码草稿。学生只看到公开题面、限制、角色要求、简洁判定和自己的CE诊断；比较细则/容器/代理字节计数说明放工程师。公开byte_limit与每角色内存限制属于题目要求。

教师上传可选择通信题、填写byte_limit及两份模板，并选择配对.in/.out；提供两份正确示例、协议规则与错误处理说明。上传拒绝单模板、额外角色、混合interactor/checker源码和非法数据。已有类型上传及页面保持兼容。

## 规范化和验收门禁

SourceUnit规范JSON使用字段role、language、source，字段顺序与Go结构体一致，固定sender、receiver数组顺序。产物规范JSON使用字段role、artifact_sha256并固定sender、receiver顺序。hash均小写SHA256十六进制；先严格验证角色集合再规范化，不能用任意排序掩盖重复角色。按sender再receiver编译，首次失败停止后续编译，已成功记录保留，未运行角色无记录。

本地验证使用可信fixture/fake工具，覆盖请求边界、角色显式绑定、双方编译失败、EOF与提前退出、通信/输出上限、墙钟/取消、清理及taint、结果可信绑定和三模式投影；旧类型回归必须通过。独立审阅先签设计，后签实现/可信测试，明确真实容器验收未执行。

补充状态约束：top artifact_sha256仅两角色成功才非空；快照或编译前SYSTEM_ERROR允许零条role_compiles、compile.status=system_error、零cases和空top artifact，仍保留规范请求源摘要。进入编译后只能实际已尝试sender前缀：一条为sender，两条固定sender/receiver；CE不能零条。case兼容exit_code映射receiver、oom_killed为任一角色真实OOM、stdout为receiver输出；通信validator使用每role退出证据，不套单进程RE/AC推断，CPU/peak仍为null。terminated_by_supervisor必须有控制终止及最终daemon状态确认，不能只凭context结束、发出kill或attach取消认定；真实OOM始终优先。提前关闭仍需可信角色状态和EOF交接，转发字节不证明选手实际消费。教师descriptor沿用编码总量≤64KiB（含双模板与题面），单份源码64KiB不承诺所有最大字段可同时上传。

环境恢复后必须真实执行双角色AC/WA/CE/RE/TLE/MLE/OLE、通信越界、双方取消/清理、无答案挂载、上传重启持久化与页面验收，并补跑旧停止/重启待办及最终完整检查。当前仅可交付本地实现和测试，不能标记产品最终PASS或以历史SPJ矩阵证明新增功能。
