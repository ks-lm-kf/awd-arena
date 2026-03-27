import { useState, useEffect } from 'react'
import { Card, Select, Input, Button, Typography, Table, Tag, message, Space, Tooltip } from 'antd'
import { ThunderboltOutlined, HistoryOutlined, AimOutlined } from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import { gameApi } from '@/api/game'
import { flagApi } from '@/api/flag'
import { rankingApi } from '@/api/ranking'
import { useWebSocket } from '@/hooks/useWebSocket'
import type { Game, FlagSubmission, RankingItem } from '@/types'
import dayjs from 'dayjs'

const { Title, Text } = Typography

export default function AttackPanelPage() {
  const user = useAuthStore((s) => s.user)
  const isAdmin = user?.role === 'admin'
  const [games, setGames] = useState<Game[]>([])
  const [selectedGame, setSelectedGame] = useState<number | null>(null)
  const [teams, setTeams] = useState<{ id: number; name: string }[]>([])
  const [targetTeam, setTargetTeam] = useState<number | null>(null)
  const [flagInput, setFlagInput] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [history, setHistory] = useState<FlagSubmission[]>([])
  const [flagFormat, setFlagFormat] = useState('')

  const { subscribe } = useWebSocket()

  useEffect(() => { gameApi.list().then(setGames).catch(() => {}) }, [])

  const activeGame = games.find((g) => g.status === 'running' || g.status === 'paused')
  useEffect(() => {
    if (activeGame && !selectedGame) setSelectedGame(activeGame.id)
  }, [activeGame, selectedGame])

  useEffect(() => {
    if (!selectedGame) return
    flagApi.history(selectedGame).then(setHistory).catch(() => {})
    const g = games.find((x) => x.id === selectedGame)
    if (g) setFlagFormat(g.flag_format || 'flag{...}')
  }, [selectedGame, games])

  useEffect(() => {
    if (!selectedGame) return
    rankingApi.list(selectedGame).then((rankings) => {
      setTeams(rankings.map((r) => ({ id: r.team_id, name: r.team_name })))
    }).catch(() => {})
  }, [selectedGame])

  useEffect(() => {
    if (!selectedGame) return
    const unsub = subscribe('flag:captured', (data: any) => {
      message.success(`${data.attacker} 夺取 ${data.target} 的 Flag！+${data.points} 分`)
      flagApi.history(selectedGame).then(setHistory).catch(() => {})
    })
    return unsub
  }, [selectedGame, subscribe])

  const submitFlag = async () => {
    if (!selectedGame || !targetTeam || !flagInput.trim()) { message.warning('请选择目标队伍并输入 flag'); return }
    setSubmitting(true)
    try {
      const result = await flagApi.submit(selectedGame, { flag: flagInput.trim(), target_team_id: targetTeam })
      if (result.is_correct) { message.success(`Flag 正确！+${result.points_earned} 分`); setFlagInput('') }
      else { message.error('Flag 错误') }
      flagApi.history(selectedGame).then(setHistory).catch(() => {})
    } catch (err: any) { message.error(err?.response?.data?.message || '提交失败') }
    finally { setSubmitting(false) }
  }

  const filteredTeams = teams.filter((t) => t.id !== user?.team_id)

  return (
    <div className="space-y-6">
      <Title level={3}>攻击面板</Title>
      {isAdmin && (
        <Card style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
          <Space>
            <Text>选择比赛：</Text>
            <Select style={{ width: 300 }} placeholder="请选择比赛" value={selectedGame} onChange={setSelectedGame}
              options={games.map((g) => ({ value: g.id, label: `${g.title} (${g.status})` }))} />
          </Space>
        </Card>
      )}
      {!isAdmin && activeGame && (
        <Card style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
          <div className="flex items-center gap-4">
            <span className="text-lg font-bold">{activeGame.title}</span>
            <Tag color={activeGame.status === 'running' ? 'success' : 'warning'}>{activeGame.status === 'running' ? '进行中' : '已暂停'}</Tag>
            <Text type="secondary">第 {activeGame.current_round}/{activeGame.total_rounds} 轮</Text>
            <Text type="secondary" className="ml-auto">Flag 格式：<code className="bg-gray-800 px-2 py-0.5 rounded text-indigo-400">{flagFormat}</code></Text>
          </div>
        </Card>
      )}
      {!selectedGame && (
        <Card style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
          <div className="text-center py-12 text-gray-400"><AimOutlined style={{ fontSize: 48 }} /><p className="mt-4">{isAdmin ? '请先选择比赛' : '暂无进行中的比赛'}</p></div>
        </Card>
      )}
      {selectedGame && (
        <>
          <Card title="提交 Flag" style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
            <div className="flex items-end gap-4 flex-wrap">
              <div className="flex-1 min-w-[200px]">
                <Text className="text-sm text-gray-400 mb-1 block">目标队伍</Text>
                <Select showSearch style={{ width: '100%' }} placeholder="选择目标队伍" value={targetTeam} onChange={setTargetTeam}
                  options={filteredTeams.map((t) => ({ value: t.id, label: t.name }))}
                  filterOption={(input, option) => (option?.label as string)?.toLowerCase().includes(input.toLowerCase())}
                  size="large" />
              </div>
              <div className="flex-[2] min-w-[300px]">
                <Text className="text-sm text-gray-400 mb-1 block">Flag</Text>
                <Input size="large" placeholder="输入 Flag 值..." value={flagInput} onChange={(e) => setFlagInput(e.target.value)}
                  onPressEnter={submitFlag}
                  suffix={<Tooltip title={flagFormat}><Text type="secondary" className="text-xs cursor-help">?</Text></Tooltip>} />
              </div>
              <Button type="primary" size="large" icon={<ThunderboltOutlined />} loading={submitting} onClick={submitFlag} style={{ minWidth: 120 }}>提交</Button>
            </div>
          </Card>
          <Card title={<span><HistoryOutlined /> 提交记录</span>} style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
            <Table dataSource={history} rowKey="id" size="small" pagination={{ pageSize: 20 }}
              columns={[
                { title: '时间', dataIndex: 'submitted_at', render: (v: string) => dayjs(v).format('HH:mm:ss'), sorter: (a: any, b: any) => new Date(a.submitted_at).getTime() - new Date(b.submitted_at).getTime(), defaultSortOrder: 'descend' },
                { title: '轮次', dataIndex: 'round' },
                { title: '目标', dataIndex: 'target_team', render: (v: number) => teams.find(t => t.id === v)?.name || `Team ${v}` },
                { title: 'Flag', dataIndex: 'flag_value', render: (v: string) => <code className="text-xs bg-gray-800 px-1 rounded">{v.length > 30 ? v.slice(0, 30) + '...' : v}</code> },
                { title: '结果', dataIndex: 'is_correct', render: (v: boolean, r: FlagSubmission) => v ? <Tag color="success">正确 +{r.points_earned}</Tag> : <Tag color="error">错误</Tag> },
              ]}
            />
          </Card>
        </>
      )}
    </div>
  )
}
