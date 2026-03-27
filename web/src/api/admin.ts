import { get, post, put, del } from './client'
import type { Game, Team } from '@/types'

export interface AdminLog {
  id: number
  user_id: number
  username: string
  action: string
  resource_type: string
  resource_id: number
  description: string
  ip_address: string
  user_agent: string
  details: string
  created_at: string
}

export interface AdjustScoreRequest {
  game_id: number
  team_id: number
  amount: number
  reason: string
}

export interface BatchImportTeamsRequest {
  teams: Array<{
    name: string
    description?: string
    avatar_url?: string
    token?: string
  }>
}

export const adminApi = {
  // Admin logs
  getLogs: (params?: { page?: number; page_size?: number; action?: string; resource_type?: string }) =>
    get<{ items: AdminLog[]; total: number; page: number; page_size: number }>('/judge/logs', params),

  // Game management
  createGame: (data: Partial<Game>) => post<Game>('/judge/games', data),
  updateGame: (id: number, data: Partial<Game>) => put<Game>(`/judge/games/${id}`, data),
  deleteGame: (id: number) => del<void>(`/judge/games/${id}`),
  startGame: (id: number) => post<void>(`/judge/games/${id}/start`),
  pauseGame: (id: number) => post<void>(`/judge/games/${id}/pause`),
  resumeGame: (id: number) => post<void>(`/judge/games/${id}/resume`),
  stopGame: (id: number) => post<void>(`/judge/games/${id}/stop`),
  resetGame: (id: number) => post<void>(`/judge/games/${id}/reset`),

  // Game teams management
  getGameTeams: (gameId: number) => get<Team[]>(`/judge/games/${gameId}/teams`),
  addTeamToGame: (gameId: number, teamId: number) => post<void>(`/judge/games/${gameId}/teams`, { team_id: teamId }),
  removeTeamFromGame: (gameId: number, teamId: number) => del<void>(`/judge/games/${gameId}/teams/${teamId}`),

  // Team management
  createTeam: (data: { name: string; description?: string; token?: string }) => post<Team>('/judge/teams', data),
  updateTeam: (id: number, data: { name?: string; description?: string; token?: string }) => put<Team>(`/judge/teams/${id}`, data),
  deleteTeam: (id: number) => del<void>(`/judge/teams/${id}`),
  batchImportTeams: (data: BatchImportTeamsRequest) => post<{ imported: Team[]; errors: string[]; total: number; success: number }>('/judge/teams/batch-import', data),
  resetTeam: (id: number) => post<Team>(`/judge/teams/${id}/reset`),

  // Score adjustment
  adjustScore: (data: AdjustScoreRequest) => post<{ team_id: number; team_name: string; old_score: number; adjustment: number; new_score: number }>('/judge/scores/adjust', data),
}
