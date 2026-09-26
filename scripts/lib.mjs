import { spawnSync } from 'node:child_process'
import { mkdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import path from 'node:path'

export const root = fileURLToPath(new URL('../', import.meta.url))
export const local = path.join(root, '.local')
export function goEnv() {
  const env = { ...process.env }
  for (const [key, dir] of Object.entries({
    GOPATH: 'go',
    GOMODCACHE: 'go/pkg/mod',
    GOCACHE: 'go-build',
    GOTMPDIR: 'tmp',
  })) {
    env[key] = path.join(local, dir)
    mkdirSync(env[key], { recursive: true })
  }
  return env
}
export function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: root,
    stdio: 'inherit',
    windowsHide: true,
    ...options,
  })
  if (result.error) console.error(result.error.message)
  if (result.status !== 0) process.exit(result.status || 1)
}
