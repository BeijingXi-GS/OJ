# 初始化验证记录

日期：2026-09-26（Asia/Shanghai）。执行者：本次初始化 Agent。此记录为实际工程检查，**不是独立 Architect / Security Review**。

## 验证基线

- Windows amd64；Go 1.25.4；Node 22.19.0；npm 10.9.3；Git 2.51.0.windows.1。
- Docker CLI 28.4.0；Compose v2.39.4-desktop.1；Linux engine 未就绪。
- Git main 已初始化；无 commit、无 remote。初始交付是未提交工作区，后续提交不得冒充已经运行过远程 CI。
- 代码与配置摘要：52 个文件；SHA-256：
  `ff37ff9427a83a874d806253b0584557cc4333860de0dabebe56ec404359c90e`。
- 摘要由 `node scripts/source-digest.mjs` 产生：按排序后的 Git 可见代码/配置文件 hash 清单再求 hash；排除 MD/docs、忽略文件、依赖、缓存和构建产物。文档后续更新不影响代码基线摘要。
- 依赖完整精确版本见 package-lock.json / go.sum；React 19.3.0、Vite 7.3.6、TypeScript 5.9.3、Monaco 0.57.0、sqlc 1.30.0。

## 结果

| 检查                                           | 实际结果     | 边界                                                                                        |
| ---------------------------------------------- | ------------ | ------------------------------------------------------------------------------------------- |
| npm 安装 + Go mod download                     | 通过         | npm ci 已执行；Go 默认代理超时后使用 goproxy.cn，未禁用 checksum                            |
| npm run format:check                           | 通过         | Prettier + gofmt                                                                            |
| npm run lint                                   | 通过         | Go vet + ESLint                                                                             |
| npm test                                       | 通过         | Run 生命周期/终态约束，HTTP readiness 正常/失败/未配置与错误信息隐藏；不涉及真实判题        |
| npm run build                                  | 通过         | Go 项目包 + TypeScript + Vite production build                                              |
| npm run smoke                                  | 通过         | API 真正启动，health/system 正常，未配置 DB 的 readyz=503，提交路由=404，完成后临时进程退出 |
| npm run sqlc                                   | 通过         | 生成 db.go / models.go / queries.sql.go，随后 Go 构建和测试覆盖该 package                   |
| npm audit                                      | 通过         | 更新 Monaco 后 0 vulnerabilities；只是当次依赖数据库结果，不等于系统安全审计                |
| Compose config --quiet                         | 通过         | 使用本地 .env 解析 PostgreSQL/Redis/MinIO 配置；不输出凭据                                  |
| 浏览器首页                                     | 通过         | 实际打开，看到“连接正常”和契约 0.1.0-draft                                                  |
| 浏览器 /editor                                 | 通过         | Monaco 编辑区及 C++ 示例、语法高亮正常渲染                                                  |
| Vite HTTP                                      | 通过         | 5173 首页返回 200                                                                           |
| API + .env                                     | 通过         | /healthz=200，/api/v1/system=200，数据库不可达时 /readyz=503                                |
| .gitignore                                     | 通过         | .env、node_modules、.local、apps/web/dist 被忽略                                            |
| Docker info / doctor                           | 阻塞         | 授权环境仍找不到 dockerDesktopLinuxEngine 管道；未启动本项目容器                            |
| PostgreSQL SELECT 1 / Redis PING / MinIO HTTP  | **未执行**   | 依赖 Docker engine，P0-005 后续执行                                                         |
| Go test -race                                  | **未能执行** | 返回“-race requires cgo”；本机没有已配置 CGO 工具链                                         |
| GitHub Actions                                 | **未执行**   | 工作流已配置，无 remote                                                                     |
| 独立 Architect Review / Linux Sandbox 安全测试 | **未执行**   | 尚无 reviewer / Sandbox 实现                                                                |

## 已修复的问题

1. Monaco 新版本调整了 exports 与语言模块路径。已使用公开导出路径与 cpp/register，TypeScript 与浏览器验证通过。
2. 首轮 npm audit 报告 Monaco 的 DOMPurify 间接依赖问题。已更新 Monaco 0.57.0，重查为 0。
3. Go 的 ./... 会扫描 node_modules 中的第三方 Go 文件。wrapper 已限定项目路径，再次测试未纳入外部代码。
4. doctor 的空白 stdout 曾掩盖错误说明，已修正并验证能显示真实 Docker 原因。

## 剩余限制

Monaco 采用路由懒加载，但编辑器 chunk 约 2.75 MB（gzip 约 713 KB），Vite 给出体积提示；这是开发骨架的非阻塞构建警告，Phase 2 可按产品性能预算优化。当前只配置 C++ 编辑预览。

Docker Desktop 曾通过桌面入口和官方启动命令尝试启动，仍未就绪；已结束等待启动的 CLI。未重置 Docker、删除任何数据或更改主机虚拟化设置。受限命令环境额外存在 Docker 用户配置读取限制；即便授权读取，engine 仍不存在，故不是单纯 CLI 语法问题。

API / Web 只在 localhost 做临时验证。按 README 的两个启动命令即可再次运行。真实 Linux Judge、队列消费、业务表、登录权限、S3 存储行为尚未实现。
