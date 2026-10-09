import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { createServer } from 'node:net'
import path from 'node:path'
import { goEnv, local, root, run } from './lib.mjs'

const compilerURL = process.env.COMPILER_URL || 'http://127.0.0.1:8081'
const health = await fetch(compilerURL + '/healthz', {
  signal: AbortSignal.timeout(5000),
})
assert.equal(
  health.status,
  200,
  'Start the real compiler with npm run compiler:up first',
)
const compiler = await health.json()
assert.match(compiler.compiler, /^GCC 14\.2\.0$/)
const listener = createServer()
await new Promise((resolve, reject) => {
  listener.once('error', reject)
  listener.listen(0, '127.0.0.1', resolve)
})
const port = listener.address().port
await new Promise((resolve) => listener.close(resolve))
const binary = path.join(
  local,
  process.platform === 'win32' ? 'api-mvp-smoke.exe' : 'api-mvp-smoke',
)
run('go', ['build', '-mod=readonly', '-o', binary, './apps/api'], {
  env: goEnv(),
})
const api = spawn(binary, [], {
  cwd: root,
  windowsHide: true,
  env: {
    ...process.env,
    API_ADDR: `127.0.0.1:${port}`,
    COMPILER_URL: compilerURL,
    DATABASE_URL: '',
  },
  stdio: 'ignore',
})
let launchError
api.on('error', (err) => {
  launchError = err
})
const base = `http://127.0.0.1:${port}`
try {
  let ready = false
  for (let i = 0; i < 60; i++) {
    if (launchError) throw launchError
    if (api.exitCode !== null) throw new Error('API exited before readiness')
    try {
      ready = (await fetch(base + '/healthz')).ok
    } catch {
      /* Startup race. */
    }
    if (ready) break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.ok(ready, 'API startup timeout')
  const detail = await (await fetch(base + '/api/v1/problems/a-plus-b')).json()
  assert.equal(detail.id, 'a-plus-b')
  const cases = [
    ['valid C++', detail.starter_code, 'success'],
    ['syntax error', 'int main() { this is not C++; }', 'compile_error'],
    [
      'link error without main',
      'int add(int a,int b) { return a+b; }',
      'compile_error',
    ],
    ['infinite loop is compiled only', 'int main() { for (;;) {} }', 'success'],
  ]
  for (const [name, source, expected] of cases) {
    const response = await fetch(base + '/api/v1/compilations', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ problem_id: detail.id, language: 'cpp', source }),
      signal: AbortSignal.timeout(22000),
    })
    const result = await response.json()
    assert.equal(response.status, 200, JSON.stringify(result))
    assert.equal(result.status, expected, name)
    assert.equal(result.standard, 'c++20')
    assert.equal(result.compiler, 'GCC 14.2.0')
    assert.ok(result.duration_ms < 15000, name)
    if (expected === 'compile_error')
      assert.ok(result.diagnostics.length > 0, name)
    console.log(`PASS: ${name} (${result.duration_ms} ms)`)
  }
  const system = await (await fetch(base + '/api/v1/system')).json()
  assert.equal(system.judge_enabled, false)
  assert.equal(system.compile_enabled, true)
  assert.equal((await fetch(base + '/readyz')).status, 503)
  console.log(
    'PASS: real API → compiler integration; no database dependency, judging remains disabled',
  )
} finally {
  if (api.exitCode === null && !launchError) {
    const stopped = new Promise((resolve) => api.once('exit', resolve))
    api.kill()
    await stopped
  }
}
