# ISSUE-2026-09-30-001：PostgreSQL 宿主机 5432 端口不可绑定

- 日期：2026-09-30（Asia/Shanghai）
- 关联：P0-005 / B-ENV-003
- 状态：**已修复并复验关闭**；范围仅限 PostgreSQL 端口 / Docker 连接，完整基础服务仍受 MinIO 阻塞。
- 执行者 / 角色：本轮 Agent / Builder（开发环境维护）
- 授权：用户要求“5432 端口被占用了，解决这个问题然后继续试试 Dockers 能不能连接，再继续进行操作”。这项修复是对异常由用户接管规则的明确范围补充。

## 背景与检查顺序

之前 Docker Linux engine 不可连接。此次获得主机执行权限后的 `docker --context desktop-linux info` 返回 `Server=29.8.1 OS=linux`，引擎连接已恢复；受限环境仍有 Docker 配置读取 / pipe 权限干扰，不能据此推断主机引擎失败。

PostgreSQL 容器 healthy，容器内 `SELECT 1` 返回 `1`；Redis healthy 且 `PING` 返回 `PONG`。但 PostgreSQL 的 `HostConfig.PortBindings` 请求 `127.0.0.1:5432`，`NetworkSettings.Ports` 却没有实际绑定，`docker port` 无输出。因此容器健康不能替代宿主机 API 的数据库联通验证。

## 已确认事实与原因边界

1. `Get-NetTCPConnection` 没有发现 `5432`、`5433`、`15432` 的监听进程。
2. `netsh interface ipv4/ipv6 show excludedportrange protocol=tcp` 均显示 `5384–5483` 被排除，`5432` 和 `5433` 都在其中。
3. 在 `127.0.0.1:5432` 做临时 TCP 绑定测试，返回 `native=10013 socket=AccessDenied`；`127.0.0.1:15432` 同样测试成功，随后立即释放测试监听。
4. 当前现象符合 Windows 排除端口导致无法绑定，而非一个已找到的普通进程占用。没有确认该排除范围由哪个系统组件创建，不将其归因于特定 WSL / Hyper-V / VPN 服务。

Microsoft 的 [排除端口与 10013 说明](https://learn.microsoft.com/en-us/troubleshoot/windows-server/networking/error-10013-wsaeacces-is-returned) 提供相关机制背景，但该文适用系统与本机不同；本机结论以以上实测为依据。

## 已授权修复方案

- OJ 宿主机 PostgreSQL 端口改为 `15432`；容器内部仍使用 `5432`，仅绑定 `127.0.0.1`。
- 同步本机 `.env` 的 `POSTGRES_PORT` 与 `DATABASE_URL`；不输出或更改凭据。
- 同步 `.env.example` 与 Compose 默认值，保证新环境初始化采用同一端口。已有 `.env` 不会被 setup 自动覆盖，其他副本需要手动同步这两个值。
- 通过 Compose 更新本项目 PostgreSQL 容器，继续挂载 `oj-dev_postgres-data` 命名卷。按 [Compose up 行为](https://docs.docker.com/reference/cli/docker/compose/up/) 配置变化会重建容器并保留挂载卷。

## 操作与实际结果

依次执行，而非仅修改文件：

1. 更新 `.env.example` 和 Compose 的宿主机默认端口为 `15432`。
2. 本机 `.env` 的 `POSTGRES_PORT` 和 `DATABASE_URL` 同步到 `15432`；其他端口 / 用户名 / 密码 / 数据库名保持原值，未打印凭据。
3. 先通过 `config --quiet`，再执行以下命令应用配置：

   ```powershell
   docker --context desktop-linux compose --project-name oj-dev --env-file .env -f deploy/docker-compose/compose.yaml up -d --no-deps --pull never --wait --wait-timeout 60 postgres
   ```

4. Compose 重建 `oj-dev-postgres-1` 并返回 healthy；实际映射为 `127.0.0.1:15432->5432/tcp`。挂载的命名卷仍为 `oj-dev_postgres-data`，路径 `/var/lib/postgresql/data`；没有删除卷或初始化替代卷。数据库内容未做完整盘点，不能把卷名核对描述为全量数据审计。
5. 临时构建并启动现有 API，让它读取本机 `.env`，仅覆盖临时 HTTP 端口 `47443`；`/healthz=200`，`/readyz=200`、响应 `{"status":"ready"}`，系统接口仍报告 `judge_enabled=false`。检查结束后已停止该临时 API；正常开发仍使用 `npm run api`。
6. `npm run doctor`、`npm run infra:config` 均通过；最终再次查询 Docker 得到 `29.8.1 / linux`。
7. 完整 `node scripts/infra-smoke.mjs` 通过 PostgreSQL 查询和 Redis PING，随后在 MinIO HTTP 检查失败。此问题独立记录，未继续尝试修复。

临时验证脚本放在工作区 `work/`，不是新增正式项目测试。未重新运行全套前端 / 单元 / race / CI 检查；本次只改端口配置和文档，没有业务代码变更。

## 后果、回退与停止位置

数据库客户端和宿主机 API 必须连接 `127.0.0.1:15432`；未来同一 Compose 网络内的服务仍应使用 `postgres:5432`，不能把宿主机端口写成容器内部端口。新副本通过 setup 使用新默认值；已有副本需同步 `.env` 的两个值，setup 不会覆盖它们。

如果未来恢复旧端口，须先确认 Windows 排除范围和绑定测试允许 `5432`，同时改回 `POSTGRES_PORT`、`DATABASE_URL`，再用上述 Compose 命令更新容器；不需要删除数据卷。当前本机不满足该回退前提。

MinIO 仍不可用，详见 [独立阻塞记录](2026-09-30-minio-unavailable.md)。已按用户规则停止后续开发，等待用户处理并通知继续。前端框架仅交付详细 MD 的限制继续有效；完整 Phase 0 验收及 P1-001 没有完成。
