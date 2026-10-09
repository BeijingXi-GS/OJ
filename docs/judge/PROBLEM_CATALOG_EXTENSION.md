# judge-mvp/0.1.0 多题增量与恢复裁定

- 2026-10-02用户要求修复并继续、增加题目；状态reviewed，独立/root/judge_reviewer已局部设计通过，采用以下实现约束。
- 原public请求 /response DTO、资源与清理profile保持；扩展problem_id受控集合，新增内部metadata按ID查询。不是开放任意服务器目录或重定义完整核心契约。

## 目录和接线

本增量覆盖旧JUDGE_MVP_CONTRACT中“只支持固定a-plus-b”的ID集合裁定；其余public请求 /结果DTO和资源profile不变。metadata新增problem_id字段以绑定请求题目，Client必须核对相等；public详情可包含同一新增字段，旧前端可兼容忽略。

新增可信internal/problems编译期目录，至少Definition包含ID /Title /Summary /Difficulty /Description /InputDescription /OutputDescription /Notes /Samples /StarterCode。List /Get /Exists作为唯一ID来源；API list/detail、旧compile请求验证、新judge Request验证以及filejudge.Load均使用同一allowlist。目录下多一个未登记文件夹不能启用题目；未知ID /../ /分隔符仍404或输入拒绝。

新增array-sum（数组元素和，用long long）、maximum-element（数组最大值）、palindrome（小写字母回文，YES /NO）、sort-numbers（整数升序，空白token比较）。原a-plus-b保持兼容。每题4..8例，输入 /答案≤64KiB、wall2000ms /memory64MiB默认，checker tokens-v1。提供完整中文题面、边界与输入输出样例；默认模板用于本机演示，可给演示解法，不引入持久化 /题库编辑API。

内部GET /problem/{id}返回指定题目的当前manifest metadata；旧GET /problem兼容A+B。MetadataProvider.Metadata(ctx,problemID)及Client固定URL encode path ID；public详情必须匹配题目ID，不能返回A+B的限额。worker只load allowlisted ID；preflight校验全部注册题目数据后再做一次通用编译 /runtime限制probe，缺失任一已注册题目的数据拒绝ready，运行时文件改变仍SYSTEM_ERROR。

## 诊断与前端

Docker控制命令依旧最多5s。诊断区分context deadline /cancel、CLI退出或启动失败、输出超限；只有成功docker info且内容非2才报告cgroup_v2_required。服务启动日志仅输出固定操作名 /错误分类，不记录原始Docker输出 /主机路径 /环境凭据；查询失败仍拒绝执行，不放宽限制。

多题列表沿用当前动态API。切题时旧请求必须取消、旧源码和判定不能串题；编辑 /恢复 /编译 /运行共用已有锁。独立reviewer审局部接线 /输入边界与真实用例正确性，root实测全部题目AC和错误解WA、Docker verdict回归、UI交互和清理。
