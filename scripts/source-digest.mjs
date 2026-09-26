import { spawnSync } from 'node:child_process'
import { readFileSync } from 'node:fs'
import { createHash } from 'node:crypto'
import path from 'node:path'
import { root } from './lib.mjs'

const result = spawnSync(
  'git',
  ['ls-files', '--cached', '--others', '--exclude-standard', '-z'],
  { cwd: root, encoding: 'utf8', windowsHide: true },
)
if (result.status !== 0) throw new Error(result.stderr)
const files = [...new Set(result.stdout.split('\0').filter(Boolean))]
  .filter((file) => !file.endsWith('.md') && !file.startsWith('docs/'))
  .sort()
const manifest = files
  .map(
    (file) =>
      createHash('sha256')
        .update(readFileSync(path.join(root, file)))
        .digest('hex') +
      '  ' +
      file,
  )
  .join('\n')
console.log(
  JSON.stringify({
    files: files.length,
    sha256: createHash('sha256').update(manifest).digest('hex'),
  }),
)
