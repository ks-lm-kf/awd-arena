import { get, put } from './client'

export interface SystemSettings {
  site_name?: string
  announcement?: string
  flag_format?: string
  initial_score?: number
  attack_weight?: number
  defense_weight?: number
  max_team_size?: number
  round_duration?: number
  break_duration?: number
}

export const settingsApi = {
  get: () => get<SystemSettings>('/settings'),
  update: (data: Partial<SystemSettings>) => put<SystemSettings>('/settings', data),
}
