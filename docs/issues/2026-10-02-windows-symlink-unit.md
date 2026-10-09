# Windows符号链接单测权限与Linux补验

- 日期：2026-10-02；JUDGE-002；非阻塞，本轮补验已完成，未修改Windows权限。
- 背景：loader明确拒绝根目录内部symlink，需要真实类型检查测试。
- 实际：本机Go TestCaseSymlinkRejected调用os.Symlink时Windows拒绝，测试记录SKIP；其他文件 /协议 /生命周期测试通过。不能把该SKIP写成Windows边界通过。
- 原因范围：Windows当前进程不具备创建符号链接的权限；未断定是否Developer Mode配置 /组策略或token限制，也未修改它们。
- 补验：交叉编译同一internal/filejudge可信Go test二进制，放在UID10002 /无网络 /无cap /NNP /只读 /有界tmpfs的Linux容器，不挂Docker socket；同一TestCaseSymlinkRejected实际PASS，全部该包测试PASS。
- 用户事项：本轮无需为了此测试调整主机权限。若用户需要Windows原生symlink覆盖，可自行配置权限后再复验；当前功能验收使用已记录Linux实测，不冒称Windows也执行成功。未执行本机race（历史CGO限制），远程CI未执行。
