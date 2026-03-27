import { useState, useEffect, useRef } from 'react'
import { Card, Row, Col, Statistic, Tag, Typography, Button, Table, Progress, Space, Tooltip, message } from 'antd'
import {
  PlayCircleOutlined, PauseCircleOutlined, StopOutlined,
  TrophyOutlined, ThunderboltOutlined, TeamOutlined,
  ClockCircleOutlined, RocketOutlined, ReloadOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import { gameApi } from '@/api/game'
import { rankingApi } from '@/api/ranking'
import { useWebSocket } from '@/hooks/useWebSocket'
import type { Game, GameStatus, RankingItem, GameMode } from '@/types'
import dayjs from 'dayjs'

const { Title } = Typography

const statusConfig: Record<GameStatus, { color: string; label: string; icon: React.ReactNode }> = {
  draft: { color: 'default', label: '未开始', icon: <ClockCircleOutlined /> },
  running: { color: 'success', label: '进行中', icon: <PlayCircleOutlined /> },
  paused: { color: 'warning', label: '已暂停', icon: <PauseCircleOutlined /> },
  finished: { color: 'error', label: '已结束', icon: <StopOutlined /> },
}

const modeLabels: Record<GameMode, string> = { awd_score: 'AWD 得分赛', awd_mix: 'AWD 混合赛', koh: 'KOH 擂台赛' }

export default function DashboardPage() {
  const user = useAuthStore((s) => s.user)
  const isAdmin = user?.role === 'admin'
  const [games, setGames] = useState<Game[]>([])
  const [activeGame, setActiveGame] = useState<Game | null>(null)
  const [rankings, setRankings] = useState<RankingItem[]>([])
  const [countdown, setCountdown] = useState<string>('')
  const timerRef = useRef<ReturnType<typeof setInterval>>(undefined)

  const { subscribe } = useWebSocket()

  const loadData = async () => {
    try {
      const g = await gameApi.list()
      setGames(g)
      const running = g.find((x) => x.status === 'running') || g.find((x) => x.status === 'paused')
      if (running) {
        setActiveGame(running)
        const r = await rankingApi.list(running.id)
        setRankings(r)
      }
    } catch { /* ignore */ }
  }

  useEffect(() => { loadData() }, [])

  useEffect(() => {
    if (!activeGame) return
    const unsub = subscribe('ranking:update', (data: any) => {
      if (Array.isArray(data)) setRankings(data)
    })
    return unsub
  }, [activeGame, subscribe])

  useEffect(() => {
    if (!activeGame || activeGame.status !== 'running') {
      setCountdown('')
      return
    }
    const tick = () => {
      if (!activeGame?.end_time) return
      const diff = dayjs(activeGame.end_time).diff(dayjs())
      if (diff <= 0) { setCountdown('已结束'); return }
      const h = Math.floor(diff / 3600000)
      const m = Math.floor((diff % 3600000) / 60000)
      const s = Math.floor((diff % 60000) / 1000)
      setCountdown(`${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`)
    }
    tick()
    timerRef.current = setInterval(tick, 1000)
    return () => { if (timerRef.current) clearInterval(timerRef.current) }
  }, [activeGame])

  const controlGame = async (id: number, action: 'start' | 'pause' | 'stop') => {
    try {
      await (gameApi as any)[action](id)
      message.success(`${action === 'start' ? '启动' : action === 'pause' ? '暂停' : '停止'}成功`)
      loadData()
    } catch (err: any) {
      message.error(err?.response?.data?.message || '操作失败')
    }
  }

  const myTeamRanking = rankings.findIndex((r) => r.team_id === user?.team_id) + 1
  const myTeamRank = rankings.find((r) => r.team_id === user?.team_id)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Title level={3} className="mb-0">{isAdmin ? '监控大屏' : '比赛概览'}</Title>
        <Button icon={<ReloadOutlined />} onClick={loadData}>刷新</Button>
      </div>

      {activeGame && (
        <Card style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
          <div className="flex items-center justify-between flex-wrap gap-4">
            <div className="flex items-center gap-4">
              <span className="text-xl font-bold text-white">{activeGame.title}</span>
              <Tag color={statusConfig[activeGame.status].color} icon={statusConfig[activeGame.status].icon}>{statusConfig[activeGame.status].label}</Tag>
              <Tag>{modeLabels[activeGame.mode]}</Tag>
            </div>
            <div className="flex items-center gap-4">
              {countdown && (
                <div className="flex items-center gap-2 text-lg font-mono text-indigo-400">
                  <ClockCircleOutlined />{countdown}
                </div>
              )}
              {isAdmin && (
                <Space>
                  {activeGame.status === 'draft' && <Button type="primary" icon={<RocketOutlined />} onClick={() => controlGame(activeGame.id, 'start')}>启动</Button>}
                  {activeGame.status === 'running' && <Button icon={<PauseCircleOutlined />} onClick={() => controlGame(activeGame.id, 'pause')}>暂停</Button>}
                  {activeGame.status === 'paused' && <Button type="primary" icon={<PlayCircleOutlined />} onClick={() => controlGame(activeGame.id, 'start')}>继续</Button>}
                  {activeGame.status !== 'finished' && <Button danger icon={<StopOutlined />} onClick={() => controlGame(activeGame.id, 'stop')}>停止</Button>}
                </Space>
              )}
            </div>
          </div>
          {activeGame.status === 'running' && (
            <div className="mt-3">
              <div className="flex items-center gap-2 text-sm text-gray-400">
                <span>第 {activeGame.current_round} / {activeGame.total_rounds} 轮</span>
                <Progress percent={Math.round((activeGame.current_round / activeGame.total_rounds) * 100)} showInfo={false} strokeColor="#6366f1" size="small" style={{ flex: 1, maxWidth: 300 }} />
              </div>
            </div>
          )}
        </Card>
      )}

      {!isAdmin && myTeamRank && (
        <Row gutter={16}>
          <Col span={6}><Card style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}><Statistic title="我的排名" value={myTeamRanking} prefix={<TrophyOutlined />} valueStyle={{ color: '#fbbf24' }} /></Card></Col>
          <Col span={6}><Card style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}><Statistic title="总得分" value={myTeamRank.total_score} valueStyle={{ color: '#6366f1' }} /></Card></Col>
          <Col span={6}><Card style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}><Statistic title="攻击得分" value={myTeamRank.attack_score} prefix={<ThunderboltOutlined />} valueStyle={{ color: '#ef4444' }} /></Card></Col>
          <Col span={6}><Card style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}><Statistic title="防御得分" value={myTeamRank.defense_score} prefix={<TeamOutlined />} valueStyle={{ color: '#10b981' }} /></Card></Col>
        </Row>
      )}

      {isAdmin && (
        <Card title="比赛列表" style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
          <Table dataSource={games} rowKey="id" pagination={false} size="small"
            columns={[
              { title: '比赛名称', dataIndex: 'title' },
              { title: '模式', dataIndex: 'mode', render: (v: GameMode) => modeLabels[v] },
              { title: '状态', dataIndex: 'status', render: (v: GameStatus) => <Tag color={statusConfig[v].color}>{statusConfig[v].label}</Tag> },
              { title: '轮次', render: (_, r) => `${r.current_round}/${r.total_rounds}` },
              { title: '创建时间', dataIndex: 'created_at', render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm') },
            ]}
          />
        </Card>
      )}

      {activeGame && rankings.length > 0 && (
        <Card title="排行榜" style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
          <Table dataSource={rankings} rowKey="team_id" pagination={false} size="small"
            rowClassName={(r) => r.team_id === user?.team_id ? 'bg-indigo-500/10' : ''}
            columns={[
              { title: '#', dataIndex: 'rank', width: 60, render: (v: number) => v === 1 ? '🥇' : v === 2 ? '🥈' : v === 3 ? '🥉' : v },
              { title: '队伍', dataIndex: 'team_name', render: (v: string, r: RankingItem) => r.team_id === user?.team_id ? <span className="text-indigo-400 font-bold">{v} (我)</span> : v },
              { title: '总得分', dataIndex: 'total_score', render: (v: number) => <span className="font-bold">{v}</span>, sorter: (a: any, b: any) => a.total_score - b.total_score },
              { title: '攻击', dataIndex: 'attack_score', render: (v: number) => <span className="text-red-400">{v}</span> },
              { title: '防御', dataIndex: 'defense_score', render: (v: number) => <span className="text-green-400">{v}</span> },
              { title: 'Flag', dataIndex: 'flag_count' },
            ]}
          />
        </Card>
      )}

      {!activeGame && (
        <Card style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
          <div className="text-center py-12 text-gray-400">
            <RocketOutlined style={{ fontSize: 48 }} />
            <p className="mt-4 text-lg">{isAdmin ? '暂无进行中的比赛，请先创建并启动' : '暂无进行中的比赛，请等待管理员开启'}</p>
          </div>
        </Card>
      )}
    </div>
  )
}
