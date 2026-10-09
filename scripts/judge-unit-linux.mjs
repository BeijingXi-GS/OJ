import { readFileSync } from 'node:fs'
import path from 'node:path'
import { goEnv, local, root, run } from './lib.mjs'

for (const pkg of ['filejudge', 'problemstore']) {
  const binary = path.join(local, `${pkg}-linux.test`)
  run(
    'go',
    ['test', '-c', '-mod=readonly', '-o', binary, `./internal/${pkg}`],
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
  const profile = JSON.parse(
    readFileSync(path.join(local, 'judge-profile.json'), 'utf8'),
  )
  // This trusted test executable uses fake Docker commands. It gets no daemon socket.
  run('docker', [
    'run',
    '--rm',
    '--init',
    '--network',
    'none',
    '--cap-drop',
    'ALL',
    '--security-opt',
    'no-new-privileges',
    '--read-only',
    '--user',
    '10002:10002',
    '--memory',
    '256m',
    '--memory-swap',
    '256m',
    '--pids-limit',
    '32',
    '--cpus',
    '1',
    '--log-driver',
    'none',
    '--tmpfs',
    '/tmp:rw,nosuid,nodev,noexec,size=67108864',
    '--mount',
    `type=bind,src=${binary},dst=/tests/filejudge.test,readonly`,
    // Trusted unit tests read the registered fixtures via ../../testdata.
    // This mount belongs only to this test binary, never a submitted program.
    '--mount',
    `type=bind,src=${path.join(root, 'testdata')},dst=/testdata,readonly`,
    '--workdir',
    '/tests',
    '--entrypoint',
    '/tests/filejudge.test',
    profile.image,
    '-test.v',
  ])
}
