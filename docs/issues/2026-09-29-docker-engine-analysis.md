# ISSUE-2026-09-29-001：Docker Linux 引擎连接复查与原因分析

- 日期：2026-09-29（Asia/Shanghai）
- 关联：B-ENV-001 / P0-005；接续 [2026-09-27 问题记录](2026-09-27-docker-engine.md)
- 状态（当日）：连接失败，具体根因尚未确认；**2026-09-30 已复验 Linux engine 可连接，B-ENV-001 关闭**。
- 执行者：本轮 Agent；处理负责人：用户
- 方法：重试连接与只读诊断；本轮没有再次启动 Desktop、修改服务/WSL/虚拟化、拉取镜像或运行 OJ 容器。
- 本轮权限边界：用户明确另行要求“前端框架写入详细 MD，不亲自搭建”；前端文档已作为独立交付推进，Docker 失败仍阻塞真实环境验证。

后续状态见 [9 月 30 日交接](../handoffs/2026-09-30-port-docker-validation.md)。以下诊断保留为 9 月 29 日历史观察；9 月 30 日 Server 返回 29.8.1 / linux，不能从恢复结果反推此前退出的根因。

## 1. 背景与变化

9 月 27 日 engine 不可连接，用户要求问题单独记录并停止受阻工作。本轮用户重新授权连接检查、失败原因分析，并明确要求完成前端实施说明。

检查发现 Docker 客户端已从之前记录的 28.4.0 更新到 **29.8.0**；Compose 从 v2.39.4 更新到 **v5.5.1**。这些是客户端版本，不能据此宣布 Docker Server 已升级并运行。

仓库基线仍为 7a952c6，当前任务分支 task/P1-001-architect-handoff。保留前轮未提交文档；未修改 .env、Compose、业务源码和 package-lock。

## 2. 检查过程与事实

| 项目                                  | 实际观察                                                                             | 可以得出的结论                                                        |
| ------------------------------------- | ------------------------------------------------------------------------------------ | --------------------------------------------------------------------- |
| 受限执行环境                          | 无法读取用户 .docker/config.json；context 显示 default，连接 docker_engine 失败      | 此结果有环境权限干扰，不能单独用来判定真实 context                    |
| 获得执行权限后的 context              | 当前为 desktop-linux；端点 npipe:////./pipe/dockerDesktopLinuxEngine                 | 用户上下文已指向 Docker Desktop Linux engine                          |
| DOCKER_HOST / DOCKER_CONTEXT          | 两个变量均未设置                                                                     | 未发现本轮环境变量覆盖 context                                        |
| Docker 客户端                         | version=29.8.0，Compose=v5.5.1，路径位于 Program Files/Docker                        | CLI 已安装且可执行                                                    |
| Docker Server                         | info/version 的 Server 部分失败，Linux engine pipe 不存在                            | engine 本轮没有提供可连接端点                                         |
| 进程                                  | 没有发现 Docker Desktop、com.docker.backend、com.docker.proxy；发现 wslservice       | Desktop/backend 当时未在运行；不能只等待现有 backend 变 ready         |
| WSL 发行版                            | OS-Workbench / docker-desktop 均为 Stopped，版本 2                                   | 没有观测到正在运行的 Docker WSL 发行版                                |
| WslService                            | Running / Automatic                                                                  | WSL 管理服务在运行                                                    |
| vmcompute                             | Stopped / Manual                                                                     | 当前没有运行该计算服务；是否阻止启动未验证                            |
| com.docker.service                    | Stopped / Manual                                                                     | 不能仅凭此判定故障；不同 backend 对该服务的要求不同                   |
| WSL 版本                              | 2.6.1.0，kernel 6.6.87.2-1                                                           | 已安装 WSL 2；不是“未安装 WSL”的证据                                  |
| Windows                               | 10.0.26200.9550                                                                      | 记录环境供用户排查                                                    |
| HypervisorPresent                     | true                                                                                 | 系统报告已有 hypervisor                                               |
| CPU WMI 的 virtualization / SLAT 字段 | 两项报告 false                                                                       | 与 HypervisorPresent 信息需结合解释；不能据此直接断言 BIOS 虚拟化关闭 |
| Docker 设置                           | AutoStart=false；本次读取的 WslEngineEnabled / UseWindowsContainers 顶层字段未返回值 | 不自动启动是线索；缺失字段不是 false，不能推断当前 backend 已切换     |

所有检查没有启停系统服务。未验证容器/数据卷状态，因为 engine 连接是前置条件。

## 3. 当前失败原文

获得执行权限后仍出现：

```text
failed to connect to the docker API at
npipe:////./pipe/dockerDesktopLinuxEngine;
check if the path is correct and if the daemon is running:
open //./pipe/dockerDesktopLinuxEngine:
The system cannot find the file specified.
```

受限环境额外出现：

```text
Error loading config file:
open C:\Users\25395\.docker\config.json: Access is denied.
```

两类现象分开记录。授权环境的失败表明：此次连接问题不能全部由 Agent 对配置文件的读取权限来解释。

## 4. 最近日志线索及时间关系

只读取了 Desktop 主进程和 backend 日志尾部的有限内容，未上传日志。

| 来源                               | 日志时间 / 文件更新时间             | 有用线索                                                                         |
| ---------------------------------- | ----------------------------------- | -------------------------------------------------------------------------------- |
| Docker Desktop.exe.log             | 2026-09-27 20:30:41 +08:00          | backend process exited: exit status 0x40010004                                   |
| com.docker.backend.exe.log         | 最后更新 2026-09-27 20:30:37 +08:00 | 记录 software_update / prepare-install / installing-update / app-relaunched 事件 |
| 同一 backend 尾部                  | 2026-09-27 20:30:37 +08:00          | 请求 modules/4.92/modules.json 返回 403，随后更新状态为 no update found          |
| docker-desktop.exe.log（CLI 文件） | 本轮 2026-09-29 有更新              | CLI 日志更新不代表后台 engine 运行                                               |

这些是两天前的最后一组主进程/backend记录，而非本轮启动尝试日志。本轮没有新启动，因此不能声称“本轮启动又崩溃”。退出码只按原样记录，未从该数字推定某个操作系统故障。

日志还包含更早的登录错误，不在本轮时间范围，也无证据关联当前 engine pipe 缺失，未据此提出登录修复。

## 5. 可能原因与优先级

| 优先级         | 假设 / 解释                                         | 支持证据                                                                 | 缺少什么，不能据此断言什么                                                              |
| -------------- | --------------------------------------------------- | ------------------------------------------------------------------------ | --------------------------------------------------------------------------------------- |
| 高             | Desktop 更新/退出后没有重新运行，或未保持运行       | 进程缺失；AutoStart=false；CLI 版本更新；最后主程序日志出现 backend 退出 | 尚不知道用户是否主动退出、重启后未打开，或打开后立即退出；不能直接说安装损坏            |
| 中             | 更新后的 backend 启动/生命周期异常                  | 更新事件附近有 0x40010004 退出记录                                       | 未拿到更新完成后的新启动日志；只是时间关联，不是已证实因果                              |
| 待验证         | WSL 虚拟机启动或计算服务链存在异常                  | docker-desktop 与 vmcompute 当前停止                                     | 未尝试启动 VM，停止状态也可能只是 Desktop 没运行的结果；不能推断缺少虚拟化或要求改 BIOS |
| 低             | 更新模块请求被代理/网络策略拒绝，影响某项组件初始化 | 历史 module 请求返回 403                                                 | 日志只证明更新检查失败，没有证据说明它阻止主 engine；不应先修改代理/关闭安全设置        |
| 已识别的干扰项 | 受限环境读不到配置导致 default/desktop-linux 不一致 | 授权前后 context 不同                                                    | 已在真实 desktop-linux 下复验仍失败，所以单纯切 context 不能解释全部现象                |

**当前最有把握的判断：Docker Desktop/backend 未运行是连接失败的直接条件；其为何未运行还需要用户侧的新启动结果确认。**

Docker context 决定客户端目标，CLI 存在不等于 daemon 存在；相关行为参考 [Docker contexts](https://docs.docker.com/engine/manage-resources/contexts/)。Docker 官方也说明安装后需要启动 Desktop，参考 [Windows 安装与启动](https://docs.docker.com/desktop/setup/install/windows-install/)。

## 6. 用户处理顺序与恢复判据

以下是建议的分诊顺序，本轮没有代为执行修复：

1. 在主机中打开 Docker Desktop，观察是否保持运行、是否出现更新后首次启动提示或具体错误。记录发生时间和界面原文即可。
2. 如果 Desktop 能持续运行，在 PowerShell 执行下面两个只读命令。用 --context 参数指定目标，不必改全局 context：
   ```powershell
   docker --context desktop-linux version
   docker --context desktop-linux info --format '{{.ServerVersion}} {{.OSType}}'
   ```
   期望得到完整 Server 信息及 linux；只显示 Client 版本不算通过。
3. 如果打开即退出，优先查看该次新生成的 Docker Desktop.exe.log / com.docker.backend.exe.log，判断是否重现旧退出码，而不是反复依赖 9 月 27 日旧日志。
4. 如果出现明确 WSL/虚拟化错误，再针对该错误检查 WSL/Windows 计算服务；当前证据不足以建议重置或重装。
5. engine 可连接后再通知 Agent 复验 oj-dev 的数据库等服务；MinIO 是独立的历史拉取问题，此时不混入 engine 根因判断。

Docker 官方提供本地日志与诊断入口，见 [Troubleshoot Docker Desktop](https://docs.docker.com/desktop/troubleshoot-and-support/troubleshoot/)；本轮未使用会上传诊断的选项。WSL backend 资料见 [Docker WSL 2](https://docs.docker.com/desktop/features/wsl/)。

## 7. 任务影响与本轮边界

- P0-005 / B-ENV-001 仍 blocked，不能运行真实 infra-smoke 或确认 PostgreSQL / Redis / MinIO 当前健康。
- 本轮没有再次发起长时间等待的 docker desktop start，没有改 AutoStart 或服务启动类型。
- 没有执行 reset / prune / down -v / unregister / 重装，没有动其他项目容器。
- 按本次用户明确范围，前端详细 MD 独立交付；真实前端实现、业务 API 联调和 Mini Judge 设计没有自动恢复。
- 关闭此问题必须记录成功的 Server linux 连接证据；仅“用户已修改设置”不能代替复验。
