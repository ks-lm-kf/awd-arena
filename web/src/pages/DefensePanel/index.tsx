import { useState, useEffect } from 'react'
import { useParams } from 'react-router'
import { Card, Table, Typography, Tag, Space, Button, message, Spin, Empty } from 'antd'
import { SafetyCertificateOutlined, ReloadOutlined, AlertOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { ColumnsType } from 'antd/es/table'
import type { TeamContainer, SecurityAlert, WSAlertNew } from '@/types'
import { containerApi } from '@/api/container'
import { gameApi } from '@/api/game'
import { useWebSocket } from '@/hooks/useWebSocket'
import { statusLabel, statusColor, formatTime } from '@/utils/format'

const { Title, Text } = Typography

const containerColumns: ColumnsType<TeamContainer> = [
  { title: '靶机', dataIndex: 'challenge_name', render: (n: string) => <Text strong>{n}</Text> },
  { title: '容器状态', dataIndex: 'status', width: 100, render: (s: string) => <Tag color={statusColor(s)}>{statusLabel(s)}</Tag> },
  { title: 'IP', dataIndex: 'ip_address', width: 130, render: (ip: string) => <code className="text-xs bg-gray-800 px-1.5 py-0.5 rounded">{ip}</code> },
  { title: '端口映射', width: 120, render: (_, r) => Object.entries(r.port_mapping).map(([c, h]) => `${c}→${h}`).join(', ') },
  { title: '容器ID', dataIndex: 'container_id', width: 120, ellipsis: true, render: (id: string) => <code className="text-xs text-gray-400">{id}</code> },
]

const alertLevelColor: Record<string, string> = { critical: 'red', warning: 'orange', info: 'blue' }

export default function DefensePanel() {
  const { id } = useParams<{ id: string }>()
  const gameId = Number(id)
  const queryClient = useQueryClient()
  const { subscribe } = useWebSocket()
  const [alerts, setAlerts] = useState<SecurityAlert[]>([])

  const { data: containers, isLoading: containersLoading } = useQuery({
    queryKey: ['containers', gameId],
    queryFn: () => containerApi.list(gameId),
    enabled: !!gameId,
  })

  const { data: apiAlerts } = useQuery({
    queryKey: ['alerts', gameId],
    queryFn: () => gameApi.alerts(gameId),
    enabled: !!gameId,
  })

  useEffect(() => {
    if (apiAlerts) setAlerts(apiAlerts)
  }, [apiAlerts])

  const restartAllMutation = useMutation({
    mutationFn: () => containerApi.restartAll(gameId),
    onSuccess: () => {
      message.success('已重启所有容器')
      queryClient.invalidateQueries({ queryKey: ['containers', gameId] })
    },
    onError: () => message.error('重启失败'),
  })

  const restartOneMutation = useMutation({
    mutationFn: (cid: number) => containerApi.restartOne(gameId, cid),
    onSuccess: () => {
      message.success('重启成功')
      queryClient.invalidateQueries({ queryKey: ['containers', gameId] })
    },
    onError: () => message.error('重启失败'),
  })

  // WebSocket for real-time alerts
  useEffect(() => {
    const unsub = subscribe('alert:new', (data: WSAlertNew) => {
      setAlerts((prev) => [
        {
          id: Date.now(),
          game_id: gameId,
          level: data.level,
          team_id: 0,
          type: 'other',
          detail: data.message,
          created_at: new Date().toISOString(),
        },
        ...prev.slice(0, 49),
      ])
    })
    return () => unsub()
  }, [subscribe, gameId])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Title level={3} style={{ margin: 0 }}>🛡️ 防御面板</Title>
        <Button
          icon={<ReloadOutlined />}
          loading={restartAllMutation.isPending}
          onClick={() => restartAllMutation.mutate()}
        >
          重启所有容器
        </Button>
      </div>

      <Card title={<span><SafetyCertificateOutlined /> 我的靶机容器</span>}>
        {containersLoading ? (
          <div className="flex items-center justify-center h-32"><Spin /></div>
        ) : !containers?.length ? (
          <Empty description="暂无容器" />
        ) : (
          <Table
            columns={[
              ...containerColumns,
              {
                title: '操作', width: 100,
                render: (_: any, r: TeamContainer) => (
                  <Button
                    size="small"
                    icon={<ReloadOutlined />}
                    loading={restartOneMutation.isPending}
                    onClick={() => restartOneMutation.mutate(r.id)}
                  >
                    重启
                  </Button>
                ),
              },
            ]}
            dataSource={containers}
            rowKey="id"
            pagination={false}
            size="small"
          />
        )}
      </Card>

      <Card title={<span><AlertOutlined /> 安全告警</span>}>
        {alerts.length === 0 ? (
          <Empty description="暂无告警" />
        ) : (
          alerts.map((alert) => (
            <div key={alert.id} className="flex items-start gap-3 py-3 border-b border-[#2a2a4a] last:border-0">
              <Tag color={alertLevelColor[alert.level]} style={{ marginTop: 2 }}>
                {alert.level.toUpperCase()}
              </Tag>
              <div className="flex-1">
                <Text className="text-sm">{alert.detail}</Text>
                <div className="text-xs text-gray-500 mt-1">
                  {alert.team_name} · {alert.type} · {formatTime(alert.created_at)}
                </div>
              </div>
            </div>
          ))
        )}
      </Card>
    </div>
  )
}
