import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider } from '@/auth/AuthContext'
import { RequireAdmin, RequireAuth } from '@/components/RequireAuth'
import { AdminRoadmapPage } from '@/pages/AdminRoadmapPage'
import { AdminUsersPage } from '@/pages/AdminUsersPage'
import { HomePage } from '@/pages/HomePage'
import { LoginPage } from '@/pages/LoginPage'
import { ModulePage } from '@/pages/ModulePage'
import { SubmodulePage } from '@/pages/SubmodulePage'

export function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route element={<RequireAuth />}>
            <Route path="/" element={<HomePage />} />
            <Route path="/modules/:moduleSlug" element={<ModulePage />} />
            <Route
              path="/modules/:moduleSlug/submodules/:submoduleSlug"
              element={<SubmodulePage />}
            />
            <Route element={<RequireAdmin />}>
              <Route path="/admin/roadmap" element={<AdminRoadmapPage />} />
              <Route path="/admin/users" element={<AdminUsersPage />} />
            </Route>
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
