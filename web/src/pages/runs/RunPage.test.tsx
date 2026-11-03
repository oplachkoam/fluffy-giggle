import { fireEvent, screen, within } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'

import { setCredentials } from '@/lib/auth'
import { runDto, stepDto } from '@/test/fixtures'
import { jsonResponse, renderRoute, sseResponse, stubApi } from '@/test/render'

const run = runDto({
  steps: [
    stepDto(),
    stepDto({
      id: 'b',
      seq: 2,
      kind: 'egress',
      status: 'blocked',
      input: { target: 'advise', payload: 'Черновик для Иванов Пётр Сергеевич' },
      output: { findings: [{ type: 'PERSON', value: 'Иванов Пётр Сергеевич' }] },
    }),
    stepDto({
      id: 'c',
      seq: 3,
      kind: 'judge',
      side: 'cloud',
      model: 'strong',
      prompt_tokens: 100,
      completion_tokens: 20,
      input: { prompt: { system: 'судья', user: 'Ответ клиенту: [PERSON_1]' } },
      output: { score: 4, reason: 'Полный ответ.' },
    }),
  ],
})

describe('RunPage', () => {
  beforeEach(() => {
    setCredentials('dGVzdDp0ZXN0')
  })

  it('shows the summary, every step with its side, and the final answer', async () => {
    stubApi({ 'GET /runs/:id': () => jsonResponse(200, run) })
    renderRoute(`/runs/${run.id}`)

    const summary = await screen.findByRole('region', { name: 'Итог' })
    expect(within(summary).getByText('готово')).toBeInTheDocument()
    expect(within(summary).getByText('совет применён')).toBeInTheDocument()
    expect(within(summary).getByText('4 / 5')).toBeInTheDocument()
    expect(within(summary).getByText('0 из 16')).toBeInTheDocument()

    const trace = screen.getByRole('region', { name: 'Трасса' })
    const items = Array.from(trace.querySelectorAll<HTMLElement>(':scope > ol > li'))
    expect(items).toHaveLength(3)
    expect(within(items[0]).getByText('Гейт 1: детекторы')).toBeInTheDocument()
    expect(within(items[0]).getByText('код')).toBeInTheDocument()
    expect(within(items[1]).getByText('заблокировано')).toBeInTheDocument()
    expect(within(items[1]).getByText(/найдено в payload/)).toBeInTheDocument()
    expect(within(items[2]).getByText('облако')).toBeInTheDocument()
    expect(within(items[2]).getAllByText(/strong/).length).toBeGreaterThan(0)

    const result = screen.getByRole('region', { name: 'Результат' })
    expect(within(result).getByText(/Уважаемый Иванов Пётр Сергеевич/)).toBeInTheDocument()
    fireEvent.click(within(result).getByRole('button', { name: 'Показать обезличенный' }))
    expect(within(result).getByText('[PERSON_1]')).toBeInTheDocument()
  })

  it('expands a step to the exact payload', async () => {
    stubApi({ 'GET /runs/:id': () => jsonResponse(200, run) })
    renderRoute(`/runs/${run.id}`)

    const trace = await screen.findByRole('region', { name: 'Трасса' })
    fireEvent.click(within(trace).getByRole('button', { name: /Оценка судьи/ }))
    expect(within(trace).getByText('Что ушло в облако')).toBeInTheDocument()
    expect(within(trace).getByText('[PERSON_1]')).toBeInTheDocument()
  })

  it('offers the review decisions and sends the chosen one', async () => {
    const parked = runDto({
      status: 'needs_review',
      advisor_outcome: 'pending',
      judge_score: null,
      review_findings: [{ text: 'рублей', type: 'OTHER', canonical: 'рублей' }],
    })
    const calls = stubApi({
      'GET /runs/:id': () => jsonResponse(200, parked),
      'POST /runs/:id/review': () => jsonResponse(200, { ...parked, status: 'running' }),
    })
    renderRoute(`/runs/${parked.id}`)

    const panel = await screen.findByRole('region', { name: 'Нужна проверка' })
    expect(within(panel).getByText('рублей')).toBeInTheDocument()
    fireEvent.click(within(panel).getByRole('button', { name: 'Это не ПДн, продолжить' }))

    await screen.findByRole('region', { name: 'Итог' })
    const post = calls.find((c) => c.method === 'POST')
    expect(post?.path).toBe(`/runs/${parked.id}/review`)
    expect(post?.body).toEqual({ decision: 'continue' })
  })

  it('appends steps that arrive over the event stream', async () => {
    const queued = runDto({
      status: 'running',
      advisor_outcome: 'pending',
      judge_score: null,
      steps: [],
    })
    stubApi({
      'GET /runs/:id': () => jsonResponse(200, queued),
      'GET /events': () =>
        sseResponse([
          { name: 'step', data: stepDto({ run_id: queued.id, kind: 'vision', side: 'local' }) },
          { name: 'run', data: { ...queued, status: 'done' } },
        ]),
    })
    renderRoute(`/runs/${queued.id}`)

    expect(await screen.findByText('Распознавание скана')).toBeInTheDocument()
    expect(await screen.findByText('готово')).toBeInTheDocument()
  })
})
