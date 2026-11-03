import { fireEvent, screen, within } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'

import { setCredentials } from '@/lib/auth'
import { runDto } from '@/test/fixtures'
import { jsonResponse, renderRoute, stubApi } from '@/test/render'

const benchmark = {
  id: 'bm1',
  created_at: '2026-11-03T12:00:00Z',
  status: 'done',
  total: 60,
  finished: 60,
}

describe('BenchmarkPage', () => {
  beforeEach(() => {
    setCredentials('dGVzdDp0ZXN0')
  })

  it('shows the latest benchmark summary per mode', async () => {
    stubApi({
      'GET /benchmarks': () => jsonResponse(200, [benchmark]),
      'GET /benchmarks/:id': () =>
        jsonResponse(200, {
          ...benchmark,
          summary: [
            {
              mode: 'off',
              runs: 20,
              done: 20,
              failed: 0,
              escalated: 1,
              blocked: 0,
              scored: 20,
              avg_score: 3.1,
              leaked: 0,
              gate2_avg: 1.5,
            },
            {
              mode: 'on',
              runs: 20,
              done: 20,
              failed: 0,
              escalated: 1,
              blocked: 2,
              scored: 18,
              avg_score: 4.25,
              leaked: 0,
              gate2_avg: 1.5,
            },
            {
              mode: 'placebo',
              runs: 20,
              done: 19,
              failed: 1,
              escalated: 1,
              blocked: 0,
              scored: 19,
              avg_score: null,
              leaked: 0,
              gate2_avg: 1.5,
            },
          ],
          runs: [runDto({ id: 'r1', benchmark_id: 'bm1', advisor_mode: 'on' })],
        }),
    })
    renderRoute('/benchmark')

    const summary = await screen.findByRole('table', { name: 'Сводка по режимам' })
    const rows = within(summary).getAllByRole('row').slice(1)
    expect(rows).toHaveLength(3)
    expect(within(rows[0]).getByText('без adviser')).toBeInTheDocument()
    expect(within(rows[0]).getByText('3.10')).toBeInTheDocument()
    expect(within(rows[1]).getByText('4.25')).toBeInTheDocument()
    expect(within(rows[2]).getByText('—')).toBeInTheDocument()
    expect(screen.getByText('Выполнено 60 из 60')).toBeInTheDocument()

    const runs = screen.getByRole('table', { name: 'Запуски бенчмарка' })
    expect(within(runs).getByRole('link', { name: 'case-01' })).toHaveAttribute('href', '/runs/r1')
  })

  it('starts a benchmark and blocks the button while one runs', async () => {
    const calls = stubApi({
      'GET /benchmarks': () =>
        calls.some((c) => c.method === 'POST')
          ? jsonResponse(200, [{ ...benchmark, status: 'running', finished: 3 }])
          : jsonResponse(200, []),
      'POST /benchmarks': () => jsonResponse(201, { ...benchmark, status: 'running', finished: 0 }),
      'GET /benchmarks/:id': () =>
        jsonResponse(200, { ...benchmark, status: 'running', finished: 3, summary: [], runs: [] }),
    })
    renderRoute('/benchmark')

    expect(await screen.findByText('Бенчмарков ещё не было.')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Прогнать набор' }))

    expect(await screen.findByText('Выполнено 3 из 60')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Прогнать набор' })).toBeDisabled()
  })
})
