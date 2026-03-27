import { get, post } from './client'
import type { FlagSubmission, SubmitFlagRequest } from '@/types'

export const flagApi = {
  submit: (gameId: number, data: SubmitFlagRequest) =>
    post<FlagSubmission>(`/games/${gameId}/flags/submit`, data),
  history: (gameId: number) => get<FlagSubmission[]>(`/games/${gameId}/flags/history`),
}
