# 特判真实验证脚本误用可选反馈字段

日期2026-10-04；处理人root；类别：本轮开发验证脚本缺口，已获用户“继续完成、解决问题、持续整理”授权修正。不是新的外部环境阻塞。

## 背景、经过与事实

Docker恢复后执行special-judge-test.mjs真实矩阵。前13项已记录通过；运行到custom-checker-reads-input-answer-output，特判按三路径成功执行，但是JS断言调用 `c.checker_feedback.includes(...)` 对未反馈的例子报 TypeError（undefined）。返回协议的checker_feedback为可选字段，没有stderr时正常省略；脚本错误地假设每例都存在。

失败发生在结果验证而不是C++特判编译/执行；不把这次进程exit1当作系统判题故障，也不把13项通过冒称全矩阵通过。finally已把脚本自建题upload-81e40ccb053a30089539102bbfde5051的源码恢复为原示例，report checker_source_restored:true。没有修改三道旧上传题、没有删除数据。

旧报告已单独保存 `.local/evidence/spj-20261004-special-matrix-first-failure.json`，原失败日志 `.local/evidence/spj-20261004-special-matrix.log`。该报告completed:false是历史事实。

## 方案与验证

只将 `.includes` 改成可选调用 `?.includes`，保持some必须找到空参考答案反馈的断言；同时明示teacher/student总体必须AC，不仅检查字段隐藏。原stdout、资源、完整快照、清理断言未放宽。

2026-10-04 16:32:38 完整重跑实际 exit0，23项全部通过，completed:true、checker_source_restored:true，每项成功查询0容器/0任务。包含本项三文件读取、空标准答案反馈和 teacher/student 实际AC。通过报告独立保存 `.local/evidence/spj-20261004-special-matrix-passed.json`，日志 `.local/evidence/spj-20261004-special-matrix-fixed.log`；新增自建题 upload-344a5e5d3e14957b7bc2ee66e858fc5b 保留。问题已关闭，旧失败证据不覆盖。

后续凡可选协议字段必须按契约处理缺省；无反馈不能判作特判运行失败。状态/最终证据看[PROGRESS](../../PROGRESS.md)和本轮验证。
