# 文件判题MVP交接

- 日期：2026-10-02；分支task/JUDGE-001-file-cases-sandbox-2026-10-01；root后端 /集成，/root/mvp_frontend前端，/root/mvp_review独立Architect+Security。
- 当前停在最终启动预检，新[B-JUDGE-002](../issues/2026-10-02-judge-final-preflight.md)交用户；不宣称最终产品验收或网页在线。
- 最终源码100文件，摘要 `3ee10a98dbff54fa0e567a6a18161a3b38aa67bab8e6190f9643fddb42bb7357`；判题镜像 `sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74`；无新增依赖 /锁文件变更，未提交 /推送。

## 交付内容

服务器testdata `.in/.out` manifest、有界类型核对 /不可变快照 /hash；可信helper只编译并导出本次artifact；独立judge-mvp固定Docker OCI限制 /逐例监督 /可信OOM /完整checker /输出展示 /清理taint；API校验和固定HTTP桥接；新前端运行按钮 /总判定 /逐例 /共享锁 /取消 /源码过期 /确认重置。

编译MVP旧接口仍保留原compile-worker，严格OCI编译profile仅覆盖新判题链路；不能把新judge的安全证据扩大到旧compile-worker。Anolix仅作为已锁源码的设计参考，使用Docker OCI/runc适配，没有直接执行Anolix CLI。

## 实际检查

完整check、基础smoke、旧真实编译mvp:smoke、受限Linux可信单测均通过；真实修正版API矩阵17项全部通过，并含动态manifest /file SYSTEM_ERROR /cancel /child cleanup；真实旧声明卷镜像启动拒绝通过。独立实现源码审阅及其fake单测通过。证据和未测范围见 [VALIDATION](../judge/VALIDATION.md)，原始结果在.local/evidence。

受控shutdown旧运行实例通过，无容器 /job目录残留；随后最终重启报cgroup_v2_required，API未上线，最后AC smoke未执行成功。浏览器只读取inventory，仍没有真实操作证据，旧B-MVP-001未关闭。所有初次失败 /跳过均有记录，未伪造首次通过。

## 下一步与Agent职责

1. 用户处理Docker info /cgroup预检新阻塞并通知继续，按问题档案恢复条件核实。
2. root仅在恢复后重启当前源码、复验最终API AC及必要受影响矩阵，再执行实际浏览器验收，更新唯一台账。
3. Frontend Builder已完成本轮apps/web，不需重搭框架；若浏览器发现实际问题，按受审DTO仅修改页面，再做typecheck /lint /build。完整框架后续按[任务包](../frontend/AGENT_TASKS.md)的路由 /查询层 /UI /账号 /历史依赖分段认领。
4. reviewer在恢复后的运行 /浏览器证据齐全时收束产品验收；此前代码审阅通过不等于浏览器通过，不重新请求已持有的并行 /审阅授权。
5. 文档 /fixture Worker若后续被明确认领，可补更多可信例子与Monaco体积优化，不改执行权限 /判定优先级 /核心契约。

本机CPU累计 /可信峰值仍null，时间是含启动开销的wall；无完整Phase1、登录、队列 /历史 /生产安全 /远程CI或本机race验收。原基础卷及.env保持不变。
