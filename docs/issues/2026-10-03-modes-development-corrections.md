# 三模式增量开发期校验修正

日期2026-10-03；处理者root与各Builder。属于当前授权新增实现、兼容性检查和审阅修正，不是新外部环境阻塞。用户要求的问题前因后果在此逐项保留，真实验收尚待完成。

## 旧题库测试的固定假设

原MVP只有5道普通题，HTTP测试固定len=5、compileCalls=5；filejudge文件oracle假定全部tokens-v1、至少4例、2000ms，metadata fake为所有题写batch manifest。新增第6道交互猜数（3例、interaction-v1、5000ms）后，首次targeted test这3项失败。并非线上判题故障。根将API count改为登记目录实际长度并增加type，catalog Builder新增秘密范围与协议fixture验证，runner Builder按题型生成metadata fake。随后httpapi/filejudge targeted tests实际exit0；Store/probs/filejudge tests与vet由catalog执行exit0。

## 上传题判题入口的旧allowlist

静态核查发现Engine.Judge在动态Store登记前仍调用旧ValidateRequest，其compiled allowlist会让合法upload-*提前404。根已交runner改为动态登记校验并要求可信fake证明上传能够到达独立compile路径。该问题尚待真实上传与重启判题复验，不以Store单测代替。

## 文本和请求限额

新input/expected展示取worker单次快照。无效UTF8转换可能扩容，需要独立input_truncated/expected_truncated标记，原始文件/hash/checker字节不受影响。multipart最初把所有底层读错转固定code，导致HTTP MaxBytesError无法辨识为413；catalog已用脱敏private readError保留Unwrap，仍不向页面透传底层路径或系统错误。适用边界测试纳入本轮检查。

本档只记录常规开发修正；新的未知环境阻塞依用户规则另开专档、交用户并停止。最终关闭依据以PROGRESS/本轮验证为准。

## 独立实现审阅发现的执行边界

roles_review只读实现与可信fake后指出：创建/检查双容器的case deadline最初被判系统错误；stdin copy错误被忽略；四路输出总量与首次超线归因不是同一临界区；worker wire只凭可选interactor_compile猜题型可能让交互请求接受batch AC；WA的participant_stopped仅凭旧Running观察，不能证明真正监督kill，且自然交互器异常可能被其他判定掩盖。

根已同步受审start_timed_out窄字段（只允许交互、双方编译成功、未开始attach、TLE/null退出/无OOM/无stop/确认清理），收束Client和API登记题型/固定源码hash验证，并增加HTTP错误worker及启动wire矩阵。runner负责实际阶段超时、输入错误、输出归因锁与监督终止事实；独立报告复验前不登记已修复。针对固定sourcehash的新strict test使旧lifecycle fake使用ffff摘要失败，属于测试fixture待同步，不放宽验证规则。

## 前端可信测试脚本初次lint

新增verify-modes脚本初次lint因Node全局标识未显式导入失败。frontend Builder改为显式Node导入/globalThis，最终Prettier/ESLint/TS/Vite及8项模式mock实际通过。未升级依赖或改变产品模式逻辑。真实页面还需root验收。

## 浏览器临时标签清理与选择控件

真实UI已完成上传，生成upload-5996de397727f9c0404617993a76c448并打开题面。随后临时标签动作报不属于浏览器会话；按工具文档只读核实该浏览器tabs为空，属于临时tab清理/失效，不据此判断浏览器断连或产品判题失败。使用同一已选浏览器新建该已发布题目的标签并markHandoff保留，无重复发布，学生默认页和运行按钮已实际恢复。题型选择getByLabel exact无匹配，读DOM后按真实combobox名称选择成功；属于定位方法修正。最终结果和截图在本轮浏览器记录中另行验收。

## 旧真实矩阵的 metadata 断言

运行旧17项脚本前静态比对发现：动态metadata的完整对象断言没有新增problem_type，而当前API按受审契约返回batch字段。只为旧测试补齐该字段，产品代码与验证条件未放宽。教师详情的原生summary在DOM中不是button角色；读取实际页面后用原生可访问控件展开成功，属于测试定位方法调整。

## 交互生命周期实际观察器

独立review指出原interactive-test取消专项只确认两个容器存在，不能证明双方已Running；原13判定和创建后取消清理通过事实保留，不升级措辞。新增严格source/run绑定与双方Running的专项。首次实际调用在尚未观察到两方前出现docker_inspect_failed，脚本取消请求；随后真实清理检查为0容器/0job。旧观察器ps选择整个run，包含可能在ps/inspect之间正常退出并清理的短命编译容器，存在查询竞态。首次未保存具体失败ID，不能断言它就是当次唯一原因；报告保留interactive-lifecycle-first-attempt.json。

runner将观察范围收束到本次run的case-a双方，仍严格检查ID/Name/owner/run/stage/Running，真正inspect错误仍失败而不盲目重试；产品核心未改。待重新执行和独立审阅后才登记双运行取消/系统故障通过。此为新增验收观察器修正，尚未观察到引擎断连或未知外部阻塞。

## 最终收束

本档各开发修正均已复验并独立签署：完整check/适用Linux单测通过；动态上传入口真实三模式及重启AC通过；首轮六项执行边界有可信与真实矩阵；第二版观察器确认双Running后取消、TERM交互器SYSTEM_ERROR和完整清理。原失败记录保留，不把推测补写为根因。页面分级展示、在途切换/保留源码、上传与手机布局均通过。详见[完整验证](../judge/MODES_VALIDATION_2026-10-03.md)和[最终独立报告](../reviews/2026-10-02-modes-interactive-review.md)。本档常规开发项已闭合；JSON响应膨胀已另档关闭，退出提示保留独立非阻塞运维项交用户。
