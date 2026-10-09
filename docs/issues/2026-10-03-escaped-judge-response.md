# 新增快照展示导致JSON响应上限不足

日期2026-10-03；发现者独立roles_review；实现者root；属于本轮模式/上传实现的必修边界，沿既有修复及独立审阅授权处理，不是新的未知外部环境阻塞。

## 前因与影响

旧文件判题Client限定内部响应6MiB。当时没有每例input/expected快照。本轮教师/工程师需显示隐藏输入和标准答案，于是每例新增各64KiB文本。JSON会把NUL等控制字符变为6字节转义，8组合法.in/.out（各60000 NUL）和相等stdout可生成超过8.6MB的内部响应。旧上限会令Client在投影前返回502；学生虽然不展示详情，也受影响。

这是独立审阅静态推导出的实现缺陷；修正前未将失败的实际判题伪造为成功，也不将静态推导当作已观测线上事故。

## 修订与校验

roles_review独立确认16MiB窄上限：按更保守的交互结构每例288KiB字符串×8×6，加双编译诊断64KiB×2×6，共14.25MiB；固定结构与有界消息小于剩余1.75MiB。仅扩大内部完整JSON接收上限，逐字段展示、文件、上传请求、原始输出、容器、时间和内存限制均不改变；超16MiB仍拒绝。

root将MaxResponse改16MiB，新增8例普通/交互最大转义完整wire测试。实际targeted TestClientMaximumEscapedSnapshots和TestClientProtocol通过，含合法>6MiB解码与16MiB+1拒绝。旧协议文档6MiB保留为历史，当前增量契约明确替代。真实8组NUL用例、最终工程检查与独立复审待补；未达到这些条件前本问题不登记关闭。

最终实证：新实例实际8组60000 NUL全部AC8/8；学生响应290字节且无cases，工程师响应8642720字节（大于旧6MiB）且输入/期望/输出全部逐字节匹配、exit0非OOM。原A+B manifest精确恢复，自建文件清理；escaped-response-smoke.json completed:true。最终npm run check exit0含全单测/vet/build。roles_review已独立重算临时数据snapshot SHA256和当前源码摘要、复核报告，边界必修已闭合；产品总签见最终review。此问题已修复并复验关闭。

## 用户接手条件

当前不需要主机操作或用户修复；若真实边界验证出现未知环境阻塞，另开外部问题档案并按用户规则停止，不弱化文件/隔离限制绕过。当前检查及最终结论由[进度](../../PROGRESS.md)和[增量契约](../judge/MODES_INTERACTIVE_CONTRACT.md)追踪。
