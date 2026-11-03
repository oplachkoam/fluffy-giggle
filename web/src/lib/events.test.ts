import { QueryClient } from '@tanstack/react-query'
import { describe, expect, it } from 'vitest'

import { applyEvent, parseChunk } from '@/lib/events'
import { type Run, toRun } from '@/lib/runs'
import { runDto, stepDto } from '@/test/fixtures'

describe('events', () => {
  it('parses an SSE chunk and ignores comments', () => {
    expect(parseChunk(': keep-alive')).toBeNull()
    expect(parseChunk('event: run\ndata: {"id":"x"}')).toEqual({ name: 'run', data: { id: 'x' } })
    expect(parseChunk('event: run\ndata: not json')).toBeNull()
  })

  it('merges steps into the cached run in order and keeps steps on run updates', () => {
    const client = new QueryClient()
    const run = toRun(runDto({ status: 'running', steps: [] }))
    client.setQueryData(['run', run.id], run)

    applyEvent(client, { name: 'step', data: stepDto({ id: 's2', seq: 2, kind: 'gate2' }) })
    applyEvent(client, { name: 'step', data: stepDto({ id: 's1', seq: 1 }) })
    applyEvent(client, { name: 'run', data: runDto({ status: 'done' }) })

    const cached = client.getQueryData<Run>(['run', run.id])
    expect(cached?.status).toBe('done')
    expect(cached?.steps.map((s) => s.seq)).toEqual([1, 2])
  })

  it('does not create a cache entry for a run the page never loaded', () => {
    const client = new QueryClient()
    applyEvent(client, { name: 'step', data: stepDto({ run_id: 'unknown' }) })
    expect(client.getQueryData(['run', 'unknown'])).toBeUndefined()
  })
})
