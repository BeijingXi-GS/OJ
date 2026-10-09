# ISSUE-2026-09-27-001：Docker Linux 引擎无法连接

- 首次记录：2026-09-27（Asia/Shanghai）
- 状态（当日）：阻塞，等待用户处理；**2026-09-30 已复验关闭 B-ENV-001**，见 [最新交接](../handoffs/2026-09-30-port-docker-validation.md)。
- 关联：P0-005 / B-ENV-001（重新打开）
- 处理负责人：用户
- 记录人：本轮 Agent
- 代码基线：`7a952c6`（Merge remote main）
- 当前分支：`task/P1-001-architect-handoff`
- 恢复条件：用户处理后明确通知继续，再由 Agent 进行只读验证和已授权的本项目服务检查。
- 本文不包含已确认的根因；当前证据只能证明本轮引擎不可连接。

## 一、背景与原计划

后续记录：用户于 2026-09-29 明确授权复查连接/分析原因及独立编写前端 MD。新的观察见 [9 月 29 日分析](2026-09-29-docker-engine-analysis.md)。本文件保留 9 月 27 日的历史状态，不作为本轮“全部工作停止”的覆盖指令。

用户说明已尝试修复环境，要求继续启动 Docker，并完成两项工作：

1. 完成下一阶段强 Agent / Architect 应承担的任务。
2. 详实说明其余 Agent 的工作，尤其是前端框架、目录、接口、职责和验收。

仓库已有 2026-09-26 的环境交接：当时 Docker engine 28.4.0 曾就绪，PostgreSQL / Redis 健康、API readiness 通过；MinIO 拉取另有历史阻塞。历史结果不能代表当前服务仍在运行。

本轮原计划先核对环境与仓库变更，再处理架构评审和 Phase 1 Mini Judge 设计、Builder/Worker 任务拆分。随后用户补充：问题须单独写 MD，由用户处理；阻碍后续进展时停止工作。

## 二、实际操作与结果

| 顺序 | 操作                                                  | 观察结果                                                                                                                                       |
| ---- | ----------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| 1    | 阅读 AGENTS、PROGRESS、任务卡及前次环境交接，检查 Git | 工作区原本干净；仓库已有 commit 和 GitHub remote，旧台账“未提交、无 remote”已过时                                                              |
| 2    | 在受限环境查询 Docker / WSL                           | Docker CLI 和 Compose 存在；读取 Docker 配置、日志及 WSL 列表出现权限错误；无法据此单独判定主机根因                                            |
| 3    | 在获得执行权限后重新查询 Desktop、engine、容器与 WSL  | Desktop 状态无法获取；engine 的 dockerDesktopLinuxEngine 管道不存在；不能列举容器与镜像；OS-Workbench 和 docker-desktop 均显示 Stopped / WSL 2 |
| 4    | 新建本地任务分支                                      | 成功创建 task/P1-001-architect-handoff，基于 7a952c6；未提交、未推送                                                                           |
| 5    | 在用户追加异常处理规则之前执行 docker desktop start   | 命令持续等待，没有成功输出；不能视为引擎已经启动                                                                                               |
| 6    | 收到补充规则后，只核实既有启动操作及 engine 状态      | engine 仍报告命名管道不存在；进程查询没有列出 Docker Desktop / com.docker.backend                                                              |
| 7    | 结束本轮等待中的启动 CLI                              | 启动命令退出，未收到成功结果；未执行 Docker 重置、WSL 关闭/重装或任何数据清理                                                                  |
| 8    | 整理问题、更新协作规范和进度                          | 只进行必要的记录与交接，不继续架构设计、功能开发或修复尝试                                                                                     |

日期只有实际观察顺序，未凭空补写分钟级时间。

## 三、关键错误原文

受限环境中的附加错误：

```text
Error loading config file: open C:\Users\25395\.docker\config.json: Access is denied.
unable to create a log file for docker-desktop.exe ... Access is denied.
Wsl/EnumerateDistributions/Service/E_ACCESSDENIED
```

在获得执行权限后，Docker 本身仍返回：

```text
Could not retrieve status. Is Docker Desktop running?
You can start Docker Desktop by running 'docker desktop start'.

error during connect:
Get "http://%2F%2F.%2Fpipe%2FdockerDesktopLinuxEngine/v1.51/info":
open //./pipe/dockerDesktopLinuxEngine:
The system cannot find the file specified.
```

WSL 的有效查询结果：OS-Workbench 与 docker-desktop 均为 Stopped，版本为 2。停止状态本身不足以解释 Docker 为何没能启动。

## 四、已确认与尚未确认

已确认：

- Docker CLI 28.4.0、Compose v2.39.4-desktop.1 已安装。
- 本轮检查时 Docker Linux engine 不可连接。
- 获得执行权限后仍出现缺失 engine 管道，问题不能仅归因于 Agent 受限环境读取配置失败。
- 发起过启动操作，但没有观测到成功。
- 本轮没有启动或重新验证 PostgreSQL、Redis、MinIO。

尚未确认：

- Docker Desktop 未持续运行的原因。
- WSL 服务、虚拟化、Docker backend 启动日志是否存在具体故障。
- 用户之前的修复具体改了什么。
- 已有容器与卷的当前运行/数据状态；没有证据表明它们被删除或损坏。
- 历史 MinIO 镜像问题本轮是否仍然存在；本轮没有重试拉取，不将旧文档中的外部原因视为新验证结论。

不得把“命名管道不存在”直接等同于“必须重装 Docker”，也不得据此删除容器或数据卷。

## 五、对任务的影响与停止位置

直接阻塞 P0-005 的基础服务启动、真实联通检查和 readiness 验证。Phase 0 的完整环境验收未完成。

Phase 1 设计和前端拆分包含可脱离 Docker 的文档工作，但用户当前明确要求在阻碍后续进展时停止。因此本轮在环境阻塞处停止整项后续工作，没有以设计文档可独立编写为由继续推进。

本轮尚未交付：

- P0-004 独立架构评审与契约冻结。
- P1-001 Mini Judge 生命周期设计。
- 本轮要求的 Builder/Worker 详细任务包和前端框架方案。

现有草案保持 proposed；不得登记为 reviewed / frozen / done。本轮没有运行新的完整构建、业务测试或远程 CI。

## 六、需要用户处理的事项

1. 在主机侧处理 Docker Desktop 的启动问题，确认使用 Linux engine。
2. 如主机界面显示错误，可按 Docker 自身诊断获得原因；本轮 Agent 不继续尝试系统级修复。
3. 可用 `docker info` 检查是否返回正常的 Server 信息，而不是仅存在 docker 命令。
4. 处理完成后通知 Agent 继续；若希望绕过环境阻塞，仅先完成设计/前端分工，请明确授权调整本次停止范围。

这些是恢复条件，不是已判定的修复方案。没有要求重装、重置、关闭安全控制或删除数据。

## 七、恢复后的顺序

1. 先阅读本文及 PROGRESS，核实用户是否已解除阻塞或调整范围。
2. 检查 engine，再仅操作 oj-dev 项目的服务；不改动其他项目。
3. 复验 PostgreSQL / Redis。若遇到 MinIO 或其他新问题，另建问题 MD，由用户处理；构成阻塞则再次停止。
4. 恢复 P0-004 的架构核查和 P1-001 设计，并如实区分作者自检与独立审阅。
5. 完成其余 Agent 任务包，重点覆盖前端目录与路由、请求和状态管理、组件、Monaco、mock/真实接口边界、测试和集成顺序。

## 八、本轮改动与保留

- 保留用户已有代码、提交历史和 origin remote。
- 新建任务分支；本轮仅新增问题记录并修改 AGENTS / DEVELOPMENT / PROGRESS 的规范和状态。
- 未修改 .env、Compose、业务代码或依赖版本。
- 未清理任何容器/卷；未执行 WSL reset、unregister、shutdown 等操作。
- 既有验证记录保持历史原貌，本轮状态以 PROGRESS 和本文为准。
