import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { fileURLToPath, URL } from 'node:url'
import { setImmediate } from 'node:timers'
import { log } from 'node:console'
import React from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import ts from 'typescript'

function load(relative, dependencies = {}) {
  const file = fileURLToPath(new URL(relative, import.meta.url))
  const source = readFileSync(file, 'utf8')
  const code = ts.transpileModule(source, {
    compilerOptions: {
      module: ts.ModuleKind.CommonJS,
      target: ts.ScriptTarget.ES2022,
      jsx: ts.JsxEmit.ReactJSX,
    },
  }).outputText
  const module = { exports: {} }
  const require = createRequire(file)
  new Function('require', 'module', 'exports', code)(
    (name) => dependencies[name] ?? require(name),
    module,
    module.exports,
  )
  return module.exports
}

const mode = load('../src/lib/mode.ts')
const api = load('../src/lib/api.ts', { './mode': mode })
const utils = load('../src/lib/utils.ts')
const specialJudge = load('../src/lib/specialJudge.ts')
const communication = load('../src/lib/communication.ts')
const JudgeFeedback = load('../src/JudgeFeedback.tsx', {
  '@/lib/mode': mode,
  '@/lib/utils': utils,
  '@/lib/communication': communication,
}).default

const seen = []
globalThis.fetch = async (url, options) => {
  seen.push({ url, mode: options.headers.get('X-OJ-Mode') })
  if (options.signal.aborted)
    throw new globalThis.DOMException('Aborted', 'AbortError')
  return { ok: true, status: 200, json: async () => ({ problems: [] }) }
}
await api.getProblems('student')
assert.equal(seen[0].mode, 'student')
let capturedSignal
globalThis.fetch = async (_, options) => {
  capturedSignal = options.signal
  return new Promise((_, reject) =>
    options.signal.addEventListener(
      'abort',
      () => reject(new globalThis.DOMException('Aborted', 'AbortError')),
      { once: true },
    ),
  )
}
const pending = api.getProblems('student')
mode.changeMode('teacher')
assert.equal(capturedSignal.aborted, true)
await assert.rejects(pending, { name: 'AbortError' })

let resolveJSON
globalThis.fetch = async (_, options) => {
  assert.equal(options.headers.get('X-OJ-Mode'), 'teacher')
  return {
    ok: true,
    status: 200,
    json: () =>
      new Promise((resolve) => {
        resolveJSON = resolve
      }),
  }
}
const stale = api.judgeSource('a-plus-b', 'int main(){}', 'teacher')
await new Promise((resolve) => setImmediate(resolve))
mode.changeMode('student')
resolveJSON({ mode: 'teacher', cases: [{ input: 'hidden-secret' }] })
await assert.rejects(stale, { name: 'AbortError' })

const controller = new globalThis.AbortController()
globalThis.fetch = async (_, options) =>
  new Promise((_, reject) =>
    options.signal.addEventListener(
      'abort',
      () => reject(new globalThis.DOMException('Aborted', 'AbortError')),
      { once: true },
    ),
  )
const cancelled = api.getProblem('a-plus-b', 'student', controller.signal)
controller.abort()
await assert.rejects(cancelled, { name: 'AbortError' })

const result = {
  verdict: 'WA',
  message: '答案错误。',
  duration_ms: 10,
  compiler: 'GCC',
  standard: 'c++20',
  compile: {
    status: 'success',
    diagnostics: '',
    diagnostics_truncated: false,
    duration_ms: 2,
  },
  passed_cases: 0,
  total_cases: 1,
  problem_type: 'interactive',
  run_id: 'hidden-run-id',
  source_sha256: 'hidden-source-hash',
  testdata_sha256: 'hidden-data-hash',
  artifact_sha256: 'hidden-artifact-hash',
  runtime_image: 'hidden-runtime-image',
  cases: [
    {
      id: 'hidden-case-id',
      verdict: 'WA',
      duration_ms: 8,
      input: 'hidden-input',
      expected: 'hidden-answer',
      input_truncated: false,
      expected_truncated: false,
      stdout: 'participant-direction',
      stderr: 'participant-error',
      output_truncated: false,
      cpu_time_ms: null,
      peak_memory_bytes: null,
      exit_code: 137,
      oom_killed: false,
      interactor: {
        stdout: 'interactor-direction',
        stderr: 'interactor-error',
        output_truncated: false,
        exit_code: 1,
        oom_killed: false,
        participant_stopped: true,
      },
    },
  ],
}
const render = (result) =>
  renderToStaticMarkup(
    React.createElement(JudgeFeedback, { result, stale: false }),
  )
const student = render({ ...result, mode: 'student' })
assert.match(student, /通过用例/)
for (const hidden of [
  'hidden-input',
  'hidden-case-id',
  'hidden-answer',
  'participant-direction',
  'participant-error',
  'interactor-direction',
  'interactor-error',
  'hidden-run-id',
  'hidden-source-hash',
  'OOMKilled',
  '退出码',
  'GCC',
  '总耗时',
  '启动开销',
  '选手编译阶段',
])
  assert.equal(
    student.includes(hidden),
    false,
    `Student output contains ${hidden}`,
  )

mode.changeMode('teacher')
const teacher = render({ ...result, mode: 'teacher' })
for (const visible of [
  'hidden-input',
  'participant-direction',
  'participant-error',
  'interactor-direction',
  'interactor-error',
])
  assert.equal(
    teacher.includes(visible),
    true,
    `Teacher output misses ${visible}`,
  )
for (const hidden of [
  'hidden-run-id',
  'hidden-source-hash',
  'OOMKilled',
  '退出码',
  '监督终止',
  'GCC',
  '启动开销',
  '选手编译阶段',
])
  assert.equal(
    teacher.includes(hidden),
    false,
    `Teacher output contains ${hidden}`,
  )

mode.changeMode('engineer')
const engineer = render({ ...result, mode: 'engineer' })
for (const visible of [
  'hidden-input',
  'hidden-run-id',
  'hidden-source-hash',
  'OOMKilled',
  '退出码',
  '监督终止',
])
  assert.equal(
    engineer.includes(visible),
    true,
    `Engineer output misses ${visible}`,
  )
mode.changeMode('student')
assert.equal(render({ ...result, mode: 'teacher' }), '')

const specialResult = {
  ...result,
  problem_type: 'batch',
  checker_compile: {
    ...result.compile,
    diagnostics: 'hidden-checker-compiler',
  },
  checker_source_sha256: 'hidden-checker-source',
  checker_artifact_sha256: 'hidden-checker-artifact',
  cases: [
    {
      ...result.cases[0],
      interactor: undefined,
      checker_feedback: '相邻数字的差不能为 1',
      checker: {
        exit_code: 1,
        duration_ms: 37,
        oom_killed: false,
        stdout: 'hidden-checker-stdout',
        stderr: 'hidden-checker-stderr',
        output_truncated: false,
      },
    },
  ],
}
for (const selectedMode of ['student', 'teacher', 'engineer']) {
  mode.changeMode(selectedMode)
  const markup = render({ ...specialResult, mode: selectedMode })
  assert.equal(
    markup.includes('相邻数字的差不能为 1'),
    selectedMode !== 'student',
  )
  for (const value of [
    'hidden-checker-compiler',
    'hidden-checker-source',
    'hidden-checker-artifact',
    'hidden-checker-stdout',
    'hidden-checker-stderr',
    '特判退出码',
  ])
    assert.equal(
      markup.includes(value),
      selectedMode === 'engineer',
      `${selectedMode}: ${value}`,
    )
}
mode.changeMode('student')
const compilerFailure = render({
  ...result,
  mode: 'student',
  verdict: 'CE',
  compile: {
    ...result.compile,
    status: 'compile_error',
    diagnostics: 'solution.cpp: missing semicolon',
  },
})
assert.ok(compilerFailure.includes('solution.cpp: missing semicolon'))
assert.equal(compilerFailure.includes('GCC'), false)

const files = (...names) => names.map((name) => ({ name, size: 6 }))
assert.deepEqual(
  specialJudge.buildCases(files('01.out', '01.in'), 'batch', 'tokens-v1'),
  [{ id: 'case-01', input: '01.in', expected: '01.out' }],
)
assert.throws(
  () => specialJudge.buildCases(files('01.in'), 'batch', 'tokens-v1'),
  /缺少/,
)
assert.deepEqual(
  specialJudge.buildCases(
    files('01.in'),
    'batch',
    'nonconsecutive-permutation-v1',
  ),
  [{ id: 'case-01', input: '01.in' }],
)
assert.throws(
  () =>
    specialJudge.buildCases(
      files('01.in', '01.out'),
      'batch',
      'nonconsecutive-permutation-v1',
    ),
  /只需/,
)
assert.deepEqual(
  specialJudge.buildCases(
    files('01.in', '02.in', '02.out'),
    'batch',
    'cpp-special-v1',
  ),
  [
    { id: 'case-01', input: '01.in' },
    { id: 'case-02', input: '02.in', expected: '02.out' },
  ],
)
assert.throws(
  () =>
    specialJudge.buildCases(
      files('01.in', '02.in', '03.in', '04.in', '05.in'),
      'batch',
      'cpp-special-v1',
    ),
  /1 至 4/,
)
assert.throws(
  () => specialJudge.buildCases(files('../01.in'), 'batch', 'cpp-special-v1'),
  /命名/,
)
assert.throws(
  () =>
    specialJudge.buildCases(files('01.in', '01.in'), 'batch', 'cpp-special-v1'),
  /重复/,
)
assert.throws(
  () =>
    specialJudge.buildCases(
      [{ name: '01.in', size: 65537 }],
      'batch',
      'cpp-special-v1',
    ),
  /64 KiB/,
)
assert.throws(
  () =>
    specialJudge.buildCases(
      files('01.in', '02.out'),
      'batch',
      'cpp-special-v1',
    ),
  /没有对应/,
)
assert.deepEqual(
  specialJudge.buildCases(files('01.in'), 'interactive', 'tokens-v1'),
  [{ id: 'case-01', input: '01.in' }],
)
for (const input of ['1', '1000\n', ' 4\r\n', '0001', '\v0004\f'])
  assert.equal(specialJudge.isConstructionInput(input), true)
for (const input of ['0', '1001', '4 5', '\u00a04', '1\0', '-1', '3x'])
  assert.equal(specialJudge.isConstructionInput(input), false)
for (const input of ['1', '1000\n', '\t7\r\n', '0001'])
  assert.equal(specialJudge.isInteractiveInput(input), true)
for (const input of ['\v4\f', '\u00a04', '0', '1001', '00001'])
  assert.equal(specialJudge.isInteractiveInput(input), false)
for (const name of [
  '_case.in',
  '-case.in',
  'con.in',
  'prn.in',
  'aux.in',
  'nul.in',
  'com1.in',
  'compare.in',
  'lpt9.in',
  'lpt-data.in',
  'case.in\n',
  `${'a'.repeat(65)}.in`,
])
  assert.throws(
    () => specialJudge.buildCases(files(name), 'batch', 'cpp-special-v1'),
    /命名|保留/,
  )
for (const name of [
  '0.in',
  'a_01.in',
  'a-01.in',
  '01con.in',
  `${'a'.repeat(64)}.in`,
])
  assert.equal(
    specialJudge.buildCases(files(name), 'batch', 'cpp-special-v1').length,
    1,
  )
specialJudge.validateCheckerSource(specialJudge.checkerExample)
assert.equal(
  specialJudge.checkerExample,
  readFileSync(
    new URL(
      '../../../testdata/special-judge-example/checker.cpp',
      import.meta.url,
    ),
    'utf8',
  ).replaceAll('\r\n', '\n'),
  'Downloaded/uploaded example differs from the canonical teacher example',
)
for (const invalid of ['', 'int\0main(){}', 'a'.repeat(65537)])
  assert.throws(() => specialJudge.validateCheckerSource(invalid))

// Render the actual statement with an editor stub; no browser or Monaco needed.
globalThis.self = globalThis
const statement = {
  id: 'sort-numbers',
  title: '排序练习',
  difficulty: '入门',
  type: 'batch',
  description: '将整数升序排列。',
  input_description: '输入整数序列。',
  output_description: '输出排序结果。',
  notes: '注意重复数字。',
  engineering_notes: '工程限定内容：默认代码是演示解法。',
  samples: [{ input: '3\n3 1 2\n', output: '1 2 3\n' }],
  starter_code: 'int main(){}',
  language: 'cpp',
  standard: 'c++20',
  judge: {
    enabled: true,
    case_count: 5,
    wall_time_ms: 2000,
    memory_limit_bytes: 67108864,
    checker: 'tokens-v1',
  },
}
let pageProblem = statement
const editorDependencies = {
  'react-router-dom': {
    Link: ({ children, to, ...props }) =>
      React.createElement('a', { href: to, ...props }, children),
  },
  '@monaco-editor/react': {
    default: ({ options, value, path }) =>
      React.createElement('textarea', {
        'aria-label': options.ariaLabel,
        value,
        readOnly: true,
        'data-path': path,
      }),
  },
  '@/components/ui/button': {
    Button: ({ children, ...props }) =>
      React.createElement('button', props, children),
  },
  './JudgeFeedback': { default: JudgeFeedback },
  '@/lib/utils': utils,
  '@/lib/mode': mode,
  '@/lib/api': api,
  '@/lib/communication': communication,
}
const CommunicationWorkspace = load(
  '../src/CommunicationWorkspace.tsx',
  editorDependencies,
).default
const EditorPage = load('../src/EditorPage.tsx', {
  '@tanstack/react-query': {
    useQuery: () => ({ data: pageProblem, isPending: false, isError: false }),
  },
  'react-router-dom': {
    Link: ({ children, to, ...props }) =>
      React.createElement('a', { href: to, ...props }, children),
    useLocation: () => ({ state: null }),
    useParams: () => ({ problemId: pageProblem.id }),
  },
  '@monaco-editor/react': { default: () => null, loader: { config: () => {} } },
  'monaco-editor/editor/editor.api.js': {},
  'monaco-editor/languages/definitions/cpp/register.js': {},
  'monaco-editor/editor/editor.worker.js?worker': { default: () => {} },
  '@/components/ui/button': {
    Button: ({ children, ...props }) =>
      React.createElement('button', props, children),
  },
  './JudgeFeedback': { default: JudgeFeedback },
  './CommunicationWorkspace': { default: CommunicationWorkspace },
  '@/lib/utils': utils,
  '@/lib/mode': mode,
  '@/lib/api': api,
}).default
for (const selectedMode of ['student', 'teacher', 'engineer']) {
  mode.changeMode(selectedMode)
  const markup = renderToStaticMarkup(React.createElement(EditorPage))
  assert.ok(markup.includes('注意重复数字。'))
  assert.ok(markup.includes('时间限制'))
  assert.ok(markup.includes('内存限制'))
  for (const technical of [
    '工程限定内容',
    'tokens-v1',
    '容器',
    '墙钟',
    '启动开销',
    '编译与链接',
  ])
    assert.equal(
      markup.includes(technical),
      selectedMode === 'engineer',
      `${selectedMode}: ${technical}`,
    )
}
pageProblem = {
  ...statement,
  type: 'interactive',
  judge: { ...statement.judge, checker: 'interaction-v1' },
}
mode.changeMode('student')
const interactiveStatement = renderToStaticMarkup(
  React.createElement(EditorPage),
)
assert.ok(interactiveStatement.includes('必须刷新输出'))
assert.ok(interactiveStatement.includes('读取回复'))

// Communication requests preserve explicit roles and use their own endpoint.
const sources = {
  receiver: communication.communicationAdditionExample.receiver,
  sender: communication.communicationAdditionExample.sender,
}
const units = communication.buildSourceUnits(sources)
assert.deepEqual(units, [
  { role: 'sender', language: 'cpp', source: sources.sender },
  { role: 'receiver', language: 'cpp', source: sources.receiver },
])
let communicationRequest
globalThis.fetch = async (url, options) => {
  communicationRequest = {
    url,
    body: JSON.parse(options.body),
    mode: options.headers.get('X-OJ-Mode'),
  }
  return { ok: true, status: 200, json: async () => ({ mode: 'student' }) }
}
await api.judgeCommunication(
  'communication-a-plus-b',
  [units[1], units[0]],
  'student',
)
assert.equal(communicationRequest.url, '/api/v1/communication-judgements')
assert.equal(communicationRequest.mode, 'student')
assert.deepEqual(communicationRequest.body, {
  problem_id: 'communication-a-plus-b',
  source_units: [units[1], units[0]],
})
for (const invalid of [
  { sender: 'int main(){}' },
  { ...sources, other: 'int main(){}' },
  { ...sources, sender: '' },
  { ...sources, receiver: 'int\0main(){}' },
  { ...sources, receiver: '中'.repeat(21846) },
])
  assert.throws(() => communication.buildSourceUnits(invalid))
assert.equal(
  communication.buildSourceUnits({ ...sources, sender: 'a'.repeat(65536) })[0]
    .source.length,
  65536,
)
for (const input of ['1', '32', '65536'])
  assert.equal(communication.validateCommunicationLimit(input), Number(input))
for (const input of ['', '0', '-1', '1.5', '65537', 'NaN'])
  assert.throws(() => communication.validateCommunicationLimit(input))
assert.equal(communication.sourcesMatch(sources, { ...sources }), true)
assert.equal(
  communication.sourcesMatch(sources, { ...sources, sender: 'changed' }),
  false,
)
assert.equal(
  communication.sourcesMatch(sources, { ...sources, receiver: 'changed' }),
  false,
)
assert.deepEqual(
  specialJudge.buildCases(
    files('01.in', '01.out'),
    'communication',
    'nonconsecutive-permutation-v1',
  ),
  [{ id: 'case-01', input: '01.in', expected: '01.out' }],
)
assert.throws(
  () => specialJudge.buildCases(files('01.in'), 'communication', 'tokens-v1'),
  /缺少/,
)
assert.throws(
  () =>
    specialJudge.buildCases(
      files(
        '01.in',
        '01.out',
        '02.in',
        '02.out',
        '03.in',
        '03.out',
        '04.in',
        '04.out',
        '05.in',
        '05.out',
      ),
      'communication',
      'tokens-v1',
    ),
  /1 至 4/,
)

let communicationSignal
globalThis.fetch = async (_, options) => {
  communicationSignal = options.signal
  return new Promise((_, reject) =>
    options.signal.addEventListener(
      'abort',
      () => reject(new globalThis.DOMException('Aborted', 'AbortError')),
      { once: true },
    ),
  )
}
const communicationPending = api.judgeCommunication(
  'communication-sum',
  units,
  'student',
)
mode.changeMode('teacher')
assert.equal(communicationSignal.aborted, true)
await assert.rejects(communicationPending, { name: 'AbortError' })
let resolveCommunicationJSON
globalThis.fetch = async () => ({
  ok: true,
  status: 200,
  json: () =>
    new Promise((resolve) => {
      resolveCommunicationJSON = resolve
    }),
})
const staleCommunication = api.judgeCommunication(
  'communication-sum',
  units,
  'teacher',
)
await new Promise((resolve) => setImmediate(resolve))
mode.changeMode('student')
resolveCommunicationJSON({
  mode: 'teacher',
  cases: [{ input: 'communication-hidden-input' }],
})
await assert.rejects(staleCommunication, { name: 'AbortError' })

const roleRuntime = {
  exit_code: 0,
  duration_ms: 30,
  oom_killed: false,
  stdout: 'communication-secret-message',
  stderr: 'communication-role-error',
  output_truncated: false,
  terminated_by_supervisor: false,
}
const commCount = {
  bytes_observed: 18,
  bytes_transferred: 18,
  byte_limit: 32,
  limit_exceeded: false,
}
const commResult = {
  ...result,
  problem_type: 'communication',
  role_compiles: [
    {
      role: 'sender',
      source_sha256: 'sender-hidden-source',
      artifact_sha256: 'sender-hidden-artifact',
      compile: result.compile,
    },
    {
      role: 'receiver',
      source_sha256: 'receiver-hidden-source',
      artifact_sha256: 'receiver-hidden-artifact',
      compile: {
        ...result.compile,
        status: 'compile_error',
        diagnostics: 'receiver.cpp: missing semicolon',
      },
    },
  ],
  cases: [
    {
      ...result.cases[0],
      interactor: undefined,
      communication: {
        ...commCount,
        sender: roleRuntime,
        receiver: roleRuntime,
      },
    },
  ],
}
for (const selectedMode of ['student', 'teacher', 'engineer']) {
  mode.changeMode(selectedMode)
  const projected =
    selectedMode === 'teacher'
      ? {
          ...commResult,
          cases: commResult.cases.map((item) => ({
            ...item,
            communication: {
              ...commCount,
              sender_stdout: roleRuntime.stdout,
              sender_stderr: roleRuntime.stderr,
              receiver_stderr: roleRuntime.stderr,
              sender_output_truncated: false,
              receiver_output_truncated: false,
            },
          })),
        }
      : commResult
  const markup = render({ ...projected, mode: selectedMode })
  assert.ok(markup.includes('receiver.cpp: missing semicolon'))
  assert.ok(markup.includes('接收端编译结果'))
  for (const value of [
    'communication-secret-message',
    'communication-role-error',
    '通信记录',
    '转发字节',
    'hidden-input',
    'hidden-answer',
  ])
    assert.equal(
      markup.includes(value),
      selectedMode !== 'student',
      `${selectedMode}: ${value}`,
    )
  for (const value of [
    'sender-hidden-source',
    'receiver-hidden-source',
    'sender-hidden-artifact',
    'receiver-hidden-artifact',
    'OOMKilled',
    '监督终止',
    '退出码',
    '实际聚合内存峰值',
    '受控输入管道',
  ])
    assert.equal(
      markup.includes(value),
      selectedMode === 'engineer',
      `${selectedMode}: ${value}`,
    )
}
// One role may have run before the peer hits the shared startup deadline.
mode.changeMode('engineer')
const partialStartup = render({
  ...commResult,
  mode: 'engineer',
  verdict: 'TLE',
  cases: commResult.cases.map((item) => ({
    ...item,
    verdict: 'TLE',
    start_timed_out: true,
    communication: {
      ...item.communication,
      receiver: { ...roleRuntime, exit_code: null },
    },
  })),
})
assert.ok(partialStartup.includes('至少一个角色未完成启动'))
assert.ok(partialStartup.includes('发送端运行证据'))
assert.equal(partialStartup.includes('程序尚未开始运行'), false)
const legacyStartup = render({
  ...result,
  mode: 'engineer',
  cases: result.cases.map((item) => ({ ...item, start_timed_out: true })),
})
assert.ok(legacyStartup.includes('程序尚未开始运行'))
mode.changeMode('teacher')
const teacherStartup = render({
  ...commResult,
  mode: 'teacher',
  verdict: 'TLE',
  cases: commResult.cases.map((item) => ({
    ...item,
    verdict: 'TLE',
    start_timed_out: true,
    communication: {
      ...commCount,
      sender_stdout: roleRuntime.stdout,
      sender_stderr: roleRuntime.stderr,
      receiver_stderr: '',
      sender_output_truncated: false,
      receiver_output_truncated: false,
    },
  })),
})
assert.ok(teacherStartup.includes('communication-secret-message'))
assert.equal(teacherStartup.includes('程序尚未开始运行'), false)
mode.changeMode('student')
const senderCE = render({
  ...commResult,
  mode: 'student',
  verdict: 'CE',
  cases: [],
  role_compiles: [
    {
      ...commResult.role_compiles[0],
      compile: {
        ...result.compile,
        status: 'compile_error',
        diagnostics: 'sender.cpp: syntax error',
      },
    },
  ],
})
assert.ok(senderCE.includes('sender.cpp: syntax error'))
assert.equal(
  senderCE.includes('接收端编译结果'),
  false,
  'Unattempted receiver compilation must not be invented',
)

pageProblem = {
  ...statement,
  id: 'communication-a-plus-b',
  title: '通信 A+B',
  type: 'communication',
  description: '发送端传两个整数，接收端计算它们的和。',
  input_description: '发送端读取 a、b 并传数；接收端只读取消息中的两个数。',
  output_description: '接收端输出 a+b。',
  samples: [{ input: '2 3\n', output: '5\n' }],
  starter_code: communication.communicationAdditionExample.sender,
  starter_codes: communication.communicationAdditionExample,
  judge: {
    ...statement.judge,
    checker: 'communication-v1',
    communication: {
      roles: ['sender', 'receiver'],
      byte_limit: 32,
      aggregate_memory_limit_bytes: 134217728,
    },
  },
}
for (const selectedMode of ['student', 'teacher', 'engineer']) {
  mode.changeMode(selectedMode)
  const markup = renderToStaticMarkup(React.createElement(EditorPage))
  for (const required of [
    '发送端 C++ 源码编辑器',
    '接收端 C++ 源码编辑器',
    '每例通信限制',
    '每角色内存限制',
    '运行测试',
    'EOF',
  ])
    assert.ok(markup.includes(required), `${selectedMode}: ${required}`)
  assert.equal(markup.includes('编译代码'), false)
  for (const value of [
    '受限容器',
    'communication-v1',
    '实际聚合内存峰值',
    '墙钟',
    '启动开销',
    '按空白分隔的 token',
  ])
    assert.equal(
      markup.includes(value),
      selectedMode === 'engineer',
      `${selectedMode}: ${value}`,
    )
  assert.ok(markup.includes('通信 A+B'))
  assert.ok(markup.includes('发送端传两个整数，接收端计算它们的和。'))
  assert.ok(markup.includes('<pre>2 3\n</pre>'))
  assert.ok(markup.includes('<pre>5\n</pre>'))
  assert.ok(markup.includes('data-path="communication-a-plus-b/sender.cpp"'))
  assert.ok(markup.includes('data-path="communication-a-plus-b/receiver.cpp"'))
}
assert.deepEqual(communication.communicationAdditionSample, {
  input: '2 3\n',
  output: '5\n',
})
assert.equal(
  communication.buildSourceUnits(communication.communicationExample).length,
  2,
  'Original sequence-sum example remains available',
)

mode.changeMode('teacher')
const UploadPage = load('../src/UploadPage.tsx', {
  ...editorDependencies,
  react: {
    ...React,
    useState: (initial) =>
      React.useState(initial === 'batch' ? 'communication' : initial),
  },
  '@tanstack/react-query': { useQueryClient: () => ({}) },
  'react-router-dom': {
    ...editorDependencies['react-router-dom'],
    useNavigate: () => () => {},
  },
  '@/lib/specialJudge': specialJudge,
}).default
const uploadMarkup = renderToStaticMarkup(React.createElement(UploadPage))
for (const value of [
  '发送端 C++20 起始模板',
  '接收端 C++20 起始模板',
  '每例通信限制',
  '填入通信求和说明与双模板',
  '填入通信 A+B 说明与双模板',
  '发送端只传 a、b，接收端读取两个数并计算它们的和',
  '1–4 组同名 .in / .out',
  '两角色共享时间限制',
  '64 KiB',
])
  assert.ok(
    uploadMarkup.includes(value),
    `Communication upload misses ${value}`,
  )
assert.equal(uploadMarkup.includes('特判源码 · checker.cpp'), false)
mode.changeMode('student')

log(
  'PASS: existing modes, cancellation, statements, special/interactive rules; communication explicit roles/endpoint, UTF-8 source and byte limits, paired data, stale/cancelled requests, per-role CE and bounded mode projections, actual dual-editor/teacher-upload rendering',
)
