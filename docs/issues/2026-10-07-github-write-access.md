# 2026-10-07：GitHub 写入授权与网络阻塞

## 2026-10-09 分支创建与原因定位

用户进一步要求分析解决原因、创建分支并合并，明确覆盖旧“不合并”的范围。root认领PUB-001，申请网络与本仓库.git写权限；工具只返回本轮network grant，file_system为null，未写本地Git元数据。

本轮实际检查：

1. Git origin仍为用户指定仓库，本地HEAD/远程main均为 `7a952c68414c1aeccef9af2475ec3e43124e292b`。远程main首次查询发生连接器后台传输错误，后续只读正常取得该SHA。
2. network grant后，`git ls-remote --heads origin` 约21秒后连接 github.com:443 失败，exit1；`curl.exe -I --connect-timeout 8 --max-time 15 https://github.com` 超时、收到0字节（curl错误28）。当前失败发生于连接阶段，未证明Git认证有效或无效。
3. 正常读取的配置显示 Git 已用 `HTTP/1.1` 和 `schannel`；该进程环境未发现HTTP_PROXY/HTTPS_PROXY/ALL_PROXY/NO_PROXY，Git未返回匹配该仓库的http.proxy。不能把“改为HTTP/1.1”再次当作有效修复，也不能由这些结果判断Windows系统代理不存在。
4. 读取Windows代理设置及DNS的命令被自动审批流中断，明确未执行。后续读取插件管理SKILL的命令也发生同类未执行错误；未换工具读取同一被阻动作或降低审批策略。已确认GitHub插件安装且启用，但启用不代表有仓库写权限。
5. 创建 `task/PUB-001-mvp-delivery-2026-10-09` 分支的第一次连接器请求因自动审批响应流断开而未执行。用户选择“恢复GitHub连接器写权限和审批连接后继续”后，正常重试相同分支创建；本次已到达GitHub并明确返回403 `Resource not accessible by integration`，接口为创建Git引用。没有创建新分支。
6. 停止继续写接口、提交、PR和合并；没有降低安全策略、改全局代理/证书、修改仓库权限或强推。当前无新远程commit/ref/PR，本地源代码保持原样，只有授权/诊断/交接MD更新。

### 原因与对应对策

**已确认的写入阻塞是集成权限不足。** GitHub官方说明该错误代表GitHub App/细粒度令牌缺少所需权限；创建引用要求相应Contents写权限。用户账号仓库admin/push=true不表示该连接器令牌具备同样权限。10-08的tree写入和10-09的branch创建均被拒，缩小请求或改接口不能解决授权问题。来源：[REST错误说明](https://docs.github.com/en/rest/using-the-rest-api/troubleshooting-the-rest-api#resource-not-accessible)、[Git引用接口](https://docs.github.com/en/rest/git/refs#create-a-reference)。

通过该连接的正常账户授权流程，核对其GitHub App/账号是否覆盖 `BeijingXi-GS/OJ`，是否提供Contents写入与创建PR需要的权限；本次提交含workflow文件，也需符合相应权限要求。若应用本身只申请只读权限，单纯重连或切换仓库不能给它增加未声明的写权限，需要使用本身支持写入的正常授权连接。插件“是否询问/自动审批”设置与GitHub OAuth/App权限不同，不改成全放行来处理403。当前工具没有用于修改GitHub App/OAuth授权范围的接口，不能代用户在后台增加权限。

**Git直连问题仍未定位到具体网络组件。** 已知443连接失败/超时，可能涉及代理、防火墙、网络链路，未取得系统代理/DNS证据；不宣称已确认是DNS污染、HTTP/2或GitHub宕机。官方建议按网络/代理/防火墙分层排查：[GitHub连接排障](https://docs.github.com/en/get-started/using-github/troubleshooting-connectivity-problems)。本轮用户选择连接器路径，因此不擅自切换到未知代理或另一身份上传。

**自动审批通信故障是第三层独立问题。** 部分诊断与第一次建分支请求未执行，属于审批响应传输失败，不是安全否决；最后一次建分支已获得GitHub实际403，也不能据此宣称审批永久修复。需要恢复正常审批连接，保留现有策略；相关诊断仍待正常运行。

### 恢复后精确接续

用户恢复连接器仓库写授权后，先核对远程main和目标分支是否存在，再从最新适用基线创建该分支；确认当前250份非MD实现与已审快照、最新MD和秘密排除，提交并推送；创建PR、核查差异/实际检查，按用户明确授权合并并读回main的merge SHA。若远程期间有新提交，正常合并解决，不强推覆盖。真实Docker、安全整改和产品验收仍pending，合并源码不能登记为生产安全通过。

当前交接见 [10-09接续记录](../handoffs/2026-10-09-github-blocked.md)。已有10-08源码包保留；本轮新增诊断/授权MD尚未上传，也不在旧包中。

## 2026-10-08 用户要求推送后的复验

用户明确再次要求推到同一仓库，root 恢复正常连接检查。本次普通 `git ls-remote --heads origin` exit 1：`getaddrinfo() thread failed to start`；随后按工具规则申请并获准沙箱外只读查询，执行后 exit 1：`Recv failure: Connection was reset`。本次没有自动审批拒绝或审批流中断，不能与旧审批故障混为一谈。

GitHub 连接器读取仓库和 main 成功：main 仍为 `7a952c68414c1aeccef9af2475ec3e43124e292b`，base tree `b989e2748f2e8e007ee6608f8bb3733159341755`。在该 tree 上写入本轮实际新增的单个安全问题文档，`create_tree` 明确返回 403 `Resource not accessible by integration`；即使小请求也失败，不能把原因只归结为批量大小。未返回新 tree SHA，未创建 commit、远程分支或 PR；root 停止写入，不切换其他写接口绕过应用权限。

已补齐开发工作包、安全方案和独立静态审阅；这些是文档交付，不是整改代码或真实安全PASS。现有非MD实现未改变。另存10-08源码包与校验清单，保留旧包。推送的接续条件仍是恢复正常 GitHub 写权限或已授权的正常 Git网络/身份路径；不需要再次授权产品范围，但本次上传尚未完成。

## 目标与准备

用户已明确授权上传 `https://github.com/BeijingXi-GS/OJ`。本轮代码修复、完整本地检查、250 份源码摘要核对、独立增量审阅均通过。上传前 Git 非忽略文件内容经过 UTF-8/常见秘密模式和本机文件排除检查；421 份源码、文档与示例准备为开发分支及 draft PR，不合并 main，不生产部署。

上述 421 是首次打包前文件数，包含后来本轮更新之前的文档；最终源码包计数以实际包清单为准。非 MD 源码仍为同一 250 份。

## 实际操作与结果

1. GitHub 连接器正常读取仓库，用户仓库权限信息含 push=true；两次核对 main 均为 `7a952c68414c1aeccef9af2475ec3e43124e292b`，等于本地 HEAD。**用户仓库权限不等于连接器应用令牌具备相同写入范围。**
2. 以该 main 的 tree 为基线发起完整 tree 对象创建，传输层报 `error sending request`，目标是连接器后台 MCP 服务，没有返回 tree SHA。不能证明该对象已成功创建；该操作尚未创建 commit、分支引用或 PR。
3. 后续只读仓库查询正常。将内容拆成较小批次，首批 tree 创建明确返回 HTTP 403：`Resource not accessible by integration`。这是 GitHub 应用集成访问拒绝，不是 Codex 自动审批故障。
4. 不再继续连接器写入；以普通 Git 的正常身份读取 `git ls-remote --heads origin`，实际 exit 1，`Recv failure: Connection was reset`。本轮此前已取得 network grant，但该结果说明此条 Git HTTPS 连接仍未成功；不能断言 Git 身份有效或远程允许 push。
5. 未创建 commit、远程分支或 PR，未写本地 `.git`，未强推、合并或删除远程内容。传输失败可能留下不可见未引用 tree 对象，但没有可交付的远程引用；不登记“已上传”。

## 原因与对策

403 明确指向当前集成无法访问写接口。可能是连接器 GitHub App/OAuth 授权缺少内容写权限、仓库选择或策略限制，**尚未读到集成授权详情，不能确定具体缺少哪一项**。用户应在 GitHub/连接器的正常授权界面核对该仓库及写入权限；本次还包括 CI workflow 文件，如果授权界面对 workflow 修改有独立限制，也需正常核对。不要粘贴 token 到聊天或代码里，不改公共仓库可见性、不关闭保护。

Git connection reset 属于另一条网络路径，可由代理/网络设备、TLS 连接或远端暂时中断等引起，当前没有足够证据确定。用户可在正常终端确认 `git ls-remote` 是否稳定成功；不要以关掉证书校验、换未知镜像、公开凭据或修改全局安全设置处理。

恢复一条正常授权的写入路径后即可接续，不需要同时解决所有通道：优先正常 GitHub 连接器；若使用 Git CLI，则还需要正常 Git 网络/身份和本地 Git 元数据写入权限。不能将本次 403 当作上传成功，也不能反复尝试其他写接口规避同一应用权限限制。

## 接续条件

1. 用户恢复连接器对该仓库的正常写权限并通知继续，或恢复正常 Git HTTPS/身份并明确可用的路径。
2. root 重新核对远程 main 和待交付内容，避免覆盖用户期间新增的提交；重新核验源码摘要，保留本地原改动。
3. 创建独立开发分支、提交和 draft PR，核实返回的 commit/ref 后更新 [交接](../handoffs/2026-10-07-recovery-delivery.md)，将实际 PR 附到会话。
4. Docker 真实验收依然单独 pending，不能因为源码上传成功而宣称产品验收通过。

本轮另外准备本地完整源码 ZIP，包含当前源码、示例和文档，不含 `.env`、`.local`、教师上传题库、node_modules 或构建产物。源码包不是 Git 历史备份，也不替代 GitHub 交付；最终路径和校验见 PROGRESS / 交接。

打包时另发现本机 PowerShell 不支持旧 `Set-Content -Encoding Byte` 参数：首次 ZIP 已写入，但清单写入失败。因此首次无 `-source` 后缀的 ZIP 不作为正式交付。脚本改用明确 UTF-8 的 .NET 文本写入，并输出新的 `-source.zip` / `-source.manifest.json`，逐项解压读回比对所有内容，再对 250 份已审源码逐项核对；只有全部通过且成功写出 verified 清单才交付。未删除或覆盖旧包。
