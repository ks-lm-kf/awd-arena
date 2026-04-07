import { useState } from 'react'
import { useNavigate } from 'react-router'
import { Card, Table, Button, Tag, Space, Modal, Form, Input, Select, InputNumber, Typography, message, Popconfirm, Spin } from 'antd'
import { PlusOutlined, PlayCircleOutlined, EyeOutlined, PauseCircleOutlined, StopOutlined, ReloadOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { ColumnsType } from 'antd/es/table'
import type { Game, GameStatus, GameMode } from '@/types'
import { gameApi } from '@/api/game'
import { formatTime, statusLabel, statusColor } from '@/utils/format'

const { Title } = Typography

const modeLabel: Record<GameMode, string> = { awd_score: 'AWD 经典', awd_mix: '攻防混合', koh: '山顶争夺' }

export default function GameManage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Game | null>(null)
  const [form] = Form.useForm()

  const { data: games, isLoading } = useQuery({
    queryKey: ['games'],
    queryFn: () => gameApi.list(),
  })

  const createMutation = useMutation({
    mutationFn: (values: any) => gameApi.create(values),
    onSuccess: () => { message.success('创建成功'); closeModal(); queryClient.invalidateQueries({ queryKey: ['games'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '创建失败'),
  })

  const updateMutation = useMutation({
    mutationFn: (values: any) => gameApi.update(editing!.id, values),
    onSuccess: () => { message.success('更新成功'); closeModal(); queryClient.invalidateQueries({ queryKey: ['games'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '更新失败'),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => gameApi.delete(id),
    onSuccess: () => { message.success('删除成功'); queryClient.invalidateQueries({ queryKey: ['games'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '删除失败'),
  })

  const actionMutation = useMutation({
    mutationFn: ({ action, id }: { action: string; id: number }) => {
      switch (action) {
        case '开始': return gameApi.start(id)
        case '暂停': return gameApi.pause(id)
        case '继续': return gameApi.resume(id)
        case '结束': return gameApi.stop(id)
        case '重置': return gameApi.reset(id)
        default: return Promise.resolve()
      }
    },
    onSuccess: (_, v) => { message.success(`${v.action}操作成功`); queryClient.invalidateQueries({ queryKey: ['games'] }) },
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
      title: '操作', width: 260,
      render: (_, r) => (
        <Space size="small" wrap>
          <Button size="small" icon={<EyeOutlined />} onClick={() => navigate(`/admin/games/${r.id}`)}>详情</Button>
          {r.status === 'draft' && <Button size="small" icon={<EditOutlined />} onClick={() => { setEditing(r); form.setFieldsValue(r); setModalOpen(true) }}>编辑</Button>}
          {r.status === 'draft' && (
            <Popconfirm title="确定开始比赛？" description="比赛开始后将自动创建容器" onConfirm={() => actionMutation.mutate({ action: '开始', id: r.id })}>
              <Button size="small" type="primary" icon={<PlayCircleOutlined />}>开始</Button>
            </Popconfirm>
          )}
          {(r.status === 'running') && (
            <Popconfirm title="确定暂停比赛？" onConfirm={() => actionMutation.mutate({ action: '暂停', id: r.id })}>
              <Button size="small" icon={<PauseCircleOutlined />}>暂停</Button>
            </Popconfirm>
          )}
          {r.status === 'paused' && (
            <Popconfirm title="确定继续比赛？" onConfirm={() => actionMutation.mutate({ action: '继续', id: r.id })}>
              <Button size="small" type="primary" icon={<PlayCircleOutlined />}>继续</Button>
            </Popconfirm>
          )}
          {(r.status === 'running' || r.status === 'paused') && (
            <Popconfirm title="确定结束比赛？" description="结束后不可恢复，请谨慎操作！" onConfirm={() => actionMutation.mutate({ action: '结束', id: r.id })}>
              <Button size="small" danger icon={<StopOutlined />}>结束</Button>
            </Popconfirm>
          )}
          {r.status === 'draft' && (
            <Popconfirm title="确定删除？不可恢复！" onConfirm={() => deleteMutation.mutate(r.id)}>
              <Button size="small" danger icon={<DeleteOutlined />} />
            </Popconfirm>
          )}
          {(r.status === 'finished' || r.status === 'draft') && (
            <Popconfirm title="确定重置？所有比赛数据将丢失！" onConfirm={() => actionMutation.mutate({ action: '重置', id: r.id })}>
              <Button size="small" icon={<ReloadOutlined />}>重置</Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  const isPending = createMutation.isPending || updateMutation.isPending

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Title level={3} style={{ margin: 0 }}>🎮 竞赛管理</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setModalOpen(true) }}>创建竞赛</Button>
      </div>

      <Card>
        {isLoading ? <div className="flex items-center justify-center h-64"><Spin /></div> : (
          <Table columns={columns} dataSource={games} rowKey="id" pagination={false} scroll={{ x: 1200 }} />
        )}
      </Card>

      <Modal
        title={editing ? '编辑竞赛' : '创建竞赛'} open={modalOpen} onCancel={closeModal}
        onOk={() => form.submit()} confirmLoading={isPending} width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleFinish}>
          <Form.Item name="title" label="竞赛名称" rules={[{ required: true, message: '请输入竞赛名称' }]}>
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
          <Form.Item name="flag_format" label="Flag 格式" initialValue="flag{...}">
            <Input placeholder="flag{...}" />
          </Form.Item>
          <Space size="large">
            <Form.Item name="initial_score" label="初始分数" initialValue={1000}><InputNumber min={0} step={100} /></Form.Item>
            <Form.Item name="attack_weight" label="攻击权重" initialValue={1.0}><InputNumber min={0} step={0.1} /></Form.Item>
            <Form.Item name="defense_weight" label="防守权重" initialValue={1.0}><InputNumber min={0} step={0.1} /></Form.Item>
          </Space>
        </Form>
      </Modal>
    </div>
  )
}
