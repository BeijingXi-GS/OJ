# 通信题本地交付与真实验收接续

日期2026-10-05；root → 用户处理环境后通知继续的执行者。本轮用户允许本地先行，已完成A+B增量实现；新完整回归仍因文件权限阻塞，依赖验收已停止。原GitHub推送与并行/独立审阅授权有效，无需再次申请同一授权；文件操作、Docker与审批恢复需实际证据。

## 已交付

固定sender/receiver双C++源码、逐角色编译与规范快照/产物绑定、单向原始字节受控管道、通信限额与共享wall、RE/OOM/输出/传输/提前退出因果、双方取消/清理/taint；三内置通信题各4例及教师双模板/配对数据上传。新增communication-a-plus-b的sender只传a、b，receiver计算和，±10¹²/32字节/普通零负大数四例；教师新增A+B说明与双模板按钮，原sum/maximum保留。API独立公开/内部端点与严格新旧题型分叉，三模式服务端白名单和双编辑器/过期结果/取消前端完成。

[通信契约](../judge/COMMUNICATION_CONTRACT.md)、[使用指南](../judge/COMMUNICATION_USER_GUIDE.md)、[实际验证](../judge/COMMUNICATION_VALIDATION_2026-10-05.md)、[独立局部报告](../reviews/2026-10-05-communication-review.md)、[开发问题与方案](../issues/2026-10-05-communication-development.md)。新增A+B前完整check历史exit0；新增后前端检查与窄独立专项PASS，当前完整回归失败，不沿用旧PASS。235非MD源码SHA25671e866f2e74ab2f8e9d107a8cacaf9240ed03ded1c87bf963821b8ba9a3cb634，证据communication-ab-20261005-local.json；工作分支/HEAD保留，不是新commit或远程发布。真实产品pending。

停止位置：[项目测试临时目录](../issues/2026-10-05-local-test-permissions.md)的rename/unlink在两轮正常授权与复验后仍拒绝；日志保留，未批量删除测试残留。最后逐例AC/WA不等于最终成功，清理失败仍SYS/taint。Docker最新是Linuxpipe权限拒绝，权限工具不能表示管道路径。

## 下一步与Agent职责

1. **用户/root环境与集成**：先恢复同一正常执行环境对.local/tmp目录重命名、假产物删除和清理的能力；两个失败专项通过后再五包回归、完整check与新摘要绑定。恢复Desktop/Linux API访问及审批后，明确查询Linux/cgroup2和本项目owner容器/jobs，再保持严格预检启动。默认配置不可访问与显式Linuxpipe权限拒绝分别验证，不能把grant或隐藏启动exit0当ready。保留旧数据卷、教师上传和工作区；不要宿主执行C++。[Docker档案](../issues/2026-10-05-local-docker-access.md)和[测试权限档案](../issues/2026-10-05-local-test-permissions.md)记录未定位原因。
2. **backend Builder**：完整本地检查恢复后，真实worker/镜像确认时先实测A+B四例双模板AC4/4、receiver错误求和WA、sender通信超限WA；再维护双方CE/natural RE/137/真OOM、共享启动及运行TLE、输出OLE、EOF/提前退出/未知IO、同时取消/双清理/taint矩阵，核实sender仅输入、自身产物与receiver仅自身产物。新增真实证据绑定当前digest/image；不自行弱化seccomp/cgroup/profile或清理要求。
3. **frontend Builder**：真实页面复验双编辑器可输入、改任一源码结果过期、切模式双草稿保留、旧请求不覆盖、角色CE；教师通信模板/byte_limit/.in.out发布与逐例反馈；学生无隐藏用例/日志/hash，工程说明只工程师。源码在apps/web，不定义新的后端判定语义。
4. **root/旧产品回归与持久化**：新旧题都可判后，继续[10-04旧交接](2026-10-04-special-judge-paused.md)的三阶段停止、六上传/34文件前后快照、下载/编辑/学生CE/工程师证据与最终smoke。Windows可信观察self-test已PASS仍不能替代真实停止；旧持久快照只before，没有after结论。
5. **独立Architect/Security**：只读实证与源码、更新自己报告；审真实挂载/资源/终止/清理及前端投影，把本地PASS和真实PASS分别写明。真实验收齐备前不能签最终产品PASS；实现者自检不冒充独立签署。
6. **root/Git交付**：真实产品/当前完整检查与独立签署齐备后，按[PUB-001](../tasks/PUB-001-github-delivery.md)核实远程、必要集成/commit/push到BeijingXi-GS/OJ，保留未提交来源与远程历史，不强推。不把本地origin URL或当前HEAD当推送成功。

新未知外部阻塞单开问题MD，写前因/实际错误/策略/影响和恢复条件；对已知Docker/审批问题采用本次用户已允许的本地先行范围。日常受审实现修正由owner完成记录，新增核心行为先审后改。实时状态仅在[PROGRESS](../../PROGRESS.md)维护。
