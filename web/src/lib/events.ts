import type { QueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'

import { apiStream } from '@/lib/api'
import { type BenchmarkDto, toBenchmark } from '@/lib/benchmarks'
import { type Run, type RunDto, type StepDto, toRun, toStep } from '@/lib/runs'

export type AppEvent =
  | { name: 'run'; data: RunDto }
  | { name: 'step'; data: StepDto }
  | { name: 'benchmark'; data: BenchmarkDto }

const reconnectDelay = 3000

/** One stream per tab; events land in the query cache and pages read the cache. */
export function useEvents(enabled: boolean, onEvent: (event: AppEvent) => void) {
  useEffect(() => {
    if (!enabled) return
    const controller = new AbortController()
    let timer: ReturnType<typeof setTimeout> | undefined

    const connect = async () => {
      try {
        const response = await apiStream('/events', controller.signal)
        await readStream(response, onEvent, controller.signal)
      } catch {
        // connection lost or refused; retried below
      }
      if (!controller.signal.aborted) {
        timer = setTimeout(connect, reconnectDelay)
      }
    }
    void connect()

    return () => {
      controller.abort()
      if (timer) clearTimeout(timer)
    }
  }, [enabled, onEvent])
}

export async function readStream(
  response: Response,
  onEvent: (event: AppEvent) => void,
  signal: AbortSignal,
) {
  const reader = response.body?.getReader()
  if (!reader) return
  const decoder = new TextDecoder()
  let buffer = ''
  while (!signal.aborted) {
    const { value, done } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    let index = buffer.indexOf('\n\n')
    while (index >= 0) {
      const chunk = buffer.slice(0, index)
      buffer = buffer.slice(index + 2)
      const event = parseChunk(chunk)
      if (event) onEvent(event)
      index = buffer.indexOf('\n\n')
    }
  }
}

export function parseChunk(chunk: string): AppEvent | null {
  let name = ''
  const data: string[] = []
  for (const line of chunk.split('\n')) {
    if (line.startsWith('event:')) name = line.slice(6).trim()
    else if (line.startsWith('data:')) data.push(line.slice(5).trim())
  }
  if (!name || data.length === 0) return null
  try {
    return { name, data: JSON.parse(data.join('\n')) } as AppEvent
  } catch {
    return null
  }
}

export function applyEvent(queryClient: QueryClient, event: AppEvent) {
  switch (event.name) {
    case 'run': {
      const run = toRun(event.data)
      queryClient.setQueryData<Run>(['run', run.id], (old) =>
        old ? { ...run, steps: old.steps } : undefined,
      )
      void queryClient.invalidateQueries({ queryKey: ['runs'] })
      if (run.benchmarkId) {
        void queryClient.invalidateQueries({ queryKey: ['benchmark', run.benchmarkId] })
      }
      break
    }
    case 'step': {
      const step = toStep(event.data)
      queryClient.setQueryData<Run>(['run', step.runId], (old) => {
        if (!old) return undefined
        const steps = old.steps.filter((s) => s.id !== step.id && s.seq !== step.seq)
        steps.push(step)
        steps.sort((a, b) => a.seq - b.seq)
        return { ...old, steps }
      })
      break
    }
    case 'benchmark': {
      const benchmark = toBenchmark(event.data)
      void queryClient.invalidateQueries({ queryKey: ['benchmark', benchmark.id] })
      void queryClient.invalidateQueries({ queryKey: ['benchmarks'] })
      break
    }
  }
}
