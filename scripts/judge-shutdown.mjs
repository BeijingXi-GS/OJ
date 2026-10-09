// Only an owned child's successful exit and a successful cleanup inspection
// confirm shutdown. A closed HTTP listener alone proves neither.
export async function shutdownOwnedJudge({
  child,
  requestShutdown,
  confirmCleanup,
  timeoutMs = 125000,
}) {
  if (!child) return
  const exited = new Promise((resolve, reject) => {
    if (child.exitCode !== null || child.signalCode !== null) {
      resolve({ code: child.exitCode, signal: child.signalCode })
      return
    }
    const timer = setTimeout(() => {
      child.removeListener('exit', finished)
      reject(
        new Error(
          'Owned judge did not confirm exit within the shutdown budget',
        ),
      )
    }, timeoutMs)
    function finished(code, signal) {
      clearTimeout(timer)
      resolve({ code, signal })
    }
    child.once('exit', finished)
  })
  // A shared Windows console can deliver Ctrl+C to both processes. The judge
  // may already be closing its listener before our control request arrives.
  if (child.exitCode === null && child.signalCode === null) {
    try {
      await requestShutdown()
    } catch {
      // Never infer success here; await the actual exit and verify cleanup.
    }
  }
  const state = await exited
  if (state.code !== 0 || state.signal !== null)
    throw new Error(
      'Owned judge exited abnormally; cleanup requires inspection',
    )
  await confirmCleanup()
}
