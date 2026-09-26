import { existsSync } from 'node:fs'
import path from 'node:path'
import { root, run } from './lib.mjs'

if (!existsSync(path.join(root, '.env'))) {
  console.error('Missing .env. Run npm run setup first.')
  process.exit(1)
}
run('docker', [
  'compose',
  '--project-name',
  'oj-dev',
  '--env-file',
  '.env',
  '-f',
  'deploy/docker-compose/compose.yaml',
  ...process.argv.slice(2),
])
