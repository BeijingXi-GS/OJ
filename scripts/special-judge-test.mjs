import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import {
  lstatSync,
  mkdirSync,
  readFileSync,
  readdirSync,
  realpathSync,
  writeFileSync,
} from 'node:fs'
import path from 'node:path'
import { local, root } from './lib.mjs'

// Root runs this serially against the local, strictly preflighted worker.
// It creates one new owned problem and only mutates that generated directory.
const base = process.env.JUDGE_TEST_API || 'http://127.0.0.1:8080'
const uploadRoot = path.resolve(
  process.env.PROBLEM_UPLOAD_ROOT || path.join(local, 'teacher-problems'),
)
const example = readFileSync(
  path.join(root, 'testdata/special-judge-example/checker.cpp'),
  'utf8',
)
const source =
  '#include <iostream>\nint main(){int n;std::cin>>n;if(n==2||n==3){std::cout<<-1;return 0;}for(int x=2;x<=n;x+=2)std::cout<<x<<" ";for(int x=1;x<=n;x+=2)std::cout<<x<<" ";}\n'
const reversed =
  '#include <iostream>\n#include <vector>\nint main(){int n;std::cin>>n;if(n==2||n==3){std::cout<<-1;return 0;}std::vector<int>v;for(int x=2;x<=n;x+=2)v.push_back(x);for(int x=1;x<=n;x+=2)v.push_back(x);for(auto i=v.rbegin();i!=v.rend();++i)std::cout<<*i<<" ";}\n'
const evidence = path.join(local, 'evidence')
mkdirSync(evidence, { recursive: true })
const report = {
  started_at: new Date().toISOString(),
  profile: JSON.parse(
    readFileSync(path.join(local, 'judge-profile.json'), 'utf8'),
  ),
  tests: [],
  completed: false,
}
const reportPath = path.join(evidence, 'special-judge-matrix.json')
function save() {
  writeFileSync(reportPath, JSON.stringify(report, null, 2) + '\n')
}
function digest(text) {
  return createHash('sha256').update(text).digest('hex')
}
function clean() {
  const query = spawnSync(
    'docker',
    ['ps', '-aq', '--filter', 'label=oj.judge.owner=file-mvp-v1'],
    { cwd: root, encoding: 'utf8', timeout: 5000, windowsHide: true },
  )
  assert.equal(query.status, 0, 'Docker owner query must succeed')
  assert.equal(query.stdout.trim(), '', 'No owned containers may remain')
  assert.deepEqual(
    readdirSync(path.join(local, 'judge-jobs')),
    [],
    'No jobs may remain',
  )
}
async function judge(problemId, code, mode = 'engineer') {
  const response = await fetch(base + '/api/v1/judgements', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-OJ-Mode': mode },
    body: JSON.stringify({
      problem_id: problemId,
      language: 'cpp',
      source: code,
    }),
    signal: AbortSignal.timeout(110000),
  })
  const result = await response.json()
  assert.equal(response.status, 200, JSON.stringify(result))
  clean()
  return result
}
async function record(name, id, code, expected, extra = () => {}) {
  const result = await judge(id, code)
  assert.equal(result.verdict, expected, `${name}: ${JSON.stringify(result)}`)
  extra(result)
  report.tests.push({
    name,
    problem_id: id,
    verdict: result.verdict,
    testdata_sha256: result.testdata_sha256,
    checker_source_sha256: result.checker_source_sha256,
    checker_artifact_sha256: result.checker_artifact_sha256,
    checker_compile_status: result.checker_compile?.status,
    cases: result.cases.map((c) => ({
      id: c.id,
      verdict: c.verdict,
      checker_exit: c.checker?.exit_code,
      checker_oom: c.checker?.oom_killed,
      checker_output_truncated: c.checker?.output_truncated,
      participant_output_truncated: c.output_truncated,
    })),
    cleanup_confirmed: true,
  })
  save()
  return result
}

save()
clean()
await record(
  'builtin-first-valid',
  'construct-permutation',
  source,
  'AC',
  (r) => assert.equal(r.cases.length, 6),
)
await record('builtin-second-valid', 'construct-permutation', reversed, 'AC')
await record(
  'builtin-adjacent-invalid',
  'construct-permutation',
  '#include <iostream>\nint main(){int n;std::cin>>n;if(n==2||n==3){std::cout<<-1;return 0;}for(int i=1;i<=n;i++)std::cout<<i<<" ";}',
  'WA',
)
await record(
  'builtin-duplicates',
  'construct-permutation',
  '#include <iostream>\nint main(){int n;std::cin>>n;if(n==2||n==3){std::cout<<-1;return 0;}for(int i=0;i<n;i++)std::cout<<1<<" ";}',
  'WA',
)
await record(
  'builtin-false-impossibility',
  'construct-permutation',
  '#include <cstdio>\nint main(){puts("-1");}',
  'WA',
)

const descriptor = {
  title: 'C++ 特判示例：构造不相邻排列',
  summary: '教师上传的特判可以接受不同的合法排列。',
  difficulty: '基础·构造',
  description:
    '构造 1 到 n 的排列，使每对相邻数字的差的绝对值都不为 1；无解则输出 -1。',
  input_description: '一个整数 n，1≤n≤1000。',
  output_description: '合法排列，或无解时单独输出 -1。',
  notes: '可以输出任何满足条件的答案。',
  engineering_notes:
    '本题由 scripts/special-judge-test.mjs 创建；使用独立 C++20 特判。',
  samples: [{ input: '4\n', output: '2 4 1 3\n' }],
  starter_code: source,
  type: 'batch',
  checker: 'cpp-special-v1',
  checker_source: 'checker.cpp',
  wall_time_ms: 2000,
  memory_limit_bytes: 67108864,
  cases: [
    { id: 'sample', input: 'sample.in', expected: 'sample.out' },
    { id: 'impossible', input: 'impossible.in' },
    { id: 'maximum', input: 'maximum.in' },
  ],
}
const form = new FormData()
form.set('descriptor', JSON.stringify(descriptor))
for (const [name, content] of Object.entries({
  'checker.cpp': example,
  'sample.in': '4\n',
  'sample.out': '2 4 1 3\n',
  'impossible.in': '3\n',
  'maximum.in': '1000\n',
}))
  form.append('files', new Blob([content], { type: 'text/plain' }), name)
const uploaded = await fetch(base + '/api/v1/problems', {
  method: 'POST',
  headers: { 'X-OJ-Mode': 'teacher' },
  body: form,
})
const problem = await uploaded.json()
assert.equal(uploaded.status, 201, JSON.stringify(problem))
assert.match(problem.id, /^upload-[a-f0-9]{32}$/)
report.uploaded_problem_id = problem.id
report.expected_checker_source_sha256 = digest(example)
const ownedDirectory = path.resolve(uploadRoot, problem.id)
const ownedRelative = path.relative(uploadRoot, ownedDirectory)
assert.equal(ownedRelative, problem.id)
assert.equal(path.dirname(ownedDirectory), uploadRoot)
const checkerPath = path.join(ownedDirectory, 'checker.cpp')
assert.equal(
  readFileSync(checkerPath, 'utf8'),
  example,
  'Only the script-created matching source may be changed',
)
function replaceOwnedChecker(checker) {
  assert.equal(lstatSync(ownedDirectory).isSymbolicLink(), false)
  assert.equal(lstatSync(checkerPath).isSymbolicLink(), false)
  assert.equal(lstatSync(checkerPath).isFile(), true)
  assert.equal(
    path.dirname(realpathSync(ownedDirectory)),
    realpathSync(uploadRoot),
  )
  writeFileSync(checkerPath, checker)
}
save()

let originalDigest
try {
  originalDigest = (
    await record('cpp-first-valid', problem.id, source, 'AC', (r) => {
      assert.equal(r.checker_source_sha256, digest(example))
      assert.equal(r.cases.length, 3)
      assert.equal(r.checker_compile.status, 'success')
      for (const c of r.cases) assert.equal(c.checker.exit_code, 0)
    })
  ).testdata_sha256
  await record('cpp-second-valid', problem.id, reversed, 'AC')
  await record(
    'cpp-WA-feedback',
    problem.id,
    '#include <iostream>\nint main(){int n;std::cin>>n;for(int i=1;i<=n;i++)std::cout<<i<<" ";}',
    'WA',
    (r) => assert.ok(r.cases.some((c) => c.checker_feedback)),
  )
  const large = source
    .replace('for(int x=2;', "std::cout<<std::string(70000,' ');for(int x=2;")
    .replace('#include <iostream>', '#include <iostream>\n#include <string>')
  await record(
    'cpp-raw-output-over-display-limit',
    problem.id,
    large,
    'AC',
    (r) =>
      assert.ok(
        r.cases.some((c) => c.output_truncated && c.checker.exit_code === 0),
      ),
  )
  await record(
    'participant-CE-before-checker',
    problem.id,
    'int main(){ this is not C++; }',
    'CE',
    (r) => {
      assert.equal(r.checker_compile, undefined)
      assert.equal(r.checker_source_sha256, digest(example))
    },
  )
  await record(
    'participant-RE-before-checker',
    problem.id,
    'int main(){return 137;}',
    'RE',
    (r) => {
      for (const c of r.cases) assert.equal(c.checker, undefined)
    },
  )
  await record(
    'participant-TLE-before-checker',
    problem.id,
    'int main(){for(;;){}}',
    'TLE',
    (r) => {
      for (const c of r.cases) assert.equal(c.checker, undefined)
    },
  )
  const referenceChecker =
    '#include <fstream>\n#include <string>\n#include <iostream>\nint main(int argc,char** argv){if(argc!=4)return 2;std::ifstream in(argv[1]),ans(argv[2]),out(argv[3]);int n;if(!(in>>n)||!ans||!out)return 2;std::string a,b;if(!(ans>>a)){std::cerr<<"Optional answer file is empty.";return 0;}do{if(!(out>>b)||a!=b)return 1;}while(ans>>a);return (out>>b)?1:0;}\n'
  replaceOwnedChecker(referenceChecker)
  await record(
    'custom-checker-reads-input-answer-output',
    problem.id,
    source,
    'AC',
    (r) => {
      assert.equal(r.checker_source_sha256, digest(referenceChecker))
      assert.notEqual(r.testdata_sha256, originalDigest)
      assert.ok(
        r.cases.some((c) => c.checker_feedback?.includes('Optional answer')),
      )
    },
  )
  const faults = [
    [
      'checker-compile-error',
      'int main(){ this is not C++; }',
      (r) => {
        assert.equal(r.checker_compile.status, 'compile_error')
        assert.equal(r.cases.length, 0)
      },
    ],
    [
      'checker-exit2',
      'int main(){return 2;}',
      (r) => assert.equal(r.cases[0].checker.exit_code, 2),
    ],
    [
      'checker-exit137-without-OOM',
      'int main(){return 137;}',
      (r) => {
        assert.equal(r.cases[0].checker.exit_code, 137)
        assert.equal(r.cases[0].checker.oom_killed, false)
      },
    ],
    [
      'checker-timeout',
      'int main(){for(;;){}}',
      (r) => assert.ok(r.cases[0].checker.duration_ms >= 1900),
    ],
    [
      'checker-OOM',
      '#include <cstring>\n#include <vector>\nint main(){std::vector<char*> blocks;for(;;){auto p=new char[1048576];memset(p,1,1048576);blocks.push_back(p);}}',
      (r) => assert.equal(r.cases[0].checker.oom_killed, true),
    ],
    [
      'checker-output-limit',
      '#include <unistd.h>\nint main(){char b[4096]={};for(;;)write(2,b,sizeof(b));}',
      (r) => assert.equal(r.cases[0].checker.output_truncated, true),
    ],
  ]
  for (const [name, checker, extra] of faults) {
    replaceOwnedChecker(checker)
    await record(name, problem.id, source, 'SYSTEM_ERROR', (r) => {
      assert.equal(r.checker_source_sha256, digest(checker))
      assert.notEqual(r.testdata_sha256, originalDigest)
      extra(r)
    })
  }
} finally {
  replaceOwnedChecker(example)
  report.checker_source_restored = readFileSync(checkerPath, 'utf8') === example
  save()
}

await record('restored-checker-AC', problem.id, source, 'AC', (r) =>
  assert.equal(r.testdata_sha256, originalDigest),
)
const teacher = await judge(problem.id, source, 'teacher')
assert.equal(teacher.verdict, 'AC')
assert.ok(teacher.cases.every((c) => typeof c.checker_feedback === 'string'))
for (const c of teacher.cases) {
  assert.equal(c.checker, undefined)
  assert.equal(c.exit_code, undefined)
}
for (const field of [
  'checker_compile',
  'checker_source_sha256',
  'checker_artifact_sha256',
  'runtime_image',
])
  assert.equal(teacher[field], undefined)
const student = await judge(problem.id, source, 'student')
assert.equal(student.verdict, 'AC')
for (const field of [
  'cases',
  'checker_compile',
  'checker_source_sha256',
  'checker_artifact_sha256',
  'testdata_sha256',
  'runtime_image',
])
  assert.equal(student[field], undefined)
report.tests.push({
  name: 'teacher-and-student-projections',
  teacher_verdict: teacher.verdict,
  student_verdict: student.verdict,
  cleanup_confirmed: true,
})
await record(
  'old-batch-smoke',
  'a-plus-b',
  '#include <iostream>\nint main(){long long a,b;std::cin>>a>>b;std::cout<<a+b;}',
  'AC',
)
const interactiveDetail = await (
  await fetch(base + '/api/v1/problems/guess-number')
).json()
await record(
  'old-interactive-smoke',
  'guess-number',
  interactiveDetail.starter_code,
  'AC',
)
report.finished_at = new Date().toISOString()
report.completed = true
save()
console.log(
  `Special judge checks passed; evidence: ${reportPath}; retained example: ${problem.id}`,
)
