# 本机三模式、上传与交互判题增量契约

版本 modes-interactive/0.1.0；2026-10-02；状态 design PASS，独立结论见 docs/reviews/2026-10-02-modes-interactive-review.md。实现与产品验收仍待完成。用户明确选择仅本机演示、手动切换，不加入登录。扩展已验收五题MVP，不替代完整Phase1/7/生产权限设计。

## 模式和公开返回

HTTP头 X-OJ-Mode 仅接受 student/teacher/engineer，缺省student，未知值400。页面显式选择并说明是演示模式，不是身份权限；本机使用者可以手动切换到教师/工程师。API仍绑定本机，不能宣称这能保护真实学生数据。

2026-10-03独立审阅通过的内部响应窄修订：Client最大JSON响应改为16MiB，覆盖8例新增input/expected、选手/交互器展示和两次编译诊断的控制字符转义。逐字段上限、64KiB文件、1MiB raw输出和上传请求及所有执行预算保持不变。保守字符串上界288KiB×8×6+2×64KiB×6=14.25MiB，剩余容纳固定结构/枚举/有界消息；超16MiB仍502拒绝。旧judge-mvp契约的6MiB仅是此前无输入快照的历史版本。

学生只得到整体verdict、通用message、duration_ms、compiler/standard、本人源码的compile状态/有界诊断，以及passed_cases/total_cases。不返回逐例数组、隐藏case名称、输入/答案、stdout/stderr、交互记录/交互器诊断、退出/OOM事实、source/testdata/artifact哈希或镜像。SYSTEM_ERROR使用通用消息不带内部原因。教师返回教学详情：逐例ID/verdict/输入/标准答案/选手输出及错误、交互器双向内容、基本时长。工程师返回上述详情并带run/hash/image、退出/OOM与交互器编译证据。公开结果带mode和problem_type；worker内部wire保留现有可信事实，API投影不可只在UI隐藏。

模式切换取消旧在途请求、清空旧结果、查询缓存包含mode，旧详细结果不能在切回学生时残留。公开样例/题面始终可见。教师和工程师可访问上传页面及POST /api/v1/problems；学生返回403。模式头不是真实认证。

## 教师上传与登记

复用existing题库框架，新增固定上传根.local/teacher-problems；API和worker同一配置 PROBLEM_UPLOAD_ROOT（缺省上述路径）。新internal/problemstore.Store{BuiltinRoot,UploadRoot}提供List/Get/Resolve/Create/ParseUpload，无可变全局注册。内置目录注册6题（原5题+guess-number）；上传仅接受服务器生成upload-<32hex>目录及严格problem.json注册记录，staging/随意目录不自动开放。Get/Resolve发现损坏登记拒绝并给脱敏错误，不忽略为已通过。

multipart/form-data：唯一非文件part descriptor是JSON；其余file part字段files，文件名来自描述内显式引用。Descriptor含title/summary/difficulty/description/input_description/output_description/notes/samples[{input,output}]/starter_code/type/checker/wall_time_ms/memory_limit_bytes/cases[{id,input,expected?}]/interactor?。无客户端id、路径、镜像、命令或flags。题面文本有界；总请求≤1MiB、descriptor和每文件≤64KiB、≤8例、至多16个数据文件。只有.in/.out普通basename，拒绝路径/保留名/重名/多余/缺失/压缩包与未知字段/追加JSON。题型和限额必须有效，采用现有wall500..5000ms、memory32..256MiB。

标准题type=batch（旧manifest空type按batch），checker tokens-v1/exact-v1，每例.in/.out。交互题type=interactive，checker interaction-v1，interactor=guess-number-v1，每例仅.in且内容为一个1..1000的秘密整数，expected为空；本次不开放任意交互器源码/二进制上传。全部验证后生成ID、写完整临时目录、写version/id/definition注册文件，再原子rename发布；失败不留公开半题，不覆盖已有题。发布后刷新列表、重启后仍可判题；运行读为不可变快照及SHA256。

filejudge.Load保持内置兼容，新增LoadWithStore(*Store,id)，Resolve给固定root后仍使用os.Root/普通文件/非symlink/有界读取。API和worker使用同一Store。客户端不能用旧compiled Exists拒绝合法上传ID，可信worker进行最终登记/来源校验；公众API先Get登记再提交。

## 交互题协议和隔离

内置guess-number题面说明：交互器先输出1000；选手flush后输出 ? x（1..1000），交互器回复 < / > / =，含义是秘密数与x比较。最多10次询问；选手用 ! x提交最终答案，正确AC，错误/非法格式/越界/超询问WA。选手stdout为协议，禁止把隐藏输入直接送给选手。本次固定guess-number-v1交互器C++源由judge编译、版本和hash可追溯，上传只改变题面/秘密数据。

选手和交互器各自独立受限编译，产物分别显式绑定；交互器编译失败SYSTEM_ERROR，选手编译失败CE。复用现有镜像/profile与固定argv，不在API/宿主执行程序。交互器只读挂自己的产物+秘密/input.txt，并固定argv[1]=/input.txt；选手仅挂自身产物，无input/答案/交互器文件/socket。双方运行均UID10002/network none/cap drop/NNP/seccomp/readonly/cgroup private/no swap/pids16/CPU1及manifest内存，不共享namespace或可写卷。

可信host协调器双向转发stdio，Docker create --interactive与start --attach --interactive；注入独立duplexCommand用于fake测试，旧batch command不变。通信/诊断原始共享1MiB停止线，展示每方向64KiB、stderr16KiB，截断不伪造协议成功。回合数由固定交互器限制10。case总wall（含双方创建/启动）与整体70s job、5s control、独立15s cleanup不放宽；关闭pipe/取消attach/分别确认清理，任一清理失败taint拒绝后续。

终态：交互器明确合法exit1时结束等待选手并WA，监督终止产生137不当成选手自然RE；若选手先自然非零/真实OOM/自身输出超限，按RE/MLE/OLE判定。总wall超限TLE（交互整体预算，不承诺定位是哪一方不flush）；交互器自然非0/1、OOM、输出异常或启动/attach不完整归SYSTEM_ERROR。AC必须协议交互器自然exit0与选手自然exit0且输出/attach/清理均完整，不接受监督强杀后AC。请求取消/整体预算/基础设施失败SYSTEM_ERROR优先，实际资源证据不编造。

worker Result新增可选interactor_compile/interactor_source_sha256/interactor_artifact_sha256；CaseResult新增有界input/expected及可选interactor证据{exit_code,oom_killed,stdout,stderr,output_truncated,participant_stopped}，Go类型统一为InteractorResult。Metadata增加problem_type，interactive checker允许interaction-v1；旧batch wire与现有必需字段保持。Client严格验证所有新字段/类型/状态矛盾，API再根据模式投影。

本轮内部wire升级要求新worker与Client一并重启：Metadata.problem_type与Case.input/expected/input_truncated/expected_truncated为必填非null，原batch字段未删除，公众compile接口兼容。judge health携带modes_contract_version；supervisor只复用该增量版本且ready的实例，旧/tainted listener不触发并行preflight或伪新验收。

独立审阅裁定：participant_stopped仅在交互器合法exit1导致监督终止选手、收束为WA时为true。其他WA非零选手退出不得例外，AC不得监督终止；无此证据且无OOM的137仍是RE。四路通信/诊断共享1MiB计量须保留触发源：选手触发OLE，交互器触发SYSTEM_ERROR。交互器exit1仅代表固定协议合法拒绝；配置、秘密输入解析或内部异常必须用非0/1退出。启动预检必须实际完成双stdio猜数probe才能ready。

教师input/expected取自本次worker结果中的不可变、有界快照，不从可变目录补读；新增input_truncated/expected_truncated标记。展示文本转UTF8与按字节截断不能改变比较、原始秘密、协议或hash。上传检查Origin为本机页面/API同源白名单且不开放CORS；无Origin本机CLI可用。这个限制及mode头均不是鉴权，也不提供真实身份隔离。

2026-10-03独立实现审阅窄补充：交互双方create/inspect的case deadline同样属于TLE，尚未启动任何attach时明确start_timed_out=true。此标记仅在interactive、双方编译success、verdict TLE、双方exit=null、无OOM/participant_stopped且确认cleanup时有效；job期限/取消/真正CLI故障仍SYSTEM_ERROR。启动后使用真实退出事实，不把cleanup期间才越过case deadline事后改判AC。Client严格拒绝非法/null marker。API根据可信登记题型核对worker交互证据，互动成功不能冒用batch AC；固定交互器source hash必须匹配guess-number-v1的实际常量digest。

## 验收与分工

root：HTTP模式/投影/上传路由、client新wire验证、真实联调/统一MD。catalog Builder：Store/loader/内置猜数/上传解析及原子发布。interactive Builder：双容器监督/固定交互器/新内部证据。Frontend Builder：切换器/查询隔离/教学与debug视图/上传表单/题型展示。独立reviewer只审设计/代码/真实证据与可信fake。

必须实际验证默认学生无详情、教师/工程师区别、模式在途切换和缓存、上传正常/拒绝/持久重启/可判、旧标准题回归、交互AC/WA/CE/RE/TLE/MLE/OLE、非法协议/超询问/不flush/EOF、取消/后代/两方清理、交互器系统故障/AC收束。工程check/适用Linux可信单测/实际浏览器和独立结论齐备才能登记done。新未知环境阻塞依用户一事一档并停止；开发实现/审阅修正继续授权范围。
