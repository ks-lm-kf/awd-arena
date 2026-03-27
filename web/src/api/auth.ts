import { get, post, put } from './client'
import type { LoginRequest, LoginResponse, User } from '@/types'

export const authApi = {
  login: (data: LoginRequest) => post<LoginResponse>('/auth/login', data),
  logout: () => post<void>('/auth/logout'),
  me: () => get<User>('/auth/me'),
  register: (data: { username: string; password: string; team_token?: string }) =>
    post<LoginResponse>('/auth/register', data),
  changePassword: (data: { old_password: string; new_password: string }) =>
    put<void>('/auth/change-password', data),
}
