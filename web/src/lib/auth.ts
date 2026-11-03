import { useSyncExternalStore } from 'react'

/**
 * The operator's login and password, kept as the ready Basic header value. The
 * server has no sessions: every request carries the pair.
 */
const STORAGE_KEY = 'credentials'
const listeners = new Set<() => void>()

function notify() {
  for (const listener of listeners) {
    listener()
  }
}

export function getCredentials(): string | null {
  try {
    return localStorage.getItem(STORAGE_KEY)
  } catch {
    return null
  }
}

export function encodeCredentials(login: string, password: string): string {
  return btoa(unescape(encodeURIComponent(`${login}:${password}`)))
}

export function setCredentials(value: string) {
  try {
    localStorage.setItem(STORAGE_KEY, value)
  } catch {
    // storage is unavailable: the login lives until reload
  }
  notify()
}

export function clearCredentials() {
  try {
    localStorage.removeItem(STORAGE_KEY)
  } catch {
    // nothing to clear
  }
  notify()
}

function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

export function useCredentials(): string | null {
  return useSyncExternalStore(subscribe, getCredentials, () => null)
}
