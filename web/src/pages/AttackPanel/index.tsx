import { useState, useEffect } from 'react'
import { Card, Select, Input, Button, Typography, Table, Tag, message, Space, Spin, Empty } from 'antd'
import { ThunderboltOutlined, HistoryOutlined, AimOutlined, DesktopOutlined, CopyOutlined } from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import { gameApi } from '@/api/game'
import { flagApi } from '@/api/flag'
import { containerApi } from '@/api/container'
import { useWebSocket } from '@/hooks/useWebSocket'
import type { Game, TeamContainer } from '@/types'
import dayjs from 'dayjs'

const { Title, Text } = Typography

export default function AttackPanelPage() {
  const user = useAuthStore((s) => s.user)
  const [games, setGames] = useState<Game[]>([])
  const [selectedGame, setSelectedGame] = useState<number | null>(null)
  const [flagInput, setFlagInput] = useState('')
  const [targetTeamId, setTargetTeamId] = useState<number>(0)
  const [submitting, setSubmitting] = useState(false)
  const [machines, setMachines] = useState<TeamContainer[]>([])
  const [machinesLoading, setMachinesLoading] = useState(false)
  const [history, setHistory] = useState<any[]>([])

  const { subscribe } = useWebSocket()

  useEffect(() => { gameApi.list().then(setGames).catch((err) => { console.error('Failed to load data:', err); message.error('加载数据失败') }) }, [])

  const activeGame = games.find((g) => g.status === 'running' || g.status === 'paused')
  useEffect(() => {
    if (activeGame && !selectedGame) setSelectedGame(activeGame.id)
  }, [activeGame, selectedGame])

  useEffect(() => {
    if (!selectedGame) return
    setMachinesLoading(true)
    containerApi.getMyMachines(selectedGame)
      .then((data) => { setMachines(data || []); setMachinesLoading(false) })
      .catch((err) => { console.error('Failed to load machines:', err); setMachinesLoading(false) })
  }, [selectedGame])

  useEffect(() => {
    if (!selectedGame) return
    flagApi.history(selectedGame).then(setHistory).catch((err) => { console.error('Failed to load data:', err); message.error('加载数据失败') })
  }, [selectedGame])

  useEffect(() => {
    if (!selectedGame) return
    const unsub = subscribe('flag:captured', () => {
      flagApi.history(selectedGame).then(setHistory).catch((err) => { console.error('Failed to load data:', err); message.error('加载数据失败') })
    })
    return unsub
  }, [selectedGame, subscribe])

  const submitFlag = async () => {
    if (!selectedGame || !flagInput.trim()) { message.warning('请输入 flag'); return }
    setSubmitting(true)
    try {
      const result = await flagApi.submit(selectedGame, { flag: flagInput.trim(), target_team_id: targetTeamId })
      if (result.is_correct) { message.success(`Flag 正确！+${result.points_earned} 分`); setFlagInput('') }
      else { message.error('Flag 错误') }
      flagApi.history(selectedGame).then(setHistory).catch((err) => { console.error('Failed to load data:', err); message.error('加载数据失败') })
    } catch (err: any) { message.error(err?.response?.data?.message || '提交失败') }
    finally { setSubmitting(false) }
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text).then(() => message.success('已复制')).catch(() => {})
  }

  const machineColumns = [
    { title: '题目', dataIndex: 'challenge_name', key: 'challenge_name', render: (n: string) => <Text strong>{n || '未知'}</Text> },
    { title: 'IP 地址', dataIndex: 'ip_address', key: 'ip_address', render: (ip: string) => <Space><code className="text-xs bg-gray-800 px-2 py-1 rounded">{ip}</code><CopyOutlined className="cursor-pointer text-gray-400 hover:text-white" onClick={() => copyToClipboard(ip)} /></Space> },
    { title: '端口', dataIndex: 'port_mapping', key: 'port_mapping', render: (m: any) => <code className="text-xs bg-gray-800 px-2 py-1 rounded">{m ? (typeof m === 'string' ? m : JSON.stringify(m)) : '-'}</code> },
    { title: 'SSH', key: 'ssh', render: (_: any, r: any) => <Space direction="vertical" size={0}>{r.ssh_user && <Text className="text-xs">用户: {r.ssh_user}</Text>}{r.ssh_password && <Space size={4}><Text className="text-xs">密码: {r.ssh_password}</Text><CopyOutlined className="cursor-pointer text-gray-400 hover:text-white text-xs" onClick={() => copyToClipboard(r.ssh_password)} /></Space>}{r.ssh_port && <Text className="text-xs">端口: {r.ssh_port}</Text>}</Space> },
    { title: '状态', dataIndex: 'status', key: 'status', render: (s: string) => <Tag color={s === 'running' ? 'success' : 'error'}>{s === 'running' ? '运行中' : s}</Tag> },
  ]

  const historyColumns = [
    { title: '时间', dataIndex: 'submitted_at', key: 'submitted_at', width: 160, render: (t: string) => dayjs(t).format('HH:mm:ss') },
    { title: 'Flag', dataIndex: 'flag', key: 'flag', ellipsis: true, render: (f: string) => <code className="text-xs">{f}</code> },
    { title: '结果', dataIndex: 'is_correct', key: 'is_correct', width: 80, render: (c: boolean) => <Tag color={c ? 'success' : 'error'}>{c ? '正确' : '错误'}</Tag> },
    { title: '得分', dataIndex: 'points_earned', key: 'points_earned', width: 80, render: (p: number) => p > 0 ? <Text className="text-green-400">+{p}</Text> : <Text className="text-gray-500">0</Text> },
  ]

  const currentGame = games.find((g) => g.id === selectedGame)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Title level={3}><AimOutlined className="mr-2" />攻击面板</Title>
        <Space><Text>选择比赛：</Text><Select style={{ width: 250 }} placeholder="选择比赛" value={selectedGame} onChange={setSelectedGame} options={games.map((g) => ({ value: g.id, label: `${g.title} (${g.status})` }))} /></Space>
      </div>
      {currentGame && <Card style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }} size="small"><Space><Text>{currentGame.title}</Text><Tag color={currentGame.status === 'running' ? 'success' : 'default'}>{currentGame.status === 'running' ? '进行中' : currentGame.status}</Tag><Text type="secondary">第 {currentGame.current_round}/{currentGame.total_rounds} 轮</Text></Space></Card>}
      <Card title={<Space><DesktopOutlined />我的靶机</Space>} style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
        {machinesLoading ? <div className="text-center py-8"><Spin /></div> : machines.length === 0 ? <Empty description="暂无靶机信息" /> : <Table dataSource={machines} columns={machineColumns} rowKey="id" pagination={false} size="small" />}
      </Card>
      <Card title={<Space><ThunderboltOutlined />提交 Flag</Space>} style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
        <Space.Compact style={{ width: '100%' }}>
          <Input placeholder={currentGame?.flag_format || 'flag{...}'} value={flagInput} onChange={(e) => setFlagInput(e.target.value)} onPressEnter={submitFlag} size="large" allowClear />
          <Button type="primary" size="large" onClick={submitFlag} loading={submitting} icon={<ThunderboltOutlined />}>提交</Button>
        </Space.Compact>
      </Card>
      <Card title={<Space><HistoryOutlined />提交历史</Space>} style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
        <Table dataSource={history} columns={historyColumns} rowKey={(r: any) => String(r.id)} pagination={{ pageSize: 10 }} size="small" />
      </Card>
    </div>
  )
}
