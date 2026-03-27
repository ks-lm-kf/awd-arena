import { get, post, put, del } from './client'
import type { Team, TeamMember } from '@/types'

export interface UpdateTeamRequest {
  name?: string
  description?: string
}

export interface AddMemberRequest {
  user_id: number
}

export const teamApi = {
  list: () => get<Team[]>('/teams'),
  get: (id: number) => get<Team>(`/teams/${id}`),
  create: (data: { name: string; description?: string }) => post<Team>('/teams', data),
  update: (id: number, data: UpdateTeamRequest) => put<Team>(`/teams/${id}`, data),
  delete: (id: number) => del<void>(`/teams/${id}`),
  members: (id: number) => get<TeamMember[]>(`/teams/${id}/members`),
  addMember: (id: number, data: AddMemberRequest) => post<void>(`/teams/${id}/members`, data),
  removeMember: (id: number, userId: number) => del<void>(`/teams/${id}/members/${userId}`),
}
