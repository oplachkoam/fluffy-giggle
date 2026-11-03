import { QueryClientProvider } from '@tanstack/react-query'
import { render } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { vi } from 'vitest'

import { createQueryClient } from '@/app/query-client'
import { routes } from '@/app/router'

export function renderRoute(path: string) {
  const router = createMemoryRouter(routes, { initialEntries: [path] })
  const client = createQueryClient({ retry: 0 })
  const result = render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
  return { ...result, router, client }
}

export type ApiCall = {
  method: string
  path: string
  headers: Headers
  body: unknown
}

type ApiHandler = (call: ApiCall) => Response | Promise<Response>

/**
 * Replaces fetch. Handlers are keyed by `METHOD /path` without `/v1`; `:param`
 * matches one segment. The event stream stays open and silent unless stubbed.
 */
export function stubApi(handlers: Record<string, ApiHandler>): ApiCall[] {
  const calls: ApiCall[] = []
  vi.stubGlobal('fetch', async (input: RequestInfo | URL, init?: RequestInit) => {
    const request = new Request(new URL(String(input), 'http://localhost'), init)
    const path = new URL(request.url).pathname.replace(/^\/v1/, '')
    let body: unknown
    if (init?.body instanceof FormData) {
      body = Object.fromEntries(init.body.entries())
    } else {
      const raw = await request.text()
      body = raw ? JSON.parse(raw) : undefined
    }
    const call: ApiCall = { method: request.method, path, headers: request.headers, body }
    calls.push(call)
    const handler = findHandler(handlers, call)
    if (!handler && call.method === 'GET' && path === '/events') {
      return sseResponse([])
    }
    if (!handler) {
      return jsonResponse(404, { code: 404, message: `no stub for ${call.method} ${path}` })
    }
    return handler(call)
  })
  return calls
}

export function sseResponse(events: { name: string; data: unknown }[], delayMs = 20): Response {
  const encoder = new TextEncoder()
  const stream = new ReadableStream<Uint8Array>({
    start(controller) {
      setTimeout(() => {
        for (const event of events) {
          controller.enqueue(
            encoder.encode(`event: ${event.name}\ndata: ${JSON.stringify(event.data)}\n\n`),
          )
        }
      }, delayMs)
    },
  })
  return new Response(stream, { status: 200, headers: { 'Content-Type': 'text/event-stream' } })
}

function findHandler(handlers: Record<string, ApiHandler>, call: ApiCall) {
  const segments = call.path.split('/')
  for (const [key, handler] of Object.entries(handlers)) {
    const [method, pattern] = key.split(' ')
    if (method !== call.method) continue
    const expected = pattern.split('/')
    if (expected.length !== segments.length) continue
    if (expected.every((part, i) => part.startsWith(':') || part === segments[i])) return handler
  }
  return undefined
}

export function jsonResponse(status: number, body?: unknown): Response {
  return new Response(body === undefined ? null : JSON.stringify(body), {
    status,
    headers: body === undefined ? {} : { 'Content-Type': 'application/json' },
  })
}

export function errorResponse(status: number, message = 'error'): Response {
  return jsonResponse(status, { code: status, message })
}
