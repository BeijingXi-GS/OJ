import { spawn, execFile } from 'node:child_process'
import { randomUUID } from 'node:crypto'
import { existsSync, readdirSync } from 'node:fs'
import { promisify } from 'node:util'
import path from 'node:path'
import { goEnv, local, root } from './lib.mjs'
import { shutdownOwnedJudge } from './judge-shutdown.mjs'

const execFileAsync = promisify(execFile)
if (existsSync(path.join(root, '.env')))
  process.loadEnvFile(path.join(root, '.env'))
const compilerURL = process.env.COMPILER_URL || 'http://127.0.0.1:8081'
const judgeURL = 'http://127.0.0.1:8082'
const children = []
const controlToken = randomUUID()
let judgeChild
let stopping = false
let stopPromise

async function healthy(url) {
  return fetch(url + '/healthz', { signal: AbortSignal.timeout(2000) })
    .then((r) => r.ok)
    .catch(() => false)
}
async function judgeState() {
  try {
    const response = await fetch(judgeURL + '/healthz', {
      signal: AbortSignal.timeout(2000),
    })
    const body = await response.json().catch(() => ({}))
    return {
      listening: true,
      ready:
        response.ok &&
        body.modes_contract_version === 'modes-interactive/0.1.0' &&
        body.special_judge_contract_version === 'special-judge/0.1.0' &&
        body.communication_contract_version === 'communication-mvp/0.1.0',
    }
  } catch {
    return { listening: false, ready: false }
  }
}
function stop(code) {
  if (stopPromise) return stopPromise
  stopping = true
  stopPromise = stopOwnedServices(code)
  return stopPromise
}
async function stopOwnedServices(code) {
  // Close the API connections first so an in-flight judge request can cancel.
  for (const child of children)
    if (
      child !== judgeChild &&
      child.exitCode === null &&
      child.signalCode === null
    )
      child.kill()
  if (judgeChild) {
    if (judgeChild.exitCode === null && judgeChild.signalCode === null) {
      // Works during preflight, before the authenticated HTTP endpoint exists.
      judgeChild.stdin?.end('shutdown\n')
    }
    try {
      await shutdownOwnedJudge({
        child: judgeChild,
        requestShutdown: async () => {
          const r = await fetch(judgeURL + '/shutdown', {
            method: 'POST',
            headers: { 'X-Judge-Control': controlToken },
            signal: AbortSignal.timeout(5000),
          })
          if (r.status !== 202) throw new Error('Judge shutdown refused')
        },
        confirmCleanup: async () => {
          const { stdout } = await execFileAsync(
            'docker',
            ['ps', '-aq', '--filter', 'label=oj.judge.owner=file-mvp-v1'],
            { cwd: root, windowsHide: true, timeout: 5000, maxBuffer: 4096 },
          )
          if (
            stdout.trim() ||
            readdirSync(path.join(local, 'judge-jobs')).length
          )
            throw new Error('Judge resources remain after exit')
        },
      })
      console.log('Owned judge stopped; container and job cleanup confirmed.')
    } catch (error) {
      console.error(
        'Judge shutdown not confirmed; manual inspection required:',
        error.message,
      )
      code = 1
    }
  }
  process.exitCode = code
}
function start(binary, args, options = {}) {
  const child = spawn(binary, args, {
    cwd: root,
    env: process.env,
    stdio: 'inherit',
    windowsHide: true,
    ...options,
  })
  children.push(child)
  child.on('error', (error) => {
    console.error(error.message)
    void stop(1)
  })
  child.on('exit', (code) => {
    if (!stopping) void stop(code === 0 ? 0 : 1)
  })
  return child
}
// Async builds allow signal handlers to finish cleanup. The shared run() helper
// uses spawnSync + process.exit on failure, which would bypass this supervisor.
async function build(output, target) {
  if (stopping) return false
  const child = spawn('go', ['build', '-mod=readonly', '-o', output, target], {
    cwd: root,
    env: goEnv(),
    stdio: 'inherit',
    windowsHide: true,
  })
  children.push(child)
  return new Promise((resolve, reject) => {
    child.once('error', reject)
    child.once('exit', (code) => {
      if (stopping) resolve(false)
      else if (code === 0) resolve(true)
      else reject(new Error('Development service build failed'))
    })
  })
}
process.on('SIGINT', () => {
  void stop(0)
})
process.on('SIGTERM', () => {
  void stop(0)
})

async function launch() {
  console.log('Checking compiler service...')
  const compilerReady = await healthy(compilerURL)
  if (stopping) return
  if (!compilerReady)
    throw new Error(
      'Compiler unavailable. Run npm run compiler:build and npm run compiler:up.',
    )
  if (!existsSync(path.join(local, 'judge-profile.json')))
    throw new Error('Run npm run judge:build first.')
  console.log('Checking judge service...')
  const existingJudge = await judgeState()
  if (stopping) return
  if (existingJudge.listening && !existingJudge.ready)
    throw new Error(
      'Existing judge is not ready for special-judge/0.1.0; stop its owning supervisor and inspect it before restarting.',
    )
  if (!existingJudge.ready) {
    const binary = path.join(
      local,
      process.platform === 'win32' ? 'judge-mvp-dev.exe' : 'judge-mvp-dev',
    )
    console.log('Building owned judge service...')
    if (!(await build(binary, './services/judge-mvp')) || stopping) return
    console.log('Starting owned judge preflight...')
    judgeChild = start(binary, [], {
      env: { ...process.env, JUDGE_CONTROL_TOKEN: controlToken },
      stdio: ['pipe', 'inherit', 'inherit'],
    })
    judgeChild.stdin.on('error', () => {
      /* exit and cleanup verification decide success */
    })
    let ready = false
    for (let i = 0; i < 140 && !stopping; i++) {
      const state = await judgeState()
      if (stopping) return
      if (state.ready) {
        ready = true
        break
      }
      await new Promise((resolve) => setTimeout(resolve, 500))
    }
    if (stopping) return
    if (!ready)
      throw new Error(
        'Judge preflight did not become ready; inspect logs, no execution fallback.',
      )
  }
  const apiBinary = path.join(
    local,
    process.platform === 'win32' ? 'api-mvp-dev.exe' : 'api-mvp-dev',
  )
  console.log('Building API service...')
  if (!(await build(apiBinary, './apps/api')) || stopping) return
  start(apiBinary, [], {
    env: {
      ...process.env,
      API_ADDR: '127.0.0.1:8080',
      COMPILER_URL: compilerURL,
      JUDGE_URL: judgeURL,
    },
  })
  start(
    process.execPath,
    [path.join(root, 'node_modules/vite/bin/vite.js'), '--host', '127.0.0.1'],
    { cwd: path.join(root, 'apps/web') },
  )
  console.log(
    'Judge MVP: http://127.0.0.1:5173; API 8080, judge 8082, compiler 8081. Ctrl+C stops owned services after cleanup.',
  )
}
try {
  await launch()
} catch (error) {
  console.error(error.message)
  await stop(1)
}
if (stopping) await stopPromise
