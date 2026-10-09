# COMM-001：通信契约、API与集成

2026-10-07 更新：测试夹具收尾已修复，当前完整检查通过，通信环境/题面/教师包已交付。当前 250 文件绑定和真实 pending 见 [新验证](../judge/RECOVERY_VALIDATION_2026-10-07.md)；此前失败保留历史。Owner root，独立增量 reviewer 为 recovery_final_review，真实联调仍待 Docker 恢复。

2026-10-06：受审 Windows 发布/清理修复与通信环境包已保存；本轮完整检查因 checker_cleanup 测试临时目录收尾失败而 exit 1，定位读取随后遇审批连接故障。用户要求只保存记录并停止。下一步和日志见 [问题档案](../issues/2026-10-06-recovery-check-and-approval-stop.md)，下方旧验证为历史。

- 状态：以[PROGRESS](../../PROGRESS.md)为准；本地实现已交审，真实产品pending。
- Owner：root Architect/Builder；Review owner：comm_review Architect/Security；风险R3。
- 范围：filejudge/contract.go、communication_contract.go及专属tests、checker_contract.go、internal/httpapi、启动器版本门禁、根MD/任务/问题/验证/交接。
- 输入：[communication-mvp/0.1.0](../judge/COMMUNICATION_CONTRACT.md)，独立局部设计PASS。用户授权继续开发、双源码通信及本地先行，已有并行授权有效。
- 目标：严格双角色请求与规范摘要、可信状态/数据绑定、独立端点、三模式服务端白名单、新worker版本门禁。
- 非目标：API运行代码、登录、持久异步提交、生产部署、绕过审批或删除历史数据。
- 验证：新增A+B前Go专项/完整check与独立实现审阅历史PASS；新增A+B窄独立专项PASS，当前五包与权限复验exit1，停止依赖验收。235文件绑定与实际检查见[验证](../judge/COMMUNICATION_VALIDATION_2026-10-05.md)；Docker真实验收待环境恢复。
- 失败：记录[开发修正](../issues/2026-10-05-communication-development.md)、[本地测试权限](../issues/2026-10-05-local-test-permissions.md)与[Docker环境](../issues/2026-10-05-local-docker-access.md)，不得把fake或单例中间AC当完整隔离验收。
