import { create } from 'zustand'
import { get } from '@/api/client'
import type { User } from '@/types'

interface AuthState {
  token: string | null
  user: User | null
  loading: boolean
  setToken: (token: string | null) => void
  setUser: (user: User | null) => void
  logout: () => void
  fetchUser: () => Promise<User>
}

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem('token'),
  user: (() => {
    try {
      const raw = localStorage.getItem('user')
      return raw ? JSON.parse(raw) : null
    } catch {
      return null
    }
  })(),
  loading: false,

  setToken: (token) => {
    if (token) localStorage.setItem('token', token)
    else localStorage.removeItem('token')
    set({ token })
  },

  setUser: (user) => {
    if (user) localStorage.setItem('user', JSON.stringify(user))
    else localStorage.removeItem('user')
    set({ user })
  },

  logout: () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    set({ token: null, user: null })
  },

  fetchUser: async () => {
    set({ loading: true })
    try {
      const user = await get<User>('/auth/me')
      localStorage.setItem('user', JSON.stringify(user))
      set({ user, loading: false })
      return user
    } catch {
      set({ loading: false, token: null, user: null })
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      throw new Error('Token invalid')
    }
  },
}))
