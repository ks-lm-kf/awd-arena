import { useState, useEffect } from 'react'
import { Card, Select, Input, Button, Typography, Table, Tag, message, Space, Descriptions, Spin, Empty, Alert } from 'antd'
import { ThunderboltOutlined, HistoryOutlined, AimOutlined, DesktopOutlined, CopyOutlined } from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import { gameApi } from '@/api/game'
import { flagApi } from '@/api/flag'
import { rankingApi } from '@/api/ranking'
import { containerApi } from '@/api/container'
import { useWebSocket } from '@/hooks/useWebSocket'
import type { Game, Container } from '@/types'
import dayjs from 'dayjs'

const { Title, Text } = Typography

export default function AttackPanelPage() {
  const user = useAuthStore((s) => s.user)
  const [games, setGames] = useState<Game[]>([])
  const [selectedGame, setSelectedGame] = useState<number | null>(null)
  const [flagInput, setFlagInput] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [machines, setMachines] = useState<Container[]>([])
  const [machinesLoading, setMachinesLoading] = useState(false)
  const [history, setHistory] = useState<any[]>([])

  const { subscribe } = useWebSocket()

  // Load games
  useEffect(() => { gameApi.list().then(setGames).catch(() => {}) }, [])

  // Auto-select active game
  const activeGame = games.find((g) => g.status === 'running' || g.status === 'paused')
  useEffect(() => {
    if (activeGame && !selectedGame) setSelectedGame(activeGame.id)
  }, [activeGame, selectedGame])

  // Load machines for selected game
  useEffect(() => {
    if (!selectedGame) return
    setMachinesLoading(true)
    // Use my-machines endpoint for player's own containers
    fetch(`/api/v1/games/${selectedGame}/my-machines`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
    })
      .then(res => res.json())
      .then(res => {
        setMachines(res.data || [])
        setMachinesLoading(false)
      })
      .catch(() => setMachinesLoading(false))
  }, [selectedGame])

  // Load flag history
  useEffect(() => {
    if (!selectedGame) return
    flagApi.history(selectedGame).then(setHistory).catch(() => {})
  }, [selectedGame])

  // WebSocket for flag captured events
  useEffect(() => {
    if (!selectedGame) return
    const unsub = subscribe('flag:captured', () => {
      flagApi.history(selectedGame).then(setHistory).catch(() => {})
    })
    return unsub
  }, [selectedGame, subscribe])

  const submitFlag = async () => {
    if (!selectedGame || !flagInput.trim()) {
      message.warning('请输入 flag')
      return
    }
    setSubmitting(true)
    try {
      const result = await flagApi.submit(selectedGame, { flag: flagInput.trim() })
      if (result.is_correct) {
        message.success(`Flag 正确！+${result.points_earned} 分`)
        setFlagInput('')
      } else {
        message.error('Flag 错误')
      }
      flagApi.history(selectedGame).then(setHistory).catch(() => {})
    } catch (err: any) {
      message.error(err?.response?.data?.message || '提交失败')
    } finally {
      setSubmitting(false)
    }
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text).then(() => message.success('已复制')).catch(() => {})
  }

  const machineColumns = [
    {
      title: '题目名称',
      dataIndex: 'challenge_name',
      key: 'challenge_name',
      render: (name: string) => <Text strong>{name || '未知'}</Text>,
    },
    {
      title: 'IP 地址',
      dataIndex: 'ip_address',
      key: 'ip_address',
      render: (ip: string) => (
        <Space>
          <code className="text-xs bg-gray-800 px-2 py-1 rounded">{ip}</code>
          <CopyOutlined className="cursor-pointer text-gray-400 hover:text-white" onClick={() => copyToClipboard(ip)} />
        </Space>
      ),
    },
    {
      title: '端口映射',
      dataIndex: 'port_mapping',
      key: 'port_mapping',
      render: (mapping: any) => {
        if (!mapping) return '-'
        const ports = typeof mapping === 'string' ? mapping : JSON.stringify(mapping)
        return <code className="text-xs bg-gray-800 px-2 py-1 rounded">{ports}</code>
      },
    },
    {
      title: 'SSH 信息',
      key: 'ssh',
      render: (_: any, record: any) => (
        <Space direction="vertical" size={0}>
          {record.ssh_user && <Text className="text-xs">用户: {record.ssh_user}</Text>}
          {record.ssh_password && (
            <Space size={4}>
              <Text className="text-xs">密码: {record.ssh_password}</Text>
              <CopyOutlined className="cursor-pointer text-gray-400 hover:text-white text-xs" onClick={() => copyToClipboard(record.ssh_password)} />
            </Space>
          )}
          {record.ssh_port && <Text className="text-xs">端口: {record.ssh_port}</Text>}
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (s: string) => {
        const color = s === 'running' ? 'success' : s === 'stopped' ? 'error' : 'default'
        const label = s === 'running' ? '运行中' : s === 'stopped' ? '已停止' : s
        return <Tag color={color}>{label}</Tag>
      },
    },
  ]

  const historyColumns = [
    {
      title: '时间',
      dataIndex: 'submitted_at',
      key: 'submitted_at',
      width: 160,
      render: (t: string) => dayjs(t).format('HH:mm:ss'),
    },
    {
      title: 'Flag',
      dataIndex: 'flag',
      key: 'flag',
      ellipsis: true,
      render: (f: string) => <code className="text-xs">{f}</code>,
    },
    {
      title: '结果',
      dataIndex: 'is_correct',
      key: 'is_correct',
      width: 80,
      render: (correct: boolean) => <Tag color={correct ? 'success' : 'error'}>{correct ? '正确' : '错误'}</Tag>,
    },
    {
      title: '得分',
      dataIndex: 'points_earned',
      key: 'points_earned',
      width: 80,
      render: (p: number) => p > 0 ? <Text className="text-green-400">+{p}</Text> : <Text className="text-gray-500">0</Text>,
    },
  ]

  const currentGame = games.find((g) => g.id === selectedGame)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Title level={3}><AimOutlined className="mr-2" />攻击面板</Title>
        <Space>
          <Text>选择比赛：</Text>
          <Select
            style={{ width: 250 }}
            placeholder="选择比赛"
            value={selectedGame}
            onChange={setSelectedGame}
            options={games.map((g) => ({ value: g.id, label: `${g.title} (${g.status})` }))}
          />
        </Space>
      </div>

      {currentGame && (
        <Card style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }} size="small">
          <Space>
            <Text>{currentGame.title}</Text>
            <Tag color={currentGame.status === 'running' ? 'success' : 'default'}>
              {currentGame.status === 'running' ? '进行中' : currentGame.status}
            </Tag>
            <Text type="secondary">第 {currentGame.current_round}/{currentGame.total_rounds} 轮</Text>
            {user?.team_name && <Tag color="blue">队伍: {user.team_name}</Tag>}
          </Space>
        </Card>
      )}

      {/* 靶机信息 */}
      <Card
        title={<Space><DesktopOutlined />我的靶机</Space>}
        style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}
      >
        {machinesLoading ? (
          <div className="text-center py-8"><Spin /></div>
        ) : machines.length === 0 ? (
          <Empty description="暂无靶机信息，请等待比赛开始" />
        ) : (
          <Table
            dataSource={machines}
            columns={machineColumns}
            rowKey="id"
            pagination={false}
            size="small"
          />
        )}
      </Card>

      {/* 提交 Flag */}
      <Card
        title={<Space><ThunderboltOutlined />提交 Flag</Space>}
        style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}
      >
        <Space.Compact style={{ width: '100%' }}>
          <Input
            placeholder={currentGame?.flag_format || 'flag{...}'}
            value={flagInput}
            onChange={(e) => setFlagInput(e.target.value)}
            onPressEnter={submitFlag}
            size="large"
            allowClear
          />
          <Button
            type="primary"
            size="large"
            onClick={submitFlag}
            loading={submitting}
            icon={<ThunderboltOutlined />}
          >
            提交
          </Button>
        </Space.Compact>
      </Card>

      {/* 提交历史 */}
      <Card
        title={<Space><HistoryOutlined />提交历史</Space>}
        style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}
      >
        <Table
          dataSource={history}
          columns={historyColumns}
          rowKey={(r: any) => r.id || Math.random()}
          pagination={{ pageSize: 10 }}
          size="small"
        />
      </Card>
    </div>
  )
}
