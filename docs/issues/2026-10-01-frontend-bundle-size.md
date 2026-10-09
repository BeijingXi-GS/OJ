# ISSUE-2026-10-01-002：前端构建通过但 Monaco 分块较大

- 日期：2026-10-01；状态：非阻塞，交由用户安排 Frontend Builder / Worker 处理。
- 关联：P0-001 / 后续 P2-110；本轮没有授权实际前端搭建，因此没有修改前端源码或隐藏警告。

## 前因与结果

修复基础服务后执行完整 `npm run check`，前端 TypeScript / Vite 构建 exit 0，815 modules transformed。Vite 报告部分压缩后 chunk 超过默认 500 kB 警戒值：

```text
EditorPage-BxG8T0qZ.js  2,745.15 kB | gzip: 712.82 kB
Some chunks are larger than 500 kB after minification.
```

初始 index chunk 为 330.28 kB / gzip 104.66 kB，editor.worker 为 275.63 kB。该警告来自现有 Monaco 编辑页，本轮只做回归构建，没有前端代码或依赖变更。

## 原因边界与影响

已知 editor route 采用单独分块，不能把其全部体积当作每次首页加载体积。没有执行浏览器性能测量，尚未核对语言贡献 / editor 功能树与实际网络传输，不把“Monaco 依赖导致警告”提升为已定位的具体冗余模块。

构建成功，此项不阻碍当前基础服务验收；可能影响编辑页首次加载，应在前端实施时制定具体资源 / 首次交互预算。不能通过单纯调高 chunkSizeWarningLimit 把性能验证登记为通过。

## 用户接手与验收

安排 Frontend Builder 在 [前端任务包](../frontend/AGENT_TASKS.md) 的 P2-110 中核对 Monaco 懒加载、必要语言 / 功能与 worker 打包，测量冷缓存和常用设备下的编辑页载入 / 可输入时间，记录优化前后体积与实际网络表现。只有用户授权实际前端修改后再实施；当前记录不要求阻断所有开发。
