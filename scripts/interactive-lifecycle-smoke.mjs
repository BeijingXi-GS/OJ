import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import {
  lstatSync,
  mkdirSync,
  readFileSync,
  readdirSync,
  writeFileSync,
} from 'node:fs'
import path from 'node:path'
import { local, root } from './lib.mjs'

// Run this script exclusively: both tests start from a verified empty worker.
// Container mutation is restricted to one inspected, source-bound interactor.
const base = process.env.JUDGE_TEST_API || 'http://127.0.0.1:8080'
const owner = 'file-mvp-v1'
const jobsRoot = path.join(local, 'judge-jobs')
const evidence = path.join(local, 'evidence')
const evidenceFile = path.join(evidence, 'interactive-lifecycle-smoke.json')
const source =
  '#include <iostream>\n#include <unistd.h>\n// OJ owned lifecycle smoke v1\nint main(){int bound;if(!(std::cin>>bound)||bound!=1000)return 2;for(;;)pause();}\n'
const sourceHash = createHash('sha256').update(source).digest('hex')
const report = {
  started_at: new Date().toISOString(),
  source_sha256: sourceHash,
  tests: [],
  queries: [],
  completed: false,
}
mkdirSync(evidence, { recursive: true })
function save() {
  writeFileSync(evidenceFile, JSON.stringify(report, null, 2) + '\n')
}
function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}
function docker(args, operation, context = {}) {
  const query = {
    at: new Date().toISOString(),
    phase: context.phase || operation,
    operation,
    ...(context.run ? { run_id: context.run } : {}),
    ...(context.candidateCount !== undefined
      ? { candidate_count: context.candidateCount }
      : {}),
    status: 'started',
  }
  report.queries.push(query)
  save()
  const result = spawnSync('docker', args, {
    cwd: root,
    encoding: 'utf8',
    timeout: 5000,
    maxBuffer: 1048576,
    windowsHide: true,
  })
  query.exit_code = result.status
  query.status = result.error || result.status !== 0 ? 'failed' : 'success'
  if (result.error) {
    query.error_kind = ['ENOENT', 'ETIMEDOUT', 'EACCES'].includes(
      result.error.code,
    )
      ? result.error.code
      : 'io_failure'
  }
  save()
  if (result.error || result.status !== 0) {
    // Do not expose raw CLI output, arguments, environment or daemon diagnostics.
    throw new Error(`docker_${operation}_failed`)
  }
  return result.stdout
}
function ownedIDs() {
  const text = docker(
    ['ps', '-aq', '--no-trunc', '--filter', `label=oj.judge.owner=${owner}`],
    'list',
    { phase: 'all-owned-containers-list' },
  )
  const ids = text.trim().split(/\s+/).filter(Boolean)
  assert.ok(
    ids.every((id) => /^[a-f0-9]{64}$/.test(id)),
    'invalid_container_id',
  )
  return ids
}
function runtimeIDs(run) {
  // Compiler containers from this run are short lived. They are deliberately
  // outside this observer's candidate set, never inspected or used as proof.
  const text = docker(
    [
      'ps',
      '-aq',
      '--no-trunc',
      '--filter',
      `label=oj.judge.owner=${owner}`,
      '--filter',
      `label=oj.judge.run=${run}`,
      '--filter',
      `name=^/oj-judge-${run}-case-a-`,
    ],
    'list',
    { phase: 'source-bound-runtime-list', run },
  )
  const ids = text.trim().split(/\s+/).filter(Boolean)
  assert.ok(
    ids.every((id) => /^[a-f0-9]{64}$/.test(id)),
    'invalid_runtime_container_id',
  )
  return ids
}
function inspect(ids, context = {}) {
  if (ids.length === 0) return []
  const containers = JSON.parse(
    docker(['inspect', ...ids], 'inspect', {
      ...context,
      candidateCount: ids.length,
    }),
  )
  assert.ok(Array.isArray(containers), 'invalid_inspect_response')
  assert.equal(containers.length, ids.length, 'incomplete_inspect_response')
  return containers
}
function safeContainer(container) {
  return {
    Name: container.Name,
    Labels: {
      'oj.judge.owner': container.Config.Labels['oj.judge.owner'],
      'oj.judge.run': container.Config.Labels['oj.judge.run'],
      'oj.judge.stage': container.Config.Labels['oj.judge.stage'],
    },
    State: {
      Running: container.State.Running,
      OOMKilled: container.State.OOMKilled,
      ExitCode: container.State.ExitCode,
      StartedAt: container.State.StartedAt,
      FinishedAt: container.State.FinishedAt,
      error_present: Boolean(container.State.Error),
    },
  }
}
function jobs() {
  return readdirSync(jobsRoot)
}
function verifyClean() {
  const ids = ownedIDs()
  const directories = jobs()
  assert.deepEqual(ids, [], 'owned_containers_exist_before_or_after_test')
  assert.deepEqual(directories, [], 'job_snapshots_exist_before_or_after_test')
  return { owned_container_count: 0, job_directory_count: 0 }
}
function requestRun() {
  const directories = jobs()
  if (directories.length === 0) return null
  assert.equal(directories.length, 1, 'exclusive_worker_required')
  const name = directories[0]
  const match = /^run-([a-f0-9]{32})-[A-Za-z0-9_-]+$/.exec(name)
  assert.ok(match, 'unexpected_job_identity')
  const directory = path.join(jobsRoot, name)
  const directoryInfo = lstatSync(directory)
  assert.ok(
    directoryInfo.isDirectory() && !directoryInfo.isSymbolicLink(),
    'unsafe_job_snapshot',
  )
  const sourcePath = path.join(directory, 'source.cpp')
  let info
  try {
    info = lstatSync(sourcePath)
  } catch (error) {
    if (error.code === 'ENOENT') return null // writer has not published it yet.
    throw new Error('source_snapshot_query_failed')
  }
  assert.ok(
    info.isFile() && !info.isSymbolicLink() && info.size <= 65536,
    'unsafe_source_snapshot',
  )
  const snapshot = readFileSync(sourcePath, 'utf8')
  if (snapshot.length < source.length && source.startsWith(snapshot))
    return null
  assert.ok(snapshot === source, 'request_source_mismatch')
  return match[1]
}
function identity(container, run, side, id) {
  const stage = `case-a-${side}`
  const labels = container.Config?.Labels
  assert.equal(container.Id, id, 'container_id_mismatch')
  assert.equal(
    container.Name,
    `/oj-judge-${run}-${stage}`,
    'container_name_mismatch',
  )
  assert.equal(labels?.['oj.judge.owner'], owner, 'container_owner_mismatch')
  assert.equal(labels?.['oj.judge.run'], run, 'container_run_mismatch')
  assert.equal(labels?.['oj.judge.stage'], stage, 'container_stage_mismatch')
}
async function runningPair(pendingState) {
  const until = Date.now() + 25000
  while (Date.now() < until) {
    assert.equal(
      pendingState.settled,
      false,
      'request_finished_before_pair_observed',
    )
    const run = requestRun()
    if (run) {
      const containers = inspect(runtimeIDs(run), {
        phase: 'source-bound-runtime-inspect',
        run,
      })
      assert.ok(
        containers.every((c) => c.Config?.Labels?.['oj.judge.run'] === run),
        'another_request_observed',
      )
      const sides = ['participant', 'interactor']
      assert.ok(containers.length <= 2, 'unexpected_runtime_candidate_count')
      for (const container of containers) {
        const side = sides.find(
          (value) =>
            container.Config?.Labels?.['oj.judge.stage'] === `case-a-${value}`,
        )
        assert.ok(side, 'unexpected_runtime_candidate_stage')
        identity(container, run, side, container.Id)
      }
      const pair = sides.map((side) =>
        containers.find(
          (c) => c.Config.Labels['oj.judge.stage'] === `case-a-${side}`,
        ),
      )
      if (pair.every(Boolean)) {
        pair.forEach((container, index) =>
          identity(container, run, sides[index], container.Id),
        )
        if (pair.every((c) => c.State.Running === true)) {
          assert.ok(
            pair.every((c) => !c.State.OOMKilled && !c.State.Error),
            'runtime_not_healthy',
          )
          assert.ok(
            pair.every(
              (c) =>
                c.State.StartedAt && !c.State.StartedAt.startsWith('0001-'),
            ),
            'runtime_not_started',
          )
          return { run, pair }
        }
      }
    }
    await sleep(75)
  }
  throw new Error('running_pair_not_observed')
}
function submit(controller) {
  const state = { settled: false }
  const promise = fetch(base + '/api/v1/judgements', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-OJ-Mode': 'engineer' },
    body: JSON.stringify({
      problem_id: 'guess-number',
      language: 'cpp',
      source,
    }),
    signal: AbortSignal.any([controller.signal, AbortSignal.timeout(110000)]),
  })
    .then(async (response) => {
      const result = await response.json()
      state.settled = true
      return { kind: 'response', status: response.status, result }
    })
    .catch((error) => {
      state.settled = true
      return { kind: 'error', name: error.name }
    })
  return { state, promise }
}
async function waitClean() {
  const until = Date.now() + 30000
  while (Date.now() < until) {
    if (ownedIDs().length === 0 && jobs().length === 0) return verifyClean()
    await sleep(100)
  }
  throw new Error('cleanup_not_confirmed')
}
save()
let activeController
try {
  const healthResponse = await fetch('http://127.0.0.1:8082/healthz', {
    signal: AbortSignal.timeout(5000),
  })
  assert.equal(healthResponse.status, 200, 'judge_not_ready')
  const health = await healthResponse.json()
  assert.equal(
    health.modes_contract_version,
    'modes-interactive/0.1.0',
    'old_worker_version',
  )
  report.modes_contract_version = health.modes_contract_version
  verifyClean()

  activeController = new AbortController()
  const cancellation = submit(activeController)
  await sleep(25)
  const observed = await runningPair(cancellation.state)
  const cancelItem = {
    name: 'cancel-confirmed-running-pair',
    run_id: observed.run,
    before_abort: observed.pair.map(safeContainer),
    passed: false,
  }
  report.tests.push(cancelItem)
  save()
  activeController.abort()
  const cancelled = await cancellation.promise
  assert.equal(cancelled.kind, 'error', 'cancellation_did_not_abort_http')
  assert.equal(cancelled.name, 'AbortError', 'unexpected_cancellation_error')
  cancelItem.http_aborted = true
  cancelItem.cleanup = await waitClean()
  cancelItem.passed = true
  save()
  console.log('PASS cancellation after both runtimes were confirmed running')

  verifyClean()
  activeController = new AbortController()
  const fault = submit(activeController)
  await sleep(25)
  const second = await runningPair(fault.state)
  const faultItem = {
    name: 'interactor-natural-term-system-error',
    run_id: second.run,
    before_term: second.pair.map(safeContainer),
    passed: false,
  }
  report.tests.push(faultItem)
  save()
  const target = second.pair[1]
  // Re-inspect immediately before mutation. Never kill a name-only match or
  // any container whose owner/run/stage differs from this submitted snapshot.
  const current = inspect([target.Id], {
    phase: 'verified-term-target-reinspect',
    run: second.run,
  })[0]
  identity(current, second.run, 'interactor', target.Id)
  assert.equal(current.State.Running, true, 'interactor_stopped_before_term')
  assert.equal(current.State.OOMKilled, false, 'interactor_oom_before_term')
  faultItem.verified_term_target = safeContainer(current)
  save()
  docker(['kill', '--signal', 'TERM', target.Id], 'term_owned_interactor', {
    phase: 'verified-interactor-term',
    run: second.run,
    candidateCount: 1,
  })
  faultItem.signal_sent = 'TERM'
  const answer = await fault.promise
  assert.equal(answer.kind, 'response', 'fault_http_failed')
  assert.equal(answer.status, 200, 'fault_http_status')
  assert.equal(answer.result.run_id, second.run, 'fault_result_run_mismatch')
  assert.equal(
    answer.result.source_sha256,
    sourceHash,
    'fault_result_source_mismatch',
  )
  assert.equal(
    answer.result.verdict,
    'SYSTEM_ERROR',
    'interactor_fault_not_system_error',
  )
  assert.equal(answer.result.cases.length, 1, 'system_error_did_not_stop_cases')
  const row = answer.result.cases[0]
  assert.equal(
    row.verdict,
    'SYSTEM_ERROR',
    'interactor_case_fault_not_system_error',
  )
  assert.equal(row.interactor.exit_code, 143, 'term_exit_evidence_missing')
  assert.equal(row.interactor.oom_killed, false, 'term_was_not_an_oom')
  assert.equal(
    row.interactor.participant_stopped,
    false,
    'system_error_claimed_wa_stop',
  )
  assert.equal(
    row.exit_code,
    137,
    'participant_cleanup_termination_not_confirmed',
  )
  assert.equal(row.oom_killed, false, 'participant_cleanup_was_not_an_oom')
  faultItem.final_result = {
    verdict: answer.result.verdict,
    run_id: answer.result.run_id,
    case: {
      id: row.id,
      verdict: row.verdict,
      exit_code: row.exit_code,
      oom_killed: row.oom_killed,
      interactor: {
        exit_code: row.interactor.exit_code,
        oom_killed: row.interactor.oom_killed,
        participant_stopped: row.interactor.participant_stopped,
      },
    },
  }
  faultItem.cleanup = await waitClean()
  faultItem.passed = true
  report.completed = true
  report.finished_at = new Date().toISOString()
  save()
  console.log(
    'PASS interactor TERM exit 143 mapped to SYSTEM_ERROR; both sides cleaned',
  )
} catch (error) {
  activeController?.abort()
  report.failure = { name: error.name, message: error.message }
  report.finished_at = new Date().toISOString()
  save()
  throw error
}
