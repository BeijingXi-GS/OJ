# 判题 MVP 契约与实现裁定

2026-10-02多题增量以受审 [目录扩展裁定](PROBLEM_CATALOG_EXTENSION.md) 覆盖下文单一A+B ID集合限制；其他执行/资源/清理契约保持。当前验证与产品状态以 [VALIDATION](VALIDATION.md) /PROGRESS为准，历史未验收描述不作为实时状态。

- 版本：judge-mvp/0.1.0；日期2026-10-02；状态reviewed，独立 /root/mvp_review 设计审阅通过，2026-10-02 采用本局部契约；运行验收尚未完成。
- 用户要求：服务器每题.in/.out文件、真实运行与AC/WA/TLE/MLE/CE/RE等结果，沙箱参考Anolix；旧compile接口保持兼容。
- Anolix已本地clone至gen/inputs/Anolix，commit `4502d2a798f3130d3c540c5c31d45251ca103623`，原源码未改。参考其namespace /capabilities /NoNewPrivileges /seccomp /cgroup /生命周期设计；不直接引入或复制其Go源码，不升级本项目Go1.25。采用Docker Engine的OCI/runc实现隔离，补上swap=0、只读rootfs、临时空间、输出限额与可验证清理；不是声称直接运行Anolix CLI。

## 服务与文件边界

API仍不编译或运行；新独立`services/judge-mvp`在本机127.0.0.1:8082监听，持有Docker CLI权限。该可信控制服务在主机管理独立容器；选手容器无Docker socket /宿主敏感目录 /答案目录，无privileged和额外capabilities。服务器不接受用户路径 /flags /镜像 /命令 /限额。旧idle judge-worker仍保留未来队列职责。

文件根`testdata/problems`，只支持固定problem_id `a-plus-b`。manifest严格JSON，显式case ID/input/expected/checker=`tokens-v1|exact-v1`；wall_time_ms=2000，memory_limit_bytes=67108864，最多8例、每个输入与标准答案最多65536bytes。case ID必须唯一，manifest及case文件必须普通文件，拒绝symbolic link /FIFO /设备文件（os.OpenRoot本身不拒绝根内symlink，须Lstat并核对已打开file.Stat）。允许限额范围wall500..5000ms、memory32..256MiB；输入输出文件名限定普通basename与.in/.out后缀，无路径或符号链接。manifest及目录使用os.OpenRoot防逃逸，所有文件限额读取；run工作目录0700且不挂载整个目录，源码 /输入0644、产物0555，启动必须真实验证Docker可读快照和Linux产物可执行；--mount参数使用CSV编码字段，支持空格 /逗号路径而非手写拼接。读取后固定字节与SHA256，期间不重读文件。至少四例：样例 /零 /负数 /大数。

每请求随机run ID，source SHA256、manifest+case内容组合SHA256、runtime image ID和compiler标识返回。源码64KiB /JSON256KiB与旧compilation.Decode规则一致。唯一工作目录在仓库.local/judge-jobs；仅root生成文件名，可写目录不挂给payload。每次编译产物属于本次run，并以SHA256绑定每个case，不复用旧产物。单worker一次一请求，繁忙429。启动校验profile /Docker Linux amd64 /cgroup v2及真实限制probe，失败不监听ready服务。

## 容器与监督

新镜像的工具层来自已有compiler镜像，最终FROM scratch复制工具层文件并加入静态judge-compile-helper，明确不继承基础镜像声明的VOLUME或日志 /入口配置；预先创建固定mount占位路径（只编译和流式导出产物），保留GCC14.2.0 /C++20。构建后profile锁实际镜像ID；所有create使用该sha256 ID，禁止运行时tag漂移。启动必须拒绝image.Config.Volumes非空；容器Mounts不得有volume类型。可信helper将g++所有诊断送stderr，成功后Lstat校验非symlink的非空普通产物≤32MiB再流式写stdout，不运行产物。控制服务将有界产物保存至本run目录，只有最终容器exit0及完整artifact /清理成功才可运行。

编译与运行共用完整UID10002 /init /network none /cap-drop ALL /NoNewPrivileges /默认seccomp /私有namespace /只读rootfs与log-driver none安全基线。编译：本run不可变源码快照文件只读挂载/source/solution.cpp；work为128MiB tmpfs，tmp16MiB；512MiB memory，同值memory-swap禁swap，pids64，cpus1；固定g++参数`-std=c++20 -O0 -fdiagnostics-color=never /source/solution.cpp -o /work/program`。总wall预算20s（含Docker attach启动），输出诊断保留64KiB /原始停止线1MiB，artifact32MiB。编译OOM /超时 /输出超限归SYSTEM_ERROR并提供编译阶段原因，正常g++编译 /链接失败归CE。helper内部失败使用保留exit125，归SYSTEM_ERROR。

运行：只读挂载本run产物与本run当前输入字节快照（不是原testdata文件）；标准答案不挂载。固定可信shell `exec /solution/program < /input.txt`，不拼接用户文本。每case新容器，UID10002、init、network none、cap-drop ALL、NoNewPrivileges、默认Docker seccomp、私有PID /IPC /UTS /mount /cgroup namespace、只读rootfs、tmp16MiB noexec tmpfs、shm8MiB、pids16、cpus1、memory题目限额与swap0、fsize16MiB、nofile64、core0。默认seccomp与network隔离不等于Anolix的自定义syscall名单；本机MVP不宣称生产安全或公网可用。

每case wall预算从发起docker start --attach开始，用主机单调时钟监督，**包含CLI /容器启动开销**，不是用户CPU时间。启动失败必须来自inspect状态 /错误归SYSTEM_ERROR；不会仅按docker CLI退出码推断程序结果。attach必须正常收完，或只有可与最终container ExitCode对应的exec.ExitError视为程序非零退出；但已记录casewall /OLE /caller cancel /job deadline等监督触发时，CLI被取消与最终ExitCode不一致按可信触发原因及最终状态映射，不能被通用attach错误覆盖；无可信监督触发原因的其他传输 /等待 /管道错误或无法对应状态都归SYSTEM_ERROR，不凭inspect exit0对不完整输出判AC。stdout/stderr各自原始字节计数，共享原始停止线1MiB；内部完整stdout保留至原始停止线供checker比较，前端仅展示stdout前64KiB /stderr前16KiB（展示截断可以AC，原始截断不可AC），超线立即取消attach并以独立cleanup context kill容器，继续drain /discard；输出未完整时不能判AC。

Dockerinspect.State提供ExitCode /OOMKilled /Error /Running。MLE只依据本容器OOMKilled事件证据，不按137推断。真实CPU累计 /内存峰值未可靠采集，返回null；不伪造0、不累加case峰值。必要启动probe必须读取memory.max、swap.max、pids.max、cpu.max、NoNewPrivs、Seccomp、CapEff并严格核对，且无挂载Docker socket。真实矩阵另证OOMKilled可支持MLE。

create /inspect /kill /list等每个控制命令独立context最多5s，stage启动前检查job未取消；整个job70s预算覆盖正常create /状态控制 /清理；job终止后的最终cleanup独立总预算15s、各命令最多5s，总调用留出额外15s，create前登记随机唯一name与run label；create超时 /失败也以成功daemon list查询定位该name+label并清理，不能把任意inspect错误当不存在。create返回后即使取消也必须清理已经创建的容器。退出 /超限后使用独立cleanup context、每个控制命令最多5s读取最终状态、必要kill并检查not-running，再`docker rm --force --volumes`并用唯一name /label检查容器已不存在；清理失败优先SYSTEM_ERROR，worker标tainted拒绝后续请求，不静默忽略。清理正常后删除本run目录，删除失败同样SYSTEM_ERROR /tainted。全局job超时归SYSTEM_ERROR reason=job_timeout；仅case自身wall预算触发为TLE。调用者取消归SYSTEM_ERROR reason=request_cancelled（不把取消判TLE）；HTTP调用已取消时不要求成功传回响应。worker进程异常退出的容器通过oj.judge.run标签可定位，重启时发现残留拒绝ready，交用户处理，不自动删除他人容器。

## 判定与接口

单case优先级：基础设施 /I/O /清理失败→SYSTEM_ERROR；其次可信OOM→MLE；监督wall超限→TLE；原始输出超线→OLE；非零exit→RE；完整输出checker匹配→AC，否则WA。观察依据最终可信状态加本次监督触发原因，不以最后出现的exit137覆盖OOM /超时。CE不运行case；普通业务失败仍继续后续case；SYSTEM_ERROR停止。总verdict为SYSTEM_ERROR优先，否则manifest顺序第一非AC，全部AC才AC。

`POST /api/v1/judgements`与内部`POST /judge`请求 `{problem_id,language:"cpp",source}`，HTTP200结果：

```json
{
  "run_id": "随机ID",
  "verdict": "AC",
  "message": "所有测试用例通过。",
  "duration_ms": 1200,
  "source_sha256": "sha256",
  "testdata_sha256": "sha256",
  "artifact_sha256": "sha256",
  "runtime_image": "sha256:imageID",
  "compiler": "GCC 14.2.0",
  "standard": "c++20",
  "compile": {
    "status": "success",
    "diagnostics": "",
    "diagnostics_truncated": false,
    "duration_ms": 300
  },
  "cases": [
    {
      "id": "sample",
      "verdict": "AC",
      "duration_ms": 120,
      "cpu_time_ms": null,
      "peak_memory_bytes": null,
      "exit_code": 0,
      "oom_killed": false,
      "stdout": "3\n",
      "stderr": "",
      "output_truncated": false
    }
  ]
}
```

result.verdict范围AC/WA/TLE/MLE/CE/RE/OLE/SYSTEM_ERROR；compile.status success/compile_error/system_error；其他必有字段与JSON类型按示例（nullable exit_code也允许基础设施未启动）。cases空数组用于CE或编译阶段失败，不捏造未运行case。checker使用原始bytes：exact-v1为bytes.Equal；tokens-v1为bytes.Fields分隔Unicode空白，逐token bytes.Equal且token数量相等。展示文本先strings.ToValidUTF8替换非法序列，再按UTF-8合法边界截断至各自byte上限，JSON转义膨胀只计入总响应上限，不改变解码后字段上限。输出展示全部是文本。reason/message提供有限中文说明，不泄露Docker原始主机路径、凭据或标准答案全文。

HTTP400输入不合法、404题目不存在、413超限、429繁忙，503服务不可用 /tainted，502协议错误；error沿用`{error:{code,message}}`。case文件加载 /运行infra失败是HTTP200 SYSTEM_ERROR，服务未启动是503。API client固定配置服务URL，超时110s，最大响应6MiB并校验枚举 /数组≤8 /每项输出上限。API /worker WriteTimeout120s，shutdown120s以容纳在途清理；整个worker请求70s，其中最多8×5s+编译20s，禁止多请求并行扩大资源。

problem详情新增`judge:{enabled:true,case_count:4,wall_time_ms:2000,memory_limit_bytes:67108864,checker:"tokens-v1"}`，仅表示已配置独立服务，不替代真实health /case运行。未配置judge服务则enabled=false。system的judge_enabled也只表示配置，worker前置严格拒绝保障无隔离时不会执行。旧/api/v1/submissions继续404，无DB /登录 /队列 /历史保存。

## 前端与验收

Frontend Builder只改apps/web，不加依赖。保留编译按钮，增加“运行测试”；编译 /判题共用请求占用锁，禁止并行。新judge状态 /结果与旧compile结果分开，修改源码标过期，恢复模板确认并清除两种结果，离页取消请求。显示文件测试数量、wall含启动开销 /内存限额、总verdict /逐例 /OOM证据 /诊断；未知CPU /峰值显示“未采集”，不显示0。不得继续写“程序从不运行”；编译区域仍准确标仅编译，判题区域明确会运行。异步加载、服务错误 /重试与响应式保持。

验收：真实AC /WA /TLE /MLE /CE（syntax和link）/RE（崩溃与return137非OOM）/OLE；token /exact差异；缺失文件 /错误manifest /沙箱缺项拒绝；无网络 /cap /权限 /只读rootfs /答案不可见 /进程后代清理 /取消 /无旧artifact；API输入 /协议 /超时 /busy /tainted单测；完整check与真实API smoke；独立实现及安全审阅。旧浏览器问题仍需复验，不伪造通过。完整Phase1 /生产隔离 /异步持久化未包含。

## 实现审阅采用的补充约束（2026-10-02）

- result增加必有artifact_sha256字段：成功编译后是实际普通产物SHA256，未得到可信产物则空字符串。每例只读挂载同一个本run产物文件；产物没有历史复用。
- 内部GET /problem返回当前服务器manifest的judge metadata，API GET /problems/{id}通过可信HTTP client读取，题目页不硬编码可变case数量 /wall /memory /checker。配置无效或服务不可用时返回503，不向页面声称旧限额。服务端文件加载失败的POST /judge仍HTTP200 SYSTEM_ERROR。
- 管理员修改testdata必须等待当前快照读取完成，维护者不得在一次读取期间并发修改目录。OpenRoot +类型核对防逃逸 /symlink，不宣称多个文件原子快照。
- 编译fsize=32MiB、runtime fsize=16MiB；启动probe同时确认GCC14.2.0，编译 /产物mount /exec全过程和含空格 /逗号的生成快照目录。
- cleanup后及最终run目录删除后重新检查caller /70s总预算；不能在清理期间超时或取消后返回AC。保留compile_error或case原始事实，但总体SYSTEM_ERROR优先，client允许compile_error +总体SYSTEM_ERROR；禁止success +CE、AC /WA非零或空exit、OOM却AC /WA /RE、SYSTEM_ERROR case之后还有case、必有字段遗漏 /非法null。
- 本机mvp supervisor在Windows不能直接kill活跃judge进程：仅对其持有随机临时控制token的服务请求POST /shutdown，服务停止接收并等待当前HTTP处理 /独立清理，120s预算。token不写日志 /页面 /公共API；API不代理该内部控制接口。无法正常关闭时保留judge并交用户检查，不强杀活跃judge。既有外部judge实例不由supervisor结束。
