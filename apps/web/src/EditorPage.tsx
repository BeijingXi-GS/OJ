import { useEffect, useMemo, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link, useLocation, useParams } from 'react-router-dom'
import Editor, { loader } from '@monaco-editor/react'
import * as monaco from 'monaco-editor/editor/editor.api.js'
import 'monaco-editor/languages/definitions/cpp/register.js'
import EditorWorker from 'monaco-editor/editor/editor.worker.js?worker'
import { Button } from '@/components/ui/button'
import JudgeFeedback from './JudgeFeedback'
import CommunicationWorkspace from './CommunicationWorkspace'
import { formatMemory } from '@/lib/utils'
import { useMode } from '@/lib/mode'
import {
  APIError,
  compileSource,
  getProblem,
  judgeSource,
  type CompilationResult,
  type JudgeResult,
  type Problem,
  type CommunicationSources,
} from '@/lib/api'

self.MonacoEnvironment = { getWorker: () => new EditorWorker() }
loader.config({ monaco })

const maxSourceBytes = 65536

function CompilationFeedback({
  result,
  stale,
}: {
  result: CompilationResult
  stale: boolean
}) {
  const mode = useMode()
  const labels = {
    success: {
      title: '编译成功',
      detail: '可以运行测试，检查你的答案。',
      icon: '✓',
    },
    compile_error: {
      title: '编译未通过',
      detail: '根据下方提示修改代码，然后重新编译。',
      icon: '!',
    },
    timeout: {
      title: '编译超时',
      detail: '编译超过时间限制，请检查代码后重试。',
      icon: '⌛',
    },
  }
  const label = labels[result.status]
  return (
    <div className={`compile-result result-${result.status}`}>
      {stale && (
        <p className="stale-notice" role="status">
          代码已修改：此结果对应修改前的代码，请重新编译。
        </p>
      )}
      <div className="result-heading" role="status" aria-live="polite">
        <span className="result-icon" aria-hidden="true">
          {label.icon}
        </span>
        <div>
          <h3>{label.title}</h3>
          <p>{label.detail}</p>
        </div>
        {mode === 'engineer' && (
          <span className="duration">{result.duration_ms} ms</span>
        )}
      </div>
      {mode === 'engineer' && (
        <p className="compiler-meta">
          {result.compiler} · {result.standard} · 仅检查编译与链接，未运行程序
        </p>
      )}
      {result.diagnostics && (
        <div className="diagnostics">
          <div className="diagnostics-heading">编译器输出</div>
          <pre tabIndex={0} aria-label="编译器诊断信息">
            {result.diagnostics}
          </pre>
        </div>
      )}
      {result.diagnostics_truncated && (
        <p className="truncated-notice">编译器输出较长，仅展示部分内容。</p>
      )}
    </div>
  )
}

function ProblemWorkspace({
  problem,
  initialSource,
  rememberSource,
}: {
  problem: Problem
  initialSource: string
  rememberSource: (source: string) => void
}) {
  const mode = useMode()
  const [source, setSource] = useState(initialSource)
  const [operation, setOperation] = useState<'compile' | 'judge' | null>(null)
  const [result, setResult] = useState<{
    data: CompilationResult
    source: string
  } | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [judgeResult, setJudgeResult] = useState<{
    data: JudgeResult
    source: string
  } | null>(null)
  const [judgeError, setJudgeError] = useState<string | null>(null)
  const controller = useRef<AbortController | null>(null)
  const isBusy = operation !== null
  const isCompiling = operation === 'compile'
  const isJudging = operation === 'judge'
  const judgeEnabled = problem.judge?.enabled ?? false
  const canReset =
    source !== problem.starter_code ||
    !!result ||
    !!judgeResult ||
    !!error ||
    !!judgeError
  const sourceBytes = useMemo(
    () => new TextEncoder().encode(source).byteLength,
    [source],
  )
  const validation = !source.trim()
    ? '请输入 C++ 代码。'
    : sourceBytes > maxSourceBytes
      ? '代码超过 64 KiB，请精简后再编译。'
      : null
  useEffect(() => () => controller.current?.abort(), [])

  async function submit(kind: 'compile' | 'judge') {
    if (controller.current || validation || (kind === 'judge' && !judgeEnabled))
      return
    const requestController = new AbortController()
    controller.current = requestController
    const submittedSource = source
    setOperation(kind)
    if (kind === 'compile') {
      setResult(null)
      setError(null)
    } else {
      setJudgeResult(null)
      setJudgeError(null)
    }
    try {
      if (kind === 'compile') {
        const data = await compileSource(
          problem.id,
          submittedSource,
          mode,
          requestController.signal,
        )
        if (!requestController.signal.aborted)
          setResult({ data, source: submittedSource })
      } else {
        const data = await judgeSource(
          problem.id,
          submittedSource,
          mode,
          requestController.signal,
        )
        if (!requestController.signal.aborted && data.mode === mode)
          setJudgeResult({ data, source: submittedSource })
      }
    } catch (failure) {
      if (!requestController.signal.aborted) {
        const message =
          failure instanceof Error ? failure.message : '请求未能完成，请重试。'
        if (kind === 'compile') setError(message)
        else setJudgeError(message)
      }
    } finally {
      if (controller.current === requestController) {
        controller.current = null
        setOperation(null)
      }
    }
  }

  function resetSource() {
    if (isBusy || !canReset) return
    if (
      !window.confirm(
        '恢复题目模板将替换当前代码，并清空编译与测试结果，确定继续吗？',
      )
    )
      return
    setSource(problem.starter_code)
    rememberSource(problem.starter_code)
    setResult(null)
    setError(null)
    setJudgeResult(null)
    setJudgeError(null)
  }

  return (
    <>
      <Link className="back-link" to="/">
        ← 返回题库
      </Link>
      <div className="problem-heading">
        <div>
          <p className="eyebrow">PROBLEM</p>
          <h1>{problem.title}</h1>
        </div>
        <span className="difficulty">{problem.difficulty}</span>
        <span className={`problem-type type-${problem.type}`}>
          {problem.type === 'interactive'
            ? '交互题'
            : problem.judge.checker === 'nonconsecutive-permutation-v1'
              ? '构造题'
              : mode !== 'student' && problem.judge.checker === 'cpp-special-v1'
                ? '特殊判题'
                : '标准题'}
        </span>
      </div>
      <div className="workspace-grid">
        <article className="statement panel" aria-labelledby="statement-title">
          <div className="panel-heading">
            <h2 id="statement-title">题目说明</h2>
            <span className="muted">C++20</span>
          </div>
          <div className="statement-body">
            <section>
              <h3>题目描述</h3>
              <p>{problem.description}</p>
            </section>
            <section>
              <h3>输入格式</h3>
              <p>{problem.input_description}</p>
            </section>
            <section>
              <h3>输出格式</h3>
              <p>{problem.output_description}</p>
            </section>
            {problem.samples.map((sample, index) => (
              <section className="sample-section" key={index}>
                <h3>样例 {index + 1}</h3>
                <div className="sample-grid">
                  <div className="sample-box">
                    <span>输入</span>
                    <pre>{sample.input}</pre>
                  </div>
                  <div className="sample-box">
                    <span>输出</span>
                    <pre>{sample.output}</pre>
                  </div>
                </div>
              </section>
            ))}
            {problem.notes && (
              <section>
                <h3>提示</h3>
                <p>{problem.notes}</p>
              </section>
            )}
            {problem.judge && (
              <section
                className="judge-limits"
                aria-labelledby="judge-limits-title"
              >
                <h3 id="judge-limits-title">题目限制</h3>
                <dl>
                  <div>
                    <dt>时间限制</dt>
                    <dd>{problem.judge.wall_time_ms} ms</dd>
                  </div>
                  <div>
                    <dt>内存限制</dt>
                    <dd>{formatMemory(problem.judge.memory_limit_bytes)}</dd>
                  </div>
                </dl>
              </section>
            )}
            {problem.type === 'interactive' && (
              <div className="statement-notice">
                <span className="status-dot" />
                <p>
                  每次询问后必须刷新输出，再读取回复。提交最终答案后结束程序。
                </p>
              </div>
            )}
            {mode === 'engineer' && (
              <section className="engineering-notes">
                <h3>工程说明</h3>
                {problem.engineering_notes && (
                  <p>{problem.engineering_notes}</p>
                )}
                <p>
                  比较器：{problem.judge.checker}；测试用例：
                  {problem.judge.case_count}{' '}
                  个。时间限制采用墙钟时间，包含容器启动开销，不等于程序 CPU
                  时间。
                </p>
                <p>
                  编译按钮仅检查编译与链接；运行测试会在独立受限容器中执行。
                  {problem.judge.checker === 'cpp-special-v1' &&
                    ' 教师特判在选手成功运行后独立执行；特判自身故障判为 SYSTEM_ERROR。'}
                </p>
              </section>
            )}
          </div>
        </article>
        <section className="coding-column" aria-labelledby="code-title">
          <div className="code-panel panel">
            <div className="panel-heading editor-heading">
              <h2 id="code-title">代码编辑器</h2>
              <Button
                variant="outline"
                className="reset-button"
                onClick={resetSource}
                disabled={isBusy || !canReset}
              >
                恢复模板
              </Button>
            </div>
            <div className="editor-filebar">
              <span>
                <span className="file-icon" aria-hidden="true">
                  {'</>'}
                </span>{' '}
                solution.cpp
              </span>
              <span>C++20</span>
            </div>
            <div className="editor">
              <Editor
                height="430px"
                language="cpp"
                theme="vs-dark"
                value={source}
                onChange={(value) => {
                  setSource(value ?? '')
                  rememberSource(value ?? '')
                  setError(null)
                  setJudgeError(null)
                }}
                loading={
                  <p className="editor-loading" role="status">
                    正在加载编辑器…
                  </p>
                }
                options={{
                  ariaLabel: 'C++ 源码编辑器',
                  editContext: false,
                  minimap: { enabled: false },
                  fontSize: 14,
                  lineHeight: 24,
                  fontFamily: "'Cascadia Code', Consolas, monospace",
                  padding: { top: 18, bottom: 18 },
                  automaticLayout: true,
                  scrollBeyondLastLine: false,
                  tabSize: 4,
                }}
              />
            </div>
            <div className="compile-toolbar">
              <div className="source-hint" id="source-hint">
                <p className={validation ? 'validation-error' : 'muted'}>
                  {validation ?? '准备好后，运行测试检查答案。'}
                </p>
                {(mode === 'engineer' ||
                  sourceBytes > maxSourceBytes * 0.9) && (
                  <span
                    className={
                      sourceBytes > maxSourceBytes
                        ? 'validation-error byte-count'
                        : 'byte-count'
                    }
                  >
                    {sourceBytes.toLocaleString()} / 65,536 字节
                  </span>
                )}
              </div>
              <div className="submission-actions">
                <Button
                  className="compile-button"
                  variant="outline"
                  onClick={() => void submit('compile')}
                  disabled={isBusy || !!validation}
                  aria-describedby="source-hint"
                >
                  {isCompiling ? (
                    <>
                      <span className="spinner" aria-hidden="true" />
                      编译中…
                    </>
                  ) : (
                    <>
                      <span aria-hidden="true">▷</span> 编译代码
                    </>
                  )}
                </Button>
                <Button
                  className="judge-button"
                  onClick={() => void submit('judge')}
                  disabled={isBusy || !!validation || !judgeEnabled}
                  aria-describedby="source-hint judge-service-hint"
                >
                  {isJudging ? (
                    <>
                      <span className="spinner" aria-hidden="true" />
                      测试中…
                    </>
                  ) : (
                    <>
                      <span aria-hidden="true">▷</span>运行测试
                    </>
                  )}
                </Button>
              </div>
            </div>
            <p className="judge-service-hint" id="judge-service-hint">
              {judgeEnabled
                ? problem.type === 'interactive'
                  ? '每次询问后请刷新输出（flush），再读取回复。'
                  : '运行测试，检查答案是否满足题目要求。'
                : '暂时无法运行测试，可以先检查编译。'}
            </p>
          </div>
          <section
            className="result-panel panel"
            aria-labelledby="result-title"
            aria-busy={isCompiling}
          >
            <div className="panel-heading">
              <h2 id="result-title">编译结果</h2>
            </div>
            {isCompiling ? (
              <div
                className="result-placeholder"
                role="status"
                aria-live="polite"
              >
                <span className="spinner" aria-hidden="true" />
                <p>正在编译代码，请稍候…</p>
              </div>
            ) : error ? (
              <div className="request-error" role="alert">
                <h3>编译请求未完成</h3>
                <p>{error}</p>
                <Button
                  variant="outline"
                  onClick={() => void submit('compile')}
                  disabled={isBusy || !!validation}
                >
                  重新尝试
                </Button>
              </div>
            ) : result ? (
              <CompilationFeedback
                result={result.data}
                stale={result.source !== source}
              />
            ) : (
              <div className="result-placeholder">
                <span className="placeholder-icon" aria-hidden="true">
                  {'{}'}
                </span>
                <div>
                  <p>准备好后，点击「编译代码」。</p>
                  <span>编译状态和诊断信息会显示在这里。</span>
                </div>
              </div>
            )}
          </section>
          <section
            className="result-panel judge-panel panel"
            aria-labelledby="judge-result-title"
            aria-busy={isJudging}
          >
            <div className="panel-heading">
              <h2 id="judge-result-title">测试结果</h2>
              {problem.judge && mode !== 'student' && (
                <span>
                  {problem.judge.case_count} 个
                  {problem.type === 'interactive' ? '交互' : '文件'}用例
                </span>
              )}
            </div>
            {isJudging ? (
              <div
                className="result-placeholder"
                role="status"
                aria-live="polite"
              >
                <span className="spinner" aria-hidden="true" />
                <div>
                  <p>正在检查你的答案，请稍候…</p>
                  {mode === 'engineer' && (
                    <span>每个用例单独运行，耗时可能包含启动开销。</span>
                  )}
                </div>
              </div>
            ) : judgeError ? (
              <div className="request-error" role="alert">
                <h3>测试请求未完成</h3>
                <p>{judgeError}</p>
                <Button
                  variant="outline"
                  onClick={() => void submit('judge')}
                  disabled={isBusy || !!validation || !judgeEnabled}
                >
                  重新运行测试
                </Button>
              </div>
            ) : judgeResult && judgeResult.data.mode === mode ? (
              <JudgeFeedback
                result={judgeResult.data}
                stale={judgeResult.source !== source}
              />
            ) : (
              <div className="result-placeholder">
                <span className="placeholder-icon" aria-hidden="true">
                  ✓
                </span>
                <div>
                  <p>
                    {judgeEnabled
                      ? '点击「运行测试」验证你的答案。'
                      : '暂时无法运行测试。'}
                  </p>
                  <span>
                    {mode === 'student'
                      ? '根据题面和公开样例，检查边界情况。'
                      : mode === 'teacher'
                        ? '显示逐例输入、答案和输出，便于讲解。'
                        : '显示逐例详情及运行证据，便于调试。'}
                  </span>
                </div>
              </div>
            )}
          </section>
        </section>
      </div>
    </>
  )
}

export default function EditorPage() {
  const { problemId = '' } = useParams()
  const location = useLocation()
  const mode = useMode()
  const [draft, setDraft] = useState<{
    id: string
    source?: string
    sources?: CommunicationSources
  } | null>(null)
  const problem = useQuery({
    queryKey: ['problem', problemId, mode],
    queryFn: ({ signal }) => getProblem(problemId, mode, signal),
    retry: (count, error) =>
      !(error instanceof APIError && error.status === 404) && count < 1,
  })
  if (problem.isPending)
    return (
      <div className="load-card" role="status">
        正在加载题目…
      </div>
    )
  if (problem.isError) {
    const missing =
      problem.error instanceof APIError && problem.error.status === 404
    return (
      <div className="load-card error-card" role="alert">
        <h1>{missing ? '题目不存在' : '题目暂时无法加载'}</h1>
        <p>{problem.error.message}</p>
        <div className="error-actions">
          <Button asChild variant="outline">
            <Link to="/">返回题库</Link>
          </Button>
          {!missing && (
            <Button
              onClick={() => void problem.refetch()}
              disabled={problem.isFetching}
            >
              {problem.isFetching ? '正在重试…' : '重新加载'}
            </Button>
          )}
        </div>
      </div>
    )
  }
  return (
    <>
      {location.state?.uploaded === true && (
        <p className="upload-success" role="status">
          题目已发布。请运行测试，验证题目数据和判定规则。
        </p>
      )}
      {problem.data.type === 'communication' ? (
        <CommunicationWorkspace
          key={`${problem.data.id}:${mode}`}
          problem={problem.data}
          initialSources={
            draft?.id === problemId && draft.sources
              ? draft.sources
              : (problem.data.starter_codes ?? {
                  sender: problem.data.starter_code,
                  receiver: '',
                })
          }
          rememberSources={(sources) => setDraft({ id: problemId, sources })}
        />
      ) : (
        <ProblemWorkspace
          key={`${problem.data.id}:${mode}`}
          problem={problem.data}
          initialSource={
            draft?.id === problemId && draft.source !== undefined
              ? draft.source
              : problem.data.starter_code
          }
          rememberSource={(source) => setDraft({ id: problemId, source })}
        />
      )}
    </>
  )
}
