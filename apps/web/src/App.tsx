import { useQuery } from '@tanstack/react-query'
import { NavLink, Route, Routes } from 'react-router-dom'
import { lazy, Suspense } from 'react'
import { Button } from '@/components/ui/button'

const EditorPage = lazy(() => import('./EditorPage'))

function Home() {
  const system = useQuery({
    queryKey: ['system'],
    queryFn: async () => {
      const response = await fetch('/api/v1/system')
      if (!response.ok) throw new Error('API 尚未连接')
      return response.json() as Promise<{
        name: string
        stage: string
        judge_enabled: boolean
        contract_version: string
      }>
    },
  })
  return (
    <>
      <p className="eyebrow">DEVELOPMENT FOUNDATION</p>
      <h1>从稳定的基础开始。</h1>
      <p className="intro">
        OJ
        工程基础已就位。这里用于检查前后端连接与编辑器环境，判题功能将按开发阶段接入。
      </p>
      <div className="cards">
        <section className="card">
          <p className="eyebrow">01 / API</p>
          <h2>
            {system.isPending
              ? '正在连接'
              : system.isError
                ? '等待启动'
                : '连接正常'}
          </h2>
          <p>
            {system.isError
              ? '请在另一个终端运行 npm run api。'
              : '通过健康接口检查控制平面的基础连接。'}
          </p>
          <Button
            variant="outline"
            onClick={() => void system.refetch()}
            disabled={system.isFetching}
          >
            重新检查
          </Button>
        </section>
        <section className="card">
          <p className="eyebrow">02 / EDITOR</p>
          <h2>代码编辑器</h2>
          <p>Monaco 本地编辑预览已配置；输入内容只保留在当前页面。</p>
          <Button asChild>
            <NavLink to="/editor">打开编辑器</NavLink>
          </Button>
        </section>
      </div>
      <section className="notice">
        <h2>接下来做什么</h2>
        <p>
          从仓库的 PROGRESS.md 领取任务，阅读 AGENTS.md 的职责边界，并按
          DEVELOPMENT.md 完成设计、实现、验证和交接。
        </p>
        <p>
          本页面不接受或运行提交。当前工程契约：
          {system.data?.contract_version ?? '等待 API 返回'}。
        </p>
      </section>
    </>
  )
}

export function App() {
  return (
    <div className="shell">
      <header>
        <NavLink className="brand" to="/">
          OJ <span>/ 开发工作台</span>
        </NavLink>
        <nav>
          <NavLink to="/">概览</NavLink>
          <NavLink to="/editor">编辑器</NavLink>
        </nav>
      </header>
      <main>
        <Suspense fallback={<p>正在加载编辑器…</p>}>
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/editor" element={<EditorPage />} />
            <Route
              path="*"
              element={
                <p>
                  页面不存在。<NavLink to="/">返回概览</NavLink>
                </p>
              }
            />
          </Routes>
        </Suspense>
      </main>
      <footer>PHASE 0 · 工程基础 · 进度以 PROGRESS.md 为准</footer>
    </div>
  )
}
