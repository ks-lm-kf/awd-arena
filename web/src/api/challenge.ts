import { get, post, put, del } from './client'
import type { Challenge, Difficulty } from '@/types'

export interface ChallengeParams {
  name: string
  description?: string
  image_name: string
  image_tag?: string
  difficulty: Difficulty
  base_score: number
  exposed_ports?: string  // e.g. "80,8080"
  cpu_limit?: number
  mem_limit?: number
}

export const challengeApi = {
  // 获取比赛的题目列表
  list: (gameId: number) => get<Challenge[]>(`/games/${gameId}/challenges`),
  
  // 添加题目到比赛
  create: (gameId: number, data: ChallengeParams) => post<Challenge>(`/games/${gameId}/challenges`, data),
  
  // 更新题目
  update: (gameId: number, challengeId: number, data: Partial<ChallengeParams>) => 
    put<Challenge>(`/games/${gameId}/challenges/${challengeId}`, data),
  
  // 删除题目
  delete: (gameId: number, challengeId: number) => 
    del<void>(`/games/${gameId}/challenges/${challengeId}`),
}
