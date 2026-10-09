import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Link, NavLink, Navigate, Route, Routes } from 'react-router-dom'
import { lazy, Suspense } from 'react'
import { Button } from '@/components/ui/button'
import { getProblems } from '@/lib/api'
import { changeMode, modeLabels, useMode, type Mode } from '@/lib/mode'

const EditorPage = lazy(() => import('./EditorPage'))
const UploadPage = lazy(() => import('./UploadPage'))
const DebugPage = lazy(() => import('./DebugPage'))

function Home() {
  const mode = useMode()
  const problems = useQuery({
    queryKey: ['problems', mode],
    queryFn: ({ signal }) => getProblems(mode, signal),
  })
  return (
    <div className="home">
      <section className="hero">
        <p className="eyebrow">C++ PRACTICE</p>
        <h1>
          从一道题，
          <br />
          开始写代码。
        </h1>
        <p className="intro">
          阅读题面，在编辑器中尝试你的解法。
          <br />
          编译检查错误，再运行测试验证答案。
        </p>
        <span className="scope-label">
          <span className="status-dot" /> C++20 · 标准题、交互题与通信题
        </span>
      </section>
      <section className="problem-list" aria-labelledby="problem-list-title">
        <div className="section-heading">
          <div>
            <p className="eyebrow">PROBLEMS</p>
            <h2 id="problem-list-title">选择一道题目</h2>
          </div>
          {problems.data && (
            <span className="muted count">
              {problems.data.problems.length} 道题目
            </span>
          )}
        </div>
        {problems.isPending ? (
          <div className="load-card" role="status">
            正在加载题目…
          </div>
        ) : problems.isError ? (
          <div className="load-card error-card" role="alert">
            <h3>题目暂时无法加载</h3>
            <p>{problems.error.message}</p>
            <Button
              variant="outline"
              onClick={() => void problems.refetch()}
              disabled={problems.isFetching}
            >
              {problems.isFetching ? '正在重试…' : '重新加载'}
            </Button>
          </div>
        ) : problems.data.problems.length === 0 ? (
          <div className="load-card">暂时没有可用题目。</div>
        ) : (
          <div className="problem-cards">
            {problems.data.problems.map((problem, index) => (
              <Link
                className="problem-card"
                to={`/problems/${encodeURIComponent(problem.id)}`}
                key={problem.id}
              >
                <span className="problem-number">
                  {String(index + 1).padStart(2, '0')}
                </span>
                <div className="problem-card-content">
                  <div className="problem-card-title">
                    <h3>{problem.title}</h3>
                    <span className="difficulty">{problem.difficulty}</span>
                    <span className={`problem-type type-${problem.type}`}>
                      {problem.type === 'communication'
                        ? '通信题'
                        : problem.type === 'interactive'
                          ? '交互题'
                          : '标准题'}
                    </span>
                  </div>
                  <p>{problem.summary}</p>
                  <span className="card-meta">
                    C++20 ·{' '}
                    {problem.type === 'communication'
                      ? '发送端与接收端协作'
                      : problem.type === 'interactive'
                        ? '与交互器通信'
                        : '编写解法，验证答案'}
                  </span>
                </div>
                <span className="card-action" aria-hidden="true">
                  ↗
                </span>
                <span className="sr-only">打开题目</span>
              </Link>
            ))}
          </div>
        )}
      </section>
      <div className="home-footnote">
        <span className="footnote-mark" aria-hidden="true">
          {'</>'}
        </span>
        <p>写下解法，检查编译，运行测试。代码与结果仅保留在当前页面。</p>
      </div>
    </div>
  )
}

export function App() {
  const mode = useMode()
  const queryClient = useQueryClient()
  return (
    <div className="shell">
      <a className="skip-link" href="#main-content">
        跳到主要内容
      </a>
      <header className="site-header">
        <Link className="brand" to="/" aria-label="OJ 首页">
          <span className="brand-mark" aria-hidden="true">
            OJ
          </span>
          <span>C++ 练习</span>
        </Link>
        <nav aria-label="主导航">
          <NavLink to="/" end>
            题库
          </NavLink>
          <NavLink to="/problems/a-plus-b">开始练习</NavLink>
          {mode !== 'student' && <NavLink to="/upload">上传题目</NavLink>}
          {mode === 'engineer' && <NavLink to="/debug">工程调试</NavLink>}
        </nav>
        <div className="mode-control">
          <label htmlFor="demo-mode">当前模式</label>
          <select
            id="demo-mode"
            value={mode}
            onChange={(event) => {
              void queryClient.cancelQueries()
              changeMode(event.target.value as Mode)
            }}
            aria-describedby="mode-help"
          >
            {Object.entries(modeLabels).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>
        </div>
      </header>
      <div className="demo-notice" id="mode-help">
        <span className="status-dot" />
        {mode === 'engineer' ? '本机演示 · 模式手动切换' : modeLabels[mode]}
        <span>切换模式会清空结果，并保留本题代码。</span>
      </div>
      <main id="main-content">
        <Suspense
          fallback={
            <div className="load-card" role="status">
              正在加载编辑器…
            </div>
          }
        >
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/problems/:problemId" element={<EditorPage />} />
            <Route
              path="/upload"
              element={
                mode === 'student' ? (
                  <Navigate to="/" replace />
                ) : (
                  <UploadPage key={mode} />
                )
              }
            />
            <Route
              path="/debug"
              element={
                mode === 'engineer' ? (
                  <DebugPage key={mode} />
                ) : (
                  <Navigate to="/" replace />
                )
              }
            />
            <Route
              path="/editor"
              element={<Navigate to="/problems/a-plus-b" replace />}
            />
            <Route
              path="*"
              element={
                <div className="load-card">
                  <h1>页面不存在</h1>
                  <p>这个地址没有对应的页面。</p>
                  <Button asChild>
                    <Link to="/">返回题库</Link>
                  </Button>
                </div>
              }
            />
          </Routes>
        </Suspense>
      </main>
      <footer className="site-footer">
        <span>OJ · C++ 练习</span>
        <span>{modeLabels[mode]} · 编译检查与运行测试</span>
      </footer>
    </div>
  )
}
