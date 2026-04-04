import { useState, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router'
import { ConfigProvider, theme, Spin, message } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MainLayout } from '@/components/Layout'
import DashboardPage from '@/pages/Dashboard'
import GameManagePage from '@/pages/GameManage'
import RankingPage from '@/pages/Ranking'
import TeamManagePage from '@/pages/TeamManage'
import AttackPanelPage from '@/pages/AttackPanel'
import DefensePanelPage from '@/pages/DefensePanel'
import SettingsPage from '@/pages/Settings'
import DockerImagesPage from '@/pages/DockerImages'
import UserManagePage from '@/pages/UserManage'
import LoginPage from '@/pages/Login'
import RegisterPage from '@/pages/Register'
import ProfilePage from '@/pages/Profile'
import ChangePasswordPage from '@/pages/ChangePassword'
import ErrorBoundary from '@/components/common/ErrorBoundary'
import { useAuthStore } from '@/stores/authStore'
// Admin pages for judges
import AdminGameManagePage from '@/pages/admin/GameManage'
import AdminTeamManagePage from '@/pages/admin/TeamManage'
import ContainerManagePage from '@/pages/admin/ContainerManage'
// New pages
import AuditPage from '@/pages/Audit'
import GameDetailPage from '@/pages/GameDetail'
// @ts-ignore
import '@/styles/index.css'

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: 1, refetchOnWindowFocus: false } },
})

function ProtectedRoute({ children, requireRole }: { children: React.ReactNode; requireRole?: string[] }) {
  const token = useAuthStore((s) => s.token)
  const user = useAuthStore((s) => s.user)
  const fetchUser = useAuthStore((s) => s.fetchUser)
  const [checking, setChecking] = useState(!user)

  useEffect(() => {
    if (!token) return
    if (user) return
    let cancelled = false
    fetchUser().catch(() => {}).finally(() => { if (!cancelled) setChecking(false) })
    return () => { cancelled = true }
  }, [token, user, fetchUser])

  if (!token) return <Navigate to="/login" replace />
  if (checking) return (
    <div className="min-h-screen flex items-center justify-center bg-gray-950">
      <Spin size="large" />
    </div>
  )
  if (requireRole && user && !requireRole.includes(user.role)) {
    message.error('无权访问该页面')
    return <Navigate to="/dashboard" replace />
  }
  return <>{children}</>
}

// Admin route guard component - ensures only admins can access /admin/* routes
function AdminRoute({ children }: { children: React.ReactNode }) {
  const user = useAuthStore((s) => s.user)
  const token = useAuthStore((s) => s.token)

  if (!token) return <Navigate to="/login" replace />
  
  if (user?.role !== 'admin') {
    message.error('需要管理员权限')
    return <Navigate to="/" replace />
  }
  
  return <>{children}</>
}

function App() {
  return (
    <ConfigProvider
      locale={zhCN}
      theme={{ algorithm: theme.darkAlgorithm, token: { colorPrimary: '#6366f1' } }}
    >
      <QueryClientProvider client={queryClient}>
        <ErrorBoundary>
          <BrowserRouter>
            <Routes>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />
              <Route path="/change-password" element={<ChangePasswordPage />} />
              <Route
                element={
                  <ProtectedRoute>
                    <MainLayout />
                  </ProtectedRoute>
                }
              >
                {/* Common routes */}
                <Route path="/dashboard" element={<DashboardPage />} />
                <Route path="/ranking" element={<RankingPage />} />
                <Route path="/attack" element={<AttackPanelPage />} />
                <Route path="/profile" element={<ProfilePage />} />

                {/* Admin only */}
                <Route path="/games" element={
                  <ProtectedRoute requireRole={['admin']}><GameManagePage /></ProtectedRoute>
                } />
                <Route path="/games/:id" element={
                  <ProtectedRoute requireRole={['admin']}><GameDetailPage /></ProtectedRoute>
                } />
                <Route path="/games/:id/ranking" element={
                  <ProtectedRoute requireRole={['admin']}><RankingPage /></ProtectedRoute>
                } />
                <Route path="/games/:id/attack" element={
                  <ProtectedRoute requireRole={['admin']}><AttackPanelPage /></ProtectedRoute>
                } />
                <Route path="/games/:id/defense" element={
                  <ProtectedRoute requireRole={['admin']}><DefensePanelPage /></ProtectedRoute>
                } />
                <Route path="/teams" element={
                  <ProtectedRoute requireRole={['admin']}><TeamManagePage /></ProtectedRoute>
                } />
                <Route path="/users" element={
                  <ProtectedRoute requireRole={['admin']}><UserManagePage /></ProtectedRoute>
                } />
                <Route path="/docker-images" element={
                  <ProtectedRoute requireRole={['admin']}><DockerImagesPage /></ProtectedRoute>
                } />
                <Route path="/settings" element={
                  <ProtectedRoute requireRole={['admin']}><SettingsPage /></ProtectedRoute>
                } />

                {/* Judge/Referee routes (admin + organizer) */}
                <Route path="/admin/games" element={
                  <ProtectedRoute requireRole={['admin', 'organizer']}><AdminGameManagePage /></ProtectedRoute>
                } />
                <Route path="/admin/teams" element={
                  <ProtectedRoute requireRole={['admin', 'organizer']}><AdminTeamManagePage /></ProtectedRoute>
                } />
                <Route path="/admin/containers" element={
                  <ProtectedRoute requireRole={['admin', 'organizer']}><ContainerManagePage /></ProtectedRoute>
                } />
                
                {/* Catch-all admin route guard - ensures all /admin/* routes require admin role */}
                <Route path="/admin/*" element={
                  <AdminRoute>
                    <Navigate to="/dashboard" replace />
                  </AdminRoute>
                } />

                {/* Audit logs - admin only */}
                <Route path="/audit" element={
                  <ProtectedRoute requireRole={['admin']}><AuditPage /></ProtectedRoute>
                } />
              </Route>
              <Route path="*" element={<Navigate to="/dashboard" replace />} />
            </Routes>
          </BrowserRouter>
        </ErrorBoundary>
      </QueryClientProvider>
    </ConfigProvider>
  )
}

export default App


