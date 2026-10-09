# 本地 Docker 配置访问与只读检查审批故障

日期2026-10-05；执行root；处理归属用户/当前权限环境。当前只记录事实，不把未验证策略登记为修复成功。

## 背景

用户要求先继续解决之前的问题，并增加通信题，统一记录开发问题与策略。已确认通信题为发送端/接收端两份C++源码。旧停止、重启和页面验收仍依赖实际Docker Linux引擎；原自动审批通信故障未最终解除。

## 操作与结果

1. 默认沙箱中实际执行Docker诊断。CLI警告无法读取 `C:\Users\25395\.docker\config.json`（Access is denied），退回default上下文，docker_engine管道不存在；没有成功引擎查询。无法读取配置可能影响选中的上下文，不能因此断言原Linux引擎本身运行/未运行。
2. 按当前granular权限流程，用request_permissions仅申请Docker配置/contexts读取及连接API所需网络权限，工具返回turn范围已授予。
3. 后续实际Docker诊断仍报同一配置访问被拒绝，context show仍default，docker_engine仍无管道。权限请求成功不等于CLI访问已恢复；未核实Windows ACL、权限向子进程传播或其他限制具体原因。
4. 并行另一条只读命令（检查已安装Docker Desktop文件/进程及项目脚本）在进程创建前返回自动审批stream disconnected/response decode。工具明确动作未执行、不是安全判定。没有把它算作Docker启动失败，没有发起后续旁路或换工具查询。

## 原因范围与候选策略

- **已确认**：配置读取失败、默认上下文、默认引擎管道不存在、授予之后仍失败，以及另一次只读检查未执行。
- **未确认**：真实当前Linux管道是否存在、配置ACL与沙箱传播、Docker Desktop当前状态、自动审批服务具体链路故障。
- **策略A**：用户恢复正常文件访问/审批通道后，首先只读核实权限真正生效、CLI选中正确上下文、Linux/cgroup2与本项目owner残留，再运行既有验收。不能以权限工具返回成功替代引擎证据。
- **策略B（尚未执行）**：如用户允许，由受审任务明确使用项目内独立Docker CLI配置和指定同一已知Linux管道，减少读取用户配置的需求；仍需正常权限渠道，不能借此绕过审核，不能在引擎不可验证时回退宿主运行。
- **开发范围策略**：已向用户询问是否允许先做不依赖Docker的通信题代码/可信测试，将容器集成验收留待环境恢复。没有答复前不按该例外继续实现；若继续沿用原停止规则，则只收束必要文档。

当前问题索引见[README](README.md)，旧审批故障分别见[10-04](2026-10-04-approval-resume-interruption.md)/[本日GitHub核实](2026-10-05-github-approval-interruption.md)。工程真实状态与答复补录到[PROGRESS](../../PROGRESS.md)。

## A+B 增量后的最新实测

用户随后继续要求加入最简单A+B通信题，sender传a/b、receiver求和，本地实现继续。新一次显式Linuxpipe查询等待后实际exit1，错误已变为permission denied while trying to connect，不再是上一查询的file not found；这只是错误状态变化，尚未取得Linux/cgroup2或ready响应。为遵循正常权限渠道，root申请仅该命名管道的双向访问；request_permissions直接返回permission path cannot be represented losslessly，权限未授予。未变换路径编码、改主机ACL/权限或换执行者绕过。

最新恢复条件：用户在正常主机会话确认Docker能连接，并恢复本执行环境对Docker API的授权/访问；当前工具不支持用上述文件路径请求命名管道权限。默认配置访问问题也需分别复验。保留真实矩阵待办，不以错误由missing变denied推断引擎已正常启动。

## A+B 增量前的恢复尝试（历史）

用户再次要求继续开发并尝试Docker，既有“不依赖Docker的本地开发、真实验收待办”答复保持有效。当时项目读写实际恢复，契约最后状态补充已成功落盘，独立设计复审PASS，前后端按登记范围并行实施；不再把更早停止记录当当时停工状态。后续新增A+B期间最新管道错误已在上节单独登记；当前另有[测试文件权限阻塞](2026-10-05-local-test-permissions.md)，不能据此历史推断当前所有读写恢复。

默认Docker仍无法读取config.json和contexts目录；本轮再次通过正常request_permissions申请这些配置/已安装Desktop文件读取和网络，返回turn授予，实际context ls仍Access denied，默认引擎pipe仍不存在。安装路径Test-Path为True，隐藏Start-Process启动命令实际exit0；不等于Desktop持续运行或引擎ready。

按用户本次连接授权实际采用项目内空的.local/docker-cli配置，并显式连接已知dockerDesktopLinuxEngine管道（正常工具权限，无替代审批路径）。该查询绕开对无关用户配置的依赖，但Linux管道同样不存在；由此可以确认本次连接失败不只来自默认上下文选择，仍未定位引擎启动/主机权限具体原因。后续查询未观察到Docker Desktop/com.docker.backend两个名称的进程，不据此推断所有服务状态。

一条拟检查WSL/service及项目代码的只读命令在创建前因自动审批stream disconnected/response decode未执行；随后另一条项目只读命令也发生同类未执行。本轮其他项目读写/可信测试仍可执行；没有把失败命令换工具/交子Agent重做。上轮setup refresh失败与最新故障补录未落盘的情况在此补记，根因不能合并断言。

处理策略：继续获准的通信本地实现、可信测试与独立代码审阅；Docker真实矩阵、挂载隔离/真OOM/取消清理、重启与浏览器验收标pending。用户在主机恢复Docker Desktop/Linux引擎并恢复审批连接后，先成功查询Linux/cgroup2与owner残留，再按既有严格预检启动；不回退宿主执行选手源码，不以版本标记或fake测试宣称真实通过。

## 用户答复与文档写入情况

用户随后明确选择“允许先做本地开发，保留真实验收待办（推荐）”。允许按受审通信契约完成前后端代码及可信测试；Docker故障仍开放，不登记真实验收通过，不在宿主编译/运行选手代码。

两次组合文档补丁（第一次含AGENTS和契约，第二次含契约和进度）在执行前发生自动审批通信/响应解码失败，未执行；具体触发条件不明，不能断定是文件或补丁长度造成。采用更窄、仅普通单份文档的补丁后，契约草案与进度登记实际成功。未把失败动作交其他Agent旁路执行，未修改受保护控制文件；当前实施权限来自用户直接答复。若源码修改再被阻止，记录实际结果，不用换执行者绕过。
