import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '@/auth/AuthContext'
import { ElectricField } from '@/components/ElectricField'

export function RequireAuth() {
  const { user, ready } = useAuth()
  const location = useLocation()

  if (!ready) {
    return (
      <div className="relative flex min-h-screen items-center justify-center">
        <ElectricField />
        <p className="relative z-10 text-muted">Загрузка…</p>
      </div>
    )
  }

  if (!user) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }

  return <Outlet />
}

export function RequireAdmin() {
  const { user } = useAuth()
  if (user?.role !== 'admin') {
    return <Navigate to="/" replace />
  }
  return <Outlet />
}
