import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { mkdirSync, readFileSync, readdirSync, writeFileSync } from 'node:fs'
import path from 'node:path'
import { local, root } from './lib.mjs'

// A real, serial acceptance run against an already-started local MVP. There is
// no fake mode, service startup, fixture mutation, upload or resource deletion.
const base = process.env.JUDGE_TEST_API || 'http://127.0.0.1:8080'
const worker = process.env.JUDGE_TEST_WORKER || 'http://127.0.0.1:8082'
const jobsRoot = path.resolve(
  process.env.JUDGE_TEST_JOBS_ROOT || path.join(local, 'judge-jobs'),
)
const profilePath = path.resolve(
  process.env.JUDGE_TEST_PROFILE || path.join(local, 'judge-profile.json'),
)
const problemId = 'communication-a-plus-b'
const fixtureRoot = path.join(root, 'testdata/problems', problemId)
const exampleRoot = path.join(root, 'examples', problemId)
const roles = ['sender', 'receiver']
const evidence = path.join(local, 'evidence')
mkdirSync(evidence, { recursive: true })
const report = {
  started_at: new Date().toISOString(),
  acceptance_kind: 'real-local-docker',
  api: base,
  worker,
  jobs_root: jobsRoot,
  profile_path: profilePath,
  problem_id: problemId,
  tests: [],
  completed: false,
}
const reportPath = path.join(
  evidence,
  `communication-matrix-${report.started_at.replace(/[:.]/g, '-')}.json`,
)
function save() {
  writeFileSync(reportPath, JSON.stringify(report, null, 2) + '\n')
}
function digest(value) {
  return createHash('sha256').update(value).digest('hex')
}
// Go encoding/json escapes HTML characters and U+2028/U+2029. Ordinary
// JSON.stringify alone would hash #include <iostream> differently.
function canonicalJSON(value) {
  return JSON.stringify(value).replace(
    /[<>&\u2028\u2029]/g,
    (character) =>
      `\\u${character.charCodeAt(0).toString(16).padStart(4, '0')}`,
  )
}
function localURL(value) {
  const url = new URL(value)
  assert.equal(url.protocol, 'http:')
  assert.ok(['127.0.0.1', 'localhost', '[::1]'].includes(url.hostname))
  assert.equal(url.username + url.password + url.search + url.hash, '')
  assert.equal(url.pathname, '/')
  return url.origin
}
function docker(args) {
  const query = spawnSync('docker', args, {
    cwd: root,
    encoding: 'utf8',
    timeout: 10000,
    maxBuffer: 1024 * 1024,
    windowsHide: true,
  })
  assert.equal(
    query.status,
    0,
    `Docker query failed (${args[0]}): ${query.error?.message || query.stderr}`,
  )
  return query.stdout.trim()
}
function cleanupSnapshot() {
  return {
    checked_at: new Date().toISOString(),
    containers: docker([
      'ps',
      '-aq',
      '--filter',
      'label=oj.judge.owner=file-mvp-v1',
    ])
      .split(/\s+/)
      .filter(Boolean),
    jobs: readdirSync(jobsRoot).sort(),
  }
}
function assertClean(snapshot) {
  assert.deepEqual(
    snapshot.containers,
    [],
    'Judge containers remain; inspect ownership manually',
  )
  assert.deepEqual(
    snapshot.jobs,
    [],
    'Judge jobs remain; do not delete unknown jobs',
  )
}
async function getJSON(url) {
  const response = await fetch(url, {
    headers: { 'X-OJ-Mode': 'engineer' },
    signal: AbortSignal.timeout(10000),
  })
  assert.equal(response.status, 200, `GET ${url}: HTTP ${response.status}`)
  return response.json()
}
function checkHealth(health, profile) {
  assert.equal(health.status, 'ok')
  assert.equal(health.profile, 'judge-mvp/0.1.0')
  assert.equal(health.communication_contract_version, 'communication-mvp/0.1.0')
  assert.equal(health.modes_contract_version, 'modes-interactive/0.1.0')
  assert.equal(health.special_judge_contract_version, 'special-judge/0.1.0')
  assert.equal(health.image, profile.image)
}
function loadSnapshot() {
  const raw = readFileSync(path.join(fixtureRoot, 'manifest.json'))
  const manifest = JSON.parse(raw)
  assert.equal(manifest.type, 'communication')
  assert.equal(manifest.checker, 'communication-v1')
  assert.deepEqual(manifest.communication, { byte_limit: 32 })
  assert.equal(manifest.wall_time_ms, 5000)
  assert.equal(manifest.memory_limit_bytes, 67108864)
  assert.deepEqual(
    manifest.cases.map((row) => row.id),
    ['normal', 'zero', 'negative', 'large'],
  )
  const hash = createHash('sha256')
  function add(bytes) {
    const size = Buffer.alloc(8)
    size.writeBigUInt64LE(BigInt(bytes.length))
    hash.update(size).update(bytes)
  }
  add(raw)
  const cases = manifest.cases.map((row) => {
    assert.equal(row.input, row.id + '.in')
    assert.equal(row.expected, row.id + '.out')
    const input = readFileSync(path.join(fixtureRoot, row.input))
    const expected = readFileSync(path.join(fixtureRoot, row.expected))
    assert.deepEqual(
      readFileSync(path.join(exampleRoot, 'data', row.input)),
      input,
    )
    assert.deepEqual(
      readFileSync(path.join(exampleRoot, 'data', row.expected)),
      expected,
    )
    add(Buffer.from(row.id))
    add(input)
    add(expected)
    return {
      id: row.id,
      input: input.toString(),
      expected: expected.toString(),
    }
  })
  return {
    manifest,
    cases,
    sha256: hash.digest('hex'),
    manifest_sha256: digest(raw),
  }
}
function validateResult(result, test, units, profile, snapshot) {
  assert.equal(result.mode, 'engineer')
  assert.equal(result.problem_type, 'communication')
  assert.equal(result.verdict, test.expected, test.name)
  assert.match(result.run_id, /^[a-z0-9][a-z0-9_-]{0,63}$/)
  assert.equal(result.runtime_image, profile.image)
  assert.equal(result.compiler, 'GCC 14.2.0')
  assert.equal(result.standard, 'c++20')
  assert.equal(result.source_sha256, digest(canonicalJSON(units)))
  assert.equal(result.testdata_sha256, snapshot.sha256)
  for (const field of [
    'interactor_compile',
    'checker_compile',
    'interactor_source_sha256',
    'checker_source_sha256',
  ])
    assert.equal(Object.hasOwn(result, field), false)
  const ce = test.expected === 'CE'
  const count = test.failedRole === 'sender' ? 1 : 2
  assert.equal(result.role_compiles.length, count)
  for (const [index, record] of result.role_compiles.entries()) {
    assert.equal(record.role, roles[index])
    assert.equal(record.source_sha256, digest(units[index].source))
    if (record.role === test.failedRole) {
      assert.equal(record.compile.status, 'compile_error')
      assert.equal(record.artifact_sha256, '')
    } else {
      assert.equal(record.compile.status, 'success')
      assert.match(record.artifact_sha256, /^[a-f0-9]{64}$/)
    }
  }
  assert.equal(result.compile.status, ce ? 'compile_error' : 'success')
  assert.equal(result.total_cases, ce ? 0 : 4)
  assert.equal(result.cases.length, ce ? 0 : 4)
  assert.equal(
    result.passed_cases,
    result.cases.filter((row) => row.verdict === 'AC').length,
  )
  if (ce) {
    assert.equal(result.artifact_sha256, '')
    return
  }
  assert.equal(
    result.artifact_sha256,
    digest(
      canonicalJSON(
        result.role_compiles.map((record) => ({
          role: record.role,
          artifact_sha256: record.artifact_sha256,
        })),
      ),
    ),
  )
  for (const [index, row] of result.cases.entries()) {
    assert.equal(row.id, snapshot.cases[index].id)
    assert.equal(row.input, snapshot.cases[index].input)
    assert.equal(row.expected, snapshot.cases[index].expected)
    assert.equal(row.verdict, test.expected)
    assert.equal(row.cpu_time_ms, null)
    assert.equal(row.peak_memory_bytes, null)
    assert.equal(row.interactor, undefined)
    assert.equal(row.checker, undefined)
    const communication = row.communication
    assert.equal(communication.byte_limit, 32)
    assert.ok(
      Number.isSafeInteger(communication.bytes_observed) &&
        communication.bytes_observed >= 0,
    )
    assert.ok(
      Number.isSafeInteger(communication.bytes_transferred) &&
        communication.bytes_transferred >= 0,
    )
    assert.ok(communication.bytes_transferred <= 32)
    assert.ok(communication.bytes_observed >= communication.bytes_transferred)
    assert.equal(
      communication.limit_exceeded,
      communication.bytes_observed > 32,
    )
    assert.equal(communication.limit_exceeded, test.overflow === true)
    assert.equal(row.stdout, communication.receiver.stdout)
    assert.equal(row.exit_code, communication.receiver.exit_code)
    assert.equal(
      row.oom_killed,
      communication.sender.oom_killed || communication.receiver.oom_killed,
    )
    for (const role of roles) {
      const run = communication[role]
      assert.equal(typeof run.terminated_by_supervisor, 'boolean')
      if (run.terminated_by_supervisor) assert.equal(run.exit_code, 137)
      if (test.expected !== 'MLE') assert.equal(run.oom_killed, false)
      if (
        test.expected === 'AC' ||
        (test.expected === 'WA' && !test.overflow)
      ) {
        assert.equal(run.exit_code, 0)
        assert.equal(run.terminated_by_supervisor, false)
      }
    }
    if (test.expected === 'AC') {
      assert.equal(row.stdout.trim(), row.expected.trim())
      assert.equal(communication.sender.stdout, row.input)
      assert.equal(communication.bytes_observed, Buffer.byteLength(row.input))
      assert.equal(
        communication.bytes_transferred,
        Buffer.byteLength(row.input),
      )
    }
    if (test.failureRole) {
      const run = communication[test.failureRole]
      if (test.expected === 'RE') {
        assert.equal(run.exit_code, test.exitCode)
        assert.equal(run.terminated_by_supervisor, false)
      } else if (test.expected === 'MLE') assert.equal(run.oom_killed, true)
      else if (test.expected === 'TLE')
        assert.equal(run.terminated_by_supervisor, true)
    }
    if (test.expected === 'OLE') assert.equal(row.output_truncated, true)
  }
  if (test.expected === 'AC') assert.equal(result.passed_cases, 4)
}

save()
try {
  const apiURL = localURL(base)
  const workerURL = localURL(worker)
  report.profile = JSON.parse(readFileSync(profilePath, 'utf8'))
  assert.equal(report.profile.version, 'judge-mvp/0.1.0')
  assert.match(report.profile.image, /^sha256:[a-f0-9]{64}$/)
  report.docker = {
    platform: docker([
      'version',
      '--format',
      '{{.Server.Os}} {{.Server.Arch}}',
    ]),
    cgroup_version: docker(['info', '--format', '{{.CgroupVersion}}']),
    image: docker([
      'image',
      'inspect',
      report.profile.image,
      '--format',
      '{{.Id}}',
    ]),
  }
  save()
  assert.equal(report.docker.platform, 'linux amd64')
  assert.equal(report.docker.cgroup_version, '2')
  assert.equal(report.docker.image, report.profile.image)
  report.initial_cleanup = cleanupSnapshot()
  save()
  assertClean(report.initial_cleanup)
  report.worker_health = await getJSON(workerURL + '/healthz')
  checkHealth(report.worker_health, report.profile)
  const snapshot = loadSnapshot()
  report.snapshot = snapshot
  const detail = await getJSON(apiURL + '/api/v1/problems/' + problemId)
  report.metadata = detail.judge
  assert.equal(detail.id, problemId)
  assert.equal(detail.type, 'communication')
  assert.deepEqual(detail.judge, {
    problem_id: problemId,
    enabled: true,
    case_count: 4,
    wall_time_ms: 5000,
    memory_limit_bytes: 67108864,
    checker: 'communication-v1',
    problem_type: 'communication',
    communication: {
      roles,
      byte_limit: 32,
      aggregate_memory_limit_bytes: 134217728,
    },
  })
  assert.deepEqual(Object.keys(detail.starter_codes).sort(), [...roles].sort())
  assert.equal(detail.starter_code, detail.starter_codes.sender)
  for (const role of roles)
    assert.equal(
      detail.starter_codes[role],
      readFileSync(path.join(exampleRoot, role + '.cpp'), 'utf8'),
      `Published ${role} template differs from example`,
    )
  const { sender, receiver } = detail.starter_codes
  const tests = [
    { name: 'A+B-AC-4-of-4', expected: 'AC' },
    { name: 'reversed-source-units-AC', expected: 'AC', reversed: true },
    {
      name: 'receiver-WA',
      expected: 'WA',
      receiver: receiver.replace('a + b', 'a + b + 1'),
    },
    {
      name: 'sender-CE',
      expected: 'CE',
      sender: 'int main(){ this is not C++; }',
      failedRole: 'sender',
    },
    {
      name: 'receiver-CE',
      expected: 'CE',
      receiver: 'int main(){ this is not C++; }',
      failedRole: 'receiver',
    },
    {
      name: 'communication-overflow-WA',
      expected: 'WA',
      overflow: true,
      sender:
        '#include <cstdio>\nint main(){for(int i=0;i<33;i++)putchar(120);}',
      receiver:
        '#include <cstdio>\nint main(){while(getchar()!=EOF){}puts("0");}',
    },
    {
      name: 'sender-RE',
      expected: 'RE',
      sender: 'int main(){return 17;}',
      failureRole: 'sender',
      exitCode: 17,
    },
    {
      name: 'receiver-RE-137-not-OOM',
      expected: 'RE',
      receiver: 'int main(){return 137;}',
      failureRole: 'receiver',
      exitCode: 137,
    },
    {
      name: 'sender-TLE',
      expected: 'TLE',
      sender: 'int main(){for(;;){}}',
      failureRole: 'sender',
    },
    {
      name: 'receiver-TLE',
      expected: 'TLE',
      receiver: 'int main(){for(;;){}}',
      failureRole: 'receiver',
    },
    {
      name: 'receiver-MLE',
      expected: 'MLE',
      failureRole: 'receiver',
      receiver:
        '#include <cstring>\n#include <vector>\nint main(){std::vector<char*>v;for(;;){auto p=new char[1048576];memset(p,1,1048576);v.push_back(p);}}',
    },
    {
      name: 'receiver-OLE',
      expected: 'OLE',
      receiver:
        '#include <unistd.h>\nint main(){char b[4096]={};for(;;)write(1,b,sizeof(b));}',
    },
  ]
  report.planned_tests = tests.map(({ name, expected }) => ({ name, expected }))
  save()
  for (const test of tests) {
    const units = roles.map((role) => ({
      role,
      language: 'cpp',
      source: test[role] ?? { sender, receiver }[role],
    }))
    const item = {
      name: test.name,
      expected: test.expected,
      source_units: units,
      passed: false,
    }
    report.tests.push(item)
    save()
    item.cleanup_before = cleanupSnapshot()
    assertClean(item.cleanup_before)
    const response = await fetch(apiURL + '/api/v1/communication-judgements', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-OJ-Mode': 'engineer' },
      body: JSON.stringify({
        problem_id: problemId,
        source_units: test.reversed ? [...units].reverse() : units,
      }),
      signal: AbortSignal.timeout(110000),
    })
    item.http_status = response.status
    item.response_text = await response.text()
    save()
    // Capture cleanup even when the response is an error or malformed JSON.
    item.cleanup_after = cleanupSnapshot()
    save()
    assertClean(item.cleanup_after)
    assert.equal(response.status, 200, `${test.name}: HTTP ${response.status}`)
    item.result = JSON.parse(item.response_text)
    delete item.response_text
    save()
    validateResult(item.result, test, units, report.profile, snapshot)
    item.worker_health = await getJSON(workerURL + '/healthz')
    checkHealth(item.worker_health, report.profile)
    assert.equal(
      loadSnapshot().sha256,
      snapshot.sha256,
      'Fixture changed during acceptance',
    )
    assert.deepEqual(
      JSON.parse(readFileSync(profilePath, 'utf8')),
      report.profile,
      'Profile changed during acceptance',
    )
    item.passed = true
    save()
    console.log(
      `PASS ${test.name}: ${item.result.verdict}; ${item.result.total_cases} cases; cleanup and bindings verified`,
    )
  }
  report.final_cleanup = cleanupSnapshot()
  assertClean(report.final_cleanup)
  report.completed = true
  report.finished_at = new Date().toISOString()
  save()
  console.log(
    `PASS real communication matrix (${tests.length} tests). Evidence: ${reportPath}`,
  )
} catch (error) {
  report.completed = false
  report.failed_at = new Date().toISOString()
  report.failure = { name: error.name, message: error.message }
  try {
    report.cleanup_at_failure = cleanupSnapshot()
  } catch (cleanupError) {
    report.cleanup_at_failure = {
      confirmed: false,
      error: cleanupError.message,
    }
  }
  save()
  console.error(
    `FAIL communication acceptance: ${error.message}\nEvidence: ${reportPath}`,
  )
  process.exitCode = 1
}
