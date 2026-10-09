import { spawnSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import {
  chmodSync,
  existsSync,
  mkdirSync,
  readFileSync,
  writeFileSync,
} from 'node:fs'
import path from 'node:path'
import { goEnv, local, root } from './lib.mjs'

const lock = JSON.parse(
  readFileSync(path.join(root, 'deploy/minio/source-lock.json'), 'utf8'),
)
const buildDir = path.join(local, 'minio-build')
mkdirSync(path.join(buildDir, 'data'), { recursive: true })
const env = {
  ...goEnv(),
  GOOS: 'linux',
  GOARCH: 'amd64',
  CGO_ENABLED: '0',
  GOTOOLCHAIN: 'local',
  GOWORK: 'off',
  GOPROXY:
    process.env.GOPROXY || 'https://goproxy.cn,https://proxy.golang.org,direct',
}

function exec(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: root,
    env,
    windowsHide: true,
    encoding: 'utf8',
    timeout: 600000,
    maxBuffer: 16 * 1024 * 1024,
    ...options,
  })
  if (result.error || result.status !== 0) {
    throw new Error(
      command + ': ' + (result.error?.message || result.stderr || 'failed'),
    )
  }
  return result.stdout?.trim()
}

const engine = JSON.parse(exec('docker', ['info', '--format', '{{json .}}']))
if (engine.OSType !== 'linux' || engine.Architecture !== 'x86_64') {
  throw new Error(
    'This pinned development image requires a Linux amd64 engine.',
  )
}
const baseCheck = spawnSync('docker', ['image', 'inspect', lock.runtimeBase], {
  cwd: root,
  env,
  windowsHide: true,
  encoding: 'utf8',
  timeout: 15000,
})
if (baseCheck.error)
  throw new Error('Checking CA image: ' + baseCheck.error.message)
if (baseCheck.status !== 0) {
  console.log('Pulling pinned CA source image: ' + lock.runtimeBase)
  exec('docker', ['pull', lock.runtimeBase], { stdio: 'inherit' })
}
console.log(
  'Downloading checksum-verified official MinIO source at ' + lock.commit,
)
const source = JSON.parse(
  exec('go', ['mod', 'download', '-json', lock.module + '@' + lock.commit]),
)
if (source.Error || !source.Dir || !source.Sum || !source.GoModSum) {
  throw new Error('Official source download did not return checksum metadata.')
}
console.log('Building MinIO ' + source.Version + ' for ' + lock.platform)
exec(
  'go',
  [
    'build',
    '-mod=readonly',
    '-trimpath',
    '-buildvcs=false',
    '-ldflags=-s -w',
    '-o',
    path.join(buildDir, 'minio'),
    '.',
  ],
  { cwd: source.Dir, stdio: 'inherit' },
)
exec(
  'go',
  [
    'build',
    '-trimpath',
    '-buildvcs=false',
    '-ldflags=-s -w',
    '-o',
    path.join(buildDir, 'healthcheck'),
    './scripts/minio-health/main.go',
  ],
  { stdio: 'inherit' },
)
// Module-cache files are read-only. Keep that cache intact and write a writable
// build output, including when repairing a prior copy on Windows.
const buildLicense = path.join(buildDir, 'LICENSE')
if (existsSync(buildLicense)) chmodSync(buildLicense, 0o644)
writeFileSync(buildLicense, readFileSync(path.join(source.Dir, 'LICENSE')), {
  mode: 0o644,
})
const metadata = {
  ...lock,
  moduleVersion: source.Version,
  moduleSum: source.Sum,
  goModSum: source.GoModSum,
  goVersion: exec('go', ['version']),
  binarySha256: createHash('sha256')
    .update(readFileSync(path.join(buildDir, 'minio')))
    .digest('hex'),
}
writeFileSync(
  path.join(buildDir, 'build-metadata.json'),
  JSON.stringify(metadata, null, 2) + '\n',
)
exec(
  'docker',
  [
    'build',
    '--pull=false',
    '--platform=' + lock.platform,
    '--build-arg',
    'RUNTIME_BASE=' + lock.runtimeBase,
    '--label',
    'org.opencontainers.image.source=https://github.com/minio/minio',
    '--label',
    'org.opencontainers.image.revision=' + lock.commit,
    '-t',
    lock.image,
    '-f',
    path.join(root, 'deploy/minio/Dockerfile'),
    buildDir,
  ],
  { stdio: 'inherit' },
)
console.log('Built local development image: ' + lock.image)
console.log('Build metadata: .local/minio-build/build-metadata.json')
