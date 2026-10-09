# COMM-002：双角色编译与单向监督器

2026-10-07 更新：本地完整回归通过，Windows 发布/清理兼容修复已纳入当前版本；后端源码交付，真实 Docker 验收 pending。新的 [验证记录](../judge/RECOVERY_VALIDATION_2026-10-07.md) 与 [下一步职责](../handoffs/2026-10-07-recovery-delivery.md) 为接续入口，下方旧摘要和结果为历史。

- 状态：以[PROGRESS](../../PROGRESS.md)为准；本地实现已交审，真实产品pending。
- Owner：comm_backend Builder；Review owner：comm_review Architect/Security；风险R3。
- 范围：filejudge除root契约文件外实现/测试、problems、problemstore、testdata；不改API/前端/共享MD/Git/依赖或控制Docker。
- 输入：[受审契约](../judge/COMMUNICATION_CONTRACT.md)；显式sender/receiver绑定、逐角色20秒编译、单向rawbyte受控管道、共享case wall和70秒job、实际退出/清理证据。
- 输出：三内置通信题12用例，包括sender传a、b而receiver求和的communication-a-plus-b；双模板上传/深拷贝/不可变摘要、独立worker路由/metadata/health、EOF/超限/RE/OOM/取消/双清理与taint。
- 可信验收：26生命周期场景、五编译失败、原八完整pipeline加A+B正确/错误两项及pipe/budget/snapshot/store/upload/worker测试。新增前wrapper Go tests/vet为历史PASS；A+B独立窄审/专项PASS，当前完整回归权限失败，不登记整体PASS，见[问题](../issues/2026-10-05-local-test-permissions.md)与[验证](../judge/COMMUNICATION_VALIDATION_2026-10-05.md)。
- 真实验收：未执行；Docker矩阵、挂载、真OOM/取消、上传重启按[指南](../judge/COMMUNICATION_USER_GUIDE.md)接续。
- 问题策略：[开发档案](../issues/2026-10-05-communication-development.md)。只执行可信fixture，不在宿主编译/运行C++。
