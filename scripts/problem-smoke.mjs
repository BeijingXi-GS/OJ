import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { mkdirSync, readFileSync, readdirSync, writeFileSync } from 'node:fs'
import path from 'node:path'
import { local, root } from './lib.mjs'

const base = process.env.JUDGE_TEST_API || 'http://127.0.0.1:8080'
// These implementations deliberately differ from the public starter strings.
// Wrong solutions model common mistakes and must compile successfully, then WA.
const solutions = [
  {
    id: 'a-plus-b',
    correct:
      '#include <cstdio>\nint main(){long long x,y;if(scanf("%lld%lld",&x,&y)!=2)return 1;printf("%lld\\n",x+y);}',
    wrong:
      '#include <cstdio>\nint main(){long long x,y;scanf("%lld%lld",&x,&y);printf("%lld\\n",x*y);}',
    boundary: 'zero / negative / ±10^9',
  },
  {
    id: 'array-sum',
    correct:
      '#include <iostream>\n#include <numeric>\n#include <vector>\nint main(){int n;std::cin>>n;std::vector<long long> a(n);for(auto &x:a)std::cin>>x;std::cout<<std::accumulate(a.begin(),a.end(),0LL)<<"\\n";}',
    wrong:
      '#include <iostream>\nint main(){int n;long long first;std::cin>>n>>first;std::cout<<first<<"\\n";}',
    boundary: 'n=1 / n=1000 / total=10^12 / positive-negative cancellation',
  },
  {
    id: 'maximum-element',
    correct:
      '#include <algorithm>\n#include <iostream>\n#include <vector>\nint main(){int n;std::cin>>n;std::vector<long long>a(n);for(auto &x:a)std::cin>>x;std::cout<<*std::max_element(a.begin(),a.end())<<"\\n";}',
    wrong:
      '#include <algorithm>\n#include <iostream>\nint main(){int n;std::cin>>n;long long best=0,x;while(n--){std::cin>>x;best=std::max(best,x);}std::cout<<best<<"\\n";}',
    boundary: 'single minimum / all-negative / duplicate / n=1000 last maximum',
  },
  {
    id: 'palindrome',
    correct:
      '#include <algorithm>\n#include <iostream>\n#include <string>\nint main(){std::string s;std::cin>>s;bool ok=std::equal(s.begin(),s.begin()+s.size()/2,s.rbegin());std::cout<<(ok?"YES":"NO")<<"\\n";}',
    wrong:
      '#include <iostream>\n#include <string>\nint main(){std::string s;std::cin>>s;std::cout<<(s.front()==s.back()?"YES":"NO")<<"\\n";}',
    boundary:
      'length=1 / odd-even palindrome / internal mismatch / length=1000',
  },
  {
    id: 'sort-numbers',
    correct:
      '#include <iostream>\n#include <set>\nint main(){int n;std::cin>>n;std::multiset<long long>a;while(n--){long long x;std::cin>>x;a.insert(x);}for(auto x:a)std::cout<<x<<" ";std::cout<<"\\n";}',
    wrong:
      '#include <iostream>\n#include <set>\nint main(){int n;std::cin>>n;std::set<long long>a;while(n--){long long x;std::cin>>x;a.insert(x);}for(auto x:a)std::cout<<x<<" ";std::cout<<"\\n";}',
    boundary: 'single minimum / duplicates / ±10^9 / n=1000 descending',
  },
]
const report = {
  started_at: new Date().toISOString(),
  tests: [],
  completed: false,
}
const evidence = path.join(local, 'evidence')
mkdirSync(evidence, { recursive: true })
function save() {
  writeFileSync(
    path.join(evidence, 'problem-smoke.json'),
    JSON.stringify(report, null, 2) + '\n',
  )
}
function clean() {
  const query = spawnSync(
    'docker',
    ['ps', '-aq', '--filter', 'label=oj.judge.owner=file-mvp-v1'],
    { cwd: root, encoding: 'utf8', timeout: 5000, windowsHide: true },
  )
  assert.equal(query.status, 0, 'Successful Docker cleanup query is required')
  assert.equal(query.stdout.trim(), '', 'No judge containers may remain')
  assert.deepEqual(readdirSync(path.join(local, 'judge-jobs')), [])
}
const listResponse = await fetch(base + '/api/v1/problems', {
  signal: AbortSignal.timeout(10000),
})
assert.equal(listResponse.status, 200)
const list = await listResponse.json()
assert.deepEqual(
  list.problems.filter((p) => !p.id.startsWith('upload-')).map((p) => p.id),
  [...solutions.map((p) => p.id), 'guess-number'],
)
save()
for (const solution of solutions) {
  const response = await fetch(base + '/api/v1/problems/' + solution.id, {
    signal: AbortSignal.timeout(10000),
  })
  assert.equal(response.status, 200)
  const detail = await response.json()
  const manifest = JSON.parse(
    readFileSync(
      path.join(root, 'testdata/problems', solution.id, 'manifest.json'),
      'utf8',
    ),
  )
  assert.equal(detail.id, solution.id)
  assert.deepEqual(detail.judge, {
    problem_id: solution.id,
    enabled: true,
    case_count: manifest.cases.length,
    wall_time_ms: manifest.wall_time_ms,
    memory_limit_bytes: manifest.memory_limit_bytes,
    checker: manifest.checker,
    problem_type: 'batch',
  })
  assert.ok(detail.samples.length >= 1)
  assert.ok(detail.description && detail.input_description && detail.notes)
  assert.notEqual(solution.correct, detail.starter_code)
  assert.notEqual(solution.wrong, detail.starter_code)
  for (const [kind, source, expected] of [
    ['independent-correct', solution.correct, 'AC'],
    ['common-mistake', solution.wrong, 'WA'],
  ]) {
    const response = await fetch(base + '/api/v1/judgements', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-OJ-Mode': 'engineer' },
      body: JSON.stringify({
        problem_id: solution.id,
        language: 'cpp',
        source,
      }),
      signal: AbortSignal.timeout(110000),
    })
    const result = await response.json()
    const item = {
      name: `${solution.id}/${kind}`,
      boundary: solution.boundary,
      expected,
      http_status: response.status,
      result,
    }
    report.tests.push(item)
    save()
    assert.equal(response.status, 200, JSON.stringify(result))
    assert.equal(result.compile.status, 'success', item.name)
    assert.equal(result.verdict, expected, item.name)
    assert.equal(result.cases.length, manifest.cases.length, item.name)
    assert.deepEqual(
      result.cases.map((c) => c.id),
      manifest.cases.map((c) => c.id),
    )
    assert.ok(
      result.cases.every((c) =>
        expected === 'AC'
          ? c.verdict === 'AC'
          : c.verdict === 'AC' || c.verdict === 'WA',
      ),
    )
    if (expected === 'WA')
      assert.ok(
        result.cases.some((c) => c.verdict === 'WA'),
        item.name,
      )
    clean()
    item.passed = true
    save()
    console.log(
      `PASS ${item.name}: ${result.verdict}, ${result.cases.length} cases`,
    )
  }
}
for (const id of ['missing', '../a-plus-b', 'a-plus-b%2f..']) {
  const response = await fetch(base + '/api/v1/judgements', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      problem_id: id,
      language: 'cpp',
      source: 'int main(){}',
    }),
    signal: AbortSignal.timeout(10000),
  })
  assert.equal(response.status, 404)
}
const unknown = await fetch(base + '/api/v1/problems/missing')
assert.equal(unknown.status, 404)
clean()
report.completed = true
report.finished_at = new Date().toISOString()
save()
console.log(
  'PASS all 5 problem details, 10 real judgements, allowlist and cleanup',
)
