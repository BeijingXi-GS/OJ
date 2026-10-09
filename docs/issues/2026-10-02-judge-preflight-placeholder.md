# 判题预检空占位文件检查修正

- 日期：2026-10-02；JUDGE-002；开发期检查实现修正，已复验关闭。
- 背景：judge镜像预先创建固定 bind mount 目标，包括空的 `/source/solution.cpp`。运行容器没有挂载源码，但空占位仍存在。
- 首次实测：11:25启动新服务，受限编译、产物流式导出及容器创建已走通，运行预检返回 `runtime_kernel_probe_failed`，未监听ready。
- 原因：可信probe使用 `test ! -e /source/solution.cpp`，与自己定义的镜像占位路径冲突。属于正在实施的probe逻辑错误，不是Docker隔离缺失、外部环境故障或新增需用户修复的阻塞。
- 修正：检查 `test ! -s /source/solution.cpp`，验证没有源码内容；容器实际mount allowlist仍禁止运行阶段挂载源码与答案。
- 复验：11:26:31服务输出 `judge ready`。编译、CSV空格/逗号快照挂载、产物exec、cgroup memory/swap/pids/CPU、CapEff0、NoNewPrivs1、Seccomp2及隔离可见性probe通过；真实API矩阵随后通过。
- 影响与处理：未启用过缺项隔离或主机执行回退。首次失败已创建的容器与job目录正常清理，无基础卷变更。用户无需处理此已关闭的开发期错误；如未来出现真正新环境阻塞仍由用户接管并停止。
