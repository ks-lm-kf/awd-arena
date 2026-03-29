import { get, post, put, del } from './client'
import type { Game, GameStatus, RoundRanking, SecurityAlert, AttackLog } from '@/types'

export const gameApi = {
  list: () => get<Game[]>('/games'),
  get: (id: number) => get<Game>(`/games/${id}`),
  create: (data: Partial<Game>) => post<Game>('/games', data),
  update: (id: number, data: Partial<Game>) => put<Game>(`/games/${id}`),
  start: (id: number) => post<void>(`/games/${id}/start`),
  pause: (id: number) => post<void>(`/games/${id}/pause`),
  resume: (id: number) => post<void>(`/games/${id}/resume`),
  stop: (id: number) => post<void>(`/games/${id}/stop`),
  delete: (id: number) => del<void>(`/games/${id}`),
  reset: (id: number) => post<void>(`/games/${id}/reset`),
  alerts: (id: number) => get<SecurityAlert[]>(`/games/${id}/alerts`),
  attacks: (id: number) => get<AttackLog[]>(`/games/${id}/attacks`),
}
