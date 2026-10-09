# 编译 MVP 实施与停止交接

- 日期：2026-10-01；分支task/MVP-001-compile-web-2026-10-01，基线7a952c6；未提交 /推送，保留既有未提交内容。
- 最新用户范围：实际前后端，显示题面、输入C++、判断编译成功，不做sandbox；已授权前端并行开发与独立review。
- 实际执行者：root（后端 /环境 /联调 /记录）、mvp_frontend（apps/web）、mvp_review（独立文档阅读后因新阻塞中止）。

## 已交付

- API：内置A+B列表和题面；校验输入并通过固定COMPILER_URL转发；judge_enabled继续false；数据库readiness不阻止无DB编译。
- 独立Linux编译worker：真实G++14.2.0、固定C++20编译链接参数、唯一临时目录、15秒超时 /Linux进程组清理、64KiB诊断输出、单次编译并发；从不运行program。
- Docker：compiler独立Compose、UID10002、127.0.0.1:8081；基础镜像摘要及g++包版本固定，apk传递依赖未完全锁定。原PG /Redis /MinIO及卷未改变。
- 前端：题库 /题面 /样例 /Monaco本地worker /真实编译状态与诊断、输入大小校验 /防重复 /过期结果提示 /重置确认 /重试、响应式CSS。静态检查不等于浏览器已验收。
- 命令：compiler:build /up /down /status /test、mvp、mvp:smoke；[使用说明与Agent后续职责](../mvp/README.md)、[接口契约](../mvp/CONTRACT.md)、[验证记录](../mvp/VALIDATION.md)、任务MVP-001/002/003与ADR-0004。

## 实际检查

`npm run check`、基础`smoke`、Windows Go单测、Linux `compiler:test`、真实GCC `mvp:smoke`均通过。真实smoke覆盖合法C++、语法错误、无main链接失败和无限循环源码仅编译。Linux专项通过输出排空 /截断、timeout、编译器子进程回收、繁忙、启动错误与目录清理。前端typecheck /lint /build均通过，Monaco体积警告仍是已知非阻塞项。

非MD源码工作区摘要：74文件，SHA256 `b9d9ee3ed07ca11d96b99efb9eae472c189a1d8757dd970a7da39b334d508364`，由scripts/source-digest.mjs生成；此后停止实现，不再修改源文件。Go /npm依赖锁文件未变。远程CI与race未执行，不冒称完整runner /安全验收。

## 新阻塞与停止

首次创建IAB标签页返回“不属于当前浏览器会话”，只读inventory无标签页，浏览器QA未完成。详见 [B-MVP-001完整前因后果](../issues/2026-10-01-mvp-browser-session.md)。按用户规则停止后续开发 /测试 /审阅，reviewer仅归档已读文档与未审范围。独立报告不代表审阅通过，见 [报告](../reviews/2026-10-01-mvp-review.md)。

## 现有服务与继续入口

服务保持运行：前端 [127.0.0.1:5173](http://127.0.0.1:5173)、API8080、compiler8081。用户可手工查看；页面HTTP200已核实，交互未验证。开发supervisor PID5952及本次日志保留在.local，重启不要重复占用端口；在交互终端使用npm run mvp时Ctrl+C停止API /web，compiler:down停止独立编译容器。

用户处理浏览器会话问题并通知继续后，先复验工具会话，再执行浏览器测试与独立代码审阅；通过后才能把MVP任务标done。完整Phase1设计仍待单独审阅，未来真实Judge任务与完整前端任务包保留，不自动视为完成。
