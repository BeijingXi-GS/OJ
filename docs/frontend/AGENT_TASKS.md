# 前端 Agent 任务包

日期：2026-09-29 · 输入：[前端框架实施指南 v1.0](IMPLEMENTATION_GUIDE.md) · 本文状态：**任务说明已编写，实施未开始**。

本轮仅写 MD，没有实际调度其他 Agent。所有实际执行者、认领日期和状态只在 [PROGRESS.md](../../PROGRESS.md) 填写。以下是任务卡集合，每个任务有唯一编号和锚点，领取时以本节为输入，可另建实施交接但不复制一份“最新状态”。

## 1. 分配与协作顺序

| 顺序 | 任务                              | 负责角色                           | 输入依赖                                          | 输出 / Review                                                |
| ---- | --------------------------------- | ---------------------------------- | ------------------------------------------------- | ------------------------------------------------------------ |
| 0    | P2-100 方案与 API 边界评审        | Architect + Frontend Builder       | 本文、现有 OpenAPI、领域模型                      | 确定 FE-A 可实施范围；真实 API 另行冻结 / Architect          |
| 1    | P2-101 目录、provider、路由与布局 | Frontend Builder                   | P2-100 的 FE-A 结论                               | 可运行应用外壳 / Builder reviewer                            |
| 2    | P2-102 请求层、类型及配置         | Frontend Builder                   | P2-101；现有健康 API 契约                         | client、ApiError、类型生成路径 / Builder；权限语义 Architect |
| 2    | P2-103 UI 基础和状态组件          | Worker-UI                          | P2-101 的公共路径与布局契约                       | common/ui 组件 / Builder                                     |
| 3    | P2-104 mock 与测试基础            | Worker-Tests                       | P2-102 请求层；P2-103 可独立随后使用              | MSW、Vitest/RTL、固定 fixtures / Builder                     |
| 4    | P2-105 业务 hooks 与模型          | Frontend Builder                   | P2-102/104；业务 DTO 提案已审                     | Query keys、view models、轮询 / Builder                      |
| 4    | P2-106 编辑器组件与源码草稿       | Worker-Editor                      | P2-101/103/104                                    | CodeEditor、SourceDraft / Builder                            |
| 5    | P2-107 题目列表/详情页面          | Worker-UI                          | P2-103/105/106                                    | problems pages / Builder                                     |
| 5    | P2-108 提交页面与 Run 视图        | Worker-UI                          | P2-103/105                                        | submissions pages / Builder                                  |
| 6    | P2-109 真实提交与 API 集成        | Frontend Builder + Backend Builder | P2-100 真实接口冻结、P2-105/106/107/108、可用后端 | FE-B 核心闭环 / Architect 审权限及幂等、Builder 审实现       |
| 7    | P2-110 浏览器回归与 CI            | Worker-Tests                       | P2-104、可用页面、P2-109 真实环境                 | E2E、CI、证据 / Builder                                      |
| 8    | P2-111 集成验收与交接             | Frontend Builder                   | 全部适用上游任务                                  | FE-A/FE-B 分开的验收结论 / 项目负责人                        |

路径隔离是后续允许协作时的执行建议，**不代表本轮自动派发并行 Agent**。P2-102 和 P2-103 可在 P2-101 后分别处理请求层与 UI；共享根 package.json、package-lock、router 和 CI 时必须由指定集成人串行改动。

FE-A 不等待真实 Docker，但必须通过方案审阅后才能由后续 Agent 实施。FE-B 遇数据库、认证、业务 API 等前置失败时停止该工作并写问题 MD；不自动降级到 mock 交付。

## 2. 各任务详细指令

### P2-100：前端方案与接口边界评审

- 角色：Architect 审接口/权限/数据一致性，Frontend Builder 审框架可实现性。
- Owner：待认领；Risk：R2（涉及身份、重复提交和源码）。
- 允许：本指南、评审记录、必要 ADR；业务 OpenAPI 只能由有范围授权的 Architect/Backend Builder 在本任务中修订。
- 禁止：直接开始整个前端、顺带重写后端、把提案 URL 标为已实现。
- 任务：
  1. 确认保留 React/Vite/Router/Query/Tailwind/shadcn/Monaco 和单一 workspace。
  2. 审阅 GUIDE 第 3–6 节，批准 FE-A 路径、依赖、mock 边界。
  3. 逐项答复第 7 节的身份、权限、分页、错误模型、源版本及 POST 幂等/未知结果恢复。
  4. 分开记录“FE-A 范围可实施”与“FE-B API 已冻结”，不要把两个结论混为一体。
  5. 指定共享文件集成人和可执行版本基线。
- 验收：独立 reviewer、日期、结论、允许的批次、未解决问题有记录；业务契约没冻结时 FE-B 不 ready。
- 验证：人工逐项对照 OpenAPI 和领域模型；不需要为纯审阅新造测试。
- 回退：撤回尚未实现的提案并保留 ADR 历史，不回滚别人的代码。
- 提示词：`你负责 P2-100。先只评审前端方案与 API 边界，输出可实施范围和待冻结清单；不得代写整个前端。按 GUIDE 第 7 节逐项确认幂等、身份、权限和版本语义。`

### P2-101：框架分层、应用入口与路由

- 角色：Frontend Builder；Risk：R1；依赖：P2-100 FE-A。
- 允许：src/main.tsx、App.tsx、app/**、pages/dev/**、pages/NotFoundPage.tsx；页面注册骨架与导航。
- 禁止：业务 fetch、登录/提交接口、修改 OpenAPI、重装 React/Vite。
- 步骤：
  1. 从 App 中提取 MainLayout 和 SystemStatusPage；将 EditorPage 搬到 dev/EditorPreviewPage，先保持原行为。
  2. 集中 QueryClient 创建和 providers；保证只有一个 BrowserRouter。
  3. 建立 GUIDE 路由矩阵；业务页暂用明确的空壳组件，不能显示假成功。
  4. 开发路由按构建模式保护，保留旧 /editor 的开发跳转。
  5. 按页懒加载，添加 route loading、NotFound、ErrorBoundary 与可访问导航。
  6. 在每次迁移后检查 @ 别名与导入路径；不要一次大规模重命名。
- 验收：FE-01/02/14；根/子路由加载，未知路由能返回，375px/1280px 合格；工程页仅开发可见。
- 命令：typecheck、web lint、web build；测试基础尚未建立时提交浏览器检查证据，后续 P2-104 补核心行为测试。
- 提示词：`你负责 P2-101，仅拆分现有入口和建立路由/layout。保留 QueryClient 与 BrowserRouter 单例；不实现题目或提交 API，提供迁移清单和构建、导航验证。`

### P2-102：统一请求层、类型生成与配置

- 角色：Frontend Builder；Risk：R1，认证语义为 R2；依赖：P2-101。
- 允许：src/api/**、app/config.ts、vite-env.d.ts；根依赖/锁文件仅由集成人改。
- 禁止：硬编码 token/userId、全局 any、在组件里创建第二套请求客户端、擅自设计认证协议。
- 步骤：
  1. 实现统一 request 入口，支持 method/body/query/signal/timeout 和安全 ApiError。
  2. 仅先接通现有 system/health/readiness；readiness=503 不是“HTTP 服务没启动”。
  3. 从真实 OpenAPI 生成类型，生成器版本精确锁定；缺失业务接口只使用明确命名的 proposed DTO。
  4. 建立必要运行时校验及 model 转换，处理未知 state/缺失字段。
  5. 公开配置校验，默认同源；禁止将 secrets 放入 VITE_*。
  6. 提供 generate:api/check:api 实际 script 后再在说明中引用。
- 验收：超时、Abort、204、HTML 错误页、无效 JSON、401/403/422/429/5xx 均不出现未捕获异常；错误不暴露正文源码。
- 检查：typecheck、lint、build；P2-104 测试工具就绪后，补上述 request 边界测试。
- 提示词：`你负责 P2-102。做一个可取消、有界超时、安全错误映射的请求层；只把真实 OpenAPI 当成已存在契约。不得实现鉴权替代方案或自动重试创建提交。`

### P2-103：通用 UI 与页面状态组件

- 角色：Worker-UI；Risk：R1；依赖：P2-101。
- 允许：components/ui/**、components/common/**、styles.css、lib/format.ts；设计 token 变更与集成人同步。
- 禁止：feature DTO、业务请求、QueryClient、领域枚举重定义。
- 输出：LoadingState、EmptyState、ErrorState、Pagination、ErrorBoundary 的展示层/约定；必需的 Button/Input/Select/Label/Alert/Tabs/Dialog。
- 步骤：复用现有 Button/cn；按需补少量 shadcn 组件；用明确 props 和回调；页面整体与组件局部错误区分；替换容易影响全站的无边界全局 CSS。
- 验收：loading/error/empty 各有语义；aria 标签、disabled 和 focus 可读；对话框键盘焦点正确；空资源不显示假 0；错误文本安全。
- 检查：typecheck、web lint/build；在 P2-104 的 RTL 环境中补关键键盘/交互测试，避免只 snapshot。
- 提示词：`你负责 P2-103，只做通用组件与样式基础；组件通过 props 接收状态，不能自己请求后端。复用已有 shadcn 路径，验证窄屏和键盘操作。`

### P2-104：Mock、单元/组件测试底座

- 角色：Worker-Tests；Risk：R1；依赖：P2-102。
- 允许：mocks/**、test/**、vitest.config.ts、公开 mock worker、开发测试脚本；依赖/入口改动交集成人处理。
- 禁止：真实 API 失败时自动开启 mock、创建真实用户或数据库记录、生产 mock。
- 步骤：
  1. 精确锁定兼容 Vitest、jsdom、RTL/user-event/jest-dom、MSW；只增加实际需要的开发依赖。
  2. 测试每例创建独立 QueryClient，关闭自动重试，结束时卸载并清理 timers/cache/handlers。
  3. 通过同一 HTTP 入口拦截 OJ API；未处理的 OJ API 在 mock 模式报错，资源加载放行。
  4. 建固定成功/慢/空/错误/状态序列数据，不引入随机 fixture。
  5. 开发页面显示 mock 标记，production guard 与 worker 生命周期可验证。
  6. 添加真实可运行的 test:unit/dev:mock scripts；不得覆盖根 Go test。
- 验收：FE-03/04；离线 mock 可开发，real 模式无后端时确实报错；每次测试互不污染；生产不激活 mock worker。
- 检查：新增 test:unit、web build、format:check；保留 peer dependency 验证记录。
- 提示词：`你负责 P2-104，为现有前端配置测试和显式 MSW 模式。mock 数据必须可复现，不能在真实模式失败时偷换数据；生产禁用 mock，未覆盖的 OJ API 不得穿透到真实写接口。`

### P2-105：查询、模型与轮询策略

- 角色：Frontend Builder；Risk：R1/R2（只遵循冻结状态语义）；依赖：P2-102/104 + DTO 提案审阅。
- 允许：features/problems/{api,queries,model}、features/submissions/{api,queries,model}、对应 tests。
- 禁止：修改后端状态机、资源计量、复用一个 submissionId key 代表所有历史 Run。
- 步骤：按 GUIDE 标准化 URL/keys；将 DTO 转 view model；读请求接 signal；配置缓存和错误重试；实现 2s 非终态轮询、终态停止、3 次失败停止和聚焦恢复；submission 与 run cache 分开。
- 验收：FE-05/09/10/11/16；晚到请求不会替换不同 key 的页面；历史 Run 不受新重判污染；无内存泄漏。
- 检查：typecheck、unit（fake timer 验证轮询与停止）、lint/build。真实 hooks 联调标到 P2-109，mock通过不等于完成 FE-B。
- 提示词：`你负责 P2-105，集中实现 feature queries 和数据映射。遵守 state/verdict 分离，轮询按 runId 停止/恢复，不把新 Run 覆盖到旧 Run 的缓存。`

### P2-106：编辑器组件与源码输入

- 角色：Worker-Editor；Risk：R1；依赖：P2-101/103/104。
- 允许：features/code-editor/**、编辑器测试；dev 预览调用新组件由集成人接线。
- 禁止：真实 POST、修改 SourceUnit 身份模型、改用外部 CDN、默认持久保存源码。
- 步骤：抽取一次性 Monaco setup；受控 value/onChange；lazy load；准确 UTF-8 长度；内存草稿；换题/Role 显式绑定；卸载清理；加载失败 textarea 保留内容；离开提醒。
- 验收：FE-06/07/15；中文源码计数正确、不 trim 修改；反复挂载无 model/listener 累积；重试不丢输入。
- 检查：组件测试可 mock Monaco 边界；浏览器必须验证真实本地 worker 和 C++ 高亮，不以 mock 代替。
- 提示词：`你负责 P2-106，复用当前已验证的 Monaco 0.57 路径抽出 CodeEditor。不得发提交请求；源码保持原样并受控，所有权和清理明确，失败降级仍保留输入。`

### P2-107：题目列表与详情页面

- 角色：Worker-UI；Risk：R1；依赖：P2-103/105/106。
- 允许：pages/problems/**、features/problems/components/**、对应测试。
- 禁止：页面直接 fetch、重定义 DTO、写死 supported languages、使用未经安全处理的 raw HTML。
- 步骤：列表 URL 搜索/分页；加载空错状态；固定版本题面、样例、limits；组合 SourceDraft；未取得必要数据不显示可提交状态；真实提交动作暂通过 props/hook 边界交给 P2-109。
- 验收：FE-05/06/12/13/14；404/403、无可用语言、过长题面都可用；移动端编辑区不挤出视口。
- 检查：RTL 状态与 URL 行为、mock浏览器检查、typecheck/lint/build。
- 提示词：`你负责 P2-107，使用已提供的 hooks 和 CodeEditor 组合题目页面；接口失败必须显式呈现，不自行补接口或把 mock 成功当作真实判题。`

### P2-108：提交列表、详情和历史 Run

- 角色：Worker-UI；Risk：R1；依赖：P2-103/105。
- 允许：pages/submissions/**、features/submissions/components/**、对应测试。
- 禁止：前端计算最终 verdict、合并多个 Run 结果、展示后端不允许查看的源码/隐藏 testcase。
- 步骤：Submission 元信息与 Run 结果分栏；URL runId 切换历史；状态文本/颜色映射；公开编译诊断纯文本；单位格式统一；只显示服务端提供的 aggregate peak。
- 验收：FE-09/10/11/12/16；未知值不当 AC；missing/null不当0；Run2完成不会覆盖Run1。
- 检查：组件边界测试、mock重判浏览器路径、typecheck/lint/build。
- 提示词：`你负责 P2-108，按 submissionId 展示提交快照、按 runId 展示判题历史。不能从当前 Submission 摘要反推历史 Run，不加总角色峰值充当聚合内存。`

### P2-109：真实提交闭环与后端联调

- 角色：Frontend Builder 主责，Backend Builder 提供已审 API；Risk：R2。
- 依赖：P2-100 的 FE-B 契约已冻结；前述页面和 hooks；可用身份、数据库和独立 Judge 路径。
- 允许：submissions/api/queries、SourceDraft 的 mutation 接线、领域数据适配、集成测试；后端缺陷另立任务。
- 禁止：以 mock替代真实验收、让浏览器/API 后端直接执行选手代码、临时加入 admin header 绕过授权。
- 步骤：
  1. 对照已冻结 OpenAPI 更新类型并检查漂移。
  2. 只 POST 已固定题目版本和完整 sources[]，服务端解析/校验 owner和hash。
  3. 成功后按返回 ID 跳转详情，适当 invalidate 我的提交列表，不清空尚未确认保存的源码。
  4. 网络结果未知按幂等查询恢复；缺失恢复契约直接写问题停止，不自动重发。
  5. 验证 real 模式全过程，包含字段失败、越权、限流、网络失败、编译错误和系统错误。
- 验收：真实“打开题目→创建一次提交→看到确定 Submission/Run→终态”；双击无两个无意提交；权限失败；未知结果可恢复；所有资源有版本追溯。
- 检查：真实 API integration + unit +浏览器；记录后端 commit和契约版本。Docker/后端阻塞交用户处理。
- 提示词：`你负责 P2-109 真实集成，仅按已冻结契约实现创建与恢复。先验证身份、幂等和独立 Judge 边界；任一前置缺失写问题 MD 并停止，不用 mock交付。`

### P2-110：E2E、响应式与 CI

- 角色：Worker-Tests；Risk：R1；依赖：P2-104和页面；真实 E2E另依赖P2-109。
- 允许：apps/web/e2e/**、playwright.config.ts、测试 scripts、CI 的前端检查步骤。
- 禁止：全局关闭 flaky断言、只等固定sleep、破坏已有Go/Compose检查或删除真实数据。
- 步骤：精确锁定 Playwright；定义 webServer 启停；按可见 role/label定位；覆盖 GUIDE矩阵；分别命名 mock UI / real integration 项目；失败保存无敏感源码/凭据的截图和trace。
- 验收：375px/1280px、路由刷新、未知路径、编辑/提交/终态、网络错误；测试产物能追到commit；CI报告不混淆两种模式。
- 检查：test:unit、test:e2e、web build；项目整体npm run check。Docker不可用时真实项目标未执行，不能把“skip”写通过。
- 提示词：`你负责 P2-110，建立可复现的前端浏览器回归和CI。明确区分mock与真实环境，按用户动作验证，不操作其他项目容器或测试数据。`

### P2-111：集成验收与交接

- 角色：Frontend Builder；Risk：R1；依赖：各适用任务完成及审阅。
- 允许：集成修订、文档、PROGRESS、验收与handoff；核心问题转Architect，不扩范围补全业务。
- 验收：GUIDE矩阵逐项记录通过/失败/不适用及原因；FE-A与FE-B各自结论；接口生成无漂移；没有未登记问题；附启动、mock、测试和部署代理说明。
- 检查：实际执行适用脚本，记录源码commit/工作区hash；独立Reviewer签署，不把自己的自检写独立评审。
- 提示词：`你负责 P2-111，整合任务结果并逐项核对验收。给出FE-A和FE-B各自完成状态，留下下一任务入口；未执行/被阻塞项不得登记完成。`

## 3. 任务共享文件锁定

| 文件                                                       | 默认修改负责人                   | 其他 Agent 做法                                |
| ---------------------------------------------------------- | -------------------------------- | ---------------------------------------------- |
| 根 package.json / package-lock.json、apps/web/package.json | Frontend Builder 集成人          | 提供依赖与script需求，避免并发改锁文件         |
| main.tsx、app/providers、router                            | P2-101 Owner / 后续集成人        | 提供路由与provider接入说明                     |
| components.json、styles.css公共tokens                      | P2-103 Owner                     | 命名与冲突先在任务中协调                       |
| packages/proto/openapi.yaml                                | P2-100 Architect/Backend Builder | 前端只能消费已审版本                           |
| .github/workflows/ci.yml                                   | P2-110 Owner                     | 保留其他阶段已有检查                           |
| PROGRESS.md                                                | 当前集成人                       | 每项任务提交状态与证据，集中更新不制造重复台账 |

## 4. 每个 Agent 的统一交付要求

1. 开始前读取 AGENTS、PROGRESS、GUIDE 和本任务；记录认领/日期/范围。
2. 修改前保留已有工作区变更。现有 task/P1-001-architect-handoff 的文档变更尚未提交，不允许 reset/clean 掉。
3. 开始实现前确认当前用户确实授权该任务；**这份任务包本身不授权当前会话自动启动其他 Agent**。
4. 提交文件清单、可复现命令、实际结果、正常和失败路径、未完成项、回退方式与交接。
5. 问题写 docs/issues，一事一档。涉及前置阻塞时按用户接管规则停止，不自行“临时绕过”。
6. 本批不要承诺已完成Sandbox、Judge、Queue、认证或数据库迁移；前端框架通过不代表后端生产可用。
