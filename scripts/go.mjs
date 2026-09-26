import { existsSync } from 'node:fs'
import { loadEnvFile } from 'node:process'
import path from 'node:path'
import { root, goEnv, run } from './lib.mjs'
if (existsSync(path.join(root, '.env'))) loadEnvFile(path.join(root, '.env'))
// Restrict recursive Go commands to project packages; node_modules can contain Go files.
const args = process.argv
  .slice(2)
  .flatMap((arg) =>
    arg === './...'
      ? ['./apps/api/...', './services/...', './internal/...']
      : [arg],
  )
run('go', args, { env: goEnv() })
