import { useSyncExternalStore } from 'react'

export type Theme = 'system' | 'light' | 'dark'

const STORAGE_KEY = 'theme'
const order: Theme[] = ['system', 'light', 'dark']

export const themeLabels: Record<Theme, string> = {
  system: 'как в системе',
  light: 'светлая',
  dark: 'тёмная',
}

const listeners = new Set<() => void>()

export function readTheme(): Theme {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    return stored === 'light' || stored === 'dark' ? stored : 'system'
  } catch {
    return 'system'
  }
}

export function applyTheme(theme: Theme = readTheme()) {
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  document.documentElement.classList.toggle(
    'dark',
    theme === 'dark' || (theme !== 'light' && prefersDark),
  )
}

export function setTheme(theme: Theme) {
  try {
    if (theme === 'system') {
      localStorage.removeItem(STORAGE_KEY)
    } else {
      localStorage.setItem(STORAGE_KEY, theme)
    }
  } catch {
    // storage is unavailable
  }
  applyTheme(theme)
  for (const listener of listeners) {
    listener()
  }
}

export function nextTheme(theme: Theme): Theme {
  return order[(order.indexOf(theme) + 1) % order.length]
}

export function useTheme(): Theme {
  return useSyncExternalStore(
    (listener) => {
      listeners.add(listener)
      return () => {
        listeners.delete(listener)
      }
    },
    readTheme,
    () => 'system',
  )
}
