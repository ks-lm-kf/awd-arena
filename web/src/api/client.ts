import axios from 'axios'
import type { APIResponse, APIError } from '@/types'

const client = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
})

// JWT token
client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 401 redirect
client.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
      return new Promise(() => {})
    }
    if (err.response?.status === 403) {
      const msg = err.response?.data?.message || ''
      if (msg.includes('password') || msg.includes('密码')) {
        window.location.href = '/change-password'
        return new Promise(() => {})
      }
    }
    return Promise.reject(err)
  },
)

// typed helpers
export async function get<T>(url: string, params?: Record<string, unknown>): Promise<T> {
  const res = await client.get<APIResponse<T>>(url, { params })
  return res.data.data
}

export async function post<T>(url: string, data?: unknown): Promise<T> {
  const res = await client.post<APIResponse<T>>(url, data)
  return res.data.data
}

export async function put<T>(url: string, data?: unknown): Promise<T> {
  const res = await client.put<APIResponse<T>>(url, data)
  return res.data.data
}

export async function del<T>(url: string): Promise<T> {
  const res = await client.delete<APIResponse<T>>(url)
  return res.data.data
}

export default client
