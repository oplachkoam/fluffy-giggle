import { clearCredentials, getCredentials } from '@/lib/auth'

export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export type ApiOptions = {
  method?: 'GET' | 'POST'
  body?: unknown
  form?: FormData
  /** Credentials to try instead of the stored ones, for the login form. */
  credentials?: string
}

export async function api<T = undefined>(path: string, options: ApiOptions = {}): Promise<T> {
  const stored = getCredentials()
  const credentials = options.credentials ?? stored
  const headers: Record<string, string> = {}
  if (credentials) {
    headers.Authorization = `Basic ${credentials}`
  }
  if (options.body !== undefined) {
    headers['Content-Type'] = 'application/json'
  }

  const response = await fetch(`/v1${path}`, {
    method: options.method ?? 'GET',
    headers,
    body: options.form ?? (options.body === undefined ? undefined : JSON.stringify(options.body)),
  })

  if (response.status === 401 && options.credentials === undefined && stored !== null) {
    clearCredentials()
  }
  if (!response.ok) {
    throw new ApiError(response.status, await errorMessage(response))
  }
  if (response.status === 204) {
    return undefined as T
  }
  return (await response.json()) as T
}

/** Fetches a binary resource with the stored credentials, for images in <img>. */
export async function apiBlob(path: string): Promise<Blob> {
  const stored = getCredentials()
  const headers: Record<string, string> = {}
  if (stored) {
    headers.Authorization = `Basic ${stored}`
  }
  const response = await fetch(`/v1${path}`, { headers })
  if (!response.ok) {
    throw new ApiError(response.status, await errorMessage(response))
  }
  return response.blob()
}

/** Opens the event stream with the stored credentials; the signal closes it. */
export async function apiStream(path: string, signal: AbortSignal): Promise<Response> {
  const stored = getCredentials()
  const headers: Record<string, string> = { Accept: 'text/event-stream' }
  if (stored) {
    headers.Authorization = `Basic ${stored}`
  }
  const response = await fetch(`/v1${path}`, { headers, signal })
  if (response.status === 401 && stored !== null) {
    clearCredentials()
  }
  if (!response.ok) {
    throw new ApiError(response.status, await errorMessage(response))
  }
  return response
}

async function errorMessage(response: Response): Promise<string> {
  try {
    const data: unknown = await response.json()
    if (typeof data === 'object' && data !== null && 'message' in data) {
      const message = (data as { message: unknown }).message
      if (typeof message === 'string') {
        return message
      }
    }
  } catch {
    // body is not JSON
  }
  return response.statusText || `HTTP ${response.status}`
}

export function errorText(error: unknown, fallback: string): string {
  if (error instanceof ApiError) {
    if (error.status === 401) return 'Неверный логин или пароль.'
    if (error.status === 404) return 'Не найдено.'
    if (error.status === 409) return 'Действие сейчас невозможно: состояние уже изменилось.'
    if (error.status === 422) return `Проверьте данные: ${error.message}.`
    if (error.status >= 500) return 'Сервер недоступен. Попробуйте позже.'
    return error.message
  }
  return fallback
}
