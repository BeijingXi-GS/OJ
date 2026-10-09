import { getMode, registerModeRequest, type Mode } from './mode'

export type ProblemType = 'batch' | 'interactive' | 'communication'
export type CommunicationRole = 'sender' | 'receiver'
export type CommunicationSources = Record<CommunicationRole, string>
export type SourceUnit = {
  role: CommunicationRole
  language: 'cpp'
  source: string
}
export type BatchChecker =
  'tokens-v1' | 'exact-v1' | 'nonconsecutive-permutation-v1' | 'cpp-special-v1'
export type ProblemSummary = {
  id: string
  title: string
  difficulty: string
  summary: string
  type: ProblemType
}
export type Problem = Omit<ProblemSummary, 'summary'> & {
  description: string
  input_description: string
  output_description: string
  notes: string
  engineering_notes?: string
  samples: { input: string; output: string }[]
  starter_code: string
  starter_codes?: CommunicationSources
  language: 'cpp'
  standard: 'c++20'
  judge: {
    enabled: boolean
    case_count: number
    wall_time_ms: number
    memory_limit_bytes: number
    checker: BatchChecker | 'interaction-v1' | 'communication-v1'
    communication?: {
      roles: CommunicationRole[]
      byte_limit: number
      aggregate_memory_limit_bytes: number
    }
  }
}
export type CompilationResult = {
  status: 'success' | 'compile_error' | 'timeout'
  diagnostics: string
  diagnostics_truncated: boolean
  duration_ms: number
  compiler: string
  standard: string
}
export type Verdict =
  'AC' | 'WA' | 'TLE' | 'MLE' | 'CE' | 'RE' | 'OLE' | 'SYSTEM_ERROR'
export type CompileEvidence = {
  status: 'success' | 'compile_error' | 'system_error'
  diagnostics: string
  diagnostics_truncated: boolean
  duration_ms: number
}
export type InteractionOutput = {
  stdout: string
  stderr: string
  output_truncated: boolean
}
export type CheckerEvidence = InteractionOutput & {
  exit_code: number | null
  duration_ms: number
  oom_killed: boolean
}
export type RoleCompile = {
  role: CommunicationRole
  compile: CompileEvidence
}
export type EngineerRoleCompile = RoleCompile & {
  source_sha256: string
  artifact_sha256?: string
}
export type CommunicationCount = {
  bytes_observed: number
  bytes_transferred: number
  byte_limit: number
  limit_exceeded: boolean
}
export type TeacherCommunication = CommunicationCount & {
  sender_stdout: string
  sender_stderr: string
  receiver_stderr: string
  sender_output_truncated: boolean
  receiver_output_truncated: boolean
}
export type RoleRuntime = InteractionOutput & {
  exit_code: number | null
  duration_ms: number
  oom_killed: boolean
  terminated_by_supervisor: boolean
}
export type CommunicationEvidence = CommunicationCount & {
  sender: RoleRuntime
  receiver: RoleRuntime
}
export type TeacherCase = {
  id: string
  verdict: Verdict
  duration_ms: number
  start_timed_out?: boolean
  input: string
  expected: string
  input_truncated: boolean
  expected_truncated: boolean
  stdout: string
  stderr: string
  output_truncated: boolean
  interactor?: InteractionOutput
  checker_feedback?: string
  communication?: TeacherCommunication
}
export type JudgeCase = Omit<TeacherCase, 'communication'> & {
  cpu_time_ms: number | null
  peak_memory_bytes: number | null
  exit_code: number | null
  oom_killed: boolean
  checker?: CheckerEvidence
  communication?: CommunicationEvidence
  interactor?: InteractionOutput & {
    exit_code: number | null
    oom_killed: boolean
    participant_stopped: boolean
  }
}
type JudgeSummary = {
  verdict: Verdict
  message: string
  duration_ms: number
  compiler: string
  standard: string
  compile: CompileEvidence
  passed_cases: number
  total_cases: number
  problem_type: ProblemType
  role_compiles?: RoleCompile[]
}
export type StudentResult = JudgeSummary & { mode: 'student' }
export type TeacherResult = JudgeSummary & {
  mode: 'teacher'
  cases: TeacherCase[]
}
export type EngineerResult = Omit<JudgeSummary, 'role_compiles'> & {
  mode: 'engineer'
  run_id: string
  source_sha256: string
  testdata_sha256: string
  artifact_sha256: string
  runtime_image: string
  cases: JudgeCase[]
  interactor_compile?: CompileEvidence
  interactor_source_sha256?: string
  interactor_artifact_sha256?: string
  checker_compile?: CompileEvidence
  checker_source_sha256?: string
  checker_artifact_sha256?: string
  role_compiles?: EngineerRoleCompile[]
}
export type JudgeResult = StudentResult | TeacherResult | EngineerResult
export type UploadDescriptor = {
  title: string
  summary: string
  difficulty: string
  description: string
  input_description: string
  output_description: string
  notes: string
  engineering_notes?: string
  samples: { input: string; output: string }[]
  starter_code: string
  starter_codes?: CommunicationSources
  type: ProblemType
  checker: Problem['judge']['checker']
  wall_time_ms: number
  memory_limit_bytes: number
  cases: { id: string; input: string; expected?: string }[]
  interactor?: 'guess-number-v1'
  checker_source?: 'checker.cpp'
  communication?: { byte_limit: number }
}
export type DebugResult = {
  mode: 'engineer'
  local_demo: boolean
  compiler_configured: boolean
  judge_configured: boolean
  problem_count: number
  judge_metadata: {
    available: boolean
    problem_id?: string
    case_count?: number
    wall_time_ms?: number
    memory_limit_bytes?: number
    checker?: string
    problem_type?: string
    error_code?: string
  }
  message: string
}
export class APIError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message)
    this.name = 'APIError'
  }
}
function errorMessage(body: unknown): string | undefined {
  if (!body || typeof body !== 'object' || !('error' in body)) return
  const error = body.error
  if (!error || typeof error !== 'object' || !('message' in error)) return
  return typeof error.message === 'string' ? error.message : undefined
}
async function request<T>(
  url: string,
  mode: Mode,
  options: RequestInit = {},
): Promise<T> {
  const controller = new AbortController()
  const unregister = registerModeRequest(controller)
  const abort = () => controller.abort()
  options.signal?.addEventListener('abort', abort, { once: true })
  if (options.signal?.aborted || getMode() !== mode) controller.abort()
  const headers = new Headers(options.headers)
  headers.set('X-OJ-Mode', mode)
  try {
    let response: Response
    try {
      response = await fetch(url, {
        ...options,
        headers,
        signal: controller.signal,
      })
    } catch (error) {
      if (controller.signal.aborted)
        throw new DOMException('Aborted', 'AbortError')
      if (error instanceof DOMException && error.name === 'AbortError')
        throw error
      throw new APIError('暂时无法连接服务，请稍后重试。', 0)
    }
    let body: unknown
    try {
      body = await response.json()
    } catch {
      if (controller.signal.aborted)
        throw new DOMException('Aborted', 'AbortError')
      throw new APIError('服务没有返回有效结果，请稍后重试。', response.status)
    }
    if (controller.signal.aborted || getMode() !== mode)
      throw new DOMException('Aborted', 'AbortError')
    if (!response.ok)
      throw new APIError(
        errorMessage(body) ?? '请求未能完成，请稍后重试。',
        response.status,
      )
    return body as T
  } finally {
    options.signal?.removeEventListener('abort', abort)
    unregister()
  }
}
export function getProblems(mode: Mode, signal?: AbortSignal) {
  return request<{ problems: ProblemSummary[] }>('/api/v1/problems', mode, {
    signal,
  })
}
export function getProblem(id: string, mode: Mode, signal?: AbortSignal) {
  return request<Problem>(`/api/v1/problems/${encodeURIComponent(id)}`, mode, {
    signal,
  })
}
export function compileSource(
  problemId: string,
  source: string,
  mode: Mode,
  signal?: AbortSignal,
) {
  return request<CompilationResult>('/api/v1/compilations', mode, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ problem_id: problemId, language: 'cpp', source }),
    signal,
  })
}
export function judgeSource(
  problemId: string,
  source: string,
  mode: Mode,
  signal?: AbortSignal,
) {
  return request<JudgeResult>('/api/v1/judgements', mode, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ problem_id: problemId, language: 'cpp', source }),
    signal,
  })
}
export function judgeCommunication(
  problemId: string,
  sourceUnits: SourceUnit[],
  mode: Mode,
  signal?: AbortSignal,
) {
  return request<JudgeResult>('/api/v1/communication-judgements', mode, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ problem_id: problemId, source_units: sourceUnits }),
    signal,
  })
}
export function uploadProblem(
  descriptor: UploadDescriptor,
  files: File[],
  mode: Mode,
  signal?: AbortSignal,
) {
  const body = new FormData()
  body.append('descriptor', JSON.stringify(descriptor))
  for (const file of files) body.append('files', file, file.name)
  return request<{ id: string; problem: Problem }>('/api/v1/problems', mode, {
    method: 'POST',
    body,
    signal,
  })
}
export function getDebug(mode: Mode, signal?: AbortSignal) {
  return request<DebugResult>('/api/v1/debug', mode, { signal })
}
