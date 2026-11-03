import { type FormEvent, useState } from 'react'
import { Navigate, useNavigate } from 'react-router'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { api, errorText } from '@/lib/api'
import { encodeCredentials, setCredentials, useCredentials } from '@/lib/auth'

export function LoginPage() {
  const credentials = useCredentials()
  const navigate = useNavigate()
  const [login, setLogin] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [pending, setPending] = useState(false)

  if (credentials !== null) {
    return <Navigate to="/" replace />
  }

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    setPending(true)
    setError(null)
    const encoded = encodeCredentials(login, password)
    try {
      await api('/auth/me', { credentials: encoded })
      setCredentials(encoded)
      void navigate('/', { replace: true })
    } catch (err) {
      setError(errorText(err, 'Не удалось войти.'))
    } finally {
      setPending(false)
    }
  }

  return (
    <main className="flex min-h-svh items-center justify-center px-4">
      <form onSubmit={submit} noValidate className="w-full max-w-sm space-y-4">
        <div>
          <h1 className="font-mono text-lg font-semibold">fluffy-giggle</h1>
          <p className="text-muted-foreground text-sm">Вход оператора</p>
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="login">Логин</Label>
          <Input
            id="login"
            autoComplete="username"
            value={login}
            onChange={(e) => setLogin(e.target.value)}
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="password">Пароль</Label>
          <Input
            id="password"
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
        {error && (
          <p role="alert" className="text-negative text-sm">
            {error}
          </p>
        )}
        <Button type="submit" className="w-full" disabled={pending || !login || !password}>
          Войти
        </Button>
      </form>
    </main>
  )
}
