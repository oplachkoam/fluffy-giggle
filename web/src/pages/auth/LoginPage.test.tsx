import { fireEvent, screen, waitFor } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { getCredentials } from '@/lib/auth'
import { errorResponse, jsonResponse, renderRoute, stubApi } from '@/test/render'

describe('LoginPage', () => {
  it('stores the credentials after the server accepts them', async () => {
    const calls = stubApi({
      'GET /auth/me': (call) =>
        call.headers.get('Authorization') === 'Basic b3BlcmF0b3I6c2VjcmV0'
          ? jsonResponse(200, { login: 'operator' })
          : errorResponse(401, 'invalid login or password'),
      'GET /cases': () => jsonResponse(200, []),
    })
    renderRoute('/login')

    fireEvent.change(screen.getByLabelText('Логин'), { target: { value: 'operator' } })
    fireEvent.change(screen.getByLabelText('Пароль'), { target: { value: 'secret' } })
    fireEvent.click(screen.getByRole('button', { name: 'Войти' }))

    await waitFor(() => expect(getCredentials()).toBe('b3BlcmF0b3I6c2VjcmV0'))
    expect(await screen.findByRole('heading', { name: 'Новый запуск' })).toBeInTheDocument()
    expect(calls[0].path).toBe('/auth/me')
  })

  it('shows an error on a wrong password and keeps the form', async () => {
    stubApi({ 'GET /auth/me': () => errorResponse(401, 'invalid login or password') })
    renderRoute('/login')

    fireEvent.change(screen.getByLabelText('Логин'), { target: { value: 'operator' } })
    fireEvent.change(screen.getByLabelText('Пароль'), { target: { value: 'wrong' } })
    fireEvent.click(screen.getByRole('button', { name: 'Войти' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('Неверный логин или пароль.')
    expect(getCredentials()).toBeNull()
  })
})
