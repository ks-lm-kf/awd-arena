import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import duration from 'dayjs/plugin/duration'

dayjs.extend(relativeTime)
dayjs.extend(duration)

export function formatTime(time: string | null | undefined): string {
  if (!time) return '-'
  return dayjs(time).format('MM-DD HH:mm:ss')
}

export function formatRelativeTime(time: string): string {
  return dayjs(time).fromNow()
}

export function formatCountdown(seconds: number): string {
  const d = dayjs.duration(seconds, 'seconds')
  return `${String(d.minutes()).padStart(2, '0')}:${String(d.seconds()).padStart(2, '0')}`
}

export function formatScore(score: number): string {
  return score.toFixed(1)
}

export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

export function formatPercent(value: number): string {
  return `${value.toFixed(1)}%`
}

export function statusColor(status: string): string {
  const map: Record<string, string> = {
    running: '#22c55e',
    draft: '#94a3b8',
    paused: '#eab308',
    finished: '#6366f1',
    error: '#ef4444',
    creating: '#3b82f6',
    stopped: '#64748b',
  }
  return map[status] || '#94a3b8'
}

export function statusLabel(status: string): string {
  const map: Record<string, string> = {
    running: '进行中',
    draft: '草稿',
    paused: '已暂停',
    finished: '已结束',
    error: '异常',
    creating: '创建中',
    stopped: '已停止',
  }
  return map[status] || status
}

export function difficultyColor(diff: string): string {
  const map: Record<string, string> = { easy: '#22c55e', medium: '#eab308', hard: '#ef4444' }
  return map[diff] || '#94a3b8'
}

export function difficultyLabel(diff: string): string {
  const map: Record<string, string> = { easy: '简单', medium: '中等', hard: '困难' }
  return map[diff] || diff
}
