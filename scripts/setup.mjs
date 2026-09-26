import { existsSync, readFileSync, writeFileSync } from 'node:fs'
import { randomBytes } from 'node:crypto'
import path from 'node:path'
import { root, goEnv, run } from './lib.mjs'

const envFile = path.join(root, '.env')
if (!existsSync(envFile)) {
  const password = randomBytes(24).toString('hex')
  const content = readFileSync(path.join(root, '.env.example'), 'utf8')
    .replace('POSTGRES_PASSWORD=CHANGE_ME', 'POSTGRES_PASSWORD=' + password)
    .replace(
      'MINIO_ROOT_PASSWORD=CHANGE_ME',
      'MINIO_ROOT_PASSWORD=' + randomBytes(24).toString('hex'),
    )
    .replace(
      '# DATABASE_URL=postgres://oj:REPLACE_ME',
      'DATABASE_URL=postgres://oj:' + password,
    )
  writeFileSync(envFile, content, { mode: 0o600 })
  console.log('Created .env with local credentials (not printed).')
} else {
  console.log('Keeping existing .env.')
}
if (!process.env.npm_execpath) {
  console.error('Run this script through npm run setup.')
  process.exit(1)
}
run(process.execPath, [process.env.npm_execpath, 'ci'])
run('go', ['mod', 'download'], { env: goEnv() })
console.log('Dependencies ready. Run npm run doctor, then npm run infra:up.')
