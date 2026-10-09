import assert from 'node:assert/strict'
import { EventEmitter } from 'node:events'
import { test } from 'node:test'
import { shutdownOwnedJudge } from './judge-shutdown.mjs'

function child() {
  const c = new EventEmitter()
  c.exitCode = null
  c.signalCode = null
  c.finish = (code, signal = null) => {
    c.exitCode = code
    c.signalCode = signal
    c.emit('exit', code, signal)
  }
  return c
}
test('shared signal: failed HTTP plus clean owned exit and inspection succeeds', async () => {
  const c = child()
  let checked = false
  await shutdownOwnedJudge({
    child: c,
    requestShutdown: async () => {
      setTimeout(() => c.finish(0), 5)
      throw new Error('listener closed')
    },
    confirmCleanup: async () => {
      checked = true
    },
  })
  assert.equal(checked, true)
})
test('202 is not completion: waits for owned process before inspection', async () => {
  const c = child()
  await shutdownOwnedJudge({
    child: c,
    requestShutdown: async () => {
      setTimeout(() => c.finish(0), 15)
    },
    confirmCleanup: async () => {
      assert.equal(c.exitCode, 0)
    },
  })
})
test('abnormal exit, timeout and failed cleanup never become success', async () => {
  for (const reason of ['exit', 'timeout', 'inspection']) {
    const c = child()
    await assert.rejects(
      shutdownOwnedJudge({
        child: c,
        timeoutMs: 20,
        requestShutdown: async () => {
          if (reason !== 'timeout')
            setTimeout(() => c.finish(reason === 'exit' ? 1 : 0), 2)
        },
        confirmCleanup: async () => {
          if (reason === 'inspection') throw new Error('docker query failed')
        },
      }),
    )
  }
})
test('already exited child still requires cleanup; unowned judge is untouched', async () => {
  const c = child()
  c.finish(0)
  let checks = 0
  const options = {
    requestShutdown: async () => {
      assert.fail('must not send')
    },
    confirmCleanup: async () => {
      checks++
    },
  }
  await shutdownOwnedJudge({ ...options, child: c })
  await shutdownOwnedJudge(options)
  assert.equal(checks, 1)
})
