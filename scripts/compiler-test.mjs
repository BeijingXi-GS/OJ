import path from 'node:path'
import { goEnv, local, run } from './lib.mjs'

const binary = path.join(local, 'compilerworker-linux.test')
run('go', ['test', '-c', '-o', binary, './internal/compilerworker'], {
  env: {
    ...goEnv(),
    GOOS: 'linux',
    GOARCH: 'amd64',
    CGO_ENABLED: '0',
    GOTOOLCHAIN: 'local',
  },
})
run('docker', [
  'run',
  '--rm',
  '--init',
  '--network',
  'none',
  '--mount',
  `type=bind,source=${binary},target=/tests/worker,readonly`,
  '--entrypoint',
  '/tests/worker',
  'oj/compile-worker:gcc14.2.0-mvp',
  '-test.v',
])
