# 开发环境与常用操作

## 工具与平台

| 项目       | 开发基线                                    | 说明                                        |
| ---------- | ------------------------------------------- | ------------------------------------------- |
| Go         | 1.25+；本机 1.25.4                          | 单一 module；以 go.mod/go.sum 锁定模块依赖  |
| Node       | 22.19+（22.x）或 24.x；本机 22.19.0         | 当前使用 npm 10.9.3                         |
| 前端依赖   | package.json 的精确版本 + package-lock.json | 复现使用 npm ci，不使用未锁定更新           |
| sqlc       | 1.30.0                                      | npm run sqlc 下载/运行并生成基础查询        |
| Docker     | Compose v2 + Linux engine                   | CLI 存在不表示 engine 已运行                |
| 实际 Judge | Linux、cgroups v2；编译器按 Phase 1 安装    | Windows 原生只验证 API/web/类型，不验证隔离 |

未来生产基线必须重新检查受支持版本和镜像 digest；本配置是本地开发环境。技术来源：[Vite 环境要求](https://vite.dev/guide/)、[Tailwind Vite 集成](https://tailwindcss.com/docs/installation/using-vite)、[shadcn 手工初始化](https://ui.shadcn.com/docs/installation/manual)、[Compose 服务定义](https://docs.docker.com/reference/compose-file/services/)。MinIO 当前开发镜像见 [官方 release](https://github.com/minio/minio/releases/tag/RELEASE.2025-09-07T16-13-09Z)。

## 初始化

从仓库根目录执行：

```sh
npm run setup
npm run doctor
npm run infra:config
npm run infra:up
node scripts/infra-smoke.mjs
```

setup 不安装或升级系统软件；它生成本项目 .env、恢复 npm 与 Go 依赖，并保留已有 .env。Go 与 npm 缓存位于 .local/，可在网络受限时复用。Go wrapper 将 ./... 限定为 apps/api、services、internal，避免误编译 node_modules 中第三方 Go 文件。

本次依赖下载使用的备用公开 Go 模块代理，可在默认代理无法访问时按需设置。保留 Go 的校验数据库验证，不关闭校验：

PowerShell：

```powershell
$env:GOPROXY = 'https://goproxy.cn,https://proxy.golang.org,direct'
npm run setup
```

Linux：

```sh
GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct npm run setup
```

这不是永久环境修改；能访问默认代理时无需设置。

## 启动与端口

| 服务          | 默认地址              | 启动 / 健康                     |
| ------------- | --------------------- | ------------------------------- |
| Web           | http://127.0.0.1:5173 | npm run dev                     |
| API           | http://127.0.0.1:8080 | npm run api；/healthz           |
| PostgreSQL    | 127.0.0.1:5432        | infra:up；pg_isready + SELECT 1 |
| Redis         | 127.0.0.1:6379        | infra:up；PING                  |
| MinIO S3      | http://127.0.0.1:9000 | infra:up；/minio/health/live    |
| MinIO Console | http://127.0.0.1:9001 | 凭据在本机 .env，未写入文档     |
| Judge Worker  | 无监听端口            | npm run worker；日志显示 idle   |

API 启动命令自动读取 .env。/healthz 只代表进程存活；/readyz 会检查数据库，不配置或连接失败返回 503。它目前不代表题库、队列或判题已经可用。Web 的“连接正常”检查基础 API，并不是数据库 readiness。

PostgreSQL / Redis / MinIO 只映射到 127.0.0.1。基础服务采用命名卷持久化。Phase 5 再建立业务 bucket、受限服务账号与对象生命周期；当前不自动创建“看似可用”的业务存储。

端口冲突：修改 .env 的相应端口；PostgreSQL 端口和密码变化后同步 DATABASE_URL，API 地址变化后同步 apps/web/vite.config.ts 中的开发代理。已存在 PostgreSQL 数据卷不会因更改 POSTGRES_PASSWORD 自动改密，应使用数据库管理流程更新；不要直接删卷。

## 验证命令

```sh
npm run check
npm run smoke
npm run sqlc
npm run infra:config
node scripts/infra-smoke.mjs
```

Linux/支持 CGO 的环境额外运行：

```sh
node scripts/go.mjs test -race ./...
```

CI 在 Ubuntu 上执行同一检查、race、sqlc 再生成差异及 Compose 基础服务健康验证。API smoke 会临时启动自己的 API 并自动停止。尚未配置 git remote，CI 文件写好不等于远程 CI 已通过。

Monaco 使用本地打包的 worker，不依赖外部 CDN；编辑器 demo 不保存代码、不发送提交。shadcn 使用手工配置与本地 button 组件，可在后续 UI 任务中按需要扩充。

## 停止与恢复

```sh
npm run infra:down
```

运行中的 API / Web / Worker 终端按 Ctrl+C。down 不删除数据卷；不要日常使用 down -v。缓存和 node_modules 可重建，.env 和数据库数据需要保留。

如果 Docker 命名管道不存在或引擎启动失败：启动 Docker Desktop 并确认 Linux engine 已就绪，然后执行 doctor / infra:up。若 WSL/虚拟化未就绪，在主机完成该前置后再验证；本仓库脚本不修改 BIOS、系统功能或 WSL 发行版。

## 真实验证记录

见 [VALIDATION.md](VALIDATION.md)。记录中明确区分已验证、失败和未执行；状态摘要见根目录 PROGRESS.md。
