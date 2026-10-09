import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function formatMemory(bytes: number) {
  const mib = bytes / (1024 * 1024)
  return `${Number.isInteger(mib) ? mib : mib.toFixed(2)} MiB`
}

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
