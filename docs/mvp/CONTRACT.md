# 编译MVP接口与任务契约

- 版本：compile-mvp/0.1.0；日期：2026-10-01；根Agent制定，按用户最新明确范围实施。
- 范围依据：[ADR-0004](../adr/0004-compile-only-mvp.md)；本次独立review在实现完成后记录。
- MVP-001后端 /环境 /集成：根Agent；MVP-002前端：独立Frontend Builder子Agent；MVP-003实现审阅：独立reviewer，未参与实现。实际姓名和状态写PROGRESS。

## 公开接口

`GET /api/v1/problems`：HTTP200，`{"problems":[{"id":"a-plus-b","title":"A + B","difficulty":"入门","summary":"读取两个整数并输出它们的和。"}]}`。

`GET /api/v1/problems/a-plus-b`：HTTP200，字段 `id,title,difficulty,description,input_description,output_description,notes` 均string；`samples`为 `[{"input":"1 2\n","output":"3\n"}]`；`starter_code`为完整C++模板；`language="cpp"`，`standard="c++20"`。未知ID为404JSON错误。

`POST /api/v1/compilations`：Content-Type application/json，body `{ "problem_id":"a-plus-b", "language":"cpp", "source":"..." }`。

有效编译请求HTTP200，body：

```json
{
  "status": "success",
  "diagnostics": "",
  "diagnostics_truncated": false,
  "duration_ms": 123,
  "compiler": "GCC 14.x",
  "standard": "c++20"
}
```

status仅为 `success` / `compile_error` / `timeout`；编译失败是正常业务响应200，diagnostics为真实有界输出。前端根据status呈现结果，不把HTTP200一律认为成功。

HTTP400非法JSON /字段 /语言 /空源码，404未知题目，413过大源码或body，429编译繁忙，502/503编译服务失败 /不可用；错误形状 `{"error":{"code":"...","message":"中文说明"}}`。源码UTF-8最多65536bytes、JSONbody最多262144bytes；只允许cpp。前端输入限制按UTF-8bytes，不按字符数。

`/api/v1/system`继续返回judge_enabled=false，增加compile_enabled字段；它只表示API功能配置，不替代一次真实编译。readyz保留数据库基础检查，编译MVP不依赖业务数据库。

## 内部compiler接口

compile-worker只接受 `POST /compile`，body `{"source":"..."}`；成功 /CE /timeout的DTO与上述结果一致，不接受compiler flags或用户文件名。GET /healthz报告worker与compiler是否可用。主API通过固定COMPILER_URL（默认127.0.0.1:8081）调用，不把用户给的URL传入。

g++参数固定 `-std=c++20 -O0 -fdiagnostics-color=never solution.cpp -o program`，独占临时目录；不运行program。编译timeout15秒，诊断保留最多65536bytes并标truncated；临时文件在请求完成后清理。worker一次只编译一个，繁忙立即429；主API调用timeout20秒，HTTP写timeout须覆盖它。

## 前端实施

允许apps/web内的组件、CSS、目录分层；不新增依赖。首页显示API题目卡片，点击进入题面与编辑器；`/editor`兼容跳转题面。同页展示描述、输入输出、样例、C++编辑区、编译按钮和结果；布局响应式，加载 /题目缺失 /服务错误可重试。保留Monaco本地worker配置和代码高亮，提供明确按钮与可访问状态。

源码空或超限时提示；编译期间按钮防重复，修改源码后结果注明或清空以避免旧结果误认新源码；重置模板先确认以免丢失输入。诊断原样以文本显示，不渲染HTML。产品文案不放内部Agent /契约 /实现细节。

## 检查与未实现

后端handler测试覆盖数据读取 /输入 /转发status与失败，真实compiler smoke覆盖有效C++、语法错误、无main链接错误、含无限循环的合法程序只编译不运行；浏览器验收题面和成功 /失败交互。完整check、基础API smoke、文档与独立review同步。无沙箱 /判答案 /程序执行 /认证 /持久化，不宣称是完整OJ。
