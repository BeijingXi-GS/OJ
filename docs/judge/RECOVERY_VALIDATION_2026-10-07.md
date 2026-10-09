# 2026-10-07 本地恢复验证与边界

Owner：root；任务 COMM-001..004、PUB-001。当前非 MD 源码共 250 份，按路径排序的 `{path,sha256}` 数组 JSON 再取 SHA256：`5ffef55d8235aa1ef4f41e58be2e0a7adb4bef6505856a8cfda76cd2fe9a8a45`。本机逐文件清单保存在 `.local/evidence/recovery-20261007-snapshot.json`，不上传 `.local`。此前 200/226/235 份摘要仅为历史。

## 修复与测试

| 检查                                  | 实际结果                             | 证据及限制                                                                                                                  |
| ------------------------------------- | ------------------------------------ | --------------------------------------------------------------------------------------------------------------------------- |
| 特判生命周期专项                      | PASS，exit 0，14 个子场景，包 3.470s | `checker-cleanup-20261007.log`；可信 fake Docker，不执行宿主 C++                                                            |
| 首次完整检查                          | FAIL，exit 1                         | `recovery-20261007-check.log`；多个本机 HTTP 客户端场景返回 judge_unavailable，后续构建未执行                               |
| 正常取得本轮 network grant 后完整检查 | PASS，exit 0                         | `recovery-20261007-network-check.log`；格式、Go vet、前端 lint、Go 测试、4 个 Node 测试、SSR、Go build、tsc、Vite build     |
| 通信矩阵脚本语法                      | PASS，exit 0                         | `node --check scripts/communication-test.mjs`；没有运行真实矩阵                                                             |
| A+B 教师包静态数据                    | PASS，4 对数据                       | 模板与 descriptor 一致、cases/限制与内置 manifest 一致、数据原字节一致、BigInt 求和一致、输入范围与 32 字节上限通过         |
| 源码交付排除检查                      | PASS                                 | Git 非忽略文件中无 `.env`、`.local`、node_modules、dist、exe；无二进制文件、已检查的令牌/私钥模式命中。不是全面秘密审计保证 |
| Docker 普通查询                       | FAIL，permission denied              | 显式 Linux pipe，无法确认引擎/cgroup/镜像/容器数                                                                            |
| Docker 正常提权查询                   | 未执行                               | 自动审批传输中断，详见问题档案；不是明确安全否决                                                                            |

上述日志均位于 `.local/evidence`。完整 Go 检查可能复用未改变包的缓存；受影响生命周期专项明确使用 `-count=1`。Windows 无符号链接权限的相关测试跳过，不能宣称 Windows 链接路径实测通过；此前 Linux 结果也不能替代当前源码的 Linux 实测。Vite 的 Monaco 大块提示仍存在，不阻止构建，本轮未新增依赖或调整警告阈值。

修复只针对测试夹具：先证明 checker_cleanup 的 SYSTEM_ERROR、taint、job 和只读产物都保留，再在测试收尾恢复两个可信假产物的属性。生产执行、资源限制、清理失败拒绝继续、未知文件不改属性等策略保持原实现。题库 pending 发布和 Windows 四种已知产物清理实现一并纳入当前全量回归。

## 通信题交付内容

内置 `communication-a-plus-b`，另保留 communication-sum 和 communication-maximum，内置题共 10 道。发送端读取 a、b 并通过单向通道传递，接收端计算和；两份源码显式按角色绑定。A+B 每角色 64 MiB、每例共享 5000 ms、通信 32 字节；四组数据覆盖普通、零、负数、大整数。

[学生题面](../../examples/communication-a-plus-b/PROBLEM.md)、[教师包](../../examples/communication-a-plus-b/README.md)、[环境与验收步骤](COMMUNICATION_ENVIRONMENT.md) 均已保存。学生无需了解的模板/工程说明不放在学生题面；工程规则见环境指南。手动模式仍是本机演示，不具有真实身份隔离。

## 仍需真实环境验收

1. 正常 Docker Linux API、镜像、cgroup v2、严格预检、服务启动和清理基线。
2. `npm run judge:communication:test` 十二项真实 A+B 矩阵，保存完整响应、角色和源码/数据/产物绑定及每次清理证据。
3. 双 attach、取消与 busy、EOF/提前退出、挂载隔离、无答案读取、教师上传与重启持久化、三模式浏览器验证。
4. 旧普通/交互/特判回归、监督器停止阶段和 Linux 链接测试；独立审阅真实证据。

因此本轮可交付源码和通过本地检查的开发分支，不能签最终真实产品 PASS，不合并 main 或生产部署。独立结论见 [10-07 审阅](../reviews/2026-10-07-recovery-review.md)，问题原因和恢复对策见 [本轮档案](../issues/2026-10-07-approval-diagnosis-and-recovery.md)。
