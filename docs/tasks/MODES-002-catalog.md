# MODES-002 教师上传与动态题库

Owner teacher_catalog Builder；允许internal/problems、internal/problemstore、filejudge/cases.go及相关tests、httpapi/problems.go DTO、新guess-number fixture、scripts/teacher-problem-smoke.mjs。禁止router/engine/docker/contract/appsweb/根MD。Store/严格multipart/验证/原子发布/持久reload/Loader快照及内置猜数题按受审契约实现，禁止任意交互源码上传。可信单测边界/fixture、root实际上传并判题/重启验证。未知外部阻塞报告并停止；开发修正不视外部故障。

## 实际交付

teacher_catalog完成内置/上传根显式Store、严格descriptor与文件配对/限额/路径/重复JSON校验、安全普通文件读取、staging原子发布和持久发现。新增内置猜数题/1到1000固定协议/3秘密数据；上传普通题支持tokens/exact，交互仅固定guess-number-v1秘密.in。老师不能指定ID、镜像、命令或任意交互器。

Store/loader/fixture可信tests和vet通过；root Linux容器实际symlink拒绝通过。root真实教师上传三模式6次与3题重启后元数据/文件/hash一致、AC2/2均通过，详见[验证](../judge/MODES_VALIDATION_2026-10-03.md)。风险R3，独立审阅roles_review，实际状态只在[PROGRESS](../../PROGRESS.md)。

## 后续交接

扩题先定义题型、题面、限制、公开样例和独立答案oracle，再登记目录及manifest并验证正确AC/常见错误WA。编辑/删除/数据版本化要新受审任务；不要覆盖运行中的文件快照或随意拷贝未注册目录。API与worker上传根必须一致，发布数据保留；提交记录尚未持久化。
