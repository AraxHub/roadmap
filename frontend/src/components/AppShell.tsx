import { Link } from 'react-router-dom'
import { useAuth } from '@/auth/AuthContext'

export function AppShell({
  children,
  title,
}: {
  children: React.ReactNode
  title?: string
}) {
  const { user, logout } = useAuth()

  return (
    <div className="relative min-h-screen">
      <header className="relative z-10 border-b border-line/80 bg-surface/80 backdrop-blur-md">
        <div className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-5 py-4">
          <div className="flex items-center gap-6">
            <Link to="/" className="brand-link font-display text-xl font-extrabold tracking-tight">
              <span className="brand-glow">Roadmap</span>
            </Link>
            {title ? <span className="hidden text-sm text-muted sm:inline">{title}</span> : null}
          </div>
          <div className="flex items-center gap-3 text-sm">
            {user?.role === 'admin' ? (
              <>
                <Link to="/admin/roadmap" className="nav-chip">
                  Роадмап
                </Link>
                <Link to="/admin/users" className="nav-chip">
                  ЛК
                </Link>
              </>
            ) : null}
            <span className="text-muted">{user?.login}</span>
            <button type="button" onClick={() => void logout()} className="btn btn-ink btn-sm">
              Выйти
            </button>
          </div>
        </div>
      </header>
      <main className="relative z-10 mx-auto max-w-6xl px-5 py-8">{children}</main>
    </div>
  )
}
