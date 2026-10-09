# JUDGE-005 已知预检恢复与最终联调

- 日期2026-10-02；Owner /root Backend Builder；Review独立Architect/Security。
- 用户明确要求“尝试解决并继续推进任务进度，增加一些题目”。授权恢复B-JUDGE-002和既有B-MVP-001，维护当前判题产品并新增题目；此前并行与独立审阅授权持续有效。
- 修改范围：internal/filejudge/docker.go、engine.go、preflight/diagnostic单测、scripts启动和真实矩阵、必要根MD /问题 /交接；不覆盖catalog或Frontend Builder认领文件。
- 交付：Docker Linux及完整预检恢复；将CLI失败与cgroup版本不符区分，脱敏有限诊断；多题metadata接线 /preflight全部题目校验；最终服务启动、API /Docker矩阵和浏览器复验。
- 保持原5s控制 /70s job /15s cleanup预算和全部隔离限制；不重置主机 /数据卷 /扩大执行权限，不把未知版本当通过。
- 验收：对应Go单测、check /smoke、真实全题AC与错误解WA、既有矩阵回归、独立审阅及实际UI；未通过必须如实登记。
- 新未知阻塞仍一事一档并停止；开发期实现和审阅修正属于已授权开发工作。

## 本轮交付证据

Docker恢复linux amd64/cgroup2、5s查询三次成功；固定诊断/缓冲和启动IO脱敏、全部题目preflight/metadata接线已实现并独立复审。18:50首次严格ready；最终19:15:06再次严格ready，API19:15:12启动，final-readiness.json四服务200、五题、独立数组和AC5/5、0残留。真实十次五题smoke、17项矩阵、工程/旧编译/Linux和页面QA通过。开发期失败/修正与未测分支见 [VALIDATION](../judge/VALIDATION.md)；本轮 [交接](../handoffs/2026-10-02-judge-recovery-problems.md)。原数据卷/限额/依赖不变。
