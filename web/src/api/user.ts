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
    const users = await get<User[]>('/admin/users', params as any)
    // 转换为前端期望的分页格式
    return { items: users, total: users.length }
  },
  get: (id: number) => get<User>(`/admin/users/${id}`),
  create: (data: CreateUserRequest) => post<User>('/admin/users', data),
  update: (id: number, data: UpdateUserRequest) => put<User>(`/admin/users/${id}`, data),
  delete: (id: number) => del<void>(`/admin/users/${id}`),
  resetPassword: (id: number) => post<void>(`/admin/users/${id}/reset-password`),
  toggleStatus: (id: number) => post<void>(`/admin/users/${id}/toggle-status`),
}

