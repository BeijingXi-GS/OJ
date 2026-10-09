# 特判增量开发及启动取消修正

日期：2026-10-03；处理人root/本轮已授权开发Agent；类别：新增实现与独立审阅发现的常规缺口，无未知外部环境阻塞。用户要求继续处理并解决，授权在本轮范围修正；此档保留前因后果，实际验收仍待补录。

## 背景与经过

1. 旧三模式产品已验收，用户本轮要求技术说明仅工程师、内置构造特判与教师上传C++特判。设计先独立审阅，明确快照绑定、4例/10秒编译/2秒特判/70秒总上限、原始合计1MiB输出及严格JSON。
2. 首轮root修复旧shutdown fetch失败提示，helper单测4项通过；独立审阅发现停止Promise未共享、预检时默认信号、用例列表未绑定。已改共享stopPromise、Go预检前信号注册且上下文可取消、结果case ID/顺序及完整性绑定、自定义4例上限。
3. 18:37:35启动预检；18:37:52服务ready附近发送Ctrl+C。当时API仍使用同步build，公共run helper遇构建被中断直接process.exit，supervisor exit1且未等待打印清理确认。实际成功查询0本项目容器、0任务，5173/8080/8082全关闭。不是观察到容器泄漏，但该父进程提前退出不符合验收。
4. 将开发启动器构建改异步受控子进程，每次等待后检查stopping，停止不再经过run的同步process.exit。18:41:20再次预检、18:41:35 ready、18:41:38 API已启动，共享Ctrl+C后打印“Owned judge stopped; container and job cleanup confirmed.”。PTY外层PowerShell仍exit1，尚不能据此断言Node自身exitcode或信号先后，需要独立观察和最终补验。
5. 增补纯null新checker/compile单变量wire测试，避免同时含unknown字段掩盖null验证。未知checker不再落入token比较；原始参与者输出供特判，展示截断另计。

## 已知事实与风险

Docker Linux/cgroup2已真实复查，本次默认工具权限读Docker失败后使用已有授权执行检查成功；这不是Docker引擎失联。真实启动前旧supervisor停止提示再次复现，仍无残留。不得把健康端口关闭当作清理证明；必须拥有子进程正常退出、Docker查询成功为空和任务目录为空。

预算达到上限是SYSTEM_ERROR与清理，不保证最慢全部规格完成。新未知外部阻塞仍单开档并停止交用户；此处均为本轮实现/审阅缺口，在已授权范围完成修正。

## 验证及当前状态

root协议/API专项两轮Go tests实际exit0；退出helper4项实际PASS。backend三包tests/vet、frontendtypecheck/lint/模式及文件专项已报告通过，完整check/真实特判矩阵/页面/重启/最终独立实现审阅尚未执行完，不标增量done。后续补录到本档及本轮验证/交接。

## 停止补录

后续新增私有stdin pipe/EOF受限控制、异步build和taint观察，已独立复读PASS；18:47:59预检开始、18:48:05真实私有pipe取消exit0并helper确认清理。三阶段脚本首轮assert日志缺口失败（tests为空/未完成），随后补保存日志，但重跑被自动审批通信失败拒绝创建进程；新外部阻塞见[B-SPJ-001](2026-10-03-spj-approval-transport.md)。完整check只到Agent编辑中的格式失败；前端和review格式后已报告整理，完整check未重跑。前端文件名保留名/首字符与交互前导0校验按独立建议已对齐。真实特判矩阵、页面和重启未执行，局部可信测试和独立实现PASS不代表产品完成。全部停止，等待用户恢复。
