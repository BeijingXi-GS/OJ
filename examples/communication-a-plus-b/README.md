# 通信 A+B 教师数据包

本包与内置 `communication-a-plus-b` 使用相同的双模板和四组输入输出。题面范围统一为 ±10¹²，通信上限 32 字节。真实 Docker 验收当前待执行；完整环境与证据要求见[通信环境指南](../../docs/judge/COMMUNICATION_ENVIRONMENT.md)。

## 文件

| 文件                               | 用途                                                       |
| ---------------------------------- | ---------------------------------------------------------- |
| [PROBLEM.md](PROBLEM.md)           | 可直接分发给学生的完整题面                                 |
| [sender.cpp](sender.cpp)           | 可复制的发送端 C++20 演示解法                              |
| [receiver.cpp](receiver.cpp)       | 可复制的接收端 C++20 演示解法                              |
| [descriptor.json](descriptor.json) | 现有教师上传接口使用的描述，不是新 schema                  |
| [data](data)                       | 仅上传其中的四对 `.in/.out`，不要上传 C++ 或 Markdown 文件 |

| 用例     | 输入                          | 答案            | 发送端模板消息长度 |
| -------- | ----------------------------- | --------------- | ------------------ |
| normal   | `1 2`                         | `3`             | 4 字节             |
| zero     | `0 0`                         | `0`             | 4 字节             |
| negative | `-7 -11`                      | `-18`           | 7 字节             |
| large    | `1000000000000 1000000000000` | `2000000000000` | 28 字节            |

最坏合法消息 `-1000000000000 -1000000000000\n` 为 30 字节，仍在 32 字节内。表中的长度包含行末一个 LF 换行。四组数据是公开教学示例；正式测验应另行编写数据，不能把这个公开仓库当作隐藏答案库。

## 教师页面上传

1. 按环境指南正常启动服务，在页面顶部切换为“教师”，进入 [上传题目](http://127.0.0.1:5173/upload)。这是本机手动模式切换，没有登录鉴权隔离。
2. 选择“通信题”。从 [descriptor.json](descriptor.json) 复制标题、简介、难度、题目描述、输入输出要求、提示和公开样例；也可依据 [PROBLEM.md](PROBLEM.md) 填写。设置时间 5000 ms、内存 64 MiB、通信上限 32 字节。
3. 分别将 [sender.cpp](sender.cpp) 和 [receiver.cpp](receiver.cpp) 复制到对应模板栏。页面内置“一键填入通信 A+B”示例可能采用 ±10⁹ 与 `2 3` 样例；使用本包时，应将题面与样例改为本包 ±10¹² 与 `1 2`，尤其不能把 large 数据与较小声明范围混用。
4. 选择 `data` 内全部八个文件。无需上传 descriptor、题面 Markdown 或 `.cpp`；模板内容由页面放进 descriptor。
5. 上传后保留返回的 `upload-…` 题目 ID。上传会创建新的持久题目，不会覆盖内置题；每次重复上传都会新增一题。打开新题运行双模板，预期 AC、4/4，通过后再确认教师逐例信息和学生整体结果。

界面会按自身规则排列文件并生成 case ID；API 包中的 case 顺序明确为 normal、zero、negative、large。两种方式都使用同样四对数据。发布后的题目目录由服务管理，不要直接覆盖或删除已有上传目录。

## 使用现有 API 上传同一数据包

在仓库根目录运行以下 PowerShell 命令；Linux/macOS 将 `curl.exe` 改为 `curl`，使用本 shell 对应的续行方式。此操作实际创建一个新题，只运行一次并保存返回 ID。API 必须已按正常路径启动。

```powershell
curl.exe --fail-with-body --request POST http://127.0.0.1:8080/api/v1/problems `
  --header 'X-OJ-Mode: teacher' `
  --form 'descriptor=<examples/communication-a-plus-b/descriptor.json' `
  --form 'files=@examples/communication-a-plus-b/data/normal.in' `
  --form 'files=@examples/communication-a-plus-b/data/normal.out' `
  --form 'files=@examples/communication-a-plus-b/data/zero.in' `
  --form 'files=@examples/communication-a-plus-b/data/zero.out' `
  --form 'files=@examples/communication-a-plus-b/data/negative.in' `
  --form 'files=@examples/communication-a-plus-b/data/negative.out' `
  --form 'files=@examples/communication-a-plus-b/data/large.in' `
  --form 'files=@examples/communication-a-plus-b/data/large.out'
```

`descriptor` 是无文件名的 JSON 文本 part；每个数据文件是名称为 `files` 的独立 part。不要自己指定 multipart boundary，也不要使用 JSON 主体或上传 ZIP。总请求不超过 1 MiB，descriptor 编码不超过 64 KiB。`starter_codes` 仅有 sender、receiver 两键，兼容字段 `starter_code` 等于 sender；本包不使用交互器、C++ 特判、任意命令或路径配置。

## 教学分工与实际判定

本题要求发送端传原始两个数，接收端做加法；两份模板体现了这个分工。当前固定通信判题器检查两程序编译/退出、消息字节上限及接收端答案，不分析源码来证明加法发生在哪一端。若课程需要评分这一要求，教师需结合代码审阅。

教师数据包上传、服务重启后的题目保留、学生/教师/工程师页面验收都需要单独记录。`scripts/communication-test.mjs` 仅验收内置题的真实判题矩阵，不上传本包，也不证明上传持久化或浏览器操作已通过。
