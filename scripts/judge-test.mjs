import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import {
  mkdirSync,
  readFileSync,
  readdirSync,
  renameSync,
  writeFileSync,
} from 'node:fs'
import path from 'node:path'
import { local, root } from './lib.mjs'

const base = process.env.JUDGE_TEST_API || 'http://127.0.0.1:8080'
const detailResponse = await fetch(base + '/api/v1/problems/a-plus-b')
assert.equal(detailResponse.status, 200)
const detail = await detailResponse.json()
assert.equal(detail.judge.enabled, true)
const profile = JSON.parse(
  readFileSync(path.join(local, 'judge-profile.json'), 'utf8'),
)
const report = {
  started_at: new Date().toISOString(),
  profile,
  tests: [],
  completed: false,
}
const evidence = path.join(local, 'evidence')
mkdirSync(evidence, { recursive: true })
function save() {
  writeFileSync(
    path.join(evidence, 'judge-matrix.json'),
    JSON.stringify(report, null, 2) + '\n',
  )
}
async function judge(source) {
  const response = await fetch(base + '/api/v1/judgements', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-OJ-Mode': 'engineer' },
    body: JSON.stringify({ problem_id: detail.id, language: 'cpp', source }),
    signal: AbortSignal.timeout(110000),
  })
  return { status: response.status, result: await response.json() }
}
function containers() {
  const r = spawnSync(
    'docker',
    ['ps', '-aq', '--filter', 'label=oj.judge.owner=file-mvp-v1'],
    { cwd: root, encoding: 'utf8', timeout: 5000, windowsHide: true },
  )
  assert.equal(r.status, 0, 'Docker query must succeed')
  return r.stdout.trim()
}
function clean() {
  assert.equal(containers(), '', 'No judge containers remain')
  assert.deepEqual(
    readdirSync(path.join(local, 'judge-jobs')),
    [],
    'No job directory remains',
  )
}
const tests = [
  ['AC', detail.starter_code, 'AC'],
  ['WA', '#include <cstdio>\nint main(){puts("42");}', 'WA'],
  ['CE-syntax', 'int main(){ this is not C++; }', 'CE'],
  ['CE-link', 'int add(int a,int b){return a+b;}', 'CE'],
  ['RE-crash', '#include <cstdlib>\nint main(){abort();}', 'RE'],
  ['RE-137-not-OOM', 'int main(){return 137;}', 'RE'],
  ['TLE', 'int main(){for(;;){}}', 'TLE'],
  [
    'MLE',
    '#include <cstring>\n#include <vector>\nint main(){std::vector<char*> blocks;for(;;){auto p=new char[1048576];memset(p,1,1048576);blocks.push_back(p);}}',
    'MLE',
  ],
  [
    'OLE',
    '#include <unistd.h>\nint main(){char b[4096]={};for(;;)write(1,b,sizeof(b));}',
    'OLE',
  ],
  [
    'display-truncation-AC',
    '#include <cstdio>\nint main(){long long a,b;scanf("%lld%lld",&a,&b);for(int i=0;i<70000;i++)putchar(32);printf("%lld\\n",a+b);}',
    'AC',
  ],
  ['invalid-UTF8-WA', '#include <cstdio>\nint main(){putchar(255);}', 'WA'],
  [
    'child-process-cleanup',
    '#include <unistd.h>\n#include <cstdio>\nint main(){if(fork()==0){for(;;)sleep(1);}long long a,b;scanf("%lld%lld",&a,&b);printf("%lld\\n",a+b);}',
    'AC',
  ],
  [
    'isolation-AC',
    '#include <unistd.h>\n#include <sys/prctl.h>\n#include <sys/socket.h>\n#include <arpa/inet.h>\n#include <cstdio>\nint main(){if(getuid()!=10002||prctl(PR_GET_NO_NEW_PRIVS,0,0,0,0)!=1||access("/var/run/docker.sock",F_OK)==0||access("/testdata",F_OK)==0||access("/expected",F_OK)==0)return 11; FILE*f=fopen("/etc/oj-write-probe","w");if(f)return 12;int s=socket(AF_INET,SOCK_STREAM,0);sockaddr_in a{};a.sin_family=AF_INET;a.sin_port=htons(80);inet_pton(AF_INET,"192.0.2.1",&a.sin_addr);if(connect(s,(sockaddr*)&a,sizeof(a))==0)return 13;long long x,y;scanf("%lld%lld",&x,&y);printf("%lld\\n",x+y);}',
    'AC',
  ],
]
save()
for (const [name, source, expected] of tests) {
  const response = await fetch(base + '/api/v1/judgements', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-OJ-Mode': 'engineer' },
    body: JSON.stringify({ problem_id: detail.id, language: 'cpp', source }),
    signal: AbortSignal.timeout(110000),
  })
  const result = await response.json()
  const item = { name, expected, http_status: response.status, result }
  report.tests.push(item)
  save()
  assert.equal(response.status, 200, JSON.stringify(result))
  assert.equal(result.verdict, expected, name)
  assert.equal(result.runtime_image, profile.image)
  assert.match(result.source_sha256, /^[a-f0-9]{64}$/)
  assert.match(result.testdata_sha256, /^[a-f0-9]{64}$/)
  if (expected === 'CE') {
    assert.equal(result.cases.length, 0)
    assert.equal(result.compile.status, 'compile_error')
  } else {
    assert.equal(result.cases.length, 4)
    assert.match(result.artifact_sha256, /^[a-f0-9]{64}$/)
  }
  for (const c of result.cases) {
    assert.equal(c.cpu_time_ms, null)
    assert.equal(c.peak_memory_bytes, null)
  }
  if (expected === 'MLE')
    assert.ok(result.cases.every((c) => c.verdict === 'MLE' && c.oom_killed))
  if (name === 'RE-137-not-OOM')
    assert.ok(
      result.cases.every(
        (c) => c.exit_code === 137 && !c.oom_killed && c.verdict === 'RE',
      ),
    )
  if (name === 'display-truncation-AC')
    assert.ok(
      result.cases.every((c) => c.output_truncated && c.verdict === 'AC'),
    )
  clean()
  item.passed = true
  save()
  console.log(
    `PASS ${name}: ${result.verdict}, ${result.duration_ms} ms, ${result.cases.length} cases, cleanup verified`,
  )
}

// Keep the first request alive until it actually enters a case, then cancel the
// API connection. It must propagate to the worker and leave no descendants.
const controller = new AbortController()
const pending = fetch(base + '/api/v1/judgements', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json', 'X-OJ-Mode': 'engineer' },
  body: JSON.stringify({
    problem_id: detail.id,
    language: 'cpp',
    source: 'int main(){for(;;){}}',
  }),
  signal: controller.signal,
}).catch(() => null)
let running = false
for (let i = 0; i < 100; i++) {
  const r = spawnSync(
    'docker',
    [
      'ps',
      '-q',
      '--filter',
      'label=oj.judge.stage=case-a',
      '--filter',
      'label=oj.judge.owner=file-mvp-v1',
    ],
    { cwd: root, encoding: 'utf8', timeout: 5000, windowsHide: true },
  )
  assert.equal(r.status, 0)
  if (r.stdout.trim()) {
    running = true
    break
  }
  await new Promise((resolve) => setTimeout(resolve, 50))
}
assert.ok(running, 'Cancellation test must reach runtime')
const busy = await fetch(base + '/api/v1/judgements', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json', 'X-OJ-Mode': 'engineer' },
  body: JSON.stringify({
    problem_id: detail.id,
    language: 'cpp',
    source: detail.starter_code,
  }),
})
assert.equal(busy.status, 429)
controller.abort()
await pending
for (let i = 0; i < 100; i++) {
  if (!containers() && readdirSync(path.join(local, 'judge-jobs')).length === 0)
    break
  await new Promise((resolve) => setTimeout(resolve, 100))
}
clean()
report.tests.push({
  name: 'busy-and-API-cancellation',
  passed: true,
  http_status: 429,
})

// Controlled fixture mutations always restore the exact original bytes. They
// verify configuration failures and that displayed metadata follows manifest.
const fixtureDir = path.join(root, 'testdata/problems/a-plus-b')
const manifestPath = path.join(fixtureDir, 'manifest.json')
const originalManifest = readFileSync(manifestPath)
const inputPath = path.join(fixtureDir, 'sample.in')
const inputBackup = path.join(local, 'judge-test-sample-input.backup')
try {
  renameSync(inputPath, inputBackup)
  try {
    const r = await judge(detail.starter_code)
    assert.equal(r.status, 200)
    assert.equal(r.result.verdict, 'SYSTEM_ERROR')
    assert.deepEqual(r.result.cases, [])
    report.tests.push({
      name: 'missing-input-SYSTEM_ERROR',
      passed: true,
      ...r,
    })
    save()
    clean()
  } finally {
    renameSync(inputBackup, inputPath)
  }
  writeFileSync(manifestPath, '{"unexpected":true}\n')
  const malformed = await judge(detail.starter_code)
  assert.equal(malformed.status, 200)
  assert.equal(malformed.result.verdict, 'SYSTEM_ERROR')
  assert.deepEqual(malformed.result.cases, [])
  report.tests.push({
    name: 'invalid-manifest-SYSTEM_ERROR',
    passed: true,
    ...malformed,
  })
  save()
  clean()
  const m = JSON.parse(originalManifest.toString())
  m.checker = 'exact-v1'
  m.wall_time_ms = 3500
  m.memory_limit_bytes = 33554432
  m.cases = m.cases.slice(0, 1)
  writeFileSync(manifestPath, JSON.stringify(m))
  const metadata = await (
    await fetch(base + '/api/v1/problems/a-plus-b')
  ).json()
  assert.deepEqual(metadata.judge, {
    problem_id: 'a-plus-b',
    enabled: true,
    case_count: 1,
    wall_time_ms: 3500,
    memory_limit_bytes: 33554432,
    checker: 'exact-v1',
    problem_type: 'batch',
  })
  const exact = await judge('#include <cstdio>\nint main(){printf("3");}')
  assert.equal(exact.status, 200)
  assert.equal(exact.result.verdict, 'WA')
  assert.equal(exact.result.cases.length, 1)
  clean()
  m.checker = 'tokens-v1'
  writeFileSync(manifestPath, JSON.stringify(m))
  const tokens = await judge('#include <cstdio>\nint main(){printf("3");}')
  assert.equal(tokens.status, 200)
  assert.equal(tokens.result.verdict, 'AC')
  assert.notEqual(tokens.result.testdata_sha256, exact.result.testdata_sha256)
  clean()
  report.tests.push({
    name: 'dynamic-metadata-and-token-exact',
    passed: true,
    metadata: metadata.judge,
    exact,
    tokens,
  })
  save()
} finally {
  writeFileSync(manifestPath, originalManifest)
}
const restored = await (await fetch(base + '/api/v1/problems/a-plus-b')).json()
assert.deepEqual(restored.judge, detail.judge)
report.completed = true
report.finished_at = new Date().toISOString()
save()
console.log(
  'PASS busy 429, cancellation cleanup, file errors, exact/tokens and dynamic metadata; evidence .local/evidence/judge-matrix.json',
)
