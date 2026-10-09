import { useEffect, useMemo, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import Editor from '@monaco-editor/react'
import { Button } from '@/components/ui/button'
import {
  judgeCommunication,
  type CommunicationSources,
  type JudgeResult,
  type Problem,
} from '@/lib/api'
import { useMode } from '@/lib/mode'
import { formatMemory } from '@/lib/utils'
import {
  buildSourceUnits,
  communicationRoles,
  roleLabels,
  sourceBytes,
  sourcesMatch,
} from '@/lib/communication'
import JudgeFeedback from './JudgeFeedback'

export default function CommunicationWorkspace({
  problem,
  initialSources,
  rememberSources,
}: {
  problem: Problem
  initialSources: CommunicationSources
  rememberSources: (sources: CommunicationSources) => void
}) {
  const mode = useMode()
  const [sources, setSources] = useState(initialSources)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<{
    data: JudgeResult
    sources: CommunicationSources
  } | null>(null)
  const controller = useRef<AbortController | null>(null)
  const starters = problem.starter_codes ?? {
    sender: problem.starter_code,
    receiver: '',
  }
  const canReset = !sourcesMatch(sources, starters) || !!result || !!error
  const validation = useMemo(() => {
    try {
      buildSourceUnits(sources)
      return null
    } catch (failure) {
      return failure instanceof Error
        ? failure.message
        : '请检查两份 C++ 代码。'
    }
  }, [sources])
  useEffect(() => () => controller.current?.abort(), [])

  async function submit() {
    if (controller.current || validation || !problem.judge.enabled) return
    const requestController = new AbortController()
    controller.current = requestController
    const submittedSources = { ...sources }
    setBusy(true)
    setError(null)
    setResult(null)
    try {
      const data = await judgeCommunication(
        problem.id,
        buildSourceUnits(submittedSources),
        mode,
        requestController.signal,
      )
      if (!requestController.signal.aborted && data.mode === mode)
        setResult({ data, sources: submittedSources })
    } catch (failure) {
      if (!requestController.signal.aborted)
        setError(
          failure instanceof Error
            ? failure.message
            : '测试请求未能完成，请重试。',
        )
    } finally {
      if (controller.current === requestController) {
        controller.current = null
        setBusy(false)
      }
    }
  }

  function resetSources() {
    if (
      busy ||
      !canReset ||
      !window.confirm(
        '恢复模板将替换发送端和接收端两份代码，并清空测试结果，确定继续吗？',
      )
    )
      return
    const next = { ...starters }
    setSources(next)
    rememberSources(next)
    setResult(null)
    setError(null)
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
        <span className="problem-type type-communication">通信题</span>
      </div>
      <div className="workspace-grid">
        <article className="statement panel" aria-labelledby="statement-title">
          <div className="panel-heading">
            <h2 id="statement-title">题目说明</h2>
            <span className="muted">C++20 · 双角色</span>
          </div>
          <div className="statement-body">
            <section>
              <h3>题目描述</h3>
              <p>{problem.description}</p>
            </section>
            <section>
              <h3>输入与通信格式</h3>
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
                    <span>发送端输入</span>
                    <pre>{sample.input}</pre>
                  </div>
                  <div className="sample-box">
                    <span>接收端输出</span>
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
            <section
              className="judge-limits"
              aria-labelledby="judge-limits-title"
            >
              <h3 id="judge-limits-title">题目限制</h3>
              <dl>
                <div>
                  <dt>整例时间限制</dt>
                  <dd>{problem.judge.wall_time_ms} ms</dd>
                </div>
                <div>
                  <dt>每角色内存限制</dt>
                  <dd>{formatMemory(problem.judge.memory_limit_bytes)}</dd>
                </div>
                <div>
                  <dt>每例通信限制</dt>
                  <dd>
                    {problem.judge.communication?.byte_limit.toLocaleString() ??
                      '未提供'}{' '}
                    字节
                  </dd>
                </div>
              </dl>
            </section>
            <div className="statement-notice">
              <span className="status-dot" />
              <p>
                发送端读取题目输入，将消息写到标准输出；接收端从标准输入读取消息，输出最终答案。通信仅从发送端到接收端，发送端结束后接收端会读到
                EOF。
              </p>
            </div>
            {mode === 'engineer' && (
              <section className="engineering-notes">
                <h3>工程说明</h3>
                {problem.engineering_notes && (
                  <p>{problem.engineering_notes}</p>
                )}
                <p>
                  固定 sender → receiver 图；{problem.judge.checker}{' '}
                  按空白分隔的 token
                  比较接收端答案。两份源码分别编译，全部成功后才运行测试。
                </p>
                <p>
                  双方在独立受限容器中执行，共享墙钟截止时间，包含启动开销。代理统计原始字节，转发量不等于接收端程序实际消费量。
                </p>
                <p>
                  双方内存限额上界：
                  {problem.judge.communication
                    ? formatMemory(
                        problem.judge.communication
                          .aggregate_memory_limit_bytes,
                      )
                    : '未提供'}
                  ；未采集实际聚合内存峰值。
                </p>
              </section>
            )}
          </div>
        </article>
        <section
          className="coding-column communication-coding"
          aria-labelledby="code-title"
        >
          <div className="code-panel panel">
            <div className="panel-heading editor-heading">
              <h2 id="code-title">双角色代码</h2>
              <Button
                variant="outline"
                className="reset-button"
                onClick={resetSources}
                disabled={busy || !canReset}
              >
                恢复双模板
              </Button>
            </div>
            {communicationRoles.map((role) => (
              <div className="role-editor" key={role}>
                <div className="editor-filebar">
                  <span>
                    {roleLabels[role]} · {role}.cpp
                  </span>
                  <span>C++20</span>
                </div>
                <p className="role-requirement">
                  {role === 'sender'
                    ? '读取题目输入，输出给接收端的消息。'
                    : '只读取发送端消息，输出题目要求的答案。'}
                </p>
                <Editor
                  height="300px"
                  language="cpp"
                  theme="vs-dark"
                  path={`${problem.id}/${role}.cpp`}
                  value={sources[role]}
                  onChange={(value) => {
                    const next = { ...sources, [role]: value ?? '' }
                    setSources(next)
                    rememberSources(next)
                    setError(null)
                  }}
                  loading={
                    <p className="editor-loading" role="status">
                      正在加载{roleLabels[role]}编辑器…
                    </p>
                  }
                  options={{
                    ariaLabel: `${roleLabels[role]} C++ 源码编辑器`,
                    editContext: false,
                    minimap: { enabled: false },
                    fontSize: 14,
                    lineHeight: 24,
                    fontFamily: "'Cascadia Code', Consolas, monospace",
                    padding: { top: 14, bottom: 14 },
                    automaticLayout: true,
                    scrollBeyondLastLine: false,
                    tabSize: 4,
                  }}
                />
                {(mode === 'engineer' ||
                  sourceBytes(sources[role]) > 65536 * 0.9) && (
                  <p
                    className={`role-byte-count ${sourceBytes(sources[role]) > 65536 ? 'validation-error' : 'muted'}`}
                  >
                    {roleLabels[role]}：
                    {sourceBytes(sources[role]).toLocaleString()} / 65,536 字节
                  </p>
                )}
              </div>
            ))}
            <div className="compile-toolbar">
              <div className="source-hint" id="source-hint">
                <p className={validation ? 'validation-error' : 'muted'}>
                  {validation ?? '运行测试会先分别编译两份代码。'}
                </p>
              </div>
              <div className="submission-actions">
                {busy && (
                  <Button
                    variant="outline"
                    onClick={() => {
                      controller.current?.abort()
                      setError('测试已取消，可以修改代码后重新运行。')
                    }}
                  >
                    取消测试
                  </Button>
                )}
                <Button
                  className="judge-button"
                  onClick={() => void submit()}
                  disabled={busy || !!validation || !problem.judge.enabled}
                  aria-describedby="source-hint judge-service-hint"
                >
                  {busy ? (
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
              {problem.judge.enabled
                ? '检查两份代码能否共同完成题目，并满足通信限制。'
                : '暂时无法运行测试，可以先编写两份代码。'}
            </p>
          </div>
          <section
            className="result-panel judge-panel panel"
            aria-labelledby="judge-result-title"
            aria-busy={busy}
          >
            <div className="panel-heading">
              <h2 id="judge-result-title">测试与编译结果</h2>
              {mode !== 'student' && (
                <span>{problem.judge.case_count} 个通信用例</span>
              )}
            </div>
            {busy ? (
              <div
                className="result-placeholder"
                role="status"
                aria-live="polite"
              >
                <span className="spinner" aria-hidden="true" />
                <p>正在检查两份代码，请稍候…</p>
              </div>
            ) : error ? (
              <div className="request-error" role="alert">
                <h3>测试请求未完成</h3>
                <p>{error}</p>
                <Button
                  variant="outline"
                  onClick={() => void submit()}
                  disabled={!!validation || !problem.judge.enabled}
                >
                  重新运行测试
                </Button>
              </div>
            ) : result && result.data.mode === mode ? (
              <JudgeFeedback
                result={result.data}
                stale={!sourcesMatch(result.sources, sources)}
              />
            ) : (
              <div className="result-placeholder">
                <span className="placeholder-icon" aria-hidden="true">
                  ✓
                </span>
                <div>
                  <p>准备好两份代码后，点击「运行测试」。</p>
                  <span>编译诊断会标明发送端或接收端。</span>
                </div>
              </div>
            )}
          </section>
        </section>
      </div>
    </>
  )
}
