import { create } from 'zustand'
import type { Game, RoundPhase } from '@/types'

interface GameState {
  currentGame: Game | null
  countdown: number // seconds remaining in current round
  setCurrentGame: (game: Game | null) => void
  setCountdown: (seconds: number) => void
}

export const useGameStore = create<GameState>((set) => ({
  currentGame: null,
  countdown: 0,
  setCurrentGame: (game) => set({ currentGame: game }),
  setCountdown: (countdown) => set({ countdown }),
}))
