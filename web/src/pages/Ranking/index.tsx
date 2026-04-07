import { useState, useEffect } from 'react'
import { Card, Select, Table, Tag, Typography, Space, message } from 'antd'
import { TrophyOutlined } from '@ant-design/icons'
import { gameApi } from '@/api/game'
import { rankingApi } from '@/api/ranking'
import { useWebSocket } from '@/hooks/useWebSocket'
import type { Game, RankingItem } from '@/types'

const { Title, Text } = Typography

export default function RankingPage() {
  const [games, setGames] = useState<Game[]>([])
  const [selectedGame, setSelectedGame] = useState<number | null>(null)
  const [rankings, setRankings] = useState<RankingItem[]>([])
  const [rounds, setRounds] = useState<number[]>([])
  const [selectedRound, setSelectedRound] = useState<number | null>(null)
  const { subscribe } = useWebSocket()

  useEffect(() => {
    gameApi.list().then((g) => {
      setGames(g)
      const running = g.find((x) => x.status === 'running') || g.find((x) => x.status === 'paused')
      if (running) {
        setSelectedGame(running.id)
        setRounds(Array.from({ length: running.current_round }, (_, i) => i + 1))
      }
    }).catch((err) => {
      console.error('Failed to load data:', err);
      message.error('加载数据失败');
    })
  }, [])

  useEffect(() => {
    if (!selectedGame) return
    ;(selectedRound ? rankingApi.round(selectedGame, selectedRound) : rankingApi.list(selectedGame))
      .then(setRankings).catch((err) => {
        console.error('Failed to load data:', err);
        message.error('加载数据失败');
      })
  }, [selectedGame, selectedRound])

  useEffect(() => {
    if (!selectedGame) return
    const unsub = subscribe('ranking:update', (data: any) => { if (Array.isArray(data)) setRankings(data) })
    return unsub
  }, [selectedGame, subscribe])

  const activeGame = games.find((g) => g.id === selectedGame)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Title level={3}>排行榜</Title>
        <Space>
          <Text>比赛：</Text>
          <Select style={{ width: 250 }} placeholder="选择比赛" value={selectedGame}
            onChange={(id) => { setSelectedGame(id); setSelectedRound(null) }}
            options={games.map((g) => ({ value: g.id, label: g.title }))} />
          {rounds.length > 0 && (
            <>
              <Text>轮次：</Text>
              <Select style={{ width: 120 }} placeholder="最新" value={selectedRound} onChange={setSelectedRound} allowClear
                options={[...rounds.map((r) => ({ value: r, label: `第 ${r} 轮` }))]} />
            </>
          )}
        </Space>
      </div>
      {activeGame && (
        <Card style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
          <div className="flex items-center gap-3 text-sm text-gray-400">
            <span>{activeGame.title}</span>
            <Tag color={activeGame.status === 'running' ? 'success' : 'default'}>{activeGame.status === 'running' ? '实时更新中' : activeGame.status}</Tag>
            <span>第 {activeGame.current_round}/{activeGame.total_rounds} 轮</span>
          </div>
        </Card>
      )}
      {selectedGame && (
        <Card style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
          <Table dataSource={rankings} rowKey="team_id" pagination={false}
            columns={[
              { title: '排名', dataIndex: 'rank', width: 80, render: (v: number) => v === 1 ? '🥇' : v === 2 ? '🥈' : v === 3 ? '🥉' : <Text type="secondary">#{v}</Text> },
              { title: '队伍', dataIndex: 'team_name', render: (v: string) => <span className="font-semibold">{v}</span> },
              { title: '总得分', dataIndex: 'total_score', render: (v: number) => <span className="text-xl font-bold text-indigo-400">{v}</span>, sorter: (a: any, b: any) => a.total_score - b.total_score, defaultSortOrder: 'ascend' as const },
              { title: '攻击', dataIndex: 'attack_score', render: (v: number) => <Tag color="red">{v}</Tag>, sorter: (a: any, b: any) => a.attack_score - b.attack_score },
              { title: '防御', dataIndex: 'defense_score', render: (v: number) => <Tag color="green">{v}</Tag>, sorter: (a: any, b: any) => a.defense_score - b.defense_score },
              { title: '首杀', dataIndex: 'first_bloods', render: (v: number) => v, sorter: (a: any, b: any) => a.first_bloods - b.first_bloods },
            ]}
          />
        </Card>
      )}
      {!selectedGame && (
        <Card style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
          <div className="text-center py-12 text-gray-400"><TrophyOutlined style={{ fontSize: 48 }} /><p className="mt-4">暂无可显示的排行榜</p></div>
        </Card>
      )}
    </div>
  )
}
