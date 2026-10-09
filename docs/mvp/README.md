# 编译 MVP 使用与开发说明（历史）

2026-10-02用户已扩展实际文件判题，当前启动与Agent职责以 [文件判题使用说明](../judge/README.md) 为准；以下保留2026-10-01只编译增量的历史边界，旧接口仍仅编译。`npm run mvp`当前会同时启用受限判题服务。

当前范围由用户 2026-10-01 最新要求确定：显示题面、编辑 C++、判断编译和链接是否成功。前端实际实现，不再只交付框架 MD。真实判题、sandbox、登录、持久化均不在此次范围。

## 启动

进入仓库根目录；已有 node_modules 与 Go 缓存的本机副本无需重复 setup。其他副本先运行 `npm run setup`。要求 Node 22.19+ /24、Go 1.25、Docker Linux amd64 与 Compose。

```sh
npm run compiler:build
npm run compiler:up
npm run mvp
```

打开 [题库](http://127.0.0.1:5173)，点击 A + B，修改 C++ 后点击编译。默认模板能够编译；删除分号可查看语法错误，删除 main 可查看链接错误。编译成功不代表答案正确，程序不会被执行。

`npm run mvp` 同时启动本机 API 8080 和前端 5173，Ctrl+C 停止这两个进程；编译容器继续运行。停止编译服务使用 `npm run compiler:down`；它与原数据库 /Redis /MinIO Compose 独立，不删除原数据卷。若分别启动，可使用两个终端运行 `npm run api` 与 `npm run dev`。不要重复占用相同端口。

本 MVP 无业务数据库依赖。保留基础 `/readyz` 的数据库检查，因此没有 DATABASE_URL 时它返回503，而题面 /编译仍可正常使用。API 默认 COMPILER_URL 为 `http://127.0.0.1:8081`；system 的 compile_enabled 仅表示配置启用，真实服务状态须实际编译或检查 worker healthz。配置值只在服务器环境设置，用户输入不能选择服务地址。

## 编译环境和边界

编译镜像使用仓库已锁定且本机已缓存的官方 Redis Alpine 基础镜像摘要，在其 Alpine3.21 上安装锁定 `g++=14.2.0-r4`；镜像只以 compile-worker 为入口，不启动 Redis。使用这个基础层是为了复用已经验证的下载来源与本机缓存。apk 的传递依赖会随上游仓库更新，此 MVP 未宣称整个镜像字节可重复；验收记录镜像实际 ID、GCC版本和源码摘要。源码 /二进制构建不新增 Go 或 npm 依赖。

编译容器 UID10002，只暴露127.0.0.1:8081；API仅调用 HTTP，编译器只在独立worker启动。固定命令为 `g++ -std=c++20 -O0 -fdiagnostics-color=never solution.cpp -o program`，同时检查编译和链接。不运行生成的 program，不调用未来 JudgeEngine 或 Sandbox 接口。Docker仅提供Linux工具环境；没有做编译或运行沙箱，不用于公网服务。

源码上限65536 UTF-8字节、请求262144字节、诊断65536字节、一次一份编译、超时15秒；API HTTP调用20秒、写响应30秒。超过输出限制继续排空输出但丢弃尾部，前端显示截断提示；Linux超时终止编译进程组并由init回收子进程，生成目录在结束时清理。

## 验证命令

```sh
npm run check
npm run smoke
npm run compiler:test
npm run mvp:smoke
npm run compiler:status
```

check含格式 /lint /Go单测 /类型检查 /构建。compiler:test在实际Linux镜像执行输出截断、超时、子进程回收、繁忙、临时文件清理测试；使用测试辅助进程而非选手程序。mvp:smoke临时启动另一个API并通过真实GCC检查合法源码、语法错误、链接错误、无限循环源码仅编译；完成后停止临时API，不影响开发页面。浏览器验收和独立审阅见 [验证记录](VALIDATION.md) 与 [MVP-003](../tasks/MVP-003-review.md)。

## Agent 修改范围和后续工作

| 执行者                | 此次职责                             | 主要文件 /操作                                                                                                                  | 后续明确任务                                                                              |
| --------------------- | ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| root /Backend Builder | API、worker、镜像、集成、验收与台账  | apps/api、internal/httpapi、internal/compilation、internal/compilerworker、services/compile-worker、deploy/compiler、scripts    | MVP后先审阅完整Phase1设计，另行实现真实runner，不把编译反馈当判题                         |
| Frontend Builder      | 实际网页及API接入                    | apps/web/src/App.tsx负责布局和列表；EditorPage.tsx负责题面 /Monaco /编译状态；lib/api.ts负责DTO和HTTP错误；styles.css负责响应式 | 后续若用户扩展题库，按API契约增加页面与查询；原完整框架任务包仍保留，未实现提交历史 /账号 |
| 独立 reviewer         | 未参与实现的代码 /契约 /运行证据审阅 | docs/reviews；验证API隔离、只编译、状态 /输入边界 /超时 /清理与UI                                                               | 报告事实和未测范围；不可把自检改写为独立审阅，不自行修复新阻塞                            |
| 后续用户指定 Worker   | 小范围文档、样式、题面样例等         | 先在PROGRESS认领具体任务，按冻结DTO工作                                                                                         | Monaco包体积优化仍是已记录非阻塞任务；不得新增sandbox或更改判题契约                       |

完整前端规划另见 [框架实施指南](../frontend/IMPLEMENTATION_GUIDE.md) 与 [详细Agent任务包](../frontend/AGENT_TASKS.md)，其中真实提交 /权限 /持久化等依赖仍未完成。此MVP采用上表实际目录，不冒称完成所有P2任务。问题一事一档写docs/issues；新阻塞出现时停止，由用户处理。
