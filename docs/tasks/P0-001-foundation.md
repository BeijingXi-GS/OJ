# P0-001：工程基础与开发环境

- 实时状态：见 PROGRESS.md
- Owner 角色 / 实际执行者：Builder + Worker / 本次初始化 Agent
- Review owner：待指派 Builder reviewer
- 输入：归档的原始计划；ADR-0001 工程范围
- 风险：R1；核心类型草案另由 P0-004 评审
- 允许修改：apps、services/judge-worker、internal/httpapi、platform、scripts、deploy/docker-compose、CI 和根工程配置
- 禁止：真实执行选手代码、开放外部服务、未经评审的业务表、生产部署
- 依赖：本机 Go / Node / Docker engine

## 输出

可启动 HTTP API、独立空闲 Worker、React/Vite/Tailwind/shadcn/TanStack/Router/Monaco 基础；OpenAPI、pgx/sqlc；Compose、跨平台命令、依赖锁文件和 CI。

## 验收

| 操作                           | 预期                                               | 证据                           |
| ------------------------------ | -------------------------------------------------- | ------------------------------ |
| npm run check                  | 格式/lint/测试/前后端构建通过                      | docs/development/VALIDATION.md |
| npm run smoke                  | API 能启动；未配置 DB 不假报 ready；无提交执行路由 | 同上                           |
| npm run sqlc                   | 生成代码可编译                                     | 同上                           |
| npm run infra:up + infra-smoke | 三项真实依赖可用                                   | 同上；未完成时拆到 P0-005      |
| Builder Review                 | 确认工程边界和复现路径                             | 未执行                         |

## 回退

停止本项目进程，infra:down 保留数据卷；回退配置版本，不删除他人的代码、.env 或持久数据。
