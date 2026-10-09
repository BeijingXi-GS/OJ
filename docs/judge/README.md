# 文件用例判题 MVP 使用与开发

最新增量为学生页面精简、内置构造题和教师C++特殊判题，详见[特判操作与规则](SPECIAL_JUDGE_USER_GUIDE.md)、[本次实际验证](SPECIAL_JUDGE_VALIDATION_2026-10-04.md)。当前13题，新增量最终验收因审批通信复发停止；以下9题与三模式已通过状态属于历史基线，不能替代本次验收，实时状态只看[PROGRESS](../../PROGRESS.md)。

## 历史基线和标准题操作

当前产品包含5道普通题和1道交互题，支持中文题面、C++编辑、编译检查和真实判题，返回AC / WA / TLE / MLE / CE / RE / OLE / SYSTEM_ERROR。新增学生/教师/工程师手动模式、教师题面与数据持久上传，具体操作和本轮Agent职责见[三模式说明](MODES_USER_GUIDE.md)。实际状态见[PROGRESS](../../PROGRESS.md)；[旧五题验证](VALIDATION.md)为历史证据，本轮页面见[三模式浏览器验收](BROWSER_QA_2026-10-03.md)。

最新增量已独立验收通过，详细[完整验证](MODES_VALIDATION_2026-10-03.md)与[交接](../handoffs/2026-10-03-modes-interactive.md)记录当前9题、教师上传持久化、真实交互与恢复实例。下文标准题文件说明适用于batch，三模式和交互使用以上增量文档。

2026-10-02 用户授权恢复后，Docker Linux 与严格预检已通过，API /Web 启动成功；上轮停止记录保留在 [问题档案](../issues/2026-10-02-judge-final-preflight.md)。最终浏览器验收结论以 PROGRESS 为准。

| 题目         | ID              | 文件用例 | 重点                             |
| ------------ | --------------- | -------- | -------------------------------- |
| A+B          | a-plus-b        | 4        | 零、负数、整数边界               |
| 数组元素和   | array-sum       | 5        | long long、正负抵消、1000 个大数 |
| 数组最大值   | maximum-element | 5        | 全负数、单元素、末尾最大值       |
| 回文字符串   | palindrome      | 5        | 奇偶长度、中间不匹配、1000 字符  |
| 整数升序排列 | sort-numbers    | 5        | 重复数、负数、极值、逆序         |

每题提供完整中文描述、输入输出说明和样例。默认 C++ 模板含可运行演示解法，便于先体验整个判题过程；可替换为自己的解法。共 24 组文件用例，每题默认 2000ms /64MiB、tokens-v1。

## 启动与访问

在仓库根目录使用已有依赖与缓存。需要 Node22.19+ /24、Go1.25、Docker Linux amd64 /cgroup v2；Docker Desktop需处于运行状态。

```sh
npm run compiler:build  # 首次准备旧编译功能的工具镜像
npm run compiler:up
npm run judge:build     # 构建受限判题镜像并锁实际image ID
npm run mvp
```

打开 [题库](http://127.0.0.1:5173)，选择题目。点“编译”只编译 /链接；点“运行测试”会编译并在受限Linux容器中执行每例，显示整体及逐例判定。页面提供编译诊断、stdout /stderr、截断提示、wall与退出 /OOM事实。源码改动后已有结果标为过期，恢复模板会先确认并清空结果。切题会重建对应编辑区、取消旧请求，不保存跨题草稿或历史结果。

API绑定8080、旧compile-worker绑定8081、独立judge-mvp绑定8082、Web绑定5173，均仅本机。`npm run mvp`会等待judge严格预检成功再启用API；已有健康judge可复用。没有profile或预检失败则拒绝启动，不回退主机执行。Ctrl+C停止supervisor拥有的服务，并通过带临时token的内部shutdown等待judge清理；外部已运行的judge与compiler不由它停止。单独启动judge可用 `npm run judge`，API另设 `JUDGE_URL=http://127.0.0.1:8082`。

无需业务数据库、登录或队列；基础API `/readyz`仍检查数据库，与题面 /判题可用性不同。保留原PG15432、Redis与MinIO及其数据卷。此本机同步MVP没有历史保存，不是完整Phase1或公网生产OJ。

## 服务器文件与判定

文件在 `testdata/problems/<已登记ID>`，manifest显式绑定case ID、input、expected、checker与资源限额。维护者应在无运行 /读取中的请求时修改它；所有文件必须普通文件，不能是符号链接、目录、FIFO或设备，不接受用户指定路径。每请求读为独立字节快照并计算SHA256，payload只挂载自己的源码或产物与当前输入，标准答案留在可信checker一侧。题目统一登记在 internal/problems，单独增加目录不会开放新题；启动必须校验全部已登记题目的文件，内部 metadata 按请求 ID 返回且客户端核对绑定。

```json
{
  "checker": "tokens-v1",
  "wall_time_ms": 2000,
  "memory_limit_bytes": 67108864,
  "cases": [{ "id": "sample", "input": "sample.in", "expected": "sample.out" }]
}
```

最多8例，每份.in/.out及manifest≤64KiB；wall500..5000ms，memory32..256MiB。页面从独立worker读取当前metadata，不固定假定四例或某个checker。tokens-v1按Unicode空白拆token并比较原始bytes，不解析数值；exact-v1严格比较全部bytes，换行也参与比较。

| 结果         | 实际含义                                              |
| ------------ | ----------------------------------------------------- |
| AC           | 编译成功且全部文件用例的完整输出匹配                  |
| WA           | 正常退出，输出与标准答案不一致                        |
| TLE          | 当前case的wall超过限额，包含容器启动开销              |
| MLE          | Docker最终状态明确OOMKilled；退出137本身不构成MLE     |
| CE           | 正常GCC语法或链接失败，不运行case                     |
| RE           | 程序非零退出或异常信号，无OOM /超时 /输出超限优先证据 |
| OLE          | stdout+stderr原始总输出超过1MiB                       |
| SYSTEM_ERROR | 文件 /基础设施 /整体70s预算 /清理故障；不会冒充WA或CE |

每例独立容器，普通失败继续其他例，基础设施失败停止。总结果按manifest第一非AC，SYSTEM_ERROR优先。stdout仅展示64KiB、stderr16KiB，但checker使用完整原始输出；展示截断可以AC。旧“编译”按钮仍走原compile-worker，仅编译不运行，未套用新judge完整OCI profile。CPU累计与可信内存峰值未采集，返回null并显示“未采集”。源码64KiB、请求256KiB；一worker一次一请求，忙时429，清理失败taint后503拒绝继续。

## 沙箱来源和边界

Anolix源码已拉取到本工程外的 `gen/inputs/Anolix`，锁commit4502d2a798f3130d3c540c5c31d45251ca103623并独立阅读核心源码。此次参考其OCI namespace /capability /NoNewPrivileges /seccomp /cgroup思路，独立实现Docker Engine的OCI/runc适配；没有直接运行Anolix CLI、复制其Go源码或升级Go依赖。

新judge链路内部的编译与运行都使用UID10002、init、network none、cap-drop ALL、NoNewPrivileges、默认Docker seccomp、私有namespace、只读根、禁swap、PID /CPU /tmpfs /输出限额，关闭daemon日志；新镜像FROM scratch复制工具文件，不继承旧Redis基础镜像的 `/data` 声明卷。实际image ID锁定在 `.local/judge-profile.json`，启动执行真实编译 /挂载 /exec及内核限制probe。API只做固定服务HTTP桥接，不持Docker执行路径；不可信容器没有socket、答案或宿主敏感目录。

镜像仍含工具链；当前wall含启动开销，Docker默认seccomp不等于Anolix自定义syscall白名单。强制终止主机进程后的崩溃恢复、生产安全、持久化与分布式调度属于后续阶段。硬退出可能留下带 `oj.judge.owner=file-mvp-v1` 标签的容器；重启会拒绝ready，应先记录并由用户核实处理，Agent不自行删除不明容器。

## 检查与Agent职责

```sh
npm run check
npm run smoke
npm run mvp:smoke
npm run judge:test       # 真实API判定、隔离、取消、文件故障与metadata矩阵
npm run judge:test:linux # 可信单测在受限Linux容器内，含symlink边界
node scripts/problem-smoke.mjs # 五题独立正确解AC /常见错误解WA及清理
npm run judge:status
```

judge:test只在本机开发实例执行；故障注入仅修改本轮fixture，并在finally恢复原字节，期间不要手动修改题目文件。原始脱敏证据保存在 `.local/evidence`，详细摘要在MD中；这不代替浏览器或独立审阅。

| Agent                                     | 本轮职责与修改范围                                                                          | 后续具体任务                                                                                                  |
| ----------------------------------------- | ------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| root /Backend Builder                     | JUDGE-005：恢复Docker、脱敏诊断、worker多题接线/preflight、真实矩阵、UI验收和统一MD         | 维护启动/清理与用例回归；后续受审Phase1异步持久化/完整资源计量另拆任务                                        |
| /root/problem_catalog /Builder            | JUDGE-006：统一目录/allowlist、API/client/loader、中文题面与24组fixture、独立解法smoke      | 新题先登记Definition，新增目录manifest和边界用例；同步oracle、详情与正确AC/错误WA验证                         |
| /root/frontend_followup /Frontend Builder | JUDGE-007：仅apps/web；移除固定题号、补artifact DTO，复核按题目重建编辑区和取消锁，无新依赖 | 题库搜索/难度筛选/分页与空态、详情加载/错误/重试、结果与键盘可访问性分别拆卡；API契约先受审，再按既有框架接线 |
| /root/judge_reviewer /Architect+Security  | 独立设计/实现阅读、fake测试和答案oracle、runtime/浏览器证据审计，仅写review                 | 复核源码/镜像绑定、路径与输入边界、判定事实和清理；不把root实测登记为自己实测                                 |
| 后续明确认领的Worker                      | fixture /样式 /文档等已定义小任务                                                           | Monaco体积优化、更多可信样例与页面可访问性；不得自行更改核心判定或执行权限                                    |

完整前端扩展步骤见 [实施指南](../frontend/IMPLEMENTATION_GUIDE.md) 与 [任务包](../frontend/AGENT_TASKS.md)。所有问题一事一档放docs/issues；新未知阻塞交用户并停止。
