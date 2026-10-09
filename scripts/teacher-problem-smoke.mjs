import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { mkdirSync, readdirSync, writeFileSync } from 'node:fs'
import path from 'node:path'
import { local, root } from './lib.mjs'

const base = process.env.JUDGE_TEST_API || 'http://127.0.0.1:8080'
const evidence = path.join(local, 'evidence')
mkdirSync(evidence, { recursive: true })
const report = {
  started_at: new Date().toISOString(),
  tests: [],
  created: [],
  completed: false,
}
function save() {
  writeFileSync(
    path.join(evidence, 'teacher-problem-smoke.json'),
    JSON.stringify(report, null, 2) + '\n',
  )
}
function clean() {
  const query = spawnSync(
    'docker',
    ['ps', '-aq', '--filter', 'label=oj.judge.owner=file-mvp-v1'],
    { cwd: root, encoding: 'utf8', timeout: 5000, windowsHide: true },
  )
  assert.equal(query.status, 0, 'Successful cleanup query is required')
  assert.equal(query.stdout.trim(), '', 'No owned judge containers may remain')
  assert.deepEqual(readdirSync(path.join(local, 'judge-jobs')), [])
}
async function call(url, options = {}) {
  const response = await fetch(base + url, {
    ...options,
    signal: AbortSignal.timeout(110000),
  })
  return { status: response.status, body: await response.json() }
}
function form(descriptor, files) {
  const body = new FormData()
  body.append('descriptor', JSON.stringify(descriptor))
  for (const [filename, content] of Object.entries(files))
    body.append('files', new Blob([content], { type: 'text/plain' }), filename)
  return body
}
function descriptor(type) {
  const interactive = type === 'interactive'
  return {
    title: interactive ? '教师上传：交互猜数字' : '教师上传：两数求和',
    summary: interactive
      ? '使用固定猜数协议寻找秘密整数。'
      : '读取两个整数，输出总和。',
    difficulty: '入门',
    description: interactive
      ? '读取初始1000后，通过最多10次? x询问找到秘密数，最后输出! x。每行需要flush。'
      : '读取两个整数a、b，输出a+b。',
    input_description: interactive
      ? '评测器先输出1000，再用<、>、=回复秘密数与询问值的比较。秘密.in不会送给选手。'
      : '一行，两个绝对值不超过100的整数。',
    output_description: interactive
      ? '输出? x询问，确定后输出! x并正常退出。'
      : '输出一个整数和。',
    notes: '由真实教师上传回归创建；仅本机演示，题面与测试文件发布后持久保留。',
    samples: [
      {
        input: interactive ? '评测器：1000\n选手：? 500\n评测器：=\n' : '1 2\n',
        output: interactive ? '选手：! 500\n' : '3\n',
      },
    ],
    starter_code: interactive
      ? '#include <iostream>\nint main(){int n;std::cin>>n;std::cout<<"! 1"<<std::endl;}\n'
      : '#include <iostream>\nint main(){long long a,b;std::cin>>a>>b;std::cout<<a+b<<"\\n";}\n',
    type,
    checker: interactive ? 'interaction-v1' : 'tokens-v1',
    wall_time_ms: interactive ? 5000 : 2000,
    memory_limit_bytes: 67108864,
    cases: interactive
      ? [
          { id: 'secret-small', input: 'small.in' },
          { id: 'secret-large', input: 'large.in' },
        ]
      : [
          { id: 'positive', input: 'positive.in', expected: 'positive.out' },
          { id: 'negative', input: 'negative.in', expected: 'negative.out' },
        ],
    ...(interactive ? { interactor: 'guess-number-v1' } : {}),
  }
}
const examples = [
  {
    type: 'batch',
    descriptor: descriptor('batch'),
    files: {
      'positive.in': '4 9\n',
      'positive.out': '13\n',
      'negative.in': '-8 2\n',
      'negative.out': '-6\n',
    },
    correct:
      '#include <cstdio>\nint main(){long long x,y;if(scanf("%lld%lld",&x,&y)!=2)return 1;printf("%lld\\n",x+y);}',
    wrong:
      '#include <iostream>\nint main(){long long a,b;std::cin>>a>>b;std::cout<<a*b<<"\\n";}',
  },
  {
    type: 'interactive',
    descriptor: descriptor('interactive'),
    files: { 'small.in': '7\n', 'large.in': '900\n' },
    correct:
      "#include <iostream>\nint main(){int upper;std::cin>>upper;int lo=1;while(lo<=upper){int q=(lo+upper)/2;std::cout<<\"? \"<<q<<std::endl;char r;if(!(std::cin>>r))return 2;if(r=='='){std::cout<<\"! \"<<q<<std::endl;return 0;}if(r=='<')upper=q-1;else if(r=='>')lo=q+1;else return 3;}return 4;}",
    wrong:
      '#include <iostream>\nint main(){int n;std::cin>>n;std::cout<<"! 1"<<std::endl;}',
  },
]
save()
for (const example of examples) {
  const denied = await call('/api/v1/problems', {
    method: 'POST',
    body: form(example.descriptor, example.files),
  })
  assert.equal(denied.status, 403, JSON.stringify(denied))
  const created = await call('/api/v1/problems', {
    method: 'POST',
    headers: { 'X-OJ-Mode': 'teacher' },
    body: form(example.descriptor, example.files),
  })
  assert.equal(created.status, 201, JSON.stringify(created))
  assert.match(created.body.id, /^upload-[a-f0-9]{32}$/)
  assert.equal(created.body.problem.id, created.body.id)
  assert.equal(created.body.problem.type, example.type)
  report.created.push({
    id: created.body.id,
    type: example.type,
    descriptor: example.descriptor,
    files: example.files,
  })
  save()
  const detail = await call('/api/v1/problems/' + created.body.id)
  assert.equal(detail.status, 200)
  assert.equal(detail.body.type, example.type)
  assert.equal(detail.body.judge.problem_type, example.type)
  assert.equal(detail.body.judge.case_count, 2)
  for (const [mode, source, expected] of [
    ['student', example.correct, 'AC'],
    ['teacher', example.wrong, 'WA'],
    ['engineer', example.correct, 'AC'],
  ]) {
    const result = await call('/api/v1/judgements', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(mode === 'student' ? {} : { 'X-OJ-Mode': mode }),
      },
      body: JSON.stringify({
        problem_id: created.body.id,
        language: 'cpp',
        source,
      }),
    })
    const item = {
      name: `${example.type}/${mode}/${expected}`,
      expected,
      ...result,
    }
    report.tests.push(item)
    save()
    assert.equal(result.status, 200, JSON.stringify(result))
    assert.equal(result.body.mode, mode)
    assert.equal(result.body.problem_type, example.type)
    assert.equal(result.body.verdict, expected)
    assert.equal(result.body.total_cases, 2)
    assert.equal(result.body.compile.status, 'success')
    if (mode === 'student') {
      assert.equal(result.body.passed_cases, 2)
      for (const key of [
        'cases',
        'run_id',
        'source_sha256',
        'testdata_sha256',
        'artifact_sha256',
        'runtime_image',
        'interactor_compile',
        'interactor_source_sha256',
      ])
        assert.equal(
          Object.hasOwn(result.body, key),
          false,
          `Student leaked ${key}`,
        )
    } else {
      assert.equal(result.body.cases.length, 2)
      assert.deepEqual(
        result.body.cases.map((c) => c.id),
        example.descriptor.cases.map((c) => c.id),
      )
      for (const c of result.body.cases) {
        const reference = example.descriptor.cases.find((x) => x.id === c.id)
        assert.equal(c.input, example.files[reference.input])
        assert.equal(
          c.expected,
          example.type === 'batch' ? example.files[reference.expected] : '',
        )
        if (example.type === 'interactive') assert.ok(c.interactor)
        if (mode === 'teacher')
          for (const key of [
            'exit_code',
            'oom_killed',
            'cpu_time_ms',
            'peak_memory_bytes',
          ])
            assert.equal(
              Object.hasOwn(c, key),
              false,
              `Teacher leaked debug field ${key}`,
            )
      }
      if (mode === 'teacher')
        for (const key of [
          'run_id',
          'source_sha256',
          'testdata_sha256',
          'artifact_sha256',
          'runtime_image',
          'interactor_compile',
        ])
          assert.equal(Object.hasOwn(result.body, key), false)
      else {
        assert.match(result.body.testdata_sha256, /^[a-f0-9]{64}$/)
        assert.ok(result.body.runtime_image)
        if (example.type === 'interactive')
          assert.equal(result.body.interactor_compile.status, 'success')
      }
    }
    clean()
    item.passed = true
    save()
    console.log(`PASS ${item.name}`)
  }
}
const before = await call('/api/v1/problems')
const invalid = descriptor('interactive')
invalid.interactor = 'interactor.cpp'
const rejected = await call('/api/v1/problems', {
  method: 'POST',
  headers: { 'X-OJ-Mode': 'teacher' },
  body: form(invalid, examples[1].files),
})
assert.equal(rejected.status, 400)
const after = await call('/api/v1/problems')
assert.deepEqual(
  after.body.problems.map((p) => p.id),
  before.body.problems.map((p) => p.id),
)
const debug = await call('/api/v1/debug', {
  headers: { 'X-OJ-Mode': 'engineer' },
})
assert.equal(debug.status, 200)
assert.equal(debug.body.problem_count, after.body.problems.length)
assert.equal(
  (await call('/api/v1/debug', { headers: { 'X-OJ-Mode': 'teacher' } })).status,
  403,
)
for (const p of report.created)
  assert.ok(after.body.problems.some((item) => item.id === p.id))
report.completed = true
report.finished_at = new Date().toISOString()
save()
console.log(
  'PASS teacher uploads, batch/interactive real judging, three server projections, rejected source interactor and cleanup',
)
// Root records a real service restart and then rechecks these saved IDs. This
// script intentionally leaves both published demonstration problems in place.
