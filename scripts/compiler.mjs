import { copyFileSync, mkdirSync } from 'node:fs'
import path from 'node:path'
import { goEnv, local, root, run } from './lib.mjs'

const action = process.argv[2]
const compose = [
  'compose',
  '-f',
  path.join(root, 'deploy/compiler/compose.yaml'),
]
if (action === 'build') {
  const output = path.join(local, 'compiler-build')
  mkdirSync(output, { recursive: true })
  run(
    'go',
    [
      'build',
      '-mod=readonly',
      '-trimpath',
      '-buildvcs=false',
      '-o',
      path.join(output, 'compile-worker'),
      './services/compile-worker',
    ],
    {
      env: {
        ...goEnv(),
        GOOS: 'linux',
        GOARCH: 'amd64',
        CGO_ENABLED: '0',
        GOTOOLCHAIN: 'local',
      },
    },
  )
  copyFileSync(
    path.join(root, 'deploy/compiler/Dockerfile'),
    path.join(output, 'Dockerfile'),
  )
  run('docker', [
    'build',
    '--platform',
    'linux/amd64',
    '-t',
    'oj/compile-worker:gcc14.2.0-mvp',
    output,
  ])
} else if (action === 'up') {
  run('docker', [...compose, 'up', '-d'])
} else if (action === 'down') {
  run('docker', [...compose, 'down'])
} else if (action === 'status') {
  run('docker', [...compose, 'ps'])
} else {
  throw new Error('Use compiler.mjs build|up|down|status')
}
