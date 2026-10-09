# 三模式、教师上传与交互MVP交接

日期2026-10-03；任务MODES-001..004；root整合，teacher_catalog/interactive_runner/modes_frontend分别实现目录/执行/页面，roles_review独立Architect/Security审阅。用户明确仅本机手动切换，不新增真实登录。

## 已交付

学生显示聚合判定/通过数量，不返回隐藏逐例信息；教师显示用例输入/标准答案/输出与双向交互日志，能上传题面与.in/.out或固定协议秘密.in；工程师另看退出码/OOM/hash/交互器编译和只读debug。交互猜数是真实双容器stdio、独立编译/限制/终态/清理；任意交互器不在本轮范围。

内置6题/27例；本轮上传3题已持久保留，当前共9题。上传重启后逐题独立解AC2/2，所有题面/文件/hash不变。原5题10次与17矩阵、交互13判定+2真实生命周期、6三mode投影、8组NUL边界、工程check/Linux单测/实际浏览器均通过，见[完整验证](../judge/MODES_VALIDATION_2026-10-03.md)。最终状态由[PROGRESS](../../PROGRESS.md)和[独立review](../reviews/2026-10-02-modes-interactive-review.md)维护。

分支task/MODES-001-interactive-2026-10-02；非MD源码176文件SHA2568785b5dd63074e25633de5f6d78e52ddbaf4b94c6b6a4df08e3349a4fbb60b1c。镜像仍sha256:188abbaca2400458eccac5fa6af052b1e8a2a0558050c0e45123f18f2dbc7a74。无新依赖/提交/推送，保留原工作区和基础数据卷。

## 使用与运行状态

当前Web5173、API8080、compiler8081、judge8082均只绑定本机，最终四服务200、9题、0本项目执行容器/临时job。浏览器产品tab已保留。新judge10:06:32严格ready、API10:06:36开启两个worker；已重新编译包含最终16MiB内部响应修正。

文档收尾再次中断后原服务/引擎未运行，按已有授权恢复；最终17:41:23 judge再次严格ready、17:41:25 API开启编译/判题，四服务/9题/零残留复验通过，上传交互题再次AC2/2且同testdata hash。上午完整已审阅实证保留，下午恢复记录另存modes-resume-readiness/judgement.json；[恢复问题已关闭](../issues/2026-10-03-final-interruption-services.md)。没有更改源码或镜像，独立绑定不变。

打开[题库](http://127.0.0.1:5173)，顶栏选择模式；操作、数据限制、固定交互协议和各Agent下一步见[使用说明](../judge/MODES_USER_GUIDE.md)。再次开机需Docker Linux引擎在线、已有compiler/profile；在仓库运行npm run mvp。上传根默认为.local/teacher-problems，API和judge配置必须一致。提交源码/结果仍仅当前页面，题目数据才持久化。

## 问题与后续

Docker未运行已恢复并真实复验；实现期审阅修正及JSON膨胀边界均已记录并关闭。Windows共享终端Ctrl+C曾导致supervisor没有获得shutdown响应；实际端口/容器/job已全部退出且重启通过。提示本身仍是交用户处理的[非阻塞运维项](../issues/2026-10-03-supervisor-stop-warning.md)，未假称握手成功、未强杀Judge或删除未知容器。

后续真实身份/权限、提交历史/异步队列、任意交互器/更多语言、资源峰值、生产安全须另立受审任务，不能把此本机MVP等同完整Phase1。前端继续在现有框架做筛选/分页/草稿等小任务，API DTO先审；[职责表与步骤](../judge/MODES_USER_GUIDE.md)和[前端任务包](../frontend/AGENT_TASKS.md)可直接认领。新未知阻塞仍单独MD、交用户并停止。
