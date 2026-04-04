// ==================== 通用响应 ====================

export interface APIResponse<T> {
  code: number
  message: string
  data: T
}

export interface PagedResponse<T> {
  code: number
  message: string
  data: T[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export interface APIError {
  code: number
  message: string
  details?: unknown
}

// ==================== 用户 ====================

export interface User {
  id: number
  username: string
  email: string
  role: 'admin' | 'organizer' | 'player'
  team_id: number | null
  team_name?: string
  must_change_password?: boolean
  created_at: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  must_change_password?: boolean
  user: User
}

// ==================== 队伍 ====================

export interface Team {
  id: number
  name: string
  score?: number
  token: string
  description: string
  avatar_url: string
  member_count?: number
  created_at: string
}

export interface TeamMember extends User {}

export interface CreateTeamRequest {
  name: string
  description?: string
}

// ==================== 竞赛 ====================

export type GameStatus = 'draft' | 'running' | 'paused' | 'finished'
export type GameMode = 'awd_score' | 'awd_mix' | 'koh'
export type RoundPhase = 'preparation' | 'running' | 'scoring' | 'break'

export interface Game {
  id: number
  title: string
  description: string
  mode: GameMode
  status: GameStatus
  round_duration: number
  break_duration: number
  total_rounds: number
  current_round: number
  current_phase: RoundPhase
  flag_format: string
  attack_weight: number
  defense_weight: number
  start_time: string | null
  end_time: string | null
  created_by: number
  created_at: string
  updated_at: string
}

export interface GameTeam {
  game_id: number
  team_id: number
  score: number
  rank: number
}

// ==================== 靶机/容器 ====================

export type Difficulty = 'easy' | 'medium' | 'hard'
export type ContainerStatus = 'creating' | 'running' | 'stopped' | 'error'

export interface Challenge {
  id: number
  game_id: number
  name: string
  description: string
  image_name: string
  image_tag: string
  difficulty: Difficulty
  base_score: number
  exposed_ports: string
  cpu_limit: number
  mem_limit: number
  created_at: string
}

export interface TeamContainer {
  id: number
  game_id: number
  team_id: number
  team_name?: string
  challenge_id: number
  challenge_name?: string
  container_id: string
  ip_address: string
  port_mapping: Record<string, number>
  status: ContainerStatus
  created_at: string
}

export interface ContainerStats {
  container_id: string
  team_id: number
  team_name: string
  cpu_percent: number
  memory_mb: number
  memory_limit_mb: number
  network_rx: number
  network_tx: number
}

// ==================== Flag ====================

export interface FlagSubmission {
  id: number
  game_id: number
  round: number
  attacker_team: number
  target_team: number
  flag_value: string
  is_correct: boolean
  points_earned: number
  submitted_at: string
}

export interface SubmitFlagRequest {
  flag: string
  target_team_id: number
}

// ==================== 排名 ====================

export interface RankingItem {
  rank: number
  team_id: number
  team_name: string
  total_score: number
  attack_score: number
  defense_score: number
  flag_count: number
  score_change?: number
}

export interface RoundRanking {
  round: number
  rankings: RankingItem[]
}

// ==================== 安全事件 ====================

export type AlertLevel = 'info' | 'warning' | 'critical'
export type AttackType = 'sql_injection' | 'xss' | 'rce' | 'ssrf' | 'deser' | 'command_injection' | 'file_inclusion' | 'other'

export interface SecurityAlert {
  id: number
  game_id: number
  level: AlertLevel
  team_id: number
  team_name?: string
  type: string
  detail: string
  created_at: string
}

export interface AttackLog {
  timestamp: string
  game_id: number
  round: number
  attacker_team: string
  target_team: string
  target_ip: string
  target_port: number
  protocol: string
  method?: string
  path?: string
  attack_type: AttackType
  severity: string
  source_ip: string
}

// ==================== AI 分析 ====================

export interface AIReport {
  team_id: number
  attack_patterns: AttackPattern[]
  vulnerabilities: Vulnerability[]
  hardening_tips: HardeningTip[]
  risk_score: number
}

export interface AttackPattern {
  type: AttackType
  count: number
  severity: string
}

export interface Vulnerability {
  name: string
  severity: string
  description: string
  fix_suggestion: string
}

export interface HardeningTip {
  priority: number
  title: string
  description: string
}

// ==================== WebSocket 事件 ====================

export type WSEventType =
  | 'round:start'
  | 'round:end'
  | 'ranking:update'
  | 'flag:captured'
  | 'alert:new'
  | 'container:status'
  | 'game:status'

export interface WSRoundStart {
  round: number
  ends_at: string
  phase: RoundPhase
}

export interface WSRoundEnd {
  round: number
  rankings: RankingItem[]
}

export interface WSFlagCaptured {
  attacker: string
  target: string
  points: number
  round: number
}

export interface WSAlertNew {
  level: AlertLevel
  team: string
  message: string
}

export interface WSContainerStatus {
  team_id: number
  status: ContainerStatus
  cpu: number
}

export interface WSGameStatus {
  status: GameStatus
  reason: string
}

export interface WSEvent<T = unknown> {
  type: WSEventType
  data: T
}
