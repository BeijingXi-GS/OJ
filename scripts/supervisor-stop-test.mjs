import assert from 'node:assert/strict'
import { spawn, spawnSync } from 'node:child_process'
import { mkdirSync, readdirSync, writeFileSync } from 'node:fs'
import { createConnection } from 'node:net'
import path from 'node:path'
import { pathToFileURL } from 'node:url'
import { local, root } from './lib.mjs'

const maxLogBytes = 1048576
const bootstrapMessage = '[supervisor-test] bootstrap'
const stoppedMessage =
  'Owned judge stopped; container and job cleanup confirmed.'
const pause = (ms) => new Promise((resolve) => setTimeout(resolve, ms))
const errorInfo = (error) => ({
  name: error.name,
  message: error.message,
})

// An exit event is not a stdio-drain notification. Inherited pipes can still
// contain output, or be held by a child of the process which just exited.
function observe(child) {
  const state = {
    exit: null,
    close: null,
    spawn_error: null,
    log_error: null,
    log_bytes_at_exit: null,
    log_bytes_at_close: null,
    logs: '',
  }
  let logBytes = 0
  let resolveExit
  let resolveClose
  const exited = new Promise((resolve) => {
    resolveExit = resolve
  })
  const closed = new Promise((resolve) => {
    resolveClose = resolve
  })
  function collect(chunk) {
    logBytes += chunk.length
    if (logBytes <= maxLogBytes) state.logs += chunk.toString()
    else state.log_error ||= new Error('Supervisor output exceeded 1 MiB')
  }
  child.stdout.on('data', collect)
  child.stderr.on('data', collect)
  child.once('error', (error) => {
    state.spawn_error = error
    // A spawn failure need not emit exit. close is still separately observed.
    resolveExit()
  })
  child.once('exit', (code, signal) => {
    state.exit = { code, signal }
    state.log_bytes_at_exit = logBytes
    resolveExit()
  })
  child.once('close', (code, signal) => {
    state.close = { code, signal }
    state.log_bytes_at_close = logBytes
    resolveClose()
  })
  return { child, state, exited, closed }
}

async function waitFor(predicate, limit = 45000, observation) {
  const started = Date.now()
  while (Date.now() - started < limit) {
    if (observation?.state.spawn_error) throw observation.state.spawn_error
    if (observation?.state.log_error) throw observation.state.log_error
    if (observation?.state.exit)
      throw new Error(
        'Supervisor exited before the required stage: ' +
          JSON.stringify(observation.state.exit),
      )
    if (await predicate()) return
    await pause(200)
  }
  throw new Error('Lifecycle condition not reached within budget')
}

async function waitUntilClosed(observation, limit = 130000) {
  let timer
  try {
    await Promise.race([
      Promise.all([observation.exited, observation.closed]),
      new Promise((_, reject) => {
        timer = setTimeout(
          () => reject(new Error('Owned supervisor exit/stdio-close timeout')),
          limit,
        )
      }),
    ])
  } finally {
    clearTimeout(timer)
  }
  if (observation.state.spawn_error) throw observation.state.spawn_error
  if (observation.state.log_error) throw observation.state.log_error
}

function throwFailures(errors) {
  if (errors.length === 1) throw errors[0]
  if (errors.length > 1)
    throw new AggregateError(
      errors,
      'Lifecycle phase failed; cleanup observations are recorded separately',
    )
}

function docker(args) {
  const result = spawnSync('docker', args, {
    cwd: root,
    windowsHide: true,
    encoding: 'utf8',
    timeout: 5000,
  })
  assert.equal(result.status, 0, 'Docker inspection must succeed')
  return result.stdout.trim()
}

function clean() {
  assert.equal(
    docker(['ps', '-aq', '--filter', 'label=oj.judge.owner=file-mvp-v1']),
    '',
    'Owned judge containers must be absent',
  )
  assert.deepEqual(
    readdirSync(path.join(local, 'judge-jobs')),
    [],
    'Owned judge jobs must be absent',
  )
}

async function servicesClosed() {
  for (const port of [5173, 8080, 8082]) {
    const refused = await new Promise((resolve, reject) => {
      const socket = createConnection({ host: '127.0.0.1', port })
      socket.setTimeout(1500, () => {
        socket.destroy(new Error('Owned service port inspection timed out'))
      })
      socket.once('connect', () => {
        socket.destroy()
        resolve(false)
      })
      socket.once('error', (error) => {
        if (error.code === 'ECONNREFUSED') resolve(true)
        else reject(error)
      })
    })
    assert.equal(refused, true, 'Owned service ' + port + ' must have stopped')
  }
}

async function realTests() {
  // Each run gets its own report so earlier failures remain available.
  const evidence = path.join(local, 'evidence')
  mkdirSync(evidence, { recursive: true })
  const startedAt = new Date().toISOString()
  const reportPath = path.join(
    evidence,
    'spj-supervisor-stop-' + startedAt.replaceAll(/[:.]/g, '-') + '.json',
  )
  const report = { started_at: startedAt, tests: [], completed: false }
  const save = () =>
    writeFileSync(reportPath, JSON.stringify(report, null, 2) + '\n')
  save()
  console.log('Supervisor lifecycle report: ' + reportPath)
  try {
    for (const mode of ['preflight', 'idle', 'running']) {
      const test = {
        mode,
        started_at: new Date().toISOString(),
        active_stage: '',
        stage_reached: false,
        cleanup_confirmed: false,
        services_closed: false,
        log: '',
        errors: [],
      }
      report.tests.push(test)
      save()
      let observation
      let request
      const errors = []
      const recordError = (phase, error) => {
        errors.push(error)
        test.errors.push({ phase, ...errorInfo(error) })
        save()
      }
      try {
        clean()
        await servicesClosed()
        const inline =
          'console.log(' +
          JSON.stringify(bootstrapMessage) +
          ");process.stdin.setEncoding('utf8');process.stdin.on('data',()=>process.emit('SIGINT'));await import('./scripts/mvp-dev.mjs')"
        observation = observe(
          spawn(process.execPath, ['--input-type=module', '-e', inline], {
            cwd: root,
            env: process.env,
            windowsHide: true,
            stdio: ['pipe', 'pipe', 'pipe'],
          }),
        )
        observation.child.stdin.on('error', (error) => {
          test.stdin_error = errorInfo(error)
        })
        const logs = () => observation.state.logs
        await waitFor(
          () => logs().includes(bootstrapMessage),
          10000,
          observation,
        )
        if (mode === 'preflight') {
          await waitFor(
            () => {
              if (!logs().includes('judge preflight started')) return false
              test.active_stage = docker([
                'ps',
                '--filter',
                'label=oj.judge.owner=file-mvp-v1',
                '--filter',
                'status=running',
                '--format',
                '{{.Label "oj.judge.stage"}}',
              ])
              return Boolean(test.active_stage)
            },
            45000,
            observation,
          )
          assert.equal(
            logs().includes('judge ready'),
            false,
            'Must cancel actual preflight, not a ready instance',
          )
        } else {
          await waitFor(
            async () =>
              logs().includes('Judge MVP:') &&
              (
                await fetch('http://127.0.0.1:8080/healthz', {
                  signal: AbortSignal.timeout(1500),
                }).catch(() => null)
              )?.ok,
            45000,
            observation,
          )
          if (mode === 'running') {
            request = fetch('http://127.0.0.1:8080/api/v1/judgements', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
                'X-OJ-Mode': 'engineer',
              },
              body: JSON.stringify({
                problem_id: 'a-plus-b',
                language: 'cpp',
                source: 'int main(){for(;;){}}',
              }),
              signal: AbortSignal.timeout(110000),
            })
              .then(async (response) => {
                await response.text()
                test.request_outcome = { status: response.status }
              })
              .catch((error) => {
                test.request_outcome = { error: errorInfo(error) }
              })
            await waitFor(
              () => {
                test.active_stage = docker([
                  'ps',
                  '--filter',
                  'label=oj.judge.owner=file-mvp-v1',
                  '--filter',
                  'status=running',
                  '--format',
                  '{{.Label "oj.judge.stage"}}',
                ])
                return test.active_stage
                  .split('\n')
                  .some((stage) => stage.startsWith('case-'))
              },
              45000,
              observation,
            )
          }
        }
        test.stage_reached = true
      } catch (error) {
        // Save the initial failure before cleanup. A finally assertion must not
        // replace a startup or stage-selection failure with a missing-log error.
        recordError('phase', error)
      } finally {
        if (observation) {
          test.stop_requested_at = new Date().toISOString()
          try {
            if (
              observation.state.exit === null &&
              !observation.state.spawn_error
            )
              observation.child.stdin.end('stop\n')
            await waitUntilClosed(observation)
          } catch (error) {
            recordError('stop', error)
          }
          if (request) await request
          test.supervisor_exit = observation.state.exit
          test.stdio_close = observation.state.close
          test.log_bytes_at_exit = observation.state.log_bytes_at_exit
          test.log_bytes_at_close = observation.state.log_bytes_at_close
          test.log = observation.state.logs
          save()
          try {
            assert.deepEqual(
              observation.state.exit,
              { code: 0, signal: null },
              'Owned supervisor must exit successfully; see the report log',
            )
            assert.deepEqual(
              observation.state.close,
              { code: 0, signal: null },
              'Owned supervisor stdio must close successfully',
            )
            assert.ok(
              test.log.includes(stoppedMessage),
              'Owned judge cleanup confirmation must be observed',
            )
          } catch (error) {
            recordError('exit-confirmation', error)
          }
        }
        try {
          clean()
          test.cleanup_confirmed = true
        } catch (error) {
          recordError('resource-inspection', error)
        }
        try {
          await servicesClosed()
          test.services_closed = true
        } catch (error) {
          recordError('service-inspection', error)
        }
        test.finished_at = new Date().toISOString()
        save()
      }
      throwFailures(errors)
    }
    report.completed = true
    report.finished_at = new Date().toISOString()
    save()
    console.log(
      'Real supervisor lifecycle tests passed: preflight, idle, running.',
    )
  } catch (error) {
    report.error = errorInfo(error)
    report.finished_at = new Date().toISOString()
    save()
    throw error
  }
}

// Trusted process-only reproductions; only owned .local fixture files are used.
// No service, Docker, uploaded data or environment-file access is involved.
async function selfTests() {
  const source =
    'console.log("module entered");process.on("SIGINT",()=>console.log("signal observed"));console.log("module ready")'
  const inline =
    'process.stdin.setEncoding("utf8");process.stdin.on("data",()=>process.emit("SIGINT"));await import(' +
    JSON.stringify('data:text/javascript,' + encodeURIComponent(source)) +
    ')'
  const dynamic = observe(
    spawn(process.execPath, ['--input-type=module', '-e', inline], {
      windowsHide: true,
      stdio: ['pipe', 'pipe', 'pipe'],
    }),
  )
  await waitFor(
    () => dynamic.state.logs.includes('module ready'),
    5000,
    dynamic,
  )
  dynamic.child.stdin.end('stop\n')
  await waitUntilClosed(dynamic, 5000)
  assert.deepEqual(dynamic.state.exit, { code: 0, signal: null })
  assert.ok(dynamic.state.logs.includes('signal observed'))

  const fixtureDir = path.join(local, 'supervisor-test-fixtures')
  mkdirSync(fixtureDir, { recursive: true })
  const fixturePath = path.join(fixtureDir, '本地 module.mjs')
  writeFileSync(
    fixturePath,
    'import { spawn } from "node:child_process";\n' +
      'console.log("local fixture entered");\n' +
      'process.on("SIGINT",()=>console.log("local fixture signal"));\n' +
      'const child=spawn(process.execPath,["-e",' +
      JSON.stringify(
        'console.log("descendant stdout");console.error("descendant stderr")',
      ) +
      '],{stdio:"inherit",windowsHide:true});\n' +
      'await new Promise((resolve,reject)=>{child.once("error",reject);child.once("exit",code=>code===0?resolve():reject(new Error("fixture descendant failed")))});\n' +
      'console.log("local fixture ready");\n',
  )
  for (const specifier of [
    './.local/supervisor-test-fixtures/本地 module.mjs',
    pathToFileURL(fixturePath).href,
  ]) {
    const localImport = observe(
      spawn(
        process.execPath,
        [
          '--input-type=module',
          '-e',
          'process.stdin.setEncoding("utf8");process.stdin.on("data",()=>process.emit("SIGINT"));await import(' +
            JSON.stringify(specifier) +
            ')',
        ],
        { cwd: root, windowsHide: true, stdio: ['pipe', 'pipe', 'pipe'] },
      ),
    )
    try {
      await waitFor(
        () => localImport.state.logs.includes('local fixture ready'),
        5000,
        localImport,
      )
    } finally {
      localImport.child.stdin.end('stop\n')
      await waitUntilClosed(localImport, 5000)
    }
    assert.deepEqual(localImport.state.exit, { code: 0, signal: null })
    for (const message of [
      'local fixture entered',
      'descendant stdout',
      'descendant stderr',
      'local fixture ready',
      'local fixture signal',
    ])
      assert.ok(localImport.state.logs.includes(message), message)
  }

  const failed = observe(
    spawn(
      process.execPath,
      ['-e', 'console.error("fixture startup failure");process.exitCode=7'],
      { windowsHide: true, stdio: ['ignore', 'pipe', 'pipe'] },
    ),
  )
  await assert.rejects(
    waitFor(() => false, 5000, failed),
    /Supervisor exited before the required stage/,
  )
  await waitUntilClosed(failed, 5000)
  assert.deepEqual(failed.state.exit, { code: 7, signal: null })
  assert.ok(failed.state.logs.includes('fixture startup failure'))

  // Pause the readable side deliberately to make the observation gap
  // deterministic. This demonstrates why exit alone cannot confirm log drain;
  // it does not claim the old real-process failure paused its pipe.
  const late = observe(
    spawn(process.execPath, ['-e', 'console.log("buffered output")'], {
      windowsHide: true,
      stdio: ['ignore', 'pipe', 'pipe'],
    }),
  )
  late.child.stdout.pause()
  await late.exited
  assert.deepEqual(late.state.exit, { code: 0, signal: null })
  assert.equal(
    late.state.log_bytes_at_exit,
    0,
    'The reproduction must observe exit before reading buffered output',
  )
  late.child.stdout.resume()
  await waitUntilClosed(late, 5000)
  assert.deepEqual(late.state.close, { code: 0, signal: null })
  assert.ok(late.state.logs.includes('buffered output'))

  const phase = new Error('original stage failure')
  const cleanup = new Error('subsequent cleanup failure')
  assert.throws(
    () => throwFailures([phase, cleanup]),
    (error) =>
      error instanceof AggregateError &&
      error.errors[0] === phase &&
      error.errors[1] === cleanup,
  )
  assert.throws(
    () => throwFailures([phase]),
    (error) => error === phase,
  )
  console.log(
    'Supervisor observation self-tests passed: data/local Unicode imports with inherited stdout/stderr, signal, early failure, exit before pipe drain, original failure preservation.',
  )
}

if (process.argv.includes('--self-test')) await selfTests()
else await realTests()
