import { get, post, put, del } from './client'
import type { User } from '@/types'

export interface CreateUserRequest {
  username: string
  password: string
  email?: string
  role?: 'admin' | 'organizer' | 'player'
  team_id?: number
}

export interface UpdateUserRequest {
  email?: string
  role?: 'admin' | 'organizer' | 'player'
  team_id?: number
}

export const userApi = {
  list: async (params?: { page?: number; page_size?: number; role?: string; search?: string }) => {
    const res = await get<any>('/admin/users', params as any)
    // Backend returns { users: User[], pagination: { total, page, page_size, total_pages } }
    if (Array.isArray(res)) {
      return { items: res, total: res.length }
    }
    const users: User[] = res.users || res.items || []
    const total: number = res.pagination?.total ?? res.total ?? users.length
    return { items: users, total }
  },
  get: (id: number) => get<User>(`/admin/users/${id}`),
  create: (data: CreateUserRequest) => post<User>('/admin/users', data),
  update: (id: number, data: UpdateUserRequest) => put<User>(`/admin/users/${id}`, data),
  delete: (id: number) => del<void>(`/admin/users/${id}`),
}

