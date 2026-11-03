import { fireEvent, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'

import { setCredentials } from '@/lib/auth'
import { runDto } from '@/test/fixtures'
import { jsonResponse, renderRoute, stubApi } from '@/test/render'

describe('RunsPage', () => {
  beforeEach(() => {
    setCredentials('dGVzdDp0ZXN0')
  })

  it('lists runs and loads the next page by cursor', async () => {
    const first = runDto({ id: 'a', case_id: 'case-01' })
    const second = runDto({ id: 'b', case_id: 'case-02', status: 'failed', judge_score: null })
    const calls = stubApi({
      'GET /runs': (call) =>
        call.headers.get('x-ignored') === null &&
        calls.filter((c) => c.path === '/runs').length <= 1
          ? jsonResponse(200, { runs: [first], next: 'a' })
          : jsonResponse(200, { runs: [second], next: '' }),
    })
    renderRoute('/runs')

    expect(await screen.findByText('case-01', { exact: false })).toBeInTheDocument()
    expect(screen.getByText('4 / 5')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Показать ещё' }))

    expect(await screen.findByText('case-02', { exact: false })).toBeInTheDocument()
    expect(screen.getByText('ошибка')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Показать ещё' })).not.toBeInTheDocument()
  })

  it('says when there is nothing yet', async () => {
    stubApi({ 'GET /runs': () => jsonResponse(200, { runs: [], next: '' }) })
    renderRoute('/runs')
    expect(await screen.findByText(/Запусков пока нет/)).toBeInTheDocument()
  })
})
