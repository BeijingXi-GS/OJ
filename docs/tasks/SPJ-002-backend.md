# SPJ-002：backend

- 状态：以 [PROGRESS](../../PROGRESS.md) 为准。
- Phase/风险：本机判题增量，R3。
- Owner/范围：special_backend Builder；允许internal/problems、internal/problemstore、internal/filejudge除contract.go与其专用tests、testdata/problems、新special测试脚本/例子。核心实现依赖SPJ-004设计PASS。禁止修改API/contract.go/apps/web/启动器。负责内置checker、构造题、上传源码/文件校验与持久化、不可变快照、独立checker编译运行/清理/资源故障。
- Review owner：spj_review。
- 输入：[special-judge/0.1.0](../judge/SPECIAL_JUDGE_CONTRACT.md)。
- 目标：完成本轮用户授权范围，原模式/交互可回归，工程细节仅工程师。
- 非目标：账号、生产部署、任意交互器、外部checker依赖、改基础数据卷、提交推送。

## 验收

实际进度以[本轮验证](../judge/SPECIAL_JUDGE_VALIDATION_2026-10-04.md)为准：完整check/可信及Linux测试/23项真实矩阵已通过相应版本，教师上传页面AC2/2；真实三阶段停止、上传重启、剩余页面和最终当前版本全量检查未完成。

## 失败与交付

新未知外部阻塞单开issues MD并停止交用户处理；本范围常规实现/审阅修正由已授权执行者完成记录。必须绑定最终源码/镜像/时间/证据，真实独立审阅完成才done。旧历史改动/数据全部保留；无需提交或推送。

## 2026-10-03实际交付与停止

本范围实现/可信测试已交付并局部独立review通过；最终产品验收未完成，当前blocked。真实停止脚本日志断言失败待定位，重跑动作因自动审批通信失败未执行；依据用户外部阻塞规则停止。已完成与未完成详见[当前交接](../handoffs/2026-10-03-special-judge-blocked.md)，恢复由用户先处理[B-SPJ-001](../issues/2026-10-03-spj-approval-transport.md)，无新测试/部署/数据删除。

## 2026-10-04恢复交付与再次停止

沿用用户继续修复/并行/独立审阅授权。实现已由root接收；lifecycle_fix仅停止脚本/下载链接/编辑器输入窄修，spj_acceptance为本次独立最终reviewer，docs_curator仅问题索引/用户指南。23项真实矩阵、Linux与对应版本完整check/两smoke通过，教师上传AC2/2。最新三阶段脚本仍失败，补诊断重跑遭审批通信中断，动作未执行；新增日志/最新停止补测未最终完整check。当前不标done，状态见PROGRESS。

[本次停止交接](../handoffs/2026-10-04-special-judge-paused.md)逐项列出恢复步骤，[当前独立报告](../reviews/2026-10-04-special-judge-acceptance.md)最终产品pending；root自检不作为独立通过。审批失败后已停止新增工作，旧失败记录/题目/文件均保留。
