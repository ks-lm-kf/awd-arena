import { useState } from 'react'
import { Card, Table, Typography, Tag, Space, Button, message, Spin, Empty, Select } from 'antd'
import { ContainerOutlined, ReloadOutlined, EyeOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { ColumnsType } from 'antd/es/table'
import { containerApi } from '@/api/container'
import { gameApi } from '@/api/game'
import { statusLabel, statusColor, formatTime } from '@/utils/format'

const { Title, Text } = Typography

interface ContainerInfo {
  id: number
  team_id: number
  team_name?: string
  challenge_id: number
  challenge_name?: string
  container_id: string
  ip_address: string
  port_mapping: string
  status: string
}

export default function ContainerManage() {
  const [selectedGameId, setSelectedGameId] = useState<number | null>(null)
  const queryClient = useQueryClient()

  // 获取所有比赛列表
  const { data: games, isLoading: gamesLoading } = useQuery({
    queryKey: ['games'],
    queryFn: () => gameApi.list(),
  })

  // 获取选中比赛的容器列表
  const { data: containers, isLoading: containersLoading } = useQuery({
    queryKey: ['containers', selectedGameId],
    queryFn: () => containerApi.list(selectedGameId!),
    enabled: !!selectedGameId,
  })

  // 重启所有容器
  const restartAllMutation = useMutation({
    mutationFn: () => containerApi.restartAll(selectedGameId!),
    onSuccess: () => {
      message.success('已重启所有容器')
      queryClient.invalidateQueries({ queryKey: ['containers', selectedGameId] })
    },
    onError: () => message.error('重启失败'),
  })

  // 重启单个容器
  const restartOneMutation = useMutation({
    mutationFn: (cid: number) => containerApi.restartOne(selectedGameId!, cid),
    onSuccess: () => {
      message.success('重启成功')
      queryClient.invalidateQueries({ queryKey: ['containers', selectedGameId] })
    },
    onError: () => message.error('重启失败'),
  })

  const columns: ColumnsType<ContainerInfo> = [
    {
      title: '队伍',
      dataIndex: 'team_name',
      width: 120,
      render: (name: string) => <Text strong>{name || `队伍 ${name}`}</Text>
    },
    {
      title: '靶机',
      dataIndex: 'challenge_name',
      width: 150,
      render: (name: string) => <Text>{name || `靶机 ${name}`}</Text>
    },
    {
      title: '容器状态',
      dataIndex: 'status',
      width: 100,
      render: (s: string) => <Tag color={statusColor(s)}>{statusLabel(s)}</Tag>
    },
    {
      title: 'IP地址',
      dataIndex: 'ip_address',
      width: 130,
      render: (ip: string) => <code className="text-xs bg-gray-800 px-1.5 py-0.5 rounded">{ip}</code>
    },
    {
      title: '端口映射',
      dataIndex: 'port_mapping',
      width: 120,
      render: (mapping: string) => <Text className="text-xs">{mapping}</Text>
    },
    {
      title: '容器ID',
      dataIndex: 'container_id',
      width: 120,
      ellipsis: true,
      render: (id: string) => <code className="text-xs text-gray-400">{id?.substring(0, 12)}</code>
    },
    {
      title: '操作',
      width: 100,
      render: (_: any, r: ContainerInfo) => (
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
  ]

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Title level={3} style={{ margin: 0 }}>
          <ContainerOutlined /> 容器管理
        </Title>
        <Space>
          <Select
            placeholder="选择比赛"
            style={{ width: 300 }}
            loading={gamesLoading}
            value={selectedGameId}
            onChange={setSelectedGameId}
            options={games?.map((g: any) => ({
              label: `${g.title} (${g.status})`,
              value: g.id,
            }))}
          />
          {selectedGameId && (
            <Button
              icon={<ReloadOutlined />}
              loading={restartAllMutation.isPending}
              onClick={() => restartAllMutation.mutate()}
              danger
            >
              重启所有容器
            </Button>
          )}
        </Space>
      </div>

      <Card>
        {!selectedGameId ? (
          <Empty description="请先选择一个比赛" />
        ) : containersLoading ? (
          <div className="flex items-center justify-center h-32">
            <Spin />
          </div>
        ) : !containers?.length ? (
          <Empty description="该比赛暂无容器" />
        ) : (
          <Table
            columns={columns}
            dataSource={containers}
            rowKey="id"
            pagination={{
              pageSize: 20,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 个容器`,
            }}
            size="small"
          />
        )}
      </Card>
    </div>
  )
}
