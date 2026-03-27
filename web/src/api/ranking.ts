import { get } from './client'
import type { RankingItem } from '@/types'

export const rankingApi = {
  list: (gameId: number) => get<RankingItem[]>(`/games/${gameId}/rankings`),
  round: (gameId: number, round: number) => get<RankingItem[]>(`/games/${gameId}/rankings/rounds/${round}`),
}
