import { spawn } from 'node:child_process'
import { mkdirSync } from 'node:fs'
import net from 'node:net'
import path from 'node:path'
import { root, local, goEnv, run } from './lib.mjs'

mkdirSync(local, { recursive: true })
const binary = path.join(
  local,
  process.platform === 'win32' ? 'api-smoke.exe' : 'api-smoke',
)
run('go', ['build', '-o', binary, './apps/api'], { env: goEnv() })
const listener = net.createServer()
await new Promise((resolve) => listener.listen(0, '127.0.0.1', resolve))
const port = listener.address().port
await new Promise((resolve) => listener.close(resolve))
const child = spawn(binary, [], {
  cwd: root,
  env: { ...goEnv(), API_ADDR: '127.0.0.1:' + port, DATABASE_URL: '' },
  stdio: 'pipe',
  windowsHide: true,
})
let logs = ''
child.stderr.on('data', (chunk) => {
  logs += chunk
})
child.stdout.on('data', (chunk) => {
  logs += chunk
})
const exited = new Promise((resolve) => child.once('exit', resolve))
try {
  const base = 'http://127.0.0.1:' + port
  let ready = false
  for (let i = 0; i < 50; i++) {
    if (child.exitCode !== null) throw new Error('API exited: ' + logs)
    try {
      const response = await fetch(base + '/healthz')
      ready = response.ok && (await response.json()).status === 'ok'
      if (ready) break
    } catch {
      /* startup retry */
    }
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  if (!ready) throw new Error('API startup timed out: ' + logs)
  const info = await fetch(base + '/api/v1/system').then((r) => r.json())
  if (info.judge_enabled !== false)
    throw new Error('Scaffold must not execute submissions')
  if ((await fetch(base + '/readyz')).status !== 503)
    throw new Error('Unconfigured database must not be ready')
  if (
    (await fetch(base + '/api/v1/submissions', { method: 'POST' })).status !==
    404
  )
    throw new Error('Submissions are not implemented')
  console.log(
    'PASS: API startup, health, metadata, fail-closed readiness, absent submission route.',
  )
} finally {
  child.kill()
  await exited
}
