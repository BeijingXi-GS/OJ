# 最终启动的Docker cgroup预检阻塞

- ID：B-JUDGE-002；日期：2026-10-02；原处理人：用户；用户授权后由root修复；状态：resolved，完整预检/实际API/浏览器已复验。
- 历史影响：最终版本judge-mvp不能ready，supervisor拒绝启动API与Web；本轮恢复证据见末节，当前不再阻塞。
- 当前日志的 `cgroup_v2_required` 是预检分支名称，**不等于已经确认Docker实际使用cgroup v1**。

## 背景与此前通过结果

用户要求继续完成文件判题，并建议先拉取源码。Anolix已本地clone并锁commit，Docker Linux按既有授权恢复；独立设计通过后实现后端 /前端。11:26首轮严格预检ready；开发期独立审阅R1–R5修正后，11:41:13 judge再次ready、11:41:15 API ready。

该运行版本真实矩阵17项全部passed /completed：AC、WA、TLE、真实OOM MLE、语法 /链接CE、崩溃RE及非OOM退出137、OLE、展示截断AC、UTF8、后代清理、busy /取消、缺失文件 /错误manifest SYSTEM_ERROR、动态metadata /exact与tokens。每项查询Docker与job目录确认清理；故障注入文件在finally恢复。

随后完整check通过（格式 /vet /ESLint /Go tests /TS /Vite build），smoke及mvp:smoke通过，受限Linux可信单测包括symlink通过，实际继承 `/data` 声明卷的镜像被新服务拒绝启动。独立代码 /本机执行边界审阅通过，产品浏览器验收尚未完成。

## 本次操作顺序和实际结果

1. 最后增加的client协议收束是禁止SYSTEM_ERROR case后仍附加case，单测与完整check通过；需要最终重新编译启动API /judge再留可访问实例。
2. 只匹配当前仓库 `.local/api-mvp-dev.exe` 的完整路径结束旧API；其supervisor按随机控制token请求拥有的judge正常shutdown。只读核对judge进程退出、`oj.judge.owner=file-mvp-v1`容器查询为空、`.local/judge-jobs`为空。受控关闭实测PASS；没有强杀活跃judge或删除基础卷。
3. 执行 `node scripts/source-digest.mjs`，最终非MD源码100文件，SHA256 `3ee10a98dbff54fa0e567a6a18161a3b38aa67bab8e6190f9643fddb42bb7357`。保留所有历史未提交改动，无提交 /推送或依赖升级。
4. 执行 `node scripts/mvp-dev.mjs`，重编译本机judge并开始严格preflight。11:54:40.203日志：

   ```text
   ERROR judge preflight failed reason=cgroup_v2_required
   Error: Judge preflight did not become ready; inspect logs, no execution fallback.
   Node.js v22.19.0
   ```

   judge退出1，supervisor结束；未进入API /Vite启动步骤。没有修改资源限制或回退执行。

5. 已发起的最终AC smoke读取8080题目时返回 `ECONNREFUSED 127.0.0.1:8080`，未取得结果，不能计为最后版本AC通过。此前矩阵不是本次最终启动成功的证据。
6. 收取上述已发起操作结果后停止。没有再次执行docker info /重启引擎 /放宽5秒控制预算 /修改cgroup /切换运行路径。浏览器此前仅重新读取接口说明与inventory：IAB id2、tabs空；未重新创建标签或完成UI操作，B-MVP-001保持未关闭。

## 已知事实、原因推测与缺失证据

已知：同一引擎此前能返回cgroup2并完成真实限制执行；最终新服务在version检查之后进入info检查并拒绝启动；该分支将“docker info控制命令失败”和“成功输出非2”统一为cgroup_v2_required。控制调用有5秒context和有界输出，错误未向网页泄露主机细节。

未确认：这次docker info是否达到5秒预算、CLI /daemon查询是否瞬时失败、连接是否受引擎负载影响、还是成功输出了非2。当前日志未保存该次命令原始stdout /stderr或退出状态，不能据此判断实际cgroup版本变化。可能原因仅是这些候选，不把它们当根因。

影响：源码 /镜像 /fixture保持完整；`.local/judge-profile.json`仍锁实际新镜像ID。原PG /Redis /MinIO卷未修改；旧compile-worker未由此次supervisor关闭，其当前健康未在阻塞后复验。5173 /8080 /8082最终新实例没有成功保留，不应向用户宣称页面目前在线。

## 用户接手与具体恢复条件

用户可检查Docker Desktop Linux引擎 /WSL状态，并在仓库终端只读核实：

```powershell
docker version --format '{{.Server.Os}} {{.Server.Arch}}'
docker info --format '{{.CgroupVersion}}'
docker ps -a --filter label=oj.judge.owner=file-mvp-v1
```

期望linux amd64、cgroup版本2、无本项目判题残留。不要删除不明容器或基础数据卷。若info稳定返回2且没有残留，通知继续并授权复验；恢复后需要重新启动严格preflight、最终API AC smoke、真实浏览器题面 /编译 /判题 /失败 /过期 /重置 /响应式验证，并按受影响范围完成独立验收。用户若明确授权修复此新阻塞，Agent可进一步核实CLI失败与daemon响应，并改善有限诊断；当前不自行重试或调整主机。

## 本轮用户授权恢复（2026-10-02）

用户已明确要求“尝试解决并继续推进任务进度，增加一些题目”，当前允许root按JUDGE-005恢复并复验，历史“不自行重试”的停止只适用于上轮。

本轮初始docker version /info均exit1，脱敏事实为DockerDesktopLinuxEngine管道不存在（system cannot find file），查询分别约197ms /1007ms /452ms。启动已有Docker Desktop隐藏进程后，version为linux amd64、info返回2；随后按原5s预算连续三次info成功，耗时651 /423 /444ms，stderr0。现时引擎未运行的故障已恢复，但不能据此追认上轮11:54同一日志的确切原因。

本轮改进预检分类：仅成功info返回非2才cgroup_v2_required；CLI退出 /查询超时 /取消 /启动失败 /输出超限分别给固定操作与分类，仍拒绝ready。保留全部原资源和时间预算，不打印原始daemon输出 /主机路径 /凭据；等待完整preflight及最终API通过后才关闭本问题。

## 恢复条件已复验

18:50:06最终新judge严格ready，18:50:10 API judge_enabled=true并启动Web。当前源码实现校验全部五题文件，再完成真实编译/只读产物挂载/执行及内核资源probe。五题10次正确AC/错误WA、原17项真实判定和故障矩阵、Linux单测与实际页面通过；19:09:34最终5173/8080/8081/8082均HTTP200，判题容器与job目录均无残留。原数据卷和.env未重置。

本轮确认并恢复的是“现时引擎未运行”；上轮11:54确切查询失败原因仍未确认，不能追认为当时必然管道缺失。诊断实现修正另见 [控制输出](2026-10-02-docker-control-buffer.md)。最终证据与限制见 [VALIDATION](../judge/VALIDATION.md) 和 [浏览器QA](../judge/BROWSER_QA_2026-10-02.md)。
