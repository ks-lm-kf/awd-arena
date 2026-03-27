import { useState } from 'react'
import { Card, Table, Button, Tag, Space, Modal, Form, Input, Select, InputNumber, Typography, message, Popconfirm, Tooltip } from 'antd'
import { PlusOutlined, PlayCircleOutlined, PauseCircleOutlined, StopOutlined, ReloadOutlined, EditOutlined, DeleteOutlined, HistoryOutlined, TeamOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { ColumnsType } from 'antd/es/table'
import type { Game, GameStatus, GameMode } from '@/types'
import { gameApi } from '@/api/game'
import { adminApi } from '@/api/admin'
import { formatTime, statusLabel, statusColor } from '@/utils/format'

const { Title, Text } = Typography

const modeLabel: Record<GameMode, string> = { awd_score: 'AWD 经典', awd_mix: '攻防混合', koh: '山顶争夺' }

export default function AdminGameManage() {
  const queryClient = useQueryClient()
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Game | null>(null)
  const [form] = Form.useForm()

  const { data: games, isLoading } = useQuery({
    queryKey: ['admin-games'],
    queryFn: () => gameApi.list(),
  })

  const createMutation = useMutation({
    mutationFn: (values: any) => adminApi.createGame(values),
    onSuccess: () => { message.success('创建成功'); closeModal(); queryClient.invalidateQueries({ queryKey: ['admin-games'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '创建失败'),
  })

  const updateMutation = useMutation({
    mutationFn: (values: any) => adminApi.updateGame(editing!.id, values),
    onSuccess: () => { message.success('更新成功'); closeModal(); queryClient.invalidateQueries({ queryKey: ['admin-games'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '更新失败'),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => adminApi.deleteGame(id),
    onSuccess: () => { message.success('删除成功'); queryClient.invalidateQueries({ queryKey: ['admin-games'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '删除失败'),
  })

  const actionMutation = useMutation({
    mutationFn: ({ action, id }: { action: string; id: number }) => {
      switch (action) {
        case '开始': return adminApi.startGame(id)
        case '暂停': return adminApi.pauseGame(id)
        case '继续': return adminApi.resumeGame(id)
        case '结束': return adminApi.stopGame(id)
        case '重置': return adminApi.resetGame(id)
        default: return Promise.resolve()
      }
    },
    onSuccess: (_, v) => { message.success(`${v.action}操作成功`); queryClient.invalidateQueries({ queryKey: ['admin-games'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '操作失败'),
  })

  const closeModal = () => { setModalOpen(false); setEditing(null); form.resetFields() }

  const handleFinish = (values: any) => {
    if (editing) updateMutation.mutate(values)
    else createMutation.mutate(values)
  }

  const columns: ColumnsType<Game> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '竞赛名称', dataIndex: 'title', ellipsis: true },
    { title: '模式', dataIndex: 'mode', width: 100, render: (m: GameMode) => modeLabel[m] },
    { title: '状态', dataIndex: 'status', width: 90, render: (s: GameStatus) => <Tag color={statusColor(s)}>{statusLabel(s)}</Tag> },
    { title: '进度', width: 100, render: (_, r) => <span>{r.current_round} / {r.total_rounds} 轮</span> },
    { title: '开始时间', dataIndex: 'start_time', width: 160, render: (t: string) => formatTime(t) },
    {
      title: '操作', width: 280, fixed: 'right',
      render: (_, r) => (
        <Space size="small" wrap>
          {r.status === 'draft' && <Button size="small" icon={<EditOutlined />} onClick={() => { setEditing(r); form.setFieldsValue(r); setModalOpen(true) }}>编辑</Button>}
          {r.status === 'draft' && (
            <Popconfirm title="确定开始比赛？" onConfirm={() => actionMutation.mutate({ action: '开始', id: r.id })}>
              <Button size="small" type="primary" icon={<PlayCircleOutlined />}>开始</Button>
            </Popconfirm>
          )}
          {r.status === 'active' && r.current_phase === 'running' && (
            <Button size="small" icon={<PauseCircleOutlined />} onClick={() => actionMutation.mutate({ action: '暂停', id: r.id })}>暂停</Button>
          )}
          {r.status === 'active' && r.current_phase === 'break' && (
            <Button size="small" type="primary" icon={<PlayCircleOutlined />} onClick={() => actionMutation.mutate({ action: '继续', id: r.id })}>继续</Button>
          )}
          {r.status === 'active' && (
            <Popconfirm title="确定结束比赛？" onConfirm={() => actionMutation.mutate({ action: '结束', id: r.id })}>
              <Button size="small" danger icon={<StopOutlined />}>结束</Button>
            </Popconfirm>
          )}
          {r.status === 'draft' && (
            <Popconfirm title="确定删除？不可恢复！" onConfirm={() => deleteMutation.mutate(r.id)}>
              <Button size="small" danger icon={<DeleteOutlined />} />
            </Popconfirm>
          )}
          <Popconfirm title="确定重置？数据将丢失！" onConfirm={() => actionMutation.mutate({ action: '重置', id: r.id })}>
            <Button size="small" icon={<ReloadOutlined />}>重置</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const isPending = createMutation.isPending || updateMutation.isPending

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Title level={3} style={{ margin: 0 }}>🎮 裁判 - 竞赛管理</Title>
        <Space>
          <Button icon={<HistoryOutlined />} onClick={() => message.info('操作日志功能开发中')}>操作日志</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setModalOpen(true) }}>创建竞赛</Button>
        </Space>
      </div>

      <Card>
        <Text type="secondary" style={{ marginBottom: 16, display: 'block' }}>
          📋 仅裁判和管理员可访问。所有操作将被记录。
        </Text>
        {isLoading ? <div className="flex items-center justify-center h-64" /> : (
          <Table columns={columns} dataSource={games} rowKey="id" scroll={{ x: 1200 }} />
        )}
      </Card>

      <Modal
        title={editing ? '编辑竞赛' : '创建竞赛'} open={modalOpen} onCancel={closeModal}
        onOk={() => form.submit()} confirmLoading={isPending} width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleFinish}>
          <Form.Item name="title" label="竞赛名称" rules={[{ required: true, message: '请输入竞赛名称' }]} >
            <Input placeholder="输入竞赛名称" />
          </Form.Item>
          <Form.Item name="mode" label="竞赛模式" initialValue="awd_score">
            <Select options={[{ value: 'awd_score', label: 'AWD 经典' }, { value: 'awd_mix', label: '攻防混合' }, { value: 'koh', label: '山顶争夺' }]} />
          </Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
          <Space size="large">
            <Form.Item name="total_rounds" label="总轮数" initialValue={20}><InputNumber min={1} max={100} /></Form.Item>
            <Form.Item name="round_duration" label="每轮时长(秒)" initialValue={300}><InputNumber min={60} step={60} /></Form.Item>
            <Form.Item name="break_duration" label="休息时长(秒)" initialValue={120}><InputNumber min={30} step={30} /></Form.Item>
          </Space>
          <Form.Item name="flag_format" label="Flag 格式" initialValue="flag{...}"><Input placeholder="flag{...}" /></Form.Item>
          <Space size="large">
            <Form.Item name="attack_weight" label="攻击权重" initialValue={1.0}><InputNumber min={0} step={0.1} /></Form.Item>
            <Form.Item name="defense_weight" label="防守权重" initialValue={0.5}><InputNumber min={0} step={0.1} /></Form.Item>
          </Space>
        </Form>
      </Modal>
    </div>
  )
}
