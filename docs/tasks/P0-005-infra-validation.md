# P0-005：本机基础服务联通验证

- 实时状态：见 PROGRESS.md
- Owner 角色：Builder / 环境维护者
- 实际执行者：初始化尝试见 VALIDATION
- 风险：R1
- 允许修改：本项目 Compose、scripts 和开发环境说明
- 禁止：重置宿主 Docker、删除既有容器/数据卷、更改生产服务
- 依赖：Docker Linux engine 可运行

## 验收

```sh
npm run doctor
npm run infra:config
npm run infra:up
node scripts/infra-smoke.mjs
```

PostgreSQL SELECT 1、Redis PONG、MinIO health 均通过；启动 API 后 /readyz 返回 200。失败时记录 engine/镜像拉取/端口/认证的实际阻塞项；不能用 Compose 语法通过代替容器启动成功。

结束时可 infra:down 停止本项目服务并保留命名卷。
