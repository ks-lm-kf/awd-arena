import { useState } from 'react'
import {
  Card, Table, Button, Space, Modal, Form, Input, Select, Typography, Tag, message, Popconfirm, Spin, Tooltip,
} from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, SearchOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { ColumnsType } from 'antd/es/table'
import { userApi, type CreateUserRequest, type UpdateUserRequest } from '@/api/user'
import type { User } from '@/types'

const { Title } = Typography

const roleColors: Record<string, string> = { admin: 'red', organizer: 'blue', player: 'green' }
const roleLabels: Record<string, string> = { admin: '管理员', organizer: '裁判', player: '选手' }

export default function UserManage() {
  const queryClient = useQueryClient()
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<User | null>(null)
  const [form] = Form.useForm()
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading } = useQuery({
    queryKey: ['users', search, page],
    queryFn: () => userApi.list({ search: search || undefined, page, page_size: 15 }),
  })

  const createMutation = useMutation({
    mutationFn: (values: CreateUserRequest) => userApi.create(values),
    onSuccess: () => { message.success('创建成功'); closeModal(); queryClient.invalidateQueries({ queryKey: ['users'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '创建失败'),
  })

  const updateMutation = useMutation({
    mutationFn: (values: UpdateUserRequest) => userApi.update(editing!.id, values),
    onSuccess: () => { message.success('更新成功'); closeModal(); queryClient.invalidateQueries({ queryKey: ['users'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '更新失败'),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => userApi.delete(id),
    onSuccess: () => { message.success('删除成功'); queryClient.invalidateQueries({ queryKey: ['users'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '删除失败'),
  })

  const closeModal = () => { setModalOpen(false); setEditing(null); form.resetFields() }

  const openEdit = (record: User) => {
    setEditing(record)
    form.setFieldsValue({ username: record.username, email: record.email, role: record.role })
    setModalOpen(true)
  }

  const columns: ColumnsType<User> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '用户名', dataIndex: 'username' },
    { title: '邮箱', dataIndex: 'email', ellipsis: true },
    { title: '角色', dataIndex: 'role', width: 90, render: (r: string) => <Tag color={roleColors[r]}>{roleLabels[r] || r}</Tag> },
    { title: '队伍', dataIndex: 'team_name', width: 100, render: (v: string) => v || <Tag>无</Tag> },
    { title: '注册时间', dataIndex: 'created_at', width: 160, render: (t: string) => t ? new Date(t).toLocaleString() : '-' },
    {
      title: '操作', width: 260,
      render: (_, r) => (
        <Space size="small">
          <Tooltip title="编辑"><Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)} /></Tooltip>
          {r.role !== 'admin' && (
            <Popconfirm title="确定删除此用户？" onConfirm={() => deleteMutation.mutate(r.id)}>
              <Tooltip title="删除"><Button size="small" danger icon={<DeleteOutlined />} /></Tooltip>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  const handleSubmit = (values: any) => {
    if (editing) {
      const { username, password, ...rest } = values
      updateMutation.mutate(rest)
    } else {
      createMutation.mutate(values)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Title level={3} style={{ margin: 0 }}>👤 用户管理</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>创建用户</Button>
      </div>

      <Card>
        <Input placeholder="搜索用户名或邮箱" prefix={<SearchOutlined />} style={{ width: 250, marginBottom: 16 }}
          allowClear value={search} onChange={(e) => { setSearch(e.target.value); setPage(1) }} />
        <Table
          columns={columns} dataSource={data?.items || []} rowKey="id" loading={isLoading}
          pagination={{ current: page, pageSize: 15, total: data?.total || 0, onChange: setPage }}
        />
      </Card>

      <Modal
        title={editing ? '编辑用户' : '创建用户'} open={modalOpen} onCancel={closeModal}
        onOk={() => form.submit()} confirmLoading={createMutation.isPending || updateMutation.isPending} width={500}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          {!editing && (
            <Form.Item name="username" label="用户名" rules={[{ required: true }]}>
              <Input placeholder="用户名" />
            </Form.Item>
          )}
          {!editing && (
            <Form.Item name="password" label="密码" rules={[{ required: true, min: 6 }]}>
              <Input.Password placeholder="至少6位" />
            </Form.Item>
          )}
          <Form.Item name="email" label="邮箱">
            <Input placeholder="email@example.com" />
          </Form.Item>
          <Form.Item name="role" label="角色" rules={[{ required: true }]}>
            <Select options={[{ value: 'admin', label: '管理员' }, { value: 'organizer', label: '裁判' }, { value: 'player', label: '选手' }]} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
