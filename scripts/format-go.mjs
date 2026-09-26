import { readdirSync } from 'node:fs'
import { spawnSync } from 'node:child_process'
import path from 'node:path'
import { root } from './lib.mjs'

function files(dir) {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    if (['.git', '.local', 'node_modules', 'dist'].includes(entry.name))
      return []
    const name = path.join(dir, entry.name)
    return entry.isDirectory()
      ? files(name)
      : entry.name.endsWith('.go')
        ? [name]
        : []
  })
}
const write = process.argv.includes('--write')
const result = spawnSync('gofmt', [write ? '-w' : '-l', ...files(root)], {
  encoding: 'utf8',
  windowsHide: true,
})
if (result.error || result.status !== 0) {
  console.error(result.error?.message || result.stderr)
  process.exit(1)
}
if (!write && result.stdout.trim()) {
  console.error(result.stdout)
  process.exit(1)
}
