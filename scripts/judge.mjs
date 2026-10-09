import { spawnSync } from 'node:child_process'
import { copyFileSync, mkdirSync, writeFileSync } from 'node:fs'
import path from 'node:path'
import { goEnv, local, root, run } from './lib.mjs'

const action = process.argv[2]
if (action === 'build') {
  const output = path.join(local, 'judge-build')
  mkdirSync(output, { recursive: true })
  run(
    'go',
    [
      'build',
      '-mod=readonly',
      '-trimpath',
      '-buildvcs=false',
      '-o',
      path.join(output, 'judge-compile-helper'),
      './services/judge-compile-helper',
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
    path.join(root, 'deploy/judge/Dockerfile'),
    path.join(output, 'Dockerfile'),
  )
  run('docker', [
    'build',
    '--platform',
    'linux/amd64',
    '-t',
    'oj/judge:gcc14.2.0-mvp',
    output,
  ])
  const inspect = spawnSync(
    'docker',
    ['image', 'inspect', 'oj/judge:gcc14.2.0-mvp', '--format', '{{.Id}}'],
    { cwd: root, encoding: 'utf8', windowsHide: true },
  )
  const image = inspect.stdout?.trim()
  if (inspect.status !== 0 || !/^sha256:[a-f0-9]{64}$/.test(image || ''))
    throw new Error('Cannot lock judge image')
  writeFileSync(
    path.join(local, 'judge-profile.json'),
    JSON.stringify(
      {
        version: 'judge-mvp/0.1.0',
        image,
        reference_commit: '4502d2a798f3130d3c540c5c31d45251ca103623',
      },
      null,
      2,
    ) + '\n',
  )
  console.log(`Locked judge profile: ${image}`)
} else if (action === 'serve') {
  run('go', ['run', '-mod=readonly', './services/judge-mvp'], { env: goEnv() })
} else if (action === 'status') {
  run('docker', ['ps', '-a', '--filter', 'label=oj.judge.owner=file-mvp-v1'])
} else throw new Error('Use judge.mjs build|serve|status')
