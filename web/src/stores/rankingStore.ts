import { create } from 'zustand'
import type { RankingItem } from '@/types'

interface RankingState {
  rankings: RankingItem[]
  setRankings: (rankings: RankingItem[]) => void
  updateRanking: (teamId: number, patch: Partial<RankingItem>) => void
}

export const useRankingStore = create<RankingState>((set) => ({
  rankings: [],
  setRankings: (rankings) => set({ rankings }),
  updateRanking: (teamId, patch) =>
    set((state) => ({
      rankings: state.rankings.map((r) =>
        r.team_id === teamId ? { ...r, ...patch } : r,
      ),
    })),
}))
