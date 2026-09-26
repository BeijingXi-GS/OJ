import { spawnSync } from 'node:child_process'
import { existsSync } from 'node:fs'
import path from 'node:path'
import { root, goEnv } from './lib.mjs'

let failed = false
console.log(
  'Node ' + process.version + '; ' + process.platform + '/' + process.arch,
)
for (const [name, args] of [
  ['git', ['--version']],
  ['go', ['version']],
  ['docker', ['compose', 'version']],
  ['docker', ['info', '--format', '{{.ServerVersion}}']],
]) {
  const r = spawnSync(name, args, {
    encoding: 'utf8',
    timeout: 15000,
    env: goEnv(),
    windowsHide: true,
  })
  const ok = r.status === 0
  console.log(
    (ok ? 'OK ' : 'BLOCKED ') +
      name +
      ' ' +
      args.join(' ') +
      ': ' +
      (r.stdout?.trim() ||
        r.error?.message ||
        r.stderr?.trim() ||
        'No diagnostic output'),
  )
  failed ||= !ok
}
for (const item of ['.env', 'node_modules', 'package-lock.json', 'go.sum']) {
  const ok = existsSync(path.join(root, item))
  console.log((ok ? 'OK ' : 'MISSING ') + item)
  failed ||= !ok
}
console.log(
  'Windows supports API/web development. Real judge isolation requires Linux + cgroups v2 (Phase 4).',
)
process.exitCode = failed ? 1 : 0
