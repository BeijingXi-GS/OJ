# 编译 MVP 验证记录

日期2026-10-01；分支task/MVP-001-compile-web-2026-10-01；保留未提交改动，无push /远程CI。实施者root与mvp_frontend；因新浏览器会话阻塞停止，独立review未完成，不以自检代替。

## 已有实际证据

- Docker29.8.1 /linux amd64；G++14.2.0，apk包14.2.0-r4；compiler容器UID10002，端口仅127.0.0.1:8081。
- compiler:build通过，镜像manifest list sha256:5f7042211a7741f1a2141b9500f9aa812f96a26fb90c0f5bdc97a1db93563c81，config0d46ae4e044876f6aa3d2ed6e645f46e711b4d9e6dda75b74ccc363043b076be。
- npm test通过；覆盖API题面、404、错误输入 /JSON /Content-Type /UTF8字节 /body限额、业务status和服务错误转发；HTTP client错误与取消 /不跟随重定向；worker输出限制、正常 /CE /超时、繁忙、启动错误、目录清理。
- compiler:test在Linux真实镜像通过；进程组超时后编译器子进程被终止 /回收，输出持续排空并截断，所有测试通过。
- mvp:smoke通过：有效模板success338ms；语法错误CE9ms；缺少main链接错误CE26ms；无限循环源码success24ms；API→真实compiler，readyz未配置DB503，judge_enabled=false。
- npm run smoke通过，原基础接口回归与/api/v1/submissions仍404。
- Frontend Agent typecheck /lint /build通过；无新依赖；已知Monaco bundle警告仍非阻塞，详见既有问题档案。
- 最终 `npm run check` 完整通过（格式 /Go vet /ESLint /Go tests /Go build /TypeScript /Vite构建，816 modules）。`git diff --check`通过；未变更Go /npm锁文件。
- 停止后只读核实原服务：5173 HTML200、API healthz200、system compile_enabled=true /judge_enabled=false、worker healthz200 /GCC14.2.0。网页HTTP200不能代替真实浏览器交互。
- 非MD源码摘要：74文件，`b9d9ee3ed07ca11d96b99efb9eae472c189a1d8757dd970a7da39b334d508364`。

## 待完成验收

- 浏览器实际题面、默认模板编译成功、修改源码编译错误、过期状态与重置确认、移动布局。
- 独立MVP-003具体代码审阅及root验收；reviewer只完成文档阅读并中止，不能登记通过。
- 浏览器阻塞完整记录见 [问题档案](../issues/2026-10-01-mvp-browser-session.md)，当前按用户规则停止。交接已保存，恢复条件见档案。

## 不属于此次验证结论

未运行任何用户生成的program，未做答案判定、sandbox /资源隔离、安全公网服务、业务DB /队列 /持久化、完整Judge矩阵、race或远程CI。Linux的短超时辅助进程测试验证取消机制；不是测量真实复杂源码稳定触发15秒超时。编译镜像并非完整可重复供应链锁定，apk传递依赖未全部固定。
