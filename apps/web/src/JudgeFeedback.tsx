import type {
  CompileEvidence,
  EngineerResult,
  JudgeCase,
  JudgeResult,
  TeacherCase,
  TeacherCommunication,
  CommunicationEvidence,
  Verdict,
} from '@/lib/api'
import { useMode } from '@/lib/mode'
import { formatMemory } from '@/lib/utils'
import { communicationRoles, roleLabels } from '@/lib/communication'

const verdictLabels: Record<Verdict, string> = {
  AC: '通过',
  WA: '答案错误',
  TLE: '运行超时',
  MLE: '内存超限',
  CE: '编译错误',
  RE: '运行错误',
  OLE: '输出超限',
  SYSTEM_ERROR: '系统异常',
}

function Output({
  title,
  content,
  label,
}: {
  title: string
  content: string
  label: string
}) {
  return (
    <div className="diagnostics">
      <div className="diagnostics-heading">{title}</div>
      <pre tabIndex={0} aria-label={label}>
        {content || '（无内容）'}
      </pre>
    </div>
  )
}

function CompileStage({
  result,
  title = '编译结果',
  engineering = false,
}: {
  result: CompileEvidence
  title?: string
  engineering?: boolean
}) {
  const labels = {
    success: '通过',
    compile_error: '编译错误',
    system_error: '系统异常',
  }
  return (
    <div className="judge-compile-stage">
      <p>
        {title}：{labels[result.status]}{' '}
        {engineering && <span>{result.duration_ms} ms</span>}
      </p>
      {result.diagnostics && (
        <Output
          title="编译器输出"
          content={result.diagnostics}
          label={`${title}诊断`}
        />
      )}
      {result.diagnostics_truncated && (
        <p className="truncated-notice">编译器输出展示已截断。</p>
      )}
    </div>
  )
}

function RuntimeEvidence({ item }: { item: JudgeCase }) {
  if (item.communication)
    return (
      <dl className="case-metrics">
        {communicationRoles.map((role) => {
          const evidence = item.communication![role]
          return (
            <div className="role-runtime-evidence" key={role}>
              <dt>{roleLabels[role]}运行证据</dt>
              <dd>
                退出码：{evidence.exit_code ?? '未获得'}；耗时：
                {evidence.duration_ms} ms
              </dd>
              <dd>
                OOMKilled：{evidence.oom_killed ? '是' : '否'}；监督终止：
                {evidence.terminated_by_supervisor ? '是' : '否'}
              </dd>
            </div>
          )
        })}
        <div>
          <dt>CPU 时间</dt>
          <dd>未采集</dd>
        </div>
        <div>
          <dt>实际聚合内存峰值</dt>
          <dd>未采集</dd>
        </div>
      </dl>
    )
  return (
    <dl className="case-metrics">
      <div>
        <dt>CPU 时间</dt>
        <dd>
          {item.cpu_time_ms === null ? '未采集' : `${item.cpu_time_ms} ms`}
        </dd>
      </div>
      <div>
        <dt>内存峰值</dt>
        <dd>
          {item.peak_memory_bytes === null
            ? '未采集'
            : formatMemory(item.peak_memory_bytes)}
        </dd>
      </div>
      <div>
        <dt>选手退出码</dt>
        <dd>{item.exit_code === null ? '未获得' : item.exit_code}</dd>
      </div>
      <div>
        <dt>选手 OOMKilled 事件</dt>
        <dd className={item.oom_killed ? 'oom-evidence' : undefined}>
          {item.oom_killed ? '是' : '否'}
        </dd>
      </div>
      {item.interactor && (
        <>
          <div>
            <dt>交互器退出码</dt>
            <dd>
              {item.interactor.exit_code === null
                ? '未获得'
                : item.interactor.exit_code}
            </dd>
          </div>
          <div>
            <dt>交互器 OOMKilled 事件</dt>
            <dd>{item.interactor.oom_killed ? '是' : '否'}</dd>
          </div>
          <div>
            <dt>协议拒绝后监督终止选手</dt>
            <dd>{item.interactor.participant_stopped ? '是' : '否'}</dd>
          </div>
        </>
      )}
      {item.checker && (
        <>
          <div>
            <dt>特判退出码</dt>
            <dd>{item.checker.exit_code ?? '未获得'}</dd>
          </div>
          <div>
            <dt>特判耗时</dt>
            <dd>{item.checker.duration_ms} ms</dd>
          </div>
          <div>
            <dt>特判 OOMKilled 事件</dt>
            <dd>{item.checker.oom_killed ? '是' : '否'}</dd>
          </div>
        </>
      )}
    </dl>
  )
}

function CommunicationOutput({
  item,
  engineering,
}: {
  item: TeacherCase | JudgeCase
  engineering: boolean
}) {
  const communication = item.communication!
  const full = communication as CommunicationEvidence
  const teacher = communication as TeacherCommunication
  const sender = engineering
    ? full.sender
    : {
        stdout: teacher.sender_stdout,
        stderr: teacher.sender_stderr,
        output_truncated: teacher.sender_output_truncated,
      }
  const receiver = engineering
    ? full.receiver
    : {
        stdout: item.stdout,
        stderr: teacher.receiver_stderr,
        output_truncated: teacher.receiver_output_truncated,
      }
  return (
    <div className="communication-output">
      <dl className="communication-count">
        <div>
          <dt>通信记录</dt>
          <dd>{communication.bytes_observed.toLocaleString()} 字节</dd>
        </div>
        <div>
          <dt>转发字节</dt>
          <dd>
            {communication.bytes_transferred.toLocaleString()} /{' '}
            {communication.byte_limit.toLocaleString()}
          </dd>
        </div>
      </dl>
      {communication.limit_exceeded && (
        <p className="truncated-notice">本例超过通信限额。</p>
      )}
      {engineering && (
        <p className="direction-note">
          计量使用原始字节；转发量表示成功写入受控输入管道的字节数，不代表接收端实际消费量。
        </p>
      )}
      <Output
        title="发送端 → 接收端（消息）"
        content={sender.stdout}
        label={`${item.id} 发送端消息`}
      />
      <Output
        title="接收端最终答案"
        content={receiver.stdout}
        label={`${item.id} 接收端最终答案`}
      />
      <Output
        title={engineering ? '发送端标准错误 stderr' : '发送端错误输出'}
        content={sender.stderr}
        label={`${item.id} 发送端错误输出`}
      />
      <Output
        title={engineering ? '接收端标准错误 stderr' : '接收端错误输出'}
        content={receiver.stderr}
        label={`${item.id} 接收端错误输出`}
      />
      {sender.output_truncated && (
        <p className="truncated-notice">发送端输出展示已截断。</p>
      )}
      {receiver.output_truncated && (
        <p className="truncated-notice">接收端输出展示已截断。</p>
      )}
    </div>
  )
}

function Trace({ result }: { result: EngineerResult }) {
  const values = [
    ['运行 ID', result.run_id],
    ['源码 SHA256', result.source_sha256],
    ['测试数据 SHA256', result.testdata_sha256],
    ['选手产物 SHA256', result.artifact_sha256],
    ['运行镜像', result.runtime_image],
    ...(result.role_compiles?.flatMap((item) => [
      [`${roleLabels[item.role]}源码 SHA256`, item.source_sha256],
      [`${roleLabels[item.role]}产物 SHA256`, item.artifact_sha256],
    ]) ?? []),
    ...(result.interactor_source_sha256
      ? [['交互器源码 SHA256', result.interactor_source_sha256]]
      : []),
    ...(result.interactor_artifact_sha256
      ? [['交互器产物 SHA256', result.interactor_artifact_sha256]]
      : []),
    ...(result.checker_source_sha256
      ? [['特判源码 SHA256', result.checker_source_sha256]]
      : []),
    ...(result.checker_artifact_sha256
      ? [['特判产物 SHA256', result.checker_artifact_sha256]]
      : []),
  ]
  return (
    <details className="debug-trace">
      <summary>工程运行证据</summary>
      <dl>
        {values.map(([label, value]) => (
          <div key={label}>
            <dt>{label}</dt>
            <dd>{value || '未生成'}</dd>
          </div>
        ))}
      </dl>
      <p className="muted">
        退出码和 OOM 事件来自运行事实；未采集的资源数值不做估计。
      </p>
    </details>
  )
}

export default function JudgeFeedback({
  result,
  stale,
}: {
  result: JudgeResult
  stale: boolean
}) {
  const mode = useMode()
  if (result.mode !== mode) return null
  return (
    <div className="judge-result">
      {stale && (
        <p className="stale-notice" role="status">
          代码已修改：此测试结果对应修改前的代码，请重新运行测试。
        </p>
      )}
      <div className="judge-result-heading" role="status" aria-live="polite">
        <span className={`verdict-badge verdict-${result.verdict}`}>
          {result.verdict}
        </span>
        <div>
          <h3>{verdictLabels[result.verdict]}</h3>
          <p>{result.message}</p>
        </div>
      </div>
      {mode === 'engineer' && (
        <p className="judge-meta">
          总耗时 {result.duration_ms} ms · {result.compiler} · {result.standard}
        </p>
      )}
      <div className="judge-summary-count">
        <span>通过用例</span>
        <strong>
          {result.passed_cases} / {result.total_cases}
        </strong>
      </div>
      {result.role_compiles?.map(
        (item) =>
          (mode !== 'student' ||
            item.compile.status !== 'success' ||
            item.compile.diagnostics) && (
            <CompileStage
              key={item.role}
              result={item.compile}
              title={`${roleLabels[item.role]}编译结果`}
              engineering={mode === 'engineer'}
            />
          ),
      )}
      {!result.role_compiles?.length &&
        (mode === 'engineer' ||
          result.compile.status !== 'success' ||
          result.compile.diagnostics) && (
          <CompileStage
            result={result.compile}
            title={mode === 'engineer' ? '选手编译阶段' : '编译结果'}
            engineering={mode === 'engineer'}
          />
        )}
      {result.mode === 'engineer' && result.interactor_compile && (
        <CompileStage
          title="交互器编译阶段"
          result={result.interactor_compile}
          engineering
        />
      )}
      {result.mode === 'engineer' && result.checker_compile && (
        <CompileStage
          title="特判编译阶段"
          result={result.checker_compile}
          engineering
        />
      )}
      {result.mode === 'student' ? (
        <p className="student-feedback-note">
          {result.verdict === 'AC'
            ? '答案通过了全部测试。'
            : '请根据题面、公开样例和自己的代码，检查边界情况。'}
        </p>
      ) : result.cases.length === 0 ? (
        <p className="no-case-results">本次没有测试用例运行结果。</p>
      ) : (
        <div className="case-results">
          <div className="case-list-heading">
            <h4>逐例结果</h4>
            <span>{result.cases.length} 个已运行用例</span>
          </div>
          {result.cases.map((item, index) => (
            <details
              className="case-result"
              key={item.id}
              open={item.verdict !== 'AC'}
            >
              <summary>
                <span className="case-index">
                  {String(index + 1).padStart(2, '0')}
                </span>
                <span className="case-name">{item.id}</span>
                <span className={`verdict-badge verdict-${item.verdict}`}>
                  {item.verdict} · {verdictLabels[item.verdict]}
                </span>
                <span className="case-duration">{item.duration_ms} ms</span>
                <span className="case-disclosure" aria-hidden="true">
                  ⌄
                </span>
              </summary>
              <div className="case-detail">
                {mode === 'engineer' && (
                  <p className="case-wall">
                    总耗时 {item.duration_ms} ms（含启动开销）
                  </p>
                )}
                {mode === 'engineer' && item.start_timed_out && (
                  <p className="truncated-notice">
                    {result.problem_type === 'communication'
                      ? '至少一个角色未完成启动；已运行情况请查看下方角色证据。'
                      : '容器启动阶段超过用例时间限制，程序尚未开始运行。'}
                  </p>
                )}
                {result.mode === 'engineer' && (
                  <RuntimeEvidence item={item as JudgeCase} />
                )}
                {item.checker_feedback && (
                  <Output
                    title="特判反馈"
                    content={item.checker_feedback}
                    label={`${item.id} 特判反馈`}
                  />
                )}
                {result.mode === 'engineer' && (item as JudgeCase).checker && (
                  <div className="checker-output">
                    <Output
                      title="特判标准输出 stdout"
                      content={(item as JudgeCase).checker!.stdout}
                      label={`${item.id} 特判标准输出`}
                    />
                    <Output
                      title="特判标准错误 stderr"
                      content={(item as JudgeCase).checker!.stderr}
                      label={`${item.id} 特判标准错误`}
                    />
                    {(item as JudgeCase).checker!.output_truncated && (
                      <p className="truncated-notice">特判输出展示已截断。</p>
                    )}
                  </div>
                )}
                <div className="case-output">
                  <Output
                    title={
                      result.problem_type === 'interactive'
                        ? '隐藏秘密输入（仅供教学 / 调试）'
                        : '测试输入'
                    }
                    content={item.input}
                    label={`${item.id} 测试输入`}
                  />
                  {item.input_truncated && (
                    <p className="truncated-notice">
                      测试输入展示已截断，判题使用完整原始数据。
                    </p>
                  )}
                  {result.problem_type !== 'interactive' && (
                    <Output
                      title="标准答案"
                      content={item.expected}
                      label={`${item.id} 标准答案`}
                    />
                  )}
                  {item.expected_truncated && (
                    <p className="truncated-notice">
                      标准答案展示已截断，比较使用完整原始数据。
                    </p>
                  )}
                  {item.communication ? (
                    <CommunicationOutput
                      item={item}
                      engineering={result.mode === 'engineer'}
                    />
                  ) : (
                    <>
                      <Output
                        title={
                          item.interactor
                            ? '选手 → 交互器（stdout）'
                            : '选手标准输出 stdout'
                        }
                        content={item.stdout}
                        label={`${item.id} 选手标准输出`}
                      />
                      <Output
                        title={
                          mode === 'engineer'
                            ? '选手标准错误 stderr'
                            : '程序错误输出'
                        }
                        content={item.stderr}
                        label={`${item.id} 选手标准错误`}
                      />
                      {item.interactor && (
                        <>
                          <Output
                            title="交互器 → 选手（stdout）"
                            content={item.interactor.stdout}
                            label={`${item.id} 交互器标准输出`}
                          />
                          <Output
                            title="交互器标准错误 stderr"
                            content={item.interactor.stderr}
                            label={`${item.id} 交互器标准错误`}
                          />
                          <p className="direction-note">
                            按两个方向展示通信内容，不代表完整交错时序。
                          </p>
                          {item.interactor.output_truncated && (
                            <p className="truncated-notice">
                              交互器输出展示已截断。
                            </p>
                          )}
                        </>
                      )}
                    </>
                  )}
                </div>
                {!item.communication && item.output_truncated && (
                  <p className="truncated-notice">
                    选手输出展示已截断，是否输出超限请以 OLE 判定为准。
                  </p>
                )}
              </div>
            </details>
          ))}
        </div>
      )}
      {result.mode === 'engineer' && <Trace result={result} />}
      {mode === 'engineer' && (
        <p className="judge-result-footnote">
          总耗时包含启动开销，不等于程序 CPU 时间。
        </p>
      )}
    </div>
  )
}
