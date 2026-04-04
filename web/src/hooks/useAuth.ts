import { useCallback } from 'react'
import { useNavigate } from 'react-router'
import { useAuthStore } from '@/stores/authStore'
import { authApi } from '@/api/auth'

export function useAuth() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const token = useAuthStore((s) => s.token)
  const setToken = useAuthStore((s) => s.setToken)
  const setUser = useAuthStore((s) => s.setUser)
  const logoutStore = useAuthStore((s) => s.logout)
  const fetchUser = useAuthStore((s) => s.fetchUser)

  const login = useCallback(
    async (username: string, password: string) => {
      const res = await authApi.login({ username, password })
      setToken(res.token)
      setUser(res.user)
      if (res.user?.must_change_password || res.must_change_password) {
        navigate('/change-password')
      } else {
        navigate('/dashboard')
      }
    },
    [setToken, setUser, navigate],
  )

  const logout = useCallback(async () => {
    try {
      await authApi.logout()
    } catch {
      /* ignore */
    }
    logoutStore()
    navigate('/login')
  }, [logoutStore, navigate])

  return { user, token, isLoggedIn: !!token, login, logout, fetchUser }
}
