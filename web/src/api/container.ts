import { get, post } from './client'
import type { TeamContainer } from '@/types'

export type Container = TeamContainer

export interface ContainerListResponse {
  items: Container[]
  total: number
}

export const containerApi = {
  // List containers for a specific game (admin view all)
  list: (gameId: number) =>
    get<TeamContainer[]>(`/games/${gameId}/containers`),

  // Get containers for current user's team
  getMyContainers: (gameId: number) =>
    get<TeamContainer[]>(`/games/${gameId}/my-machines`),

  // Get my machines for attack panel (player view)
  getMyMachines: (gameId: number) =>
    get<TeamContainer[]>(`/games/${gameId}/my-machines`),

  // Get container detail
  getContainerDetail: (containerId: number) =>
    get<TeamContainer>(`/containers/${containerId}`),

  // Restart a single container
  restartOne: (gameId: number, containerId: number) =>
    post<void>(`/games/${gameId}/containers/${containerId}/restart`),

  // Alias for restartOne
  restartContainer: (gameId: number, containerId: number) =>
    post<void>(`/games/${gameId}/containers/${containerId}/restart`),

  // Restart all containers
  restartAll: (gameId: number) =>
    post<void>(`/games/${gameId}/containers/restart`),

  // Alias for restartAll
  restartAllContainers: (gameId: number) =>
    post<void>(`/games/${gameId}/containers/restart`),

  // Get container stats
  stats: (gameId: number) =>
    get<any>(`/games/${gameId}/containers/stats`),

  // Alias for stats
  getContainerStats: (gameId: number) =>
    get<any>(`/games/${gameId}/containers/stats`),
}

// Legacy exports for backward compatibility
export function getMyContainers(gameId: number) {
  return containerApi.getMyContainers(gameId)
}

export function getContainerDetail(containerId: number) {
  return containerApi.getContainerDetail(containerId)
}

