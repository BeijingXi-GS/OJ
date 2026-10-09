# 通信题环境与真实验收

本指南提供可复现的本机通信题环境和 [A+B 教师数据包](../../examples/communication-a-plus-b/README.md)。固定采用发送端、接收端两份 C++20 源码，沿用 [communication-mvp/0.1.0 契约](COMMUNICATION_CONTRACT.md)。目标页面为 [通信 A+B](http://127.0.0.1:5173/problems/communication-a-plus-b)。

**2026-10-07 交付状态：当前完整本地检查、脚本语法和四组示例静态核验通过；真实 Docker 通信矩阵待执行。** 本轮 Docker Linux pipe 查询 permission denied，提权只读诊断因审批传输故障未执行，尚无当前版本 Linux/cgroup v2、严格预检、服务 ready 或资源零残留证据。见[本轮原因与对策](../issues/2026-10-07-approval-diagnosis-and-recovery.md)、[当前验证](RECOVERY_VALIDATION_2026-10-07.md)和[实时进度](../../PROGRESS.md)。本地检查不能替代真实容器执行；后续结果必须绑定当次源码与镜像。

## 正常启动路径

需要 Node 22.19+（22 系列）或 24、Go 1.25，以及可访问的 Docker Linux amd64 引擎和 cgroup v2。Windows 使用 Docker Desktop 的 Linux 容器环境；Windows 能编辑源码或运行 Go/Node 不等于具备 Linux 隔离。执行前阅读当前 PROGRESS 和已有环境问题，保留 `.env`、数据卷和教师上传题。

在仓库根目录依次运行：

```powershell
npm run setup
npm run compiler:build
npm run compiler:up
npm run judge:build
npm run mvp
```

`setup` 安装锁定的 Node/Go 依赖并保留已有 `.env`；已完成 setup 的同一副本可跳过此步。`compiler:build` 和 `compiler:up` 准备既有独立编译服务。`judge:build` 构建 Linux 判题镜像，将实际 image ID 写入 `.local/judge-profile.json`。不要手填虚构镜像 ID、修改预检使其放行，或在宿主机直接编译/运行选手程序。以上步骤可能下载依赖或镜像，需要对应正常网络与 Docker 访问权限。

`mvp` 在现有监督器下等待 judge 严格预检，通过后启动 API 和页面。预检检查 Linux/amd64、cgroup v2、镜像与容器限制、真实编译/产物挂载/执行，以及旧容器与任务目录残留；缺失 profile、旧 worker 版本不符、查询失败或预检失败都拒绝启动。worker 健康接口需要 `communication_contract_version=communication-mvp/0.1.0`，版本字段本身不能证明通信矩阵已通过。

| 组件     | 本机地址                | 用途                                   |
| -------- | ----------------------- | -------------------------------------- |
| 页面     | `http://127.0.0.1:5173` | 学生双编辑器、教师上传、工程师信息     |
| API      | `http://127.0.0.1:8080` | 题目与提交转发；不编译、不运行选手代码 |
| compiler | `http://127.0.0.1:8081` | 既有独立编译功能                       |
| judge    | `http://127.0.0.1:8082` | 严格预检后的独立编译及双角色执行       |

服务默认只绑定本机。不为远程 CI 自动启动或部署此环境，不把本机手动切换模式当作公网账号隔离。旧基础数据库/Redis/MinIO的完整验收仍依 [主 README](../../README.md) 单独执行，API `/healthz` 也不等于数据库 `/readyz` 已通过。

服务启动成功后打开 [通信 A+B](http://127.0.0.1:5173/problems/communication-a-plus-b)。结束时在拥有 `mvp` 的终端使用 Ctrl+C，等待监督器确认判题资源清理；它只负责自己启动的服务。外部预先启动的 judge 和 compiler 需由各自拥有者停止。遇到未知残留先记录与检查，不删除不明容器、任务目录、上传题或数据卷。

## 学生操作

在顶部选择学生模式，打开通信 A+B。题面说明：发送端读取 a、b 并发送原始两个数；接收端读取消息后计算和。输入范围 ±10¹²、每例通信 32 字节、共享 5000 ms、每角色 64 MiB。

将 [sender.cpp](../../examples/communication-a-plus-b/sender.cpp) 与 [receiver.cpp](../../examples/communication-a-plus-b/receiver.cpp) 分别粘贴到对应编辑器，点击“运行测试”。正常环境预期为 AC、4/4；更改任一程序后需要重新运行。学生可看到整体判定、通过数量和本次代码的角色编译诊断，不显示隐藏测试数据、通信日志或工程摘要。

## 教师操作与题面

切换教师模式，进入[上传页](http://127.0.0.1:5173/upload)，按 [教师包 README](../../examples/communication-a-plus-b/README.md) 填写 [完整题面](../../examples/communication-a-plus-b/PROBLEM.md)、双模板、限制和四对输入输出。现有 [descriptor.json](../../examples/communication-a-plus-b/descriptor.json) 也可通过原 multipart 接口上传；不要发明 ZIP 或另一套 manifest 上传 API。

教师可查看逐例输入、标准答案、接收端输出、方向日志和通信量。教师日志有长度上限，截断会标注；教师投影不提供角色退出/OOM/控制终止和源码产物摘要。上传生成新的 `upload-…` ID，默认持久目录为 `.local/teacher-problems`，已有 `PROBLEM_UPLOAD_ROOT` 配置继续有效。

本包四组数据是公开示例。题面要求的“发送端只传原始数、接收端加法”由模板示范，通用判题器不会静态证明源码计算位置；需要评分这项分工时由教师审代码。

## 工程师执行真实验收

先按正常路径启动并保持服务运行，再打开第二个终端，从同一仓库根目录串行执行。期间不要在浏览器或别的终端发起判题，以免混淆全局 owner 容器与任务目录检查。

```powershell
npm run judge:communication:test
```

等价命令为 `node scripts/communication-test.mjs`。脚本连接既有服务，不启动或部署服务，不上传或改写题目，不操作 Git，不运行宿主 C++，不删除资源。没有 fake 模式，不能在缺少 Docker 或 API 时给出真实 PASS。

配置沿用已有 `JUDGE_TEST_API`，并提供显式的核验路径：

| 环境变量               | 默认值                      | 约束                                        |
| ---------------------- | --------------------------- | ------------------------------------------- |
| `JUDGE_TEST_API`       | `http://127.0.0.1:8080`     | 仅 http 本机 origin，不允许凭据/查询/子路径 |
| `JUDGE_TEST_WORKER`    | `http://127.0.0.1:8082`     | 同一已启动 worker 的健康接口 origin，限本机 |
| `JUDGE_TEST_JOBS_ROOT` | `.local/judge-jobs`         | 实际 worker 任务目录；只读检查，必须存在    |
| `JUDGE_TEST_PROFILE`   | `.local/judge-profile.json` | 实际 worker 的锁定镜像 profile；只读检查    |

相对目录值按当前工作目录解析，请从仓库根目录运行。正常 `mvp` 当前使用固定 `.local` 路径；环境变量只改变验收观察位置，不能重新配置 worker，也不能指向另一个空目录来假称清理成功。Docker CLI 必须连接服务使用的同一 Linux 引擎；不要让另一个终端切换 Docker context。

脚本先核实真实 Docker Linux/amd64、cgroup v2、profile 对应镜像 ID、无 owner 容器和空 jobs 目录，以及 worker health。Docker 查询失败是验收失败，不能解释为零容器。随后校验通信 metadata、两个公开模板、兼容模板字段和示例数据与内置题一致。

十二项矩阵包含：

| 场景                        | 预期                                           |
| --------------------------- | ---------------------------------------------- |
| A+B 双模板、四组数据        | AC，4/4                                        |
| source_units 请求数组倒序   | AC，仍显式按角色绑定                           |
| 接收端输出和加一            | WA，0/4                                        |
| 发送端语法错误              | CE，仅实际尝试 sender，零运行例                |
| 接收端语法错误              | CE，保留 sender 成功与 receiver 失败，零运行例 |
| 发送 33 字节到 32 字节通道  | WA，limit_exceeded，实际转发不超 32            |
| sender 自然退出 17          | RE，有对应角色自然退出证据                     |
| receiver 自然退出 137       | RE，不伪报 OOM 或监督器终止                    |
| sender 无限循环             | TLE，并验证受控终止                            |
| receiver 无限循环           | TLE，并验证受控终止                            |
| receiver 持续分配并写入内存 | MLE，真实角色 OOM 标记                         |
| receiver 持续输出           | OLE，输出截断证据                              |

每次提交保存实际响应；要求源码规范 SHA256、逐角色源码与成功产物摘要、产物聚合摘要、当前本地 manifest/用例原始字节摘要、镜像 ID、case 顺序与输入答案都匹配。规范源码 JSON 与 Go 一样转义 `<`、`>`、`&`、U+2028/U+2029；不会仅用“hash 看起来有 64 个字符”代替绑定。每次运行前后检查 owner 容器与 jobs，之后再次确认 worker 健康与 profile/数据未变。

证据保存在 `.local/evidence/communication-matrix-<时间>.json`。初始化即写 `completed:false`，每项默认 `passed:false`；只有该项响应、角色证据、绑定和清理都通过后才标记通过。所有十二项及最终清理完成才写 `completed:true`。失败保留当次响应、错误和可查询到的残留状态，退出码为 1；中断或启动前失败的未完成报告也不能作为成功。证据目录不可写时无法可靠落盘，需修复正常目录权限后重新验收，不能补造记录。

该矩阵只覆盖内置 A+B 的以上判题路径，**不等于整套通信产品最终验收**。当前仍需另行真实检查双 attach/取消与 busy、挂载隔离/无答案可读、提前退出/EOF、教师上传与重启持久化、三模式页面及数据投影、旧普通/交互/特判回归和监督器停止阶段。完成记录与独立审阅前，不关闭这些待办，不把历史 SPJ 结果当作通信证据。

## 当前可执行的静态检查

```powershell
node --check scripts/communication-test.mjs
node node_modules/prettier/bin/prettier.cjs --check scripts/communication-test.mjs examples/communication-a-plus-b/README.md examples/communication-a-plus-b/PROBLEM.md examples/communication-a-plus-b/descriptor.json docs/judge/COMMUNICATION_ENVIRONMENT.md
```

静态检查无需 Docker，不导入或执行验收脚本，也不编译 C++。数据检查可确认四对输入输出的数学结果、32 字节预算、双模板与 descriptor 一致，以及示例文件与内置 fixture 一致。静态结果由本轮交接记录；真实执行继续标为 pending，直到上述真实报告与服务证据存在。
