# 学生信息与特判页面实际验收（部分完成）

日期2026-10-04；执行root；浏览器Codex In-app Browser，同一实际标签。不是最终页面PASS，当前停止于自动审批通信复发。

## 已执行结果

1. 学生打开整数排序：中文题面、输入输出、公开样例、重复元素提示及时间/内存限制正常。没有token比较、容器、墙钟/启动开销、默认演示解法、GCC版本或阶段时间等工程说明。
2. 切换工程师：上述token/空白比较、编译/运行区别、容器/墙钟与启动开销说明显示于工程说明区域；规则和调试信息可见。
3. 教师进入真实上传页面，选择“教师上传 C++ 特判”，实际文件选择器读取 canonical checker.cpp（2368字节）。填写题面、数学条件、公开样例、工程说明和228字节完整解法模板；选sample.in、sample.out、impossible.in，页面确认2组/12字节，验证标准答案可省略。
4. 发布返回真实题目 `upload-29db51251c6398020722c527cfd55470`，标题“不相邻排列 · 教师特判示例”。点击运行，实际AC2/2。展开case-01，反馈“Correctly reported no solution.”、输入3、标准答案为空、选手输出-1正常显示。没有特判源码/产物hash或编译器内部证据泄入教师区。

真实发布后的页面地址：`http://127.0.0.1:5173/problems/upload-29db51251c6398020722c527cfd55470`。服务目前已停；不能把缓存显示当作在线健康证明。

## 开发修正与未完成

初次等待错误标题“上传题目”超时，实际标题为“上传一道题目”，随后用真实DOM继续；只是QA定位错误。原下载按钮两次等待download事件超时；Monaco native EditContext导致DOM控制输入失败，textarea压键没有修改内容，未提交错误答案。两项修正见[独立问题](../issues/2026-10-04-browser-special-controls.md)：常驻标准下载链接、editContext:false。

修正后类型/lint/模式自检和独立代码复读通过；HMR后实际DOM已显示传统“C++源码编辑器”textbox，但未实际输入/提交，因此不能说编辑验收通过。仍需：下载文件与当前源码字节一致、教师WA及具体反馈、学生CE诊断/隐藏反馈、工程师特判编译/运行证据、移动布局和最终截图。API角色投影的23项真实矩阵已通过，但不是这些页面操作的替代。

停止时DOM保存在 `.local/evidence/spj-20261004-teacher-browser-snapshot.txt`。截图请求失败，单独[记录](../issues/2026-10-04-browser-screenshot.md)，没有有效截图或最终截图交付。浏览器标签保留为handoff。

最终继续顺序见[交接](../handoffs/2026-10-04-special-judge-paused.md)和[PROGRESS](../../PROGRESS.md)。
