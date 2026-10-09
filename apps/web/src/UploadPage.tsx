import {
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  type FormEvent,
} from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { Link, useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import {
  uploadProblem,
  type BatchChecker,
  type ProblemType,
  type UploadDescriptor,
} from '@/lib/api'
import { useMode } from '@/lib/mode'
import {
  buildSourceUnits,
  communicationExample,
  communicationAdditionExample,
  communicationAdditionSample,
  validateCommunicationLimit,
} from '@/lib/communication'
import {
  buildCases,
  checkerExample,
  constructionRule,
  isConstructionInput,
  isInteractiveInput,
  validateCheckerSource,
} from '@/lib/specialJudge'

const batchStarter =
  '#include <iostream>\nusing namespace std;\n\nint main() {\n    ios::sync_with_stdio(false);\n    cin.tie(nullptr);\n    // 读取输入，编写解法并输出答案。\n    return 0;\n}\n'
const interactiveStarter =
  '#include <iostream>\n#include <string>\nusing namespace std;\n\nint main() {\n    int upper;\n    if (!(cin >> upper)) return 0;\n    int low = 1, high = upper;\n    while (low <= high) {\n        int mid = (low + high) / 2;\n        cout << "? " << mid << endl; // endl 刷新输出\n        string reply;\n        if (!(cin >> reply)) return 0;\n        if (reply == "=") {\n            cout << "! " << mid << endl;\n            return 0;\n        }\n        if (reply == "<") high = mid - 1;\n        else if (reply == ">") low = mid + 1;\n        else return 0;\n    }\n    return 0;\n}\n'
const textFields = [
  { key: 'title', label: '题目名称', limit: 256, rows: 1 },
  { key: 'summary', label: '题库简介', limit: 1024, rows: 2 },
  { key: 'difficulty', label: '难度', limit: 64, rows: 1 },
  { key: 'description', label: '题目描述', limit: 16384, rows: 6 },
  { key: 'input_description', label: '输入 / 通信格式', limit: 8192, rows: 4 },
  { key: 'output_description', label: '输出 / 回答格式', limit: 8192, rows: 4 },
  { key: 'notes', label: '提示（可选）', limit: 8192, rows: 3 },
  {
    key: 'engineering_notes',
    label: '工程说明（可选，仅工程师可见）',
    limit: 8192,
    rows: 3,
  },
  { key: 'starter_code', label: 'C++20 起始模板', limit: 65536, rows: 12 },
] as const
const encoder = new TextEncoder()
const bytes = (value: string) => encoder.encode(value).byteLength
const initialText = {
  title: '',
  summary: '',
  difficulty: '入门',
  description: '',
  input_description: '',
  output_description: '',
  notes: '',
  engineering_notes: '',
  starter_code: batchStarter,
}

export default function UploadPage() {
  const mode = useMode()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [type, setType] = useState<ProblemType>('batch')
  const [checker, setChecker] = useState<BatchChecker>('tokens-v1')
  const [checkerSource, setCheckerSource] = useState('')
  const [checkerFileName, setCheckerFileName] = useState<string | null>(null)
  const [readingChecker, setReadingChecker] = useState(false)
  const checkerReadVersion = useRef(0)
  const checkerDownloadLink = useRef<HTMLAnchorElement | null>(null)
  const [text, setText] = useState(initialText)
  const [samples, setSamples] = useState([{ input: '', output: '' }])
  const [wall, setWall] = useState('2000')
  const [memory, setMemory] = useState('64')
  const [communicationLimit, setCommunicationLimit] = useState('32')
  const [receiverStarter, setReceiverStarter] = useState('')
  const [files, setFiles] = useState<File[]>([])
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const controller = useRef<AbortController | null>(null)
  useEffect(
    () => () => {
      controller.current?.abort()
      checkerReadVersion.current++
    },
    [],
  )
  const custom = type === 'batch' && checker === 'cpp-special-v1'
  const communication = type === 'communication'
  const construction =
    type === 'batch' && checker === 'nonconsecutive-permutation-v1'
  let checkerSourceValid = true
  try {
    validateCheckerSource(checkerSource)
  } catch {
    checkerSourceValid = false
  }
  const checkerDownloadEnabled =
    custom && checkerSourceValid && !readingChecker && !busy
  useLayoutEffect(() => {
    const link = checkerDownloadLink.current
    if (!link || !checkerDownloadEnabled) return
    const url = URL.createObjectURL(
      new Blob([checkerSource], { type: 'text/x-c++src;charset=utf-8' }),
    )
    // Update before paint so an edited source cannot download an older Blob.
    link.href = url
    return () => {
      link.removeAttribute('href')
      URL.revokeObjectURL(url)
    }
  }, [checkerSource, checkerDownloadEnabled])
  let fileSummary: string | null = null
  let fileError: string | null = null
  if (files.length > 0) {
    try {
      fileSummary = `${buildCases(files, type, checker).length} 组测试数据，${files.reduce((sum, file) => sum + file.size, 0).toLocaleString()} 字节`
    } catch (failure) {
      fileError =
        failure instanceof Error ? failure.message : '文件不符合要求。'
    }
  }

  function changeType(value: ProblemType) {
    checkerReadVersion.current++
    setReadingChecker(false)
    setType(value)
    setWall(value === 'batch' ? '2000' : '5000')
    setError(null)
  }

  function fillProtocol() {
    setText((current) => ({
      ...current,
      title: current.title || '猜数练习',
      summary: current.summary || '用至多十次询问找出 1 到 1000 之间的秘密数。',
      description:
        '交互器持有一个 1 到 1000 之间的秘密整数。通过至多 10 次询问确定秘密数，并提交最终答案。',
      input_description:
        '先读取交互器给出的上界 1000。每次询问 ? x 后，读取 <、> 或 =：分别表示秘密数小于、大于或等于 x。',
      output_description:
        '输出 ? x 询问（1 ≤ x ≤ 1000），至多 10 次。输出 ! x 提交最终答案。每行输出后必须刷新缓冲区。',
      notes: '每行输出后必须刷新缓冲区，提交最终答案后结束程序。',
      engineering_notes:
        '固定 guess-number-v1 协议。隐藏 .in 中存放一个 1..1000 的秘密整数，不直接作为选手输入；默认代码是完整二分查找演示解法。',
      starter_code: interactiveStarter,
    }))
    setSamples([{ input: '1000\n>\n=\n', output: '? 500\n? 750\n! 750\n' }])
  }

  function fillConstruction() {
    setText((current) => ({
      ...current,
      title: current.title || '构造不相邻排列',
      summary: current.summary || '构造相邻数字之差不为 1 的排列。',
      description: constructionRule,
      input_description: '输入一行，一个整数 n（1 ≤ n ≤ 1000）。',
      output_description:
        '有解时输出一个满足要求的排列，用空格分隔；无解时只输出 -1。',
      notes: '公开样例只是一种合法答案；满足条件的其他排列同样正确。',
      engineering_notes:
        '使用 nonconsecutive-permutation-v1 内置规则或同规则 C++ 特判；不以固定输出顺序比较答案。',
      starter_code: batchStarter,
    }))
    setSamples([
      { input: '4\n', output: '2 4 1 3\n' },
      { input: '3\n', output: '-1\n' },
    ])
  }

  function fillCommunication() {
    setText((current) => ({
      ...current,
      title: current.title || '通信求和',
      summary: current.summary || '发送端压缩整数序列，接收端还原它们的和。',
      description:
        '发送端能读取一个整数序列，接收端只能读取发送端的消息。协作计算全部整数的和，每例从发送端到接收端的消息最多 32 字节。',
      input_description:
        '发送端输入：第一行 n（1 ≤ n ≤ 1000），随后 n 个整数（绝对值不超过 10^9）。接收端标准输入仅包含发送端标准输出的消息；发送端退出后会收到 EOF，没有反向通信。',
      output_description:
        '接收端输出一个整数，表示序列中全部整数的和。发送端标准输出只用于发送消息。',
      notes:
        '换行也计入通信字节。两个角色都应正常结束；用 long long 保存求和结果。',
      engineering_notes:
        '固定 sender → receiver 单向图，receiver 答案按 token 比较。示例发送十进制求和结果加换行，不转发整个数组；原始字节计量与展示长度不同。',
      starter_code: communicationExample.sender,
    }))
    setReceiverStarter(communicationExample.receiver)
    setCommunicationLimit('32')
    setSamples([{ input: '4\n1 -2 3 8\n', output: '10\n' }])
    setError(null)
  }

  function fillCommunicationAddition() {
    setText((current) => ({
      ...current,
      title: current.title || '通信 A+B',
      summary: current.summary || '发送端传两个整数，接收端计算它们的和。',
      description:
        '发送端读取两个整数 a 和 b，将两个数传给接收端。接收端从消息中读取 a 和 b，计算并输出 a+b。每例通信最多 32 字节。',
      input_description:
        '发送端输入：两个整数 a、b（绝对值不超过 10^9）。示例发送端输出 a、一个空格、b 和换行。接收端只从发送端消息中读取两个数，不能直接读取原题输入。',
      output_description:
        '接收端输出一个整数 a+b 和换行。发送端负责传数，接收端负责求和；两个程序正常结束。',
      notes:
        '消息里的空格和换行也占通信字节。没有反向通信，发送端退出后接收端收到 EOF。',
      engineering_notes:
        '固定 sender → receiver 单向图，接收端答案按 token 比较。A+B 示例传递原始两个数，而非由发送端预先求和；消息按原始字节计量。',
      starter_code: communicationAdditionExample.sender,
    }))
    setReceiverStarter(communicationAdditionExample.receiver)
    setCommunicationLimit('32')
    setSamples([{ ...communicationAdditionSample }])
    setError(null)
  }

  function fillCheckerExample() {
    checkerReadVersion.current++
    setReadingChecker(false)
    setCheckerSource(checkerExample)
    setCheckerFileName('内置示例 → checker.cpp')
    setError(null)
  }

  async function readChecker(file?: File) {
    const version = ++checkerReadVersion.current
    if (!file) return
    setReadingChecker(true)
    setCheckerSource('')
    setCheckerFileName(null)
    setError(null)
    try {
      if (!file.name.endsWith('.cpp')) throw new Error('请选择 .cpp 源文件。')
      if (file.size > 65536) throw new Error('特判源码超过 64 KiB，请精简。')
      const source = new TextDecoder('utf-8', { fatal: true }).decode(
        await file.arrayBuffer(),
      )
      if (version !== checkerReadVersion.current) return
      validateCheckerSource(source)
      setCheckerSource(source)
      setCheckerFileName(`${file.name} → checker.cpp`)
    } catch (failure) {
      if (version === checkerReadVersion.current)
        setError(
          failure instanceof Error
            ? failure.message
            : '无法读取特判源码，请使用 UTF-8 文件。',
        )
    } finally {
      if (version === checkerReadVersion.current) setReadingChecker(false)
    }
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (busy || readingChecker || controller.current || mode === 'student')
      return
    setError(null)
    const requestController = new AbortController()
    controller.current = requestController
    setBusy(true)
    try {
      for (const field of textFields) {
        const value = text[field.key]
        if (
          field.key !== 'notes' &&
          field.key !== 'engineering_notes' &&
          !value.trim()
        )
          throw new Error(`请填写${field.label}。`)
        if (bytes(value) > field.limit)
          throw new Error(
            `${field.label}超过 ${field.limit.toLocaleString()} 字节，请精简。`,
          )
      }
      for (const sample of samples) {
        if (bytes(sample.input) > 4096 || bytes(sample.output) > 4096)
          throw new Error('每个公开样例的输入和输出分别最多 4 KiB。')
      }
      const wallMS = Number(wall),
        memoryMiB = Number(memory)
      if (!Number.isInteger(wallMS) || wallMS < 500 || wallMS > 5000)
        throw new Error('每例耗时限额须为 500 至 5000 ms 的整数。')
      if (!Number.isInteger(memoryMiB) || memoryMiB < 32 || memoryMiB > 256)
        throw new Error('内存限额须为 32 至 256 MiB 的整数。')
      const cases = buildCases(files, type, checker)
      const byteLimit = communication
        ? validateCommunicationLimit(communicationLimit)
        : undefined
      if (communication)
        buildSourceUnits({
          sender: text.starter_code,
          receiver: receiverStarter,
        })
      if (type === 'interactive' || construction) {
        for (const file of files) {
          const secret = await file.text()
          if (requestController.signal.aborted) return
          if (
            !(construction
              ? isConstructionInput(secret)
              : isInteractiveInput(secret))
          )
            throw new Error(
              `${file.name} 必须只包含一个 1 到 1000 之间的${construction ? '整数 n' : '秘密整数'}。`,
            )
        }
      }
      const uploadFiles = [...files]
      if (custom) {
        validateCheckerSource(checkerSource)
        uploadFiles.push(
          new File([checkerSource], 'checker.cpp', { type: 'text/x-c++src' }),
        )
      }
      const descriptor: UploadDescriptor = {
        ...text,
        samples,
        type,
        checker: communication
          ? 'communication-v1'
          : type === 'interactive'
            ? 'interaction-v1'
            : checker,
        wall_time_ms: wallMS,
        memory_limit_bytes: memoryMiB * 1024 * 1024,
        cases,
        ...(type === 'interactive'
          ? { interactor: 'guess-number-v1' as const }
          : {}),
        ...(custom ? { checker_source: 'checker.cpp' as const } : {}),
        ...(communication
          ? {
              communication: { byte_limit: byteLimit! },
              starter_codes: {
                sender: text.starter_code,
                receiver: receiverStarter,
              },
            }
          : {}),
      }
      const descriptorBytes = bytes(JSON.stringify(descriptor))
      if (descriptorBytes > 65536)
        throw new Error(
          '题目描述整体超过 64 KiB，请精简题面、样例或代码模板（通信题包含双方模板）。',
        )
      const estimatedRequestBytes =
        descriptorBytes +
        uploadFiles.reduce(
          (sum, file) => sum + file.size + bytes(file.name) + 512,
          0,
        ) +
        1024
      if (estimatedRequestBytes > 1048576)
        throw new Error('题目与文件总大小接近或超过 1 MiB，请精简后上传。')
      if (requestController.signal.aborted) return
      const created = await uploadProblem(
        descriptor,
        uploadFiles,
        mode,
        requestController.signal,
      )
      if (requestController.signal.aborted) return
      await queryClient.invalidateQueries({ queryKey: ['problems'] })
      if (!requestController.signal.aborted)
        navigate(`/problems/${encodeURIComponent(created.id)}`, {
          state: { uploaded: true },
        })
    } catch (failure) {
      if (!requestController.signal.aborted)
        setError(
          failure instanceof Error ? failure.message : '上传未完成，请重试。',
        )
    } finally {
      if (controller.current === requestController) {
        controller.current = null
        if (!requestController.signal.aborted) setBusy(false)
      }
    }
  }

  return (
    <div className="upload-page">
      <Link className="back-link" to="/">
        ← 返回题库
      </Link>
      <div className="page-heading">
        <p className="eyebrow">TEACHING</p>
        <h1>上传一道题目</h1>
        <p>填写题面与公开样例，再选择本地测试数据。发布后即可在题库中练习。</p>
      </div>
      <form className="upload-form" onSubmit={(event) => void submit(event)}>
        <fieldset className="upload-section panel" disabled={busy}>
          <legend>题型与判定</legend>
          <div className="form-grid">
            <label className="form-field">
              <span>题型</span>
              <select
                value={type}
                onChange={(event) =>
                  changeType(event.target.value as ProblemType)
                }
              >
                <option value="batch">标准题 / 构造题</option>
                <option value="interactive">交互题 · 固定猜数协议</option>
                <option value="communication">通信题 · 发送端 / 接收端</option>
              </select>
            </label>
            <label className="form-field">
              <span>答案判定</span>
              {communication ? (
                <input value="单向通信 · 接收端答案比较" readOnly />
              ) : type === 'interactive' ? (
                <input value="猜数交互协议" readOnly />
              ) : (
                <select
                  value={checker}
                  onChange={(event) => {
                    checkerReadVersion.current++
                    setReadingChecker(false)
                    setChecker(event.target.value as typeof checker)
                    setError(null)
                  }}
                >
                  <option value="tokens-v1">按空白分隔比较</option>
                  <option value="exact-v1">精确字节比较</option>
                  <option value="nonconsecutive-permutation-v1">
                    内置规则 · 不相邻排列
                  </option>
                  <option value="cpp-special-v1">教师上传 C++ 特判</option>
                </select>
              )}
            </label>
            <label className="form-field">
              <span>时间限制（ms）</span>
              <input
                type="number"
                min="500"
                max="5000"
                step="1"
                required
                value={wall}
                onChange={(event) => setWall(event.target.value)}
              />
              <small>
                500–5000 ms。{mode === 'engineer' && '包含启动和通信。'}
              </small>
            </label>
            <label className="form-field">
              <span>
                {communication ? '每角色内存（MiB）' : '每例内存（MiB）'}
              </span>
              <input
                type="number"
                min="32"
                max="256"
                step="1"
                required
                value={memory}
                onChange={(event) => setMemory(event.target.value)}
              />
              <small>32–256 MiB。</small>
            </label>
            {communication && (
              <label className="form-field">
                <span>每例通信限制（字节）</span>
                <input
                  type="number"
                  min="1"
                  max="65536"
                  step="1"
                  required
                  value={communicationLimit}
                  onChange={(event) =>
                    setCommunicationLimit(event.target.value)
                  }
                />
                <small>1–65,536 字节；包含消息中的空格和换行。</small>
              </label>
            )}
          </div>
          {communication && (
            <div className="protocol-notice">
              <p>
                只支持发送端 → 接收端的单向消息。发送端读取
                .in，接收端读取发送端输出并给出答案；发送端结束后接收端收到
                EOF。每例的两角色共享时间限制，需正常退出，超出通信量判 WA。
              </p>
              <p>
                数据必须配对 .in / .out，最多 4
                例。发布后先验证双方正确代码，再测试错误答案、超通信量、无法编译和等待不结束等情况。暂不支持自定义通信图或混合
                C++ 特判。
              </p>
              <p>
                最简单的 A+B 示例：发送端只传
                a、b，接收端读取两个数并计算它们的和。也可选择整数序列求和示例，由发送端先压缩求和结果。
              </p>
              <div className="checker-actions">
                <Button
                  type="button"
                  variant="outline"
                  onClick={fillCommunicationAddition}
                >
                  填入通信 A+B 说明与双模板
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={fillCommunication}
                >
                  填入通信求和说明与双模板
                </Button>
              </div>
            </div>
          )}
          {type === 'interactive' && (
            <div className="protocol-notice">
              <p>
                交互器先发 1000；选手发 ? x 并刷新输出，收到秘密数相对 x 的 &lt;
                / &gt; / =；至多 10 问，最后发 ! x。仅支持该固定协议。
              </p>
              <Button type="button" variant="outline" onClick={fillProtocol}>
                填入猜数题说明与模板
              </Button>
            </div>
          )}
          {(construction || custom) && (
            <div className="protocol-notice construction-notice">
              <h3>构造题规则示例</h3>
              <p>{constructionRule}</p>
              <p>n = 4 时，2 4 1 3 与 3 1 4 2 都正确；1 2 3 4 错误。</p>
              <Button
                type="button"
                variant="outline"
                onClick={fillConstruction}
              >
                填入构造题说明与样例
              </Button>
              {custom && (
                <p className="checker-scope-hint">
                  下面的示例特判对应这条规则。上传自己的特判时，请同时修改题面、样例和数据，使它们描述同一套要求。
                </p>
              )}
            </div>
          )}
        </fieldset>
        {custom && (
          <fieldset
            className="upload-section checker-section panel"
            disabled={busy}
          >
            <legend>C++ 特判程序</legend>
            <p className="form-section-hint">
              特判接收输入文件、标准答案文件、选手输出文件三个路径参数（argv[1..3]）。
              返回 0 表示 AC，1 表示 WA；其他退出值、超时或程序故障表示
              SYSTEM_ERROR。 可把简短反馈写到 stderr，供教师检查。仅支持自包含
              C++20 源码，无外部依赖。
            </p>
            <p className="checker-warning" role="note">
              发布只保存源码，不代表特判已经编译通过。发布后请先用正确答案和错误答案运行测试，确认判定符合题目规则。
            </p>
            <div className="checker-actions">
              <Button
                type="button"
                variant="outline"
                onClick={fillCheckerExample}
              >
                填入示例特判
              </Button>
              <Button
                asChild
                variant="outline"
                className="text-xs aria-disabled:opacity-50"
              >
                <a
                  ref={checkerDownloadLink}
                  download="checker.cpp"
                  aria-disabled={!checkerDownloadEnabled}
                  tabIndex={checkerDownloadEnabled ? undefined : -1}
                  onClick={(event) => {
                    if (!checkerDownloadEnabled) event.preventDefault()
                  }}
                >
                  下载当前 checker.cpp
                </a>
              </Button>
            </div>
            <label className="file-picker">
              <span>读取本地 .cpp 源文件（UTF-8，最多 64 KiB）</span>
              <input
                type="file"
                accept=".cpp"
                disabled={readingChecker}
                onChange={(event) => void readChecker(event.target.files?.[0])}
              />
            </label>
            {readingChecker && (
              <p className="file-summary" role="status">
                正在读取源码…
              </p>
            )}
            {checkerFileName && (
              <p className="file-summary">{checkerFileName}</p>
            )}
            <label className="form-field checker-source">
              <span>特判源码 · checker.cpp</span>
              <textarea
                className="code-input"
                rows={20}
                value={checkerSource}
                disabled={readingChecker}
                required
                onChange={(event) => {
                  setCheckerSource(event.target.value)
                  setCheckerFileName(null)
                  setError(null)
                }}
              />
              <small
                className={
                  bytes(checkerSource) > 65536 ? 'validation-error' : undefined
                }
              >
                {bytes(checkerSource).toLocaleString()} / 65,536
                字节；下载和发布使用当前编辑内容。
              </small>
            </label>
            <details className="checker-rules">
              <summary>数据文件与特判编写规则</summary>
              <p>
                每例必须上传 .in，最多 4 例。同名 .out
                可选；缺省时传给特判一个空标准答案文件。特判应检查完整选手输出，包括多余内容。
              </p>
              <p>
                示例只读取 n
                与选手输出，验证排列的范围、唯一性和相邻差，不需要标准答案。题目数据不合法应返回
                2，不要把题目配置错误判为学生 WA。
              </p>
              <p>
                源码发布为固定名称
                checker.cpp，不接受二进制、脚本、其他头文件或自定义编译参数。特判失败会中止本次判题，请在教师模式查看反馈、在工程师模式检查编译与运行详情。
              </p>
            </details>
          </fieldset>
        )}
        <fieldset className="upload-section panel" disabled={busy}>
          <legend>题面与代码模板</legend>
          <div className="form-grid">
            {textFields.map((field) => (
              <label
                className={`form-field ${field.rows > 2 ? 'full-width' : ''}`}
                key={field.key}
              >
                <span>
                  {communication && field.key === 'starter_code'
                    ? '发送端 C++20 起始模板'
                    : field.label}
                </span>
                {field.rows === 1 ? (
                  <input
                    value={text[field.key]}
                    required
                    onChange={(event) =>
                      setText({ ...text, [field.key]: event.target.value })
                    }
                  />
                ) : (
                  <textarea
                    className={
                      field.key === 'starter_code' ? 'code-input' : undefined
                    }
                    rows={field.rows}
                    value={text[field.key]}
                    required={
                      field.key !== 'notes' && field.key !== 'engineering_notes'
                    }
                    onChange={(event) =>
                      setText({ ...text, [field.key]: event.target.value })
                    }
                  />
                )}
                <small
                  className={
                    bytes(text[field.key]) > field.limit
                      ? 'validation-error'
                      : undefined
                  }
                >
                  {bytes(text[field.key]).toLocaleString()} /{' '}
                  {field.limit.toLocaleString()} 字节
                </small>
                {field.key === 'notes' && (
                  <small>
                    学生也能看到：只填写解题提示、边界条件或协议要求。
                  </small>
                )}
              </label>
            ))}
            {communication && (
              <label className="form-field full-width">
                <span>接收端 C++20 起始模板</span>
                <textarea
                  className="code-input"
                  rows={12}
                  required
                  value={receiverStarter}
                  onChange={(event) => setReceiverStarter(event.target.value)}
                />
                <small
                  className={
                    bytes(receiverStarter) > 65536
                      ? 'validation-error'
                      : undefined
                  }
                >
                  {bytes(receiverStarter).toLocaleString()} / 65,536 字节
                </small>
                <small>
                  两份模板随题目一起保存；题目描述编码总量（含两模板）最多 64
                  KiB。
                </small>
              </label>
            )}
          </div>
        </fieldset>
        <fieldset className="upload-section panel" disabled={busy}>
          <legend>公开样例</legend>
          <p className="form-section-hint">
            公开样例会显示给所有模式。1–4 组，每项最多 4
            KiB；交互题可填写一段示例通信，通信题填写发送端输入和接收端最终答案。
          </p>
          {samples.map((sample, index) => (
            <div className="sample-form" key={index}>
              <div className="sample-form-heading">
                <h3>样例 {index + 1}</h3>
                {samples.length > 1 && (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() =>
                      setSamples(samples.filter((_, n) => n !== index))
                    }
                  >
                    移除
                  </Button>
                )}
              </div>
              <div className="form-grid">
                {(['input', 'output'] as const).map((direction) => (
                  <label className="form-field" key={direction}>
                    <span>
                      {communication
                        ? direction === 'input'
                          ? '样例发送端输入'
                          : '样例接收端答案'
                        : direction === 'input'
                          ? '样例输入 / 交互器回复'
                          : '样例输出 / 选手消息'}
                    </span>
                    <textarea
                      rows={4}
                      value={sample[direction]}
                      onChange={(event) =>
                        setSamples(
                          samples.map((value, n) =>
                            n === index
                              ? { ...value, [direction]: event.target.value }
                              : value,
                          ),
                        )
                      }
                    />
                    <small>
                      {bytes(sample[direction]).toLocaleString()} / 4096 字节
                    </small>
                  </label>
                ))}
              </div>
            </div>
          ))}
          <Button
            type="button"
            variant="outline"
            disabled={samples.length >= 4}
            onClick={() => setSamples([...samples, { input: '', output: '' }])}
          >
            增加公开样例
          </Button>
        </fieldset>
        <fieldset className="upload-section panel" disabled={busy}>
          <legend>隐藏测试数据</legend>
          <p className="form-section-hint">
            {communication
              ? '选择 1–4 组同名 .in / .out 文件；.in 仅给发送端，.out 是接收端应输出的答案。'
              : type === 'interactive'
                ? '选择 1–8 个 .in 文件，每个只放一个 1..1000 的秘密整数，不需要 .out。'
                : construction
                  ? '选择 1–8 个 .in 文件，每个只放一个整数 n（1..1000），不需要 .out。'
                  : custom
                    ? '选择 1–4 个 .in 文件，可选同名 .out；不提供 .out 时传给特判空标准答案文件。'
                    : '选择 1–8 组同名 .in / .out 文件，例如 01.in 和 01.out。'}{' '}
            文件名主体须以小写字母或数字开头，其后可用小写字母、数字、下划线或短横线（最多
            64 字符）。 不使用 con、prn、aux、nul 或 com / lpt
            开头的主体。每个文件最多 64 KiB，题目与文件合计最多 1 MiB。
          </p>
          <label className="file-picker">
            <span>选择测试数据文件</span>
            <input
              type="file"
              multiple
              accept={
                type === 'interactive' || construction ? '.in' : '.in,.out'
              }
              onChange={(event) => {
                setFiles(Array.from(event.target.files ?? []))
                setError(null)
              }}
            />
          </label>
          {files.length > 0 && (
            <ul className="selected-files">
              {files.map((file, index) => (
                <li key={`${file.name}:${index}`}>
                  <span>{file.name}</span>
                  <span>{file.size.toLocaleString()} 字节</span>
                </li>
              ))}
            </ul>
          )}
          {fileSummary && <p className="file-summary">✓ {fileSummary}</p>}
          {fileError && (
            <p className="validation-error file-summary" role="alert">
              {fileError}
            </p>
          )}
        </fieldset>
        {error && (
          <div className="upload-error" role="alert">
            <h3>题目尚未发布</h3>
            <p>{error}</p>
          </div>
        )}
        <div className="upload-actions">
          <p>
            {busy
              ? '正在检查并发布，请稍候…'
              : '发布后数据保存在本机；学生结果页不会展示隐藏用例。'}
          </p>
          <Button
            type="submit"
            disabled={busy || readingChecker || !!fileError}
          >
            {busy ? (
              <>
                <span className="spinner" aria-hidden="true" /> 正在发布…
              </>
            ) : (
              '发布题目'
            )}
          </Button>
        </div>
      </form>
    </div>
  )
}
