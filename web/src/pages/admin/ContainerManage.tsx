import { useState } from 'react'
import { Card, Table, Typography, Tag, Space, Button, message, Spin, Empty, Select, Popconfirm, Descriptions } from 'antd'
import { ContainerOutlined, ReloadOutlined, EyeOutlined, InfoCircleOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { ColumnsType } from 'antd/es/table'
import type { TeamContainer } from '@/types'
import { containerApi } from '@/api/container'
import { gameApi } from '@/api/game'
import { statusLabel, statusColor, formatTime } from '@/utils/format'

const { Title, Text } = Typography

export default function ContainerManage() {
  const [selectedGameId, setSelectedGameId] = useState<number | null>(null)
  const [detailContainer, setDetailContainer] = useState<TeamContainer | null>(null)
  const queryClient = useQueryClient()

  const { data: games, isLoading: gamesLoading } = useQuery({
    queryKey: ['games'],
    queryFn: () => gameApi.list(),
  })

  const { data: containers, isLoading: containersLoading } = useQuery({
    queryKey: ['containers', selectedGameId],
    queryFn: () => containerApi.list(selectedGameId!),
    enabled: !!selectedGameId,
  })

  const restartAllMutation = useMutation({
    mutationFn: () => containerApi.restartAll(selectedGameId!),
    onSuccess: () => {
      message.success('批量重启指令已发送')
      queryClient.invalidateQueries({ queryKey: ['containers', selectedGameId] })
    },
    onError: (err: any) => message.error(err.response?.data?.message || '批量重启失败'),
  })

  const restartOneMutation = useMutation({
    mutationFn: (cid: number) => containerApi.restartOne(selectedGameId!, cid),
    onSuccess: () => {
      message.success('容器重启成功（已扣分）')
      queryClient.invalidateQueries({ queryKey: ['containers', selectedGameId] })
    },
    onError: (err: any) => message.error(err.response?.data?.message || '重启失败'),
  })

  const selectedGame = games?.find((g: any) => g.id === selectedGameId)
  const isGameFinished = selectedGame?.status === 'finished'

  const columns: ColumnsType<TeamContainer> = [
    {
      title: '队伍',
      dataIndex: 'team_name',
      width: 120,
      render: (name: string, r: TeamContainer) => <Text strong>{name || `队伍 #${r.team_id}`}</Text>
    },
    {
      title: '题目',
      dataIndex: 'challenge_name',
      width: 150,
      render: (name: string, r: TeamContainer) => <Text>{name || `题目 #${r.challenge_id}`}</Text>
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
      render: (ip: string) => ip ? <code className="text-xs bg-gray-800 px-1.5 py-0.5 rounded">{ip}</code> : <Text type="secondary">-</Text>
    },
    {
      title: '端口映射',
      dataIndex: 'port_mapping',
      width: 120,
      render: (mapping: Record<string, number>) => mapping && Object.keys(mapping).length > 0 ? <Text className="text-xs">{Object.entries(mapping).map(([k, v]) => `${k}:${v}`).join(', ')}</Text> : <Text type="secondary">-</Text>
    },
    {
      title: '容器ID',
      dataIndex: 'container_id',
      width: 120,
      ellipsis: true,
      render: (id: string) => id ? <code className="text-xs text-gray-400">{id.substring(0, 12)}</code> : <Text type="secondary">-</Text>
    },
    {
      title: '操作',
      width: 160,
      render: (_: any, r: TeamContainer) => (
        <Space size="small">
          <Popconfirm title="确认重启该容器？" description="重启将扣除50分" onConfirm={() => restartOneMutation.mutate(r.id)}>
            <Button size="small" icon={<ReloadOutlined />} loading={restartOneMutation.isPending}>重启</Button>
          </Popconfirm>
          <Button size="small" icon={<EyeOutlined />} onClick={() => setDetailContainer(r)}>详情</Button>
        </Space>
      ),
    },
  ]

  const getEmptyContent = () => {
    if (isGameFinished) {
      return (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description={
            <span>
              <InfoCircleOutlined style={{ marginRight: 8 }} />
              比赛已结束，容器已清理
            </span>
          }
        />
      )
    }
    return <Empty description="该比赛暂无容器，可能尚未开始或未配置容器" />
  }

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
              label: `${g.title} (${statusLabel(g.status)})`,
              value: g.id,
            }))}
          />
          {selectedGameId && !isGameFinished && (
            <Popconfirm title="确认重启所有容器？" description="所有容器数据将被重置" onConfirm={() => restartAllMutation.mutate()}>
              <Button
                icon={<ReloadOutlined />}
                loading={restartAllMutation.isPending}
                danger
              >
                批量重启所有
              </Button>
            </Popconfirm>
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
          getEmptyContent()
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

      {detailContainer && (
        <Card title={`容器详情 - ${detailContainer.container_id?.substring(0, 12) || 'N/A'}`} style={{ marginTop: 16 }}>
          <Descriptions column={2} size="small">
            <Descriptions.Item label="队伍ID">{detailContainer.team_id}</Descriptions.Item>
            <Descriptions.Item label="队伍名称">{detailContainer.team_name || '-'}</Descriptions.Item>
            <Descriptions.Item label="题目ID">{detailContainer.challenge_id}</Descriptions.Item>
            <Descriptions.Item label="题目名称">{detailContainer.challenge_name || '-'}</Descriptions.Item>
            <Descriptions.Item label="状态"><Tag color={statusColor(detailContainer.status)}>{statusLabel(detailContainer.status)}</Tag></Descriptions.Item>
            <Descriptions.Item label="IP地址"><code>{detailContainer.ip_address || '-'}</code></Descriptions.Item>
            <Descriptions.Item label="端口映射">{detailContainer.port_mapping && Object.keys(detailContainer.port_mapping).length > 0 ? Object.entries(detailContainer.port_mapping).map(([k, v]) => `${k}:${v}`).join(', ') : '-'}</Descriptions.Item>
            <Descriptions.Item label="容器ID"><code className="text-xs">{detailContainer.container_id || '-'}</code></Descriptions.Item>
            {detailContainer.ssh_user && <Descriptions.Item label="SSH用户">{detailContainer.ssh_user}</Descriptions.Item>}
            {detailContainer.ssh_password && <Descriptions.Item label="SSH密码"><code>{detailContainer.ssh_password}</code></Descriptions.Item>}
          </Descriptions>
          <div style={{ marginTop: 16 }}>
            <Button onClick={() => setDetailContainer(null)}>关闭</Button>
          </div>
        </Card>
      )}
    </div>
  )
}
