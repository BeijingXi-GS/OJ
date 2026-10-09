# 前端框架实施指南

版本：1.0 · 编写日期：2026-09-29 · 适用仓库：OJ · 设计状态：**proposed，待前端 Builder / 涉及 API 的 Architect 评审**。

本文件是交给后续 Agent 的详细实施指令，**本轮未修改前端源码、安装依赖或搭建页面**。最新任务状态只在 [PROGRESS.md](../../PROGRESS.md) 维护；具体分配见 [前端任务包](AGENT_TASKS.md)。

## 1. 交付边界与实施顺序

不要重新运行 create-vite，不要另建前端项目。沿用现有 `apps/web`，逐步把 Phase 0 演示骨架整理成能够承载 Phase 2 最小 Web OJ 的应用。

分成两个可单独验收的批次：

| 批次          | 本批交付                                                                                        | 依赖                                                               | 能否在 Docker 不可用时进行                            |
| ------------- | ----------------------------------------------------------------------------------------------- | ------------------------------------------------------------------ | ----------------------------------------------------- |
| FE-A 框架     | 分层目录、应用布局、路由、通用状态组件、请求层、类型边界、Monaco 组件、测试工具、显式 mock 模式 | 本文评审、已安装 Node/npm、接口适配边界                            | 可以；只验证页面与模拟流程，不能宣称真实提交/判题成功 |
| FE-B 业务接入 | 题目列表/详情、创建提交、提交列表/详情、状态轮询、真实错误处理                                  | Phase 2 OpenAPI 冻结、后端服务、身份/权限边界与独立 Judge 方案验收 | 真实联调不可以；遇依赖问题按 AGENTS 记录并交用户处理  |

FE-A 不等于 Phase 2 完成。Phase 1 Mini Judge 设计和核心契约评审仍是独立工作，本文件不替代它们。比赛、榜单、管理后台、交互题 transcript、通信题图编辑器、SSE、国际化和主题系统本批不实现。

## 2. 当前仓库事实

以下是 2026-09-29 读取现有文件的结果，不是建议升级版本：

| 内容              | 当前文件 / 版本                               | 实施要求                                                       |
| ----------------- | --------------------------------------------- | -------------------------------------------------------------- |
| React / ReactDOM  | 19.3.0                                        | 保留根 StrictMode，清理副作用要能承受开发期重复挂载            |
| TypeScript / Vite | 5.9.3 / 7.3.6                                 | 保留 strict，不能用 any 或关闭检查绕过错误                     |
| React Router      | react-router-dom 7.18.4                       | 沿用声明式 SPA 路由，避免同时引入另一套路由框架                |
| 服务端状态        | TanStack Query 5.104.0                        | 一个 QueryClient，查询与提交都经过 feature hooks               |
| 样式              | Tailwind 4.3.3 + @tailwindcss/vite            | 不照搬 Tailwind v3 的初始化配置                                |
| shadcn 风格基础   | components.json、components/ui/button.tsx     | 已有配置与 Button；不代表完整组件库已生成                      |
| Monaco            | monaco-editor 0.57.0、React wrapper 4.7.0     | 现有导出路径与本地 worker 已验证；按需抽取，不能改回旧深层路径 |
| 当前页面          | App.tsx：工程首页；EditorPage.tsx：C++ 预览   | 保留能力，逐步迁移；迁移期间每一步都能构建                     |
| 开发服务          | 127.0.0.1:5173 → proxy 127.0.0.1:8080         | 开发代理只用于开发；部署另配同源 /api 转发                     |
| 当前真实 API      | GET /healthz、GET /readyz、GET /api/v1/system | 仅这三项已存在                                                 |
| 测试工具          | 无前端 unit / browser test 命令               | 不得把根 npm test 的 Go 测试当成前端测试                       |

当前文件入口：

- `apps/web/src/main.tsx`：React 根、QueryClientProvider、BrowserRouter。
- `apps/web/src/App.tsx`：布局、工程首页、路由目前混在一起。
- `apps/web/src/EditorPage.tsx`：Monaco 配置和示例。
- `apps/web/src/components/ui/button.tsx`、`src/lib/utils.ts`：复用基础。
- `packages/proto/openapi.yaml`：真实 HTTP 契约入口，目前没有题目/提交 API。
- `apps/web/tsconfig.json`：strict 与 @/* 别名已配好。
- 根 `package-lock.json` 是 npm workspace 唯一锁文件；不要另建子目录锁文件。

## 3. 目标目录与依赖方向

以下目录是**待创建**结构；不存在的文件不能标为已经实现。

```text
apps/web/
├── public/
│   └── mockServiceWorker.js            # 仅在采用 MSW 后生成
├── src/
│   ├── main.tsx                       # 启动；先完成可选 mock 初始化，再 render
│   ├── App.tsx                        # 过渡入口，迁移后仅组合 providers/router
│   ├── app/
│   │   ├── providers.tsx              # 单一 QueryClient、Router、全局错误边界
│   │   ├── router.tsx                 # 路由声明、懒加载、fallback
│   │   ├── query-client.ts            # 默认缓存/重试策略
│   │   ├── config.ts                  # 环境参数解析；不含业务 DTO
│   │   └── layouts/
│   │       └── MainLayout.tsx         # Header、导航、Outlet、页面容器
│   ├── pages/
│   │   ├── problems/ProblemListPage.tsx
│   │   ├── problems/ProblemDetailPage.tsx
│   │   ├── submissions/SubmissionListPage.tsx
│   │   ├── submissions/SubmissionDetailPage.tsx
│   │   ├── dev/SystemStatusPage.tsx
│   │   ├── dev/EditorPreviewPage.tsx
│   │   └── NotFoundPage.tsx
│   ├── features/
│   │   ├── problems/
│   │   │   ├── api.ts                 # 该领域请求，不放 JSX
│   │   │   ├── queries.ts             # query keys + hooks
│   │   │   ├── model.ts               # UI view model / DTO adapter
│   │   │   └── components/
│   │   ├── submissions/
│   │   │   ├── api.ts
│   │   │   ├── queries.ts
│   │   │   ├── model.ts
│   │   │   └── components/            # 状态、Run 选择、结果列表
│   │   └── code-editor/
│   │       ├── monaco-setup.ts        # 单一注册位置
│   │       ├── CodeEditor.tsx
│   │       ├── SourceDraft.tsx        # 内存草稿与提交表单的组合
│   │       └── model.ts
│   ├── api/
│   │   ├── client.ts                  # fetch、超时、错误转换、signal
│   │   ├── errors.ts                  # ApiError，禁止泄露原始服务端栈
│   │   ├── generated/                # 由已审 OpenAPI 生成，禁止手改
│   │   └── contracts/                # 必需的运行时响应校验
│   ├── components/
│   │   ├── ui/                        # Button/Input/Select/Dialog 等基础组件
│   │   └── common/                    # Loading/Empty/Error/Pagination/Boundary
│   ├── lib/
│   │   ├── utils.ts                   # 保留现有 cn
│   │   ├── format.ts                  # 时间、bytes/ms 显示
│   │   └── url-state.ts               # URL 参数解析、默认值、重置规则
│   ├── mocks/
│   │   ├── browser.ts
│   │   ├── handlers.ts
│   │   ├── fixtures/                  # 标准 DTO，固定样例
│   │   └── scenarios/                 # slow/empty/error/queued/finished
│   ├── test/
│   │   ├── setup.ts
│   │   └── render.tsx                 # 每测试独立 QueryClient/MemoryRouter
│   ├── styles.css                     # 全局 tokens + Tailwind 入口
│   └── vite-env.d.ts                  # 允许的公开 VITE_* 类型
├── e2e/                               # 浏览器用户流程；不直接改业务数据库
├── components.json
├── vite.config.ts
├── vitest.config.ts                   # P2-104 才添加
└── playwright.config.ts               # P2-110 才添加
```

依赖方向：`app → pages → features → api/components/lib`。app 负责组合；page 只负责读取路由参数、组合功能、显示状态；feature 负责业务请求和转换；通用 UI 不导入 pages/features；api 层不导入 React。

禁止：

- 页面中散落 fetch、写死后端地址、重复定义多个 QueryClient。
- 把题目、提交 DTO、Monaco、路由和登录状态堆进一个 App.tsx。
- 不同 feature 相互导入内部组件形成循环；公共显示组件升到 common，公共数据契约来自 api。
- 为“以后可能用”建立大量空目录；按当前任务实际产物建立。
- 同时修改 Router 的定义方式、技术栈和业务契约，造成无法审阅的大提交。

## 4. 路由与布局

沿用 BrowserRouter + Routes/Route；全应用只能有一个 BrowserRouter。嵌套布局用 Outlet，页面用 React.lazy + Suspense。每个懒加载边界同时有 loading 和真正的错误边界；Suspense 本身不处理渲染异常。

| URL                        | 页面              | 行为                                                     | 阶段               |
| -------------------------- | ----------------- | -------------------------------------------------------- | ------------------ |
| /                          | 入口              | redirect 到 /problems；FE-A 未完成题目页时可显示明确占位 | FE-A               |
| /problems                  | 题目列表          | 搜索、分页、加载/空/失败状态                             | FE-B；FE-A 可 mock |
| /problems/:problemId       | 题目详情 + 提交区 | 固定题目版本、展示题面、编辑代码                         | FE-B；FE-A 可 mock |
| /submissions               | 我的提交          | URL 记录筛选和分页，不能显示未授权源码                   | FE-B               |
| /submissions/:submissionId | 提交详情          | SourceUnit 快照、JudgeRun 历史及选定 Run                 | FE-B               |
| /dev/status                | 现有工程状态页    | 保留真实 /api/v1/system 基础检查                         | 仅开发模式         |
| /dev/editor                | 现有编辑器预览    | 不保存、不发真实提交                                     | 仅开发模式         |
| *                          | NotFound          | 有明确返回入口                                           | FE-A               |

当前 /editor 在开发模式兼容重定向至 /dev/editor；生产环境不暴露工程预览路由。生产首页和业务页不展示 npm 命令、内部任务 ID 或契约文件路径。

MainLayout 保持语义化 header/nav/main；当前导航 aria-current 可识别，键盘焦点可见。375px 视口下，双栏题面/编辑区改上下布局；表格局部滚动，不让整页横向溢出。开发桌面目标 1280px，同样支持键盘操作。浏览器刷新子路由依赖部署端 SPA fallback，不能把部署配置问题伪装成路由已验收。

## 5. 状态放在哪里

| 状态                               | 所有者                                    | 规则                                             |
| ---------------------------------- | ----------------------------------------- | ------------------------------------------------ |
| 题目、语言、提交、Run 等服务端数据 | TanStack Query                            | 不复制进另一个全局 store                         |
| 页码、筛选、搜索、选定 runId       | URL search params                         | 可刷新/分享；过滤变化时 page 回到 1              |
| 编辑器正文、表单输入、当前选项     | feature 内组件 state / 必要的局部 context | 输入流畅，不每个按键触发全局查询                 |
| 弹窗、菜单、tab                    | 最近使用者                                | 只在确有跨页需求时提升                           |
| 已登录用户及权限                   | 已审 session API + Query                  | FE-A 不伪造用户，不用 localStorage role 冒充权限 |
| 环境/mock 模式                     | app/config                                | 启动时决定，不在网络失败后自动切换               |

暂不新增 Redux/Zustand。已有库足够支撑本批框架，未来确有跨页面客户端状态再评估。

Query keys 由各 feature 工厂生成，例如：`['problems','list',{q,page,pageSize}]`、`['problems','detail',problemId,versionId]`、`['submissions','detail',submissionId]`、`['judge-runs','detail',runId]`。参数需标准化后进入 key；用 ID 而非标题。多账户能力接入时 key 隔离或切换账户彻底移除私有 cache。

建议初始策略（是项目选择，不是库的默认保证）：

- 题目列表 staleTime 30s；版本化详情 60s；语言配置 5min。
- 网络错误 / 5xx 的 GET 最多重试 1 次；401/403/404/409/422/429 不自动重试。
- 创建提交 mutation 的 retry=false。禁用按钮只是防误触，服务端仍须处理重复请求。
- Run 非终态每 2s 拉取；停止于 FINISHED/CANCELLED。后台页面暂停轮询，重新聚焦时再拉取。
- 轮询连续 3 次请求失败后停止自动轮询，显示“状态同步中断 / 重试”；不能覆盖已知 verdict。用户重试成功后清零失败计数再恢复。
- 当前 Run 以 URL runId 为准，重判产生新 Run 时不偷偷覆盖用户正在看的历史 Run；显示“有新的判题记录”供切换。
- Query 提供的 AbortSignal 传给 fetch，路由切换取消无用读请求。取消浏览器请求不等于服务端取消判题。

库取消机制参考 [TanStack Query Cancellation](https://tanstack.com/query/latest/docs/framework/react/guides/query-cancellation)；缓存、重试与后台刷新需要显式配置，参考 [Important Defaults](https://tanstack.com/query/latest/docs/framework/react/guides/important-defaults)。

## 6. HTTP、类型和错误边界

只允许 feature api 调用统一 `api/client.ts`。默认相对地址 /api/v1，现有 healthz/readyz 用独立明确路径；本地通过 Vite proxy，同源部署通过反向代理。浏览器不能直连 PostgreSQL、Redis、MinIO 管理端或 Judge Worker。

请求层统一处理：

1. path 参数编码，query 参数用 URLSearchParams；不拼接未经验证的 URL。
2. 10s 的读取超时；组合 timeout 和传入 AbortSignal，必须释放监听器/计时器。
3. 先检查 status 与 Content-Type，再解析 JSON；HTML 错误页、204、空 body 都不能导致未捕获异常。
4. 返回 `ApiError { kind, status?, code?, safeMessage, requestId?, fieldErrors? }`。不直接把原始 HTML、SQL 错误、stack 或响应对象展示给用户。
5. 网络失败、超时、取消、业务拒绝分开；取消不弹“系统错误”。
6. 401：按已审登录方案提示/跳转；403：权限不足；404：对象不存在或无权访问；409：版本/重复冲突；422：字段校验；429：限流提示及 retry-after 信息；5xx：可重试提示和可复制 requestId。
7. 不记录正文源码、密码、token、隐藏 testcase、授权 header；requestId 可用于排查。
8. 认证方案未冻结前，不能添加假的 bearer token、硬编码 userId 或管理员身份。后续跨域、cookie、CSRF 由 Architect 决定并写 ADR。

类型来源：

- 已实现接口以 `packages/proto/openapi.yaml` 为准。采用 openapi-typescript 生成 TypeScript 类型时，由 P2-100/102 确认兼容版本后精确锁定；本轮不安装。
- 生成文件进入 api/generated 并注明命令、来源版本；不可手改。
- TypeScript 断言不是运行时验证。P2-102 为关键响应建立边界校验，可采用经审阅的轻量 schema 库；不能靠 `as DTO` 将任意 JSON 当作正确对象。
- 暂无业务 OpenAPI 的 mock 数据使用明确命名的 proposed DTO，由同一适配层转换；不能另立一套看起来已冻结的生产协议。
- 页面只用 UI view model，wire snake_case → camelCase 的转换集中在 model.ts；不复制后端状态机。

环境变量只允许公开配置，例如 `VITE_API_BASE_URL=/api/v1`、`VITE_ENABLE_MOCKS=true`；不把数据库连接串或 S3 secret 放入 VITE_*。Vite 会把此类变量暴露给客户端，参见 [Vite 环境变量](https://vite.dev/guide/env-and-mode)。

## 7. 业务数据与待冻结 API 清单

以下业务 URL/字段是 **Phase 2 契约提案，不是已存在接口，也未获冻结**。前端任务不得据此直接编造后端实现。P2-100 需与 Architect/Backend Builder 对齐并修改真正 OpenAPI，然后才接通真实模式。

| 提案接口                                        | 用途               | 最小响应/请求约束                                                                                   |
| ----------------------------------------------- | ------------------ | --------------------------------------------------------------------------------------------------- |
| GET /api/v1/problems?q=&page=&page_size=        | 题目列表           | items、page、page_size、total；稳定排序 created_at + id，page_size 上限由服务端定义                 |
| GET /api/v1/problems/{problemId}?version_id=    | 固定版本题面       | problem_id、problem_version_id、title、statement、samples、limits、allowed_roles、allowed_languages |
| GET /api/v1/languages                           | 可选语言与版本     | language_id、display_name、compiler/runtime 版本、可用性；不靠前端常量冒充后端支持                  |
| POST /api/v1/submissions                        | 创建不可变提交     | problem_version_id、sources[]；返回 submission_id、initial_run_id；认证、幂等与错误契约先冻结       |
| GET /api/v1/submissions?page=&page_size=&state= | 我的提交           | 明确权限、分页和 current_run 摘要；不是全站任意源码查询                                             |
| GET /api/v1/submissions/{submissionId}          | 提交与源码快照     | id、problem_version_id、created_at、可查看的 sources[]、current_run_id                              |
| GET /api/v1/submissions/{submissionId}/runs     | 重判记录           | Run ID、版本、状态、verdict、创建时间；排序和分页必须明确                                           |
| GET /api/v1/judge-runs/{runId}                  | 指定 Run 详情/轮询 | state、verdict、版本、公开诊断、case 摘要、资源计量；隐藏测试内容不返回                             |

请求提案形状：

```json
{
  "problem_version_id": "pv_example",
  "sources": [
    {
      "role": "main",
      "language_id": "cpp23",
      "compiler_version_id": "compiler_example",
      "runtime_version_id": "runtime_example",
      "source": "#include <iostream>\nint main() { return 0; }\n"
    }
  ]
}
```

此示例中 ID 为占位值；字段是否由用户选择或服务端解析由 P2-100 决定。首批 UI 只显示单 Role main / C++，内部仍保留 sources[]。不能把 SourceUnit 拆成互不相关的 Submission，也不能把源码放进 URL。哈希、用户归属、Run ID、资源限制和 verdict 由服务端确定，前端不作为信任来源。

契约评审必须明确这些问题后才打开真实提交：

- 认证凭据、所有者/权限以及 401/403 行为。
- 创建成功返回 201 还是 202、是否有 current_run/initial_run、状态版本字段。
- POST 幂等标识（请求 header 或字段）、有效期、同标识不同内容的 409 规则，以及结果未知时的查询恢复方式。
- 源码最大 UTF-8 字节数、行结束符是否原样保存、Role/语言有效集、题目版本过期提示。
- error envelope、request_id、field_errors；网络超时是否已创建资源的确认路径。
- 题目、提交、Run 列表的分页协议及稳定排序。
- 所有 UTC 时间用带时区的 ISO 8601 字符串，ID 按字符串处理；计量单位 ms/bytes。
- 无计量数据为 null/缺失，不把 unknown 填成 0；聚合内存由后端返回，前端不相加 per-role peak。

提交超时或连接断开时显示“提交结果尚未确认”，保留源码，不自动重新 POST。冻结的幂等查询可用后先查询原请求；只有确定未创建或用户明确重新提交才进行下一次请求。这是 P2-100 的阻塞验收点，不能靠禁用按钮解决。

## 8. Monaco 与源码提交区

P2-106 将现有 EditorPage 的配置抽到 code-editor feature：

- 懒加载 editor route / CodeEditor，不让题目列表首页承担 Monaco 包体。
- monaco-setup 只配置一次本地 worker 与 C++ 语言，保留当前公开导出路径；不可恢复 CDN loader。
- CodeEditor 用明确 props：value、onChange、languageId、readOnly、ariaLabel、height；不直接 fetch 或创建提交。
- source 按原始内容保留；空白校验可用 trim 判断是否为空，但不能 trim 后再提交或静默改换行。
- 源码长度按 UTF-8 bytes 校验，不能把 JS string.length 当字节数。界面显示限制，服务端再次校验。
- 由上层 SourceDraft 保存输入与选项；初版草稿只在内存，未保存离开页面有提醒。页面明确告知刷新会丢失草稿。
- 不默认将源码写入 localStorage、URL、日志或分析服务；需要持久草稿时单独任务评审。
- 组件卸载时由唯一所有者释放 model/listener，wrapper 自动释放的对象不重复释放；切题不能遗留上题代码到新题。
- model URI 按 problemVersion + role + draft ID 区分，不能依据数组下标绑定源码。
- 编辑器加载失败保留 value，显示重试入口，可降级为同一 value 控制的 textarea；降级不改变提交内容。
- Tab 在编辑器内输入缩进，同时提供可访问的退出编辑器焦点方式；按钮有文本标签。
- 不将 Compile 失败自动当作语法高亮结果；编译诊断只来自选定 JudgeRun，按纯文本显示并截断到契约上限。

提交按钮的启用条件：真实模式可用 + 已确认身份 + 题目版本/语言/Role 配置有效 + 非空/未超限源码 + 无进行中的创建请求。mock 模式按钮必须显示“模拟提交”，且不能向真实创建接口发送请求。

## 9. 页面行为和通用组件

| 页面 / 组件           | 实现要求                                                             | 必须覆盖的失败或边界                                                          |
| --------------------- | -------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| ProblemList           | 搜索 q、page/page_size 在 URL，输入 debounce 300ms，筛选变化重置页码 | 首次 loading、空列表、请求失败、页码非法、快速切换查询、旧请求晚返回          |
| ProblemDetail         | 确定 problem_version_id 后才加载对应提交配置，题面和编辑器可分别失败 | 404/403、版本不可用、语言为空、样例很长、题面未加载时禁提交                   |
| SubmissionList        | 展示本用户提交与当前 Run 摘要，时间/语言/状态可读                    | 空历史、分页返回、未知 state、接口无权、系统失败                              |
| SubmissionDetail      | 分开显示 Submission 不可变信息与选定 JudgeRun，列出 Run 历史         | rejudge 新 Run 出现、所选历史 Run 不存在、SourceUnit 不可查看、后端返回不一致 |
| RunStatusBadge        | 文本 + 图形/样式，不能只靠红绿                                       | 所有 lifecycle/terminal verdict、未知值统一“未知状态”                         |
| ResourceSummary       | CPU/wall 用 ms，内存 MiB（bytes / 1024²），保留原始值                | null 显示“未提供”，0 只在后端明确提供 0 时展示                                |
| LoadingState          | 用于首次无数据加载，必要区域 aria-busy                               | 后台刷新保留旧内容，加轻量标识，不整页闪烁                                    |
| EmptyState            | 区分“无题目”和“筛选无结果”                                           | 提供清空筛选/返回入口                                                         |
| ErrorState            | 安全提示、重试按钮、requestId                                        | 重试中禁重复点击；401/403 不无限重试                                          |
| Pagination            | 受控页码和条数，操作同步 URL                                         | 前后端页码边界、筛选后 page=1，不能依赖数组长度猜 total                       |
| ErrorBoundary         | 页面渲染失败时显示恢复入口                                           | 保留可恢复的草稿所有者；不能把错误详情直接渲染                                |
| CodeBlock / Statement | 默认纯文本；若使用 Markdown，原始 HTML 禁用并验证危险链接            | script/onerror/javascript:、超长内容、代码块溢出                              |

前端显示生命周期与 verdict 分离：PENDING/QUEUED/COMPILING/JUDGING 表示处理中；FINISHED 必须查看 verdict；CANCELLED 显示已取消。未知枚举不能映射成 AC。任何从服务端来的非法组合进入“数据暂不可用/待刷新”，记录不含敏感数据的诊断，不私自改后端状态。

Tailwind 负责布局与通用样式；全局 styles.css 只定义主题变量、基础排版和第三方必要覆盖。不要继续用全局 h1/nav/section 规则给所有业务组件隐式上样式。保留现有 components/ui 路径，新增 shadcn 组件前检查是否已有；优先少量 Button、Input、Select、Label、Alert、Tabs、Dialog，不一次引入大批未使用组件。

## 10. Mock 与真实服务严格区分

采用 MSW 作为推荐方案，由 P2-104 精确锁定测试/模拟依赖。启动模式必须显式：

- 默认 `VITE_ENABLE_MOCKS=false`；只在 `import.meta.env.DEV && VITE_ENABLE_MOCKS==='true'` 时初始化 browser worker。
- 在 render 之前 await worker.start，避免首个请求先漏到真实服务。
- 画面顶部固定“模拟数据”标识，模拟创建提示不代表真实判题。
- mock 覆盖需要的 /api/v1/problem/submission/run 相关请求；未覆盖的 **OJ API** 请求报开发错误并阻止漏网请求，Vite 资源请求正常放行。
- 真实模式后端不可用就显示失败，不能自动切到 mock；测试不能因此假通过。
- production 构建禁止启用 mock；worker 代码使用条件动态 import。交付时验证生产页面无 mock 标识/初始化请求，也没有激活的 MSW worker 控制 API；对已有开发 worker 需按 MSW 官方生命周期处理。
- mock 保存在开发环境本地内存，刷新重置；不用真实数据库记录假提交，不修改业务状态机。
- 固定 fixture 构造可复现的 loading/empty/error/queued/compiling/finished/cancelled，延迟由场景参数确定，不使用随机失败。
- 协议冻结后，handlers 与生成 DTO 对齐，测试保证必填字段和枚举一致；不能分别维护两套 error 模型。

浏览器 worker 初始化方式参见 [MSW Browser Integration](https://mswjs.io/docs/integrations/browser/)。Mock 框架本轮未安装。

## 11. 开发和未来验证命令

当前已经存在的命令（仓库根目录执行）：

```sh
npm run dev
npm run typecheck
npm run lint --workspace @oj/web
npm run build --workspace @oj/web
npm run format:check
```

这些命令本轮没有执行前端构建测试，因为本轮只交付 MD；历史构建结果不能代表后续新代码通过。

P2-104 / P2-110 完成后，才新增并使用下列命令；现在不要假定它们已存在：

| 待新增 script（@oj/web） | 内容                                         | 谁负责 |
| ------------------------ | -------------------------------------------- | ------ |
| test:unit                | Vitest run + jsdom + React Testing Library   | P2-104 |
| test:unit:watch          | Vitest watch                                 | P2-104 |
| dev:mock                 | 用单一 Vite mock mode 启动，仍绑定 127.0.0.1 | P2-104 |
| generate:api             | 从仓库 OpenAPI 生成类型，版本固定            | P2-102 |
| check:api                | 重新生成后检查无漂移                         | P2-102 |
| test:e2e                 | Playwright；webServer 复用/启停规则明确      | P2-110 |

新增依赖前读取对应官方文档，检查当前 Node/React/Vite 的 peer compatibility，使用精确版本和根 package-lock。不要把“升级所有依赖”当作框架任务。验证工具为开发依赖；生产请求层新增 schema 库须说明包体和使用范围。

最终工程检查仍使用 `npm run check`；根 npm test 当前只运行 Go，前端 test:unit/test:e2e 必须在相应任务中接入 CI，不能仅在 README 声称已有。

## 12. 验收矩阵

| 验收 ID | 操作                              | 通过条件                                      | 批次                     |
| ------- | --------------------------------- | --------------------------------------------- | ------------------------ |
| FE-01   | 打开 /problems 并浏览器刷新       | 页面正常，路由无双 Router、无整页跳转         | A/mock，B/real           |
| FE-02   | 导航未知路径、键盘遍历导航        | NotFound 有出口，当前导航可读、焦点可见       | A                        |
| FE-03   | 后端不可连接且关闭 mock           | 明确连接错误，无模拟成功结果                  | A                        |
| FE-04   | 开启 mock，再 production build    | 开发有标識；生产无法初始化 mock               | A                        |
| FE-05   | 搜索后快速改条件/翻页             | URL 与页面一致，无旧响应覆盖新条件            | B/mock+real              |
| FE-06   | 切换题目、语言并编辑非 ASCII 源码 | 正确绑定版本与 Role，UTF-8 上限正确           | A/component，B/real      |
| FE-07   | 编辑器加载失败                    | 原输入仍可用；重试/textarea 正常              | A                        |
| FE-08   | 双击提交、慢响应、网络断开        | 不并发 POST；未知结果不自动重提，源码保留     | B                        |
| FE-09   | QUEUED→COMPILING→JUDGING→FINISHED | 状态正确；终态停止轮询                        | B                        |
| FE-10   | 连续轮询失败、隐藏标签页再回来    | 停止/恢复规则符合第 5 节；旧结果不消失        | B                        |
| FE-11   | 重判产生 Run2，同时查看 Run1      | 历史 Run1 保持；显式切换才显示 Run2           | B，rejudge mock 可先验证 |
| FE-12   | 401/403/404/409/422/429/5xx       | 分别处理，不无限重试，不泄漏原始错误          | B                        |
| FE-13   | 恶意题面 HTML/链接、超长编译日志  | 内容不执行，合理换行/截断                     | A/component，B/real      |
| FE-14   | 375px 和 1280px                   | 内容可用、无整体横溢、触控可操作              | A+B                      |
| FE-15   | 路由反复进出编辑器                | model/listener 不持续增长；无未处理 rejection | A                        |
| FE-16   | 未提供资源值和未知 verdict        | 显示未提供/未知，不擅自显示 0 或 AC           | A+B                      |
| FE-17   | 新增测试脚本、类型生成和 CI       | 命令真实可执行，结果绑定 commit / 工作区摘要  | A+B                      |

实现型任务必须提交实际命令、浏览器证据与版本。不得用“页面看起来完成”替代 FE-B 的真实 API 证据，也不能把 mock 的 AC 当作 Judge 通过。

## 13. 交付和回退

每个任务提交范围明确的 diff、使用说明、适用检查结果和未完成项。遵守 [AGENTS.md](../../AGENTS.md) 的问题接管规则：出现问题单开 MD，由用户处理；涉及该任务的前置阻塞就停止，不自行用假数据掩盖。

合入顺序：依赖边界与框架 → client/类型 → mock/测试基础 → feature hooks → editor/页面 → 集成/E2E。不做全量 rewrite。回退按任务提交撤销，不能覆盖用户其他未提交修改；依赖回退同时恢复根锁文件。前端框架任务不执行数据库迁移，不删除 Docker 卷。

技术依据另见 [React Router 声明式路由](https://reactrouter.com/start/declarative/routing)、[shadcn Vite 配置](https://ui.shadcn.com/docs/installation/vite)。以上资料用于核实框架能力；本文的目录、任务顺序、时间阈值和验收项是针对本项目的设计选择。
