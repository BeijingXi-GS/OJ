import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { readFileSync, writeFileSync } from 'node:fs'
import path from 'node:path'
import { goEnv, local, root, run } from './lib.mjs'

const binary = path.join(
  local,
  process.platform === 'win32' ? 'judge-refusal.exe' : 'judge-refusal',
)
run('go', ['build', '-mod=readonly', '-o', binary, './services/judge-mvp'], {
  env: goEnv(),
})
const inspect = spawnSync(
  'docker',
  [
    'image',
    'inspect',
    'oj/compile-worker:gcc14.2.0-mvp',
    '--format',
    '{{.Id}}',
  ],
  { cwd: root, encoding: 'utf8', timeout: 5000, windowsHide: true },
)
assert.equal(inspect.status, 0)
const profilePath = path.join(local, 'judge-profile.json')
const original = readFileSync(profilePath)
const p = JSON.parse(original.toString())
p.image = inspect.stdout.trim()
let result
try {
  writeFileSync(profilePath, JSON.stringify(p))
  // Existing ready worker keeps its immutable in-memory profile. This new
  // process must refuse before opening the already-used listener.
  result = spawnSync(binary, [], {
    cwd: root,
    encoding: 'utf8',
    timeout: 70000,
    windowsHide: true,
  })
  assert.equal(result.status, 1)
  assert.match(result.stdout, /image_volumes_forbidden/)
  assert.doesNotMatch(result.stdout, /judge ready/)
} finally {
  writeFileSync(profilePath, original)
}
writeFileSync(
  path.join(local, 'evidence/judge-startup-refusal.json'),
  JSON.stringify(
    {
      tested_at: new Date().toISOString(),
      candidate_image: p.image,
      expected_reason: 'image_volumes_forbidden',
      exit_code: result.status,
      stdout: result.stdout,
      passed: true,
    },
    null,
    2,
  ) + '\n',
)
console.log(
  'PASS real preflight refuses inherited /data VOLUME; profile restored',
)
