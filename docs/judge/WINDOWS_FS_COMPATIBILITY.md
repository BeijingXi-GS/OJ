# Windows 文件发布与判题清理兼容设计

日期2026-10-05；任务COMM-001/002；root提出，comm_review独立审阅。范围是已复现文件操作兼容性；不改变题目判定、沙箱隔离或真实验收门禁。设计审阅与实际实现结果在后续段落补录。

## 已复现事实

Go1.25.4/windows amd64、项目内新建可信测试目录：普通os.Rename空目录成功；os.Root.Rename空目录及关闭子Root后的目录均返回invalid handle；os.Root.Remove空目录成功。os.Remove只读文件成功；os.RemoveAll含0600文件的目录成功，含0555产物时返回unlinkat AccessDenied。Root.Open文件后的File.Chmod因为句柄缺少属性写入访问仍失败；Root.Chmod固定单文件名清除readonly后，递归清理成功。证据.local/evidence/windows-filesystem-probe-20261005.log；这些现象不证明主机ACL或安全软件的唯一根因。

## 题目发布：受限目录中的待提交标记

沿用随机生成的upload-ID和version1注册格式，所有操作继续通过同一个受限os.Root，禁止失败后退回拼接绝对路径rename。

1. 创建空目录`.staging-<ID>`作为pending标记，再独占创建最终ID目录。两次创建均不得覆盖既有目录；只有本次成功创建的目录才取得回滚所有权。
2. 在最终ID的Root内写入全部数据、manifest和problem.json，逐文件Sync/Close；全部成功并关闭最终ID的Root后，以Root.Remove删除空pending目录作为可见性提交点。
3. registered读取开始检查pending；打开已存在的最终ID Root后、读取任何内容前再次检查。后一检查避免读者先看到无pending、然后与新目录创建交错而读到部分文件。任何最终success/error返回前复查pending：合法空目录标记存在则隐藏(false,nil)，畸形标记或查询失败则unavailable。无pending的旧version1题目按原严格规则读取。
4. 写入/关闭/提交失败时，先清理本次确实创建的最终ID目录，确认删除成功才删除pending。删除失败保留pending，重启后仍不可见；不自动回收来源不明的旧pending。

提交前即使problem.json已完整也不能对外可见。读者只在不存在pending且严格注册有效时获得题目。这里只承诺进程并发/中断时的发布可见性；不新增断电后文件系统持久性保证。

必须验证：成功上传/重启、旧version1兼容；写入期间Get/List隐藏、不因部分JSON污染列表；完整文件但未提交仍隐藏；并发读写；提交/回滚失败的隐藏与所有权；原非法数据、链接根/注册文件与通信双模板约束。

## Windows 判题任务清理

保持执行期间产物0555及容器只读挂载。只有已确认全部容器清理、Engine未taint、既有直属run目录检查通过，才进入文件清理。Linux现有行为保持。

从jobsRoot的Root锚定直属Job，校验目录非链接及身份；仅处理program、checker-build/program、sender-build/program、receiver-build/program四种固定路径。未编译阶段缺失可跳过，其他错误失败。固定子目录同样校验非链接与打开后身份；产物必须普通文件，持有打开文件并核对SameFile。

Windows Root.Chmod使用Go的O_OPEN_REPARSE/属性写入路径，末级链接不会被跟随；在固定子Root中只对单名program恢复0600，并再次核对身份。关闭文件及子Root后，由jobsRoot的Root删除该直属Job。不得全树chmod或递归修改外部文件，不改变运行期写权限。任一检查/恢复/关闭/删除失败仍按原路径SYS、taint并保留未清理现场。

必须验证：四种只读产物实际清理、缺失角色、taint现场不修改、非普通文件/目录/链接拒绝、越界目标拒绝，以及双编译失败和通信全流程清理回归。Windows无法创建链接时只能记SKIP，Linux检查另记，不能伪称执行。
