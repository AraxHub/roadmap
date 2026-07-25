import { FormEvent, useState } from 'react'
import { Navigate, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '@/auth/AuthContext'
import { ElectricField } from '@/components/ElectricField'

export function LoginPage() {
  const { user, ready, login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const from = (location.state as { from?: string } | null)?.from || '/'

  const [loginName, setLoginName] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [pending, setPending] = useState(false)

  if (ready && user) {
    return <Navigate to={from} replace />
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setPending(true)
    try {
      await login(loginName.trim(), password)
      navigate(from, { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка входа')
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="relative flex min-h-screen items-center justify-center px-5 py-10">
      <ElectricField />
      <form
        onSubmit={(e) => void onSubmit(e)}
        className="relative z-10 w-full max-w-md rounded-3xl border border-electric/25 bg-surface/90 p-8 shadow-[0_24px_80px_rgba(0,119,182,0.12)] backdrop-blur"
      >
        <p className="text-sm uppercase tracking-[0.2em] text-electric-deep">вход</p>
        <h1 className="mt-2 font-display text-4xl font-extrabold leading-none">
          <span className="brand-glow">Roadmap</span>
        </h1>
        <p className="mt-3 text-muted">Логин и пароль личного кабинета.</p>

        <label className="mt-8 block text-sm text-muted">
          Логин
          <input
            className="mt-1.5 w-full rounded-xl border border-line bg-bg px-3 py-2.5 outline-none transition focus:border-electric"
            value={loginName}
            onChange={(e) => setLoginName(e.target.value)}
            autoComplete="username"
            required
          />
        </label>
        <label className="mt-4 block text-sm text-muted">
          Пароль
          <input
            type="password"
            className="mt-1.5 w-full rounded-xl border border-line bg-bg px-3 py-2.5 outline-none transition focus:border-electric"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="current-password"
            required
          />
        </label>

        {error ? <p className="mt-4 text-sm text-danger">{error}</p> : null}

        <button
          type="submit"
          disabled={pending}
          className="btn btn-primary btn-lg mt-6 w-full"
        >
          {pending ? 'Входим…' : 'Войти'}
        </button>
      </form>
    </div>
  )
}
