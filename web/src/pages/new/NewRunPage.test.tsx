import { fireEvent, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'

import { setCredentials } from '@/lib/auth'
import { runDto } from '@/test/fixtures'
import { jsonResponse, renderRoute, stubApi } from '@/test/render'

describe('NewRunPage', () => {
  beforeEach(() => {
    setCredentials('dGVzdDp0ZXN0')
  })

  it('redirects to login without credentials', async () => {
    localStorage.clear()
    stubApi({})
    renderRoute('/')
    expect(await screen.findByRole('button', { name: 'Войти' })).toBeInTheDocument()
  })

  it('sends a fixture case with the chosen mode and opens the run', async () => {
    const run = runDto({ status: 'queued' })
    const calls = stubApi({
      'GET /cases': () =>
        jsonResponse(200, [
          { id: 'case-01', title: 'Задержка выплаты', has_image: false, canaries: 16 },
        ]),
      'POST /runs': () => jsonResponse(201, run),
      'GET /runs/:id': () => jsonResponse(200, run),
    })
    renderRoute('/')

    fireEvent.click(screen.getByRole('tab', { name: 'Кейс из набора' }))
    const select = await screen.findByLabelText('Кейс')
    await screen.findByRole('option', { name: /Задержка выплаты/ })
    fireEvent.change(select, { target: { value: 'case-01' } })
    fireEvent.change(screen.getByLabelText('Adviser'), { target: { value: 'placebo' } })
    fireEvent.click(screen.getByRole('button', { name: 'Запустить' }))

    expect(await screen.findByRole('region', { name: 'Итог' })).toBeInTheDocument()
    const post = calls.find((c) => c.method === 'POST')
    expect(post?.body).toEqual({ advisor_mode: 'placebo', case_id: 'case-01' })
  })

  it('keeps the button disabled until there is text', () => {
    stubApi({ 'GET /cases': () => jsonResponse(200, []) })
    renderRoute('/')

    const button = screen.getByRole('button', { name: 'Запустить' })
    expect(button).toBeDisabled()
    fireEvent.change(screen.getByLabelText('Текст обращения'), { target: { value: 'Претензия' } })
    expect(button).toBeEnabled()
  })
})
