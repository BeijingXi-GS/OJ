# 三模式与教师题目上传使用说明

本轮增加学生、教师、工程师三种模式与真实交互题。仅供本机演示，在页面顶栏手动切换，无账号登录。任何本机使用者都能选择教师或工程师，因此不能用这种模式选择保护真实课堂的隐藏数据。实际验收与Agent任务以 [PROGRESS](../../PROGRESS.md) 为准。

## 页面使用

打开 http://127.0.0.1:5173。默认学生模式；题库包含原5道标准题和“交互猜数字”，上传成功后会追加新题。

| 模式   | 运行结果                                                                                        | 额外页面                |
| ------ | ----------------------------------------------------------------------------------------------- | ----------------------- |
| 学生   | 整体AC/WA/TLE/MLE/CE/RE/OLE/SYSTEM_ERROR、通过数量、本人编译诊断与总耗时；API不返回隐藏逐例数据 | 题库、题面、C++20编辑器 |
| 教师   | 每例输入、标准答案、实际stdout/stderr、判定与时间；交互题显示选手和交互器两个方向的文本         | 上传题目与数据          |
| 工程师 | 教师详情、运行ID/源码与数据/产物hash、镜像、退出码/OOM、交互器编译与监督终止事实                | 上传题目、工程调试      |

切换模式会取消在途请求、清空结果并按模式隔离查询缓存；同题未提交源码仍保留。切题按题目重建编辑区。工程调试只读取配置与judge metadata，不提交代码，显示“已配置”不能当作真实编译或运行通过。CPU累计和峰值内存仍未采集，工程师看到null。

“编译”只做C++编译/链接；“运行测试”走独立受限judge。在交互题中，运行测试会真实启动选手和固定交互器、转发stdio并判定，不是把秘密.in作为选手stdin。

## 上传普通题

1. 切教师或工程师模式，进入“上传题目”。填写标题、摘要、难度、题面、输入输出说明、公开样例和C++起始源码。备注可空；公开样例1–4组。题面按纯文本显示，不执行脚本或HTML。
2. 选择标准题，checker可选tokens-v1（按空白分token，原始token严格相等）或exact-v1（全部字节严格相等，含换行）。设置墙钟500–5000ms、内存32–256MiB，默认2000ms/64MiB。
3. 上传同名配对文件，如 `sample.in`/`sample.out`、`negative.in`/`negative.out`。页面自动生成case绑定，文件名仅限安全basename，不能带目录。最多8组/16份，每份≤64KiB，整个multipart请求≤1MiB。
4. 点击发布。服务器验证全部描述和数据后生成upload-随机ID、完整写入临时目录，再原子发布。成功跳转新题，立即可判，API/worker重启后仍保留。不会覆盖内置题或已发布题。错误响应不会开放半题。

服务器上传根默认 `.local/teacher-problems`，API和worker的 `PROBLEM_UPLOAD_ROOT` 必须一致。目录中有严格problem.json注册文件和manifest.json；随意拷贝目录不会自动注册。上传数据作为题库持久保留，提交源码和运行结果尚不持久保存。

## 上传和编写交互题

本轮支持固定 `guess-number-v1` 猜数协议。教师选择交互题，填写与该协议一致的题面，上传1–8份.in文件；每份仅含一个1–1000整数（允许末尾空白），作为该例的秘密数。无需.out，也不上传交互器源码。checker固定interaction-v1，默认5000ms/64MiB。

协议步骤：

1. 选手从stdin读取交互器初始输出1000。
2. 选手输出 `? x` 并立即flush。交互器回复 `<` / `>` / `=`，表示秘密数小于 / 大于 / 等于x。x必须在1–1000，最多10次询问。
3. 选手输出 `! x` 并flush，然后正常退出。猜对且双方正常结束为AC；错误最终答案、非法格式、越界、超过询问次数或最终答案前EOF为WA。

例如秘密数500时，选手输出 `? 500`，读到 `=`，再输出 `! 500`。C++可使用 `std::endl` 或显式 `std::flush`；只写换行且关闭cin/cout自动flush可能造成TLE。选手不能读取隐藏.in；秘密只挂给独立交互器容器。页面两个方向是分别保存的日志，不承诺全局先后顺序。

选手自然异常退出为RE（137本身不等于MLE）；真OOM为MLE；双方协议/诊断共享输出预算1MiB，选手超限为OLE、交互器超限为SYSTEM_ERROR。固定交互器异常、通信不完整或清理失败也为SYSTEM_ERROR。交互器合法拒绝后可能监督终止仍等待的选手，此时保留真实退出码与participant_stopped证据并判WA。AC必须双方自然exit0和完整通信/清理。

## API与开发检查

API请求头 `X-OJ-Mode: student|teacher|engineer`，缺省student，未知或重复值400。POST `/api/v1/problems` 要教师/工程师，multipart唯一文本字段descriptor为严格JSON，其余文件字段均files。服务器不接受客户端ID、路径、镜像、命令、任意交互器或额外字段。浏览器跨源上传拒绝，不开放CORS；无Origin的本机CLI可用。该限制同样不是身份认证。

```sh
npm run check
node apps/web/scripts/verify-modes.mjs
npm run judge:test:linux
npm run judge:test
node scripts/problem-smoke.mjs
node scripts/interactive-test.mjs
node scripts/interactive-lifecycle-smoke.mjs
node scripts/teacher-problem-smoke.mjs
```

真实脚本必须在本机已启动、没有其他提交的测试实例顺序执行。teacher-problem-smoke会持久保留两道演示上传题并记录ID供重启复验。judge:test会注入文件故障再finally还原，期间不要手动编辑该文件。检查不能代替实际浏览器和独立审阅。

## 本轮Agent交付与后续职责

| Agent                           | 本轮修改范围与交付                                                                            | 后续具体工作                                                                                                         |
| ------------------------------- | --------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| root Backend Builder            | httpapi模式/投影/上传/debug、API初始化、Client新wire/题型和sourcehash绑定、真实回归/浏览器/MD | 维护总体验收、启动与安全退出、动态题库配置、未知阻塞档案；将真实登录或异步提交另立受审契约                           |
| teacher_catalog Builder         | problemstore、problems、loader、猜数题面与3例、上传解析/原子发布/持久发现与专项测试           | 扩题先定义题型/限制/题面和独立答案oracle；受审后支持编辑/删除或版本化，不直接覆盖在途快照                            |
| interactive_runner Builder      | 独立双方编译/运行、固定交互器、双向stdio/输出预算/终态与清理、真实启动probe                   | 每次改协议都更新固定源码版本/hash和交互矩阵；任意交互器上传、更多语言和资源计量需要新设计                            |
| modes_frontend Builder          | apps/web三模式、query与请求隔离、三层反馈、上传表单、debug页、响应式和可信mock检查            | 在React/TS现有框架下做搜索/筛选/分页、表单草稿、提交记录；先定API DTO，再接query/mutation、取消、空态/错误与可访问性 |
| roles_review Architect/Security | 只读设计/源码/可信fake/真实证据，独立结论写review                                             | 复核隔离/协议/上传原子/题型绑定及学生投影；源码/镜像/实际证据绑定齐后才登记产品PASS                                  |

完整Phase1异步持久化、真实身份权限、任意交互器和公网生产部署未纳入本次本机增量。新的未知外部阻塞仍单独写MD、交用户处理并停止。

前端后续扩展应沿用现有React/TypeScript/Vite、React Router、TanStack Query和Monaco框架。具体目录分层、provider/request/DTO、页面空态和错误态、取消/缓存/可访问性及逐卡验收步骤见[实施指南](../frontend/IMPLEMENTATION_GUIDE.md)和[详细任务包](../frontend/AGENT_TASKS.md)；先核对本轮已完成页面，避免重复搭建骨架或把旧planned任务当成当前未实现功能。
