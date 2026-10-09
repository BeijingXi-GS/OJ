# ADR-0005：文件用例判题与Docker OCI适配

- 日期2026-10-02；root /Architect；reviewed，独立 /root/mvp_review 设计通过后采用。
- 范围：[judge-mvp/0.1.0契约](../judge/JUDGE_MVP_CONTRACT.md)。只覆盖本机同步文件判题，不冻结完整domain /sandbox.Runtime /Graph或实现未来持久化。

用户明确要求参考Anolix，已拉取其commit4502d2a798f3130d3c540c5c31d45251ca103623并核对config /launcher /测试。它采用Go1.27.1，rootfs默认可写，memory swap暂未映射，强制delete与bundle清理忽略错误，没有OJ答案checker /逐例结果协议。此MVP参考其隔离模式，独立实现Docker Engine OCI/runc适配，不直接引入 /复制Anolix源码，也不为此升级本项目Go。

Docker提供当前Windows开发机上已验证的Linux OCI、seccomp与cgroup能力，避免在主机或服务里申请privileged、修改全局WSL安全配置、给选手暴露Docker socket。可信judge-mvp服务持Docker权限，API仍只有HTTP桥接；payload仅得到固定程序和当前输入。编译与每case分开资源域，swap0、只读根文件系统、有界tmpfs /输出和已检查清理；缺项拒绝执行。

此次时间限制明确为wall且含容器启动开销；CPU累计和真实峰值没有可信采集时为null。MLE依据容器OOM事件，不能把所有137都判MLE。编译失败 /基础设施失败分开，case答案采用显式tokens-v1或exact-v1，manifest快照可追溯。

旧compile接口兼容，新judge接口独立；旧前端“仅编译”文案只保留在编译路径。无业务DB迁移 /队列 /外部部署 /生产安全宣称。回退时停止judge-mvp并移除API JUDGE_URL配置，旧编译功能保留；不删除基础卷。Anolix外部源副本保持原样，不纳入项目代码或依赖锁。
