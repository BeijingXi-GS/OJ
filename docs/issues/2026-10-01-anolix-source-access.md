# Anolix 核心源码获取阻塞

- ID：B-JUDGE-001；发现日期：2026-10-01（Asia/Shanghai）；处理人：用户。
- 当前状态：未解决；已停止新增实现 /验证，只整理本次已发起操作及交接。
- 影响任务：JUDGE-001参考实现核对 /安全设计审阅，以及后续判题worker、前端判题结果和真实AC/WA/TLE/MLE/CE/RE矩阵。
- 当前功能：旧题面 /编译MVP保留，本轮没有启用用户程序执行，也没有修改应用代码或Docker配置。

## 前因

用户要求在已有产品增加读取文件测试、运行和判定，参考 https://github.com/Sh1roganeMiyuki/Anolix 。随后明确选择服务器读取每题目录中的 `.in/.out` 文件，前端展示逐例判定。此次授权将旧“仅编译”范围扩展为实际运行，但不能把未核实的沙箱行为登记为已实现或通过。

准备接入前，需要核对参考项目的OCI配置 /seccomp /runc生命周期 /取消 /清理实现及构建要求，并对真实Linux限制、OOM证据与安全边界完成独立审阅。README不是核心代码审阅证据。

## 已执行顺序与结果

1. 本机阅读当前AGENTS、PROGRESS、开发流程、compile契约、Phase1设计、基础不变量以及已有API /worker代码，保留所有旧未提交内容。
2. 网页工具成功获取Anolix仓库首页README及 `go.mod` 页面。README描述Go +runc的OCI沙箱、namespace /seccomp /NoNewPrivileges /能力集清空 /超时清理；go.mod声明 `go 1.27.1` 与runtime-spec依赖。本机 `go version` 为 `go1.25.4 windows/amd64`。若直接采用原工程，构建版本需要另行解决；本轮未擅自改其go.mod或升级工具链。
3. 网页工具尝试读取main的launcher/spec.go、launcher/launcher.go、config/policy.go及相关目录，返回 `Failed to fetch ...: Cache miss`。这些响应没有实际源码，不能据此审阅实现。
4. 通过已获执行许可的shell运行 `git ls-remote https://github.com/Sh1roganeMiyuki/Anolix.git refs/heads/main`，失败：`fatal: unable to access 'https://github.com/Sh1roganeMiyuki/Anolix.git/': Recv failure: Connection was reset`。未得到commit SHA、没有clone成功或下载源码。
5. 同一轮此前发起的Docker限制预检完成：Server29.8.1 /Linux amd64，cgroup v2 /cgroupfs，builtin seccomp /cgroup namespace；临时容器使用既有compiler镜像、UID10002、network none、cap-drop ALL、no-new-privileges、read-only、64MiB内存、同值memory-swap、16PIDs、1CPU与受限tmpfs。
6. 临时容器实际输出：memory.max=67108864；memory.swap.max=0；pids.max=16；cpu.max=`100000 100000`；CapEff全0；NoNewPrivs=1；Seccomp=2。容器自动移除，现有PG /Redis /MinIO /compiler及卷未改变。这是限制预检，不是AC/MLE判题或沙箱逃逸验证。
7. 确认源码访问问题后报告用户，未再次clone、切换下载来源、改代理 /防火墙 /系统配置，未自行采用另一运行实现绕过参考核对。建立本问题、范围 /任务和停止交接，不进入代码实施或启动新子Agent。

## 已知事实与原因推测

已知事实：仓库首页和go.mod的网页读取成功；核心代码的网页读取未得到内容；本机git的GitHub HTTPS访问被重置；源码版本未锁定；Docker连通且限制预检通过。

未确认原因：本机到GitHub的链路、代理或连接中间层重置，或远端暂时中断；网页工具的cache miss是独立观察，不能证明与git失败完全同源。没有足够证据归因于凭据、仓库私有、用户代码、Docker或本机Go版本。

本机Go版本与原工程声明不相同是兼容性事实，不是本次网络错误的原因。用户说“参考”不等于必须直接引入其依赖；最终采用何种适配仍须在取得源码 /审阅后明确，当前未做裁定。

## 后果与停止位置

判题扩展尚未实施，AC/WA/TLE/MLE/CE/RE均没有新增真实运行证据。读取用例文件、checker、独立judge-worker及前端逐例结果仍待实现。API仍不执行用户程序，原编译MVP的judge_enabled继续false。原浏览器会话B-MVP-001仍未关闭，不冒称前端验收通过。

当前分支为task/JUDGE-001-file-cases-sandbox-2026-10-01；仅新增或更新任务 /范围 /问题 /交接MD，保留旧改动，未提交 /推送。新停止规则继续来自用户，不是自动审批拒绝；本轮shell权限审批没有拒绝。

## 用户处理与具体恢复条件

请恢复本机访问该GitHub仓库的能力，或把包含完整源文件的Anolix源码副本放到工作区并说明对应commit /版本，再通知继续。建议本地副本路径：`D:\新建文件夹 (4)\2026-09-26\gen\inputs\Anolix`；此路径本轮未创建，也不是已取得的源代码。

恢复后先确认可读取launcher、config、CLI、测试及go.mod，记录版本 /文件摘要；再审阅并确定原工程适配或独立实现方案及Go版本策略。之后依次实现文件用例、受限编译 /运行、结果映射、API /前端，执行真实AC/WA/TLE/MLE/CE/RE/OLE和清理 /隔离矩阵及独立审阅。未经复验不关闭B-JUDGE-001；不得因普通compile smoke通过而放行运行。

## 2026-10-02 用户重新授权后的恢复

用户要求“继续完成，并且针对这种情况，你可以考虑先拉取源码，而不是等着链接重置”。本轮首先直接git clone，不再等待网页源码缓存。命令使用HTTP/1.1与lowSpeedTime=20 /lowSpeedLimit=1000，depth1；clone成功，源码在 `D:\新建文件夹 (4)\2026-09-26\gen\inputs\Anolix`。

commit为 `4502d2a798f3130d3c540c5c31d45251ca103623`，参考工作区干净；已实际读取launcher/spec.go、launcher/launcher.go、config/config.go和测试 /资源实验文档。核心文件SHA256：spec fcd3181be7f47afe9a6599d90711eca07cf4acc6d29ac938b3478465634aca54；launcher 2bbd29ed2227380d5cb2f57515aa07d54e61361d24eeec10159621008d5c7a6b；config dd7122b86a8d43d0b89a55c24efdb600abeefb90389f1be45d270f6b252e530e；go.mod 62884a65d01a0a2b67fd6e3f03133b95ef37cb258364d8c0f80e87a38f97cb0f。

源码获取阻塞B-JUDGE-001经实际clone /读取复验关闭，不声称网络根因已定位或以后不再失败。原Go1.27.1声明保持未改；新方案参考隔离模式，独立Docker OCI适配而不直接依赖原库，见 [新契约](../judge/JUDGE_MVP_CONTRACT.md)。独立设计review与实现 /真实判题尚待完成；此关闭只针对源码取得。

Docker Desktop此次开始未运行，按原已知环境修复授权隐藏启动并有界检查，已恢复Server29.8.1 Linux amd64。原服务与卷保留；旧浏览器阻塞不在本关闭范围内。
