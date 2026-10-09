# 通信题使用与接续指南

当前固定双角色MVP本地实现；真实Docker通信验收待环境恢复，状态以[PROGRESS](../../PROGRESS.md)为准。设计见[通信契约](COMMUNICATION_CONTRACT.md)，问题与修正见[开发档案](../issues/2026-10-05-communication-development.md)。仅本机手动切换模式，没有账号权限隔离。

## 学生：两份代码共同完成一道题

选择通信题后，分别编辑“发送端”和“接收端”。发送端读取题目输入，并将消息写到自己的标准输出；接收端只能读取这些消息，将最终答案写到自己的标准输出。两份程序都使用C++20，并遵守题目公开的通信字节、时间和每角色内存限制。

点击“运行测试”会先分别编译两份代码，全部成功后运行用例。任一代码修改都会让旧结果过期；切题或切模式取消旧请求，切模式保留双源码草稿。学生只显示整体结果、通过数量以及自己实际已编译角色的CE诊断；不会显示隐藏用例、答案或对方的运行日志。首次编译失败后另一角色未编译，不伪造成功。

新增最简单的内置题 **communication-a-plus-b**：发送端读取并发送a、b，接收端计算a+b。输入范围±10¹²，通信上限32字节，四组服务器数据覆盖普通、零、负数和大数。原communication-sum（整数序列总和，32字节）与communication-maximum（最大值，16字节）保留；三道通信题各4例，内置题总数10。

例如发送端读取`1 2`，发送`1 2\n`；接收端读取这两个数，输出`3\n`。下面是该题的默认双模板；真实容器执行尚待环境恢复。

```cpp
// 发送端 sender.cpp
#include <iostream>
int main() {
    long long a, b;
    if (!(std::cin >> a >> b)) return 1;
    std::cout << a << ' ' << b << '\n';
    return 0;
}
```

```cpp
// 接收端 receiver.cpp
#include <iostream>
int main() {
    long long a, b;
    if (!(std::cin >> a >> b)) return 1;
    std::cout << a + b << '\n';
    return 0;
}
```

## 教师：创建题面、消息协议和配对数据

上传页面选择通信题，填写题面、两份模板、1..65536字节通信限额，并上传同名.in/.out，最多4组。点击“填入通信A+B说明与双模板”，即可得到发送端传两数、接收端求和的示例。该教师示例范围±10⁹，样例`2 3`→`5`，限额32字节；例如上传sample.in内容`2 3\n`和sample.out内容`5\n`。原整数序列求和示例仍可选择。

通信编码由题目公开要求决定，两角色只能单向通信；本次不支持反向消息、任意进程图、第三角色、交互器或C++特殊判题混用。判定以接收端输出和对应.out为准。输入数据仅发送端可读，标准答案不挂到任何选手容器。

教师能查看逐例输入/答案、接收端输出、方向日志与通信量。日志有界：stdout最多64KiB、stderr最多16KiB，截断会标注；教师不获得角色exit/OOM/终止证据或源码/产物hash。通信越界为WA，正常完成后答案错误为WA；CE/RE/TLE/MLE/OLE/SYSTEM_ERROR按对应原因显示。

上传总量上限1MiB，descriptor编码上限64KiB包括题面与双模板。每份源码原始上限64KiB不表示两份最大模板和最大题面可以同时上传。starter_codes固定sender/receiver键，兼容starter_code必须等于sender模板；缺失、重复或额外角色及非法文件路径/配对均拒绝。发布后保存在既有教师题库目录，不覆盖旧题。

## 工程师：传输、资源与可信结果

固定图sender stdout → 可信受控输入管道 → receiver stdin，EOF关闭输入，不建立反向通道。receiver stdout采用tokens-v1的空白token比较。bytes_observed计原始读取字节，bytes_transferred计成功写入受控输入管道的字节，不代表程序实际消费；越界只转发剩余额度内prefix并停止双方。sender stderr与receiver stdout/stderr共享1MiB执行输出线，sender stdout独立受通信限额约束。

每例两独立容器共享墙钟截止时间，含创建/启动开销。每角色内存/pids独立受限，aggregate_memory_limit_bytes是两角色限额的受限上界，不是实际聚合峰值；CPU和peak未采集保持null。先基础设施/清理错误SYS，再真实OOM MLE、wall TLE、输出OLE、通信越界WA、自然非零RE，双方正常才比较答案。natural137不等于监督器停止；terminated_by_supervisor只有控制kill和最终daemon exit137被确认才为true。

通信提交使用POST /api/v1/communication-judgements，JSON为problem_id和source_units:[{role,language,source}]，固定两角色/语言cpp，每份64KiB，总编码≤1MiB。Worker端点为POST /internal/communication-judge。旧提交/编译端点仍只接受单源码，旧262144字节上限不变。显式角色绑定后规范sender/receiver顺序计算源摘要，只有两个成功产物才有聚合产物摘要；Client/API严格验证角色、源、产物、数据、规则与用例顺序。

启动器要求worker health有communication_contract_version，并保持现有严格沙箱预检；不能复用缺标记的旧worker。API不编译、不运行任何源码。双attach收束、取消、EOF和双清理由Worker负责，清理不可信则taint并停止接新请求。该版本标记不证明真实通信矩阵通过。

## 真实环境恢复后的验收

先成功连接Docker Linux/cgroup2，再核实本项目owner容器和jobs；保留旧上传与数据卷。按正常预检启动本机MVP，实测双角色AC/答案WA/通信越界WA/双方CE/自然RE与137/TLE/真OOM/OLE、提前退出/EOF/双方取消/清理和输入/答案挂载隔离。

再通过教师页面上传通信题、逐例结果、学生不泄露隐藏数据和工程师完整证据；完成重启后题目/数据摘要与判定复验。旧SPJ停止三阶段、六上传题持久化及页面待办继续保留。只有对应当前源码的真实证据和独立签署齐备才关闭产品验收，不以本地fake或历史SPJ矩阵替代。
