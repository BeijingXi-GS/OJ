import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { mkdirSync, readFileSync, readdirSync, writeFileSync } from 'node:fs'
import path from 'node:path'
import { local, root } from './lib.mjs'

const base = process.env.JUDGE_TEST_API || 'http://127.0.0.1:8080'
const evidence = path.join(local, 'evidence')
mkdirSync(evidence, { recursive: true })
const report = {
  started_at: new Date().toISOString(),
  tests: [],
  completed: false,
}
function save() {
  writeFileSync(
    path.join(evidence, 'interactive-matrix.json'),
    JSON.stringify(report, null, 2) + '\n',
  )
}
function containers() {
  const result = spawnSync(
    'docker',
    ['ps', '-aq', '--filter', 'label=oj.judge.owner=file-mvp-v1'],
    { cwd: root, encoding: 'utf8', timeout: 5000, windowsHide: true },
  )
  assert.equal(result.status, 0, 'Cleanup query must succeed')
  return result.stdout.trim().split(/\s+/).filter(Boolean)
}
function clean() {
  assert.deepEqual(containers(), [], 'No owned judge containers may remain')
  assert.deepEqual(
    readdirSync(path.join(local, 'judge-jobs')),
    [],
    'No job snapshot may remain',
  )
}
async function submit(
  source,
  mode = 'engineer',
  signal = AbortSignal.timeout(110000),
) {
  const response = await fetch(base + '/api/v1/judgements', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-OJ-Mode': mode },
    body: JSON.stringify({
      problem_id: 'guess-number',
      language: 'cpp',
      source,
    }),
    signal,
  })
  return { status: response.status, result: await response.json() }
}
const detailResponse = await fetch(base + '/api/v1/problems/guess-number', {
  signal: AbortSignal.timeout(10000),
})
assert.equal(detailResponse.status, 200)
const detail = await detailResponse.json()
assert.equal(detail.type, 'interactive')
assert.equal(detail.judge.problem_type, 'interactive')
assert.equal(detail.judge.checker, 'interaction-v1')
assert.equal(detail.judge.case_count, 3)
report.profile = JSON.parse(
  readFileSync(path.join(local, 'judge-profile.json'), 'utf8'),
)
const cases = [
  ['AC-binary-search', detail.starter_code, 'AC'],
  [
    'WA-final',
    '#include <iostream>\nint main(){int n;std::cin>>n;std::cout<<"! 0"<<std::endl;}',
    'WA',
  ],
  [
    'WA-supervised-stop',
    '#include <iostream>\nint main(){int n;std::cin>>n;std::cout<<"! 0"<<std::endl;for(;;){}}',
    'WA',
  ],
  [
    'WA-malformed',
    '#include <iostream>\nint main(){int n;std::cin>>n;std::cout<<"bad protocol"<<std::endl;}',
    'WA',
  ],
  [
    'WA-query-limit',
    '#include <iostream>\nint main(){int n;std::cin>>n;for(int i=0;i<11;i++){std::cout<<"? 500"<<std::endl;char c;if(!(std::cin>>c))return 0;}}',
    'WA',
  ],
  ['WA-EOF', 'int main(){return 0;}', 'WA'],
  ['CE', 'int main(){ this is not C++; }', 'CE'],
  ['RE-crash', '#include <cstdlib>\nint main(){abort();}', 'RE'],
  ['RE-137-not-OOM', 'int main(){return 137;}', 'RE'],
  [
    'TLE-no-flush',
    '#include <iostream>\nint main(){std::ios::sync_with_stdio(false);std::cin.tie(nullptr);int n;std::cin>>n;std::cout<<"? 500\\n";char r;std::cin>>r;}',
    'TLE',
  ],
  [
    'MLE',
    '#include <cstring>\n#include <vector>\nint main(){std::vector<char*> v;for(;;){auto p=new char[1048576];memset(p,1,1048576);v.push_back(p);}}',
    'MLE',
  ],
  [
    'OLE-stderr',
    '#include <unistd.h>\nint main(){char b[4096]={};for(;;)write(2,b,sizeof(b));}',
    'OLE',
  ],
  [
    'WA-descendant-cleanup',
    '#include <iostream>\n#include <unistd.h>\nint main(){int n;std::cin>>n;if(fork()==0){for(;;)pause();}std::cout<<"! 0"<<std::endl;for(;;)pause();}',
    'WA',
  ],
]
save()
for (const [name, source, expected] of cases) {
  const answer = await submit(source)
  const item = { name, expected, ...answer }
  report.tests.push(item)
  save()
  assert.equal(answer.status, 200, JSON.stringify(answer.result))
  assert.equal(answer.result.verdict, expected, name)
  if (expected !== 'CE') {
    assert.equal(answer.result.compile.status, 'success')
    assert.equal(answer.result.interactor_compile.status, 'success')
    assert.equal(answer.result.cases.length, 3)
    for (const row of answer.result.cases) {
      assert.equal(row.verdict, expected, `${name}/${row.id}`)
      assert.ok(row.interactor)
      if (expected === 'AC') {
        assert.equal(row.exit_code, 0)
        assert.equal(row.interactor.exit_code, 0)
        assert.equal(row.interactor.participant_stopped, false)
      }
      if (expected === 'MLE') assert.equal(row.oom_killed, true)
      if (expected === 'RE') assert.equal(row.oom_killed, false)
      if (expected === 'WA') {
        assert.equal(row.interactor.exit_code, 1)
        assert.equal(row.oom_killed, false)
        if (row.exit_code !== 0)
          assert.equal(row.interactor.participant_stopped, true)
      }
      if (name === 'WA-supervised-stop' || name === 'WA-descendant-cleanup')
        assert.equal(row.interactor.participant_stopped, true)
    }
  }
  clean()
  item.passed = true
  save()
  console.log(`PASS ${name}: ${expected}`)
}

// Observe both runtime sides before aborting, then require cleanup with no retry
// which might hide a previous execution's leftover container or snapshot.
const controller = new AbortController()
const pending = submit(
  'int main(){for(;;){}}',
  'engineer',
  controller.signal,
).then(
  () => false,
  (error) => error.name === 'AbortError',
)
let paired = false
for (let i = 0; i < 100; i++) {
  if (containers().length >= 2) {
    paired = true
    break
  }
  await new Promise((resolve) => setTimeout(resolve, 200))
}
assert.equal(
  paired,
  true,
  'Must observe the two runtime containers before cancellation',
)
controller.abort()
assert.equal(await pending, true)
let cleaned = false
for (let i = 0; i < 100; i++) {
  if (
    containers().length === 0 &&
    readdirSync(path.join(local, 'judge-jobs')).length === 0
  ) {
    cleaned = true
    break
  }
  await new Promise((resolve) => setTimeout(resolve, 200))
}
assert.equal(cleaned, true, 'Both sides must be cleaned after cancellation')
clean()
report.tests.push({ name: 'cancel-paired-runtimes', passed: true })
report.completed = true
report.finished_at = new Date().toISOString()
save()
console.log(
  'PASS 13 real interactive judgements, paired cancellation and cleanup',
)
