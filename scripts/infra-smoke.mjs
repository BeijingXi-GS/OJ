import { loadEnvFile } from 'node:process'
import { spawnSync } from 'node:child_process'
import { root } from './lib.mjs'

loadEnvFile('.env')
const args = [
  'compose',
  '--project-name',
  'oj-dev',
  '--env-file',
  '.env',
  '-f',
  'deploy/docker-compose/compose.yaml',
]
function exec(service, command) {
  const r = spawnSync('docker', [...args, 'exec', '-T', service, ...command], {
    cwd: root,
    encoding: 'utf8',
    windowsHide: true,
    timeout: 15000,
  })
  if (r.status !== 0)
    throw new Error(service + ': ' + (r.error?.message || r.stderr))
  return r.stdout.trim()
}
if (
  !exec('postgres', [
    'psql',
    '-U',
    process.env.POSTGRES_USER,
    '-d',
    process.env.POSTGRES_DB,
    '-Atc',
    'SELECT 1',
  ]).includes('1')
)
  throw new Error('PostgreSQL query failed')
if (exec('redis', ['redis-cli', 'ping']) !== 'PONG')
  throw new Error('Redis ping failed')
const response = await fetch(
  'http://127.0.0.1:' + (process.env.MINIO_PORT || 9000) + '/minio/health/live',
  { signal: AbortSignal.timeout(10000) },
)
if (!response.ok) throw new Error('MinIO health check failed')
console.log(
  'PASS: PostgreSQL SELECT 1, Redis PING, MinIO liveness. S3 persistence tests belong to Phase 5.',
)
