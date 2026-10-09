# 多题 Linux 单测文件挂载适配

- 日期：2026-10-02；处理人：root；独立审阅：/root/judge_reviewer。
- 状态：已适配并复验通过；属于新增测试的开发修正，没有外部环境阻塞。

## 背景与实际失败

新增 TestRegisteredFixtureAnswers 使用仓库 testdata/problems 校验五题文件答案。原 judge-unit-linux 脚本只挂载交叉编译的可信 Go 测试二进制，容器内没有仓库 fixture。首轮 Linux 执行该测试的五个题目子项失败，错误 testdata_root_invalid；原 symlink、隔离预检、metadata 等单测通过。此失败表明测试输入未配置，不能解释为题目文件或 Docker 引擎坏了。

## 修正顺序与最终结果

1. 为可信测试二进制增加 testdata 的只读 /testdata 挂载，并设置 /tests 工作目录，使 ../../testdata 相对路径有效。
2. 初次适配漏导入脚本 root，运行 ReferenceError 后补充导入；此时未执行容器，不计为测试通过。
3. 重跑 node scripts/judge-unit-linux.mjs，exit0；全部单测通过，包括 Linux symlink 拒绝、五题共24组答案 oracle、deadline 分类、跨题 metadata 拒绝与完整 fixture 预检。
4. 独立 reviewer 确认该挂载仅用于可信 Go test，命令适配是 fake，不挂 Docker socket；实际选手程序容器仍不能读取标准答案，提交配置没有放宽。

无需用户处理。未来增加使用仓库文件的可信测试时，维护对应测试输入，不可用跳过 oracle 掩盖输入缺失。详见 [验证记录](../judge/VALIDATION.md)。
