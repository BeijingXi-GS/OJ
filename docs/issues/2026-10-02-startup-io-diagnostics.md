# 启动IO诊断脱敏收束

- 日期2026-10-02；发现者独立/root/judge_reviewer，修正者root；状态已修正并复验，不构成外部环境阻塞。

## 前因、影响与修正

本轮多题裁定要求启动日志只打印固定操作/分类。Docker control已脱敏，但独立审阅最后发现NewEngine的filepath.Abs/MkdirAll与Preflight的MkdirTemp等分支仍直接返回OS error，而服务会log err.Error()；真实启动IO失败时可能包含本机路径。这是审阅发现的代码失败路径，不是当前主机已发生目录故障，不能伪造实际泄露事件。

root将data_root、jobs_root及probe identity/directory失败改为固定分类，保留拒绝启动、目录边界、预算和权限。新增StartupIOErrorDoesNotExposeHostPath，使用临时普通文件及其子路径模拟不可创建jobs根，断言固定jobs_root_create_failed/不含临时路径。root和独立reviewer分别执行专项fake测试通过。

最终在无活跃任务时仅匹配本项目API可执行文件的完整路径关闭API，让supervisor通过内部控制接口正常关闭拥有的judge，确认0容器/0job残留；supervisor因API被主动结束返回1，是本次受控重启行为，不是预检失败。重新编译启动最后源码：19:15:06 judge严格ready、19:15:12 API开启判题。独立数组和解5例AC、四服务200及零残留通过；详细证据在.local/evidence/final-readiness.json。

本改动只影响启动失败日志，不改变执行profile。17项矩阵/五题十次smoke/浏览器是在此前相同执行路径版本完成，最后版另经专项单测、完整工程检查、严格预检及AC验证；不冒称全部矩阵重新执行。无需用户处理，后续新未知阻塞仍按用户规则单开MD并停止。
