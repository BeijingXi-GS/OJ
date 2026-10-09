import { useSyncExternalStore } from 'react'

export type Mode = 'student' | 'teacher' | 'engineer'

let currentMode: Mode = 'student'
const listeners = new Set<() => void>()
const requests = new Set<AbortController>()

export function getMode() {
  return currentMode
}

export function changeMode(mode: Mode) {
  if (mode === currentMode) return
  for (const controller of requests) controller.abort()
  requests.clear()
  currentMode = mode
  for (const listener of listeners) listener()
}

export function registerModeRequest(controller: AbortController) {
  requests.add(controller)
  return () => requests.delete(controller)
}

function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

export function useMode() {
  return useSyncExternalStore(subscribe, getMode, getMode)
}

export const modeLabels: Record<Mode, string> = {
  student: '学生模式',
  teacher: '教师模式',
  engineer: '工程师模式',
}
