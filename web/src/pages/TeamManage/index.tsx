import { useState } from 'react'
import { Card, Table, Button, Modal, Form, Input, Space, Typography, Tag, message, Popconfirm, Spin, Select, AutoComplete } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, UserAddOutlined, SearchOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { ColumnsType } from 'antd/es/table'
import type { Team, TeamMember } from '@/types'
import { teamApi } from '@/api/team'
import { adminApi } from '@/api/admin'
import { userApi } from '@/api/user'
import { formatTime } from '@/utils/format'

const { Title } = Typography

export default function TeamManage() {
  const queryClient = useQueryClient()
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [memberOpen, setMemberOpen] = useState(false)
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null)
  const [createForm] = Form.useForm()
  const [editForm] = Form.useForm()
  const [addMemberForm] = Form.useForm()
  const [search, setSearch] = useState('')

  const { data: teams, isLoading } = useQuery({
    queryKey: ['teams'],
    queryFn: () => teamApi.list(),
  })

  // Filter teams by search
  const filteredTeams = teams?.filter((t) =>
    !search || t.name.toLowerCase().includes(search.toLowerCase()) || t.description?.toLowerCase().includes(search.toLowerCase())
  )

  const { data: members, isLoading: membersLoading } = useQuery({
    queryKey: ['team-members', selectedTeam?.id],
    queryFn: () => teamApi.members(selectedTeam!.id),
    enabled: !!selectedTeam,
  })

  const { data: allUsers } = useQuery({
    queryKey: ['users-all'],
    queryFn: () => userApi.list({ page: 1, page_size: 200 }),
  })

  const createMutation = useMutation({
    mutationFn: (values: { name: string; description?: string }) => teamApi.create(values),
    onSuccess: () => { message.success('创建成功'); setCreateOpen(false); createForm.resetFields(); queryClient.invalidateQueries({ queryKey: ['teams'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '创建失败'),
  })

  // Use adminApi for update to support token editing
  const updateMutation = useMutation({
    mutationFn: (values: { name?: string; description?: string; token?: string }) => adminApi.updateTeam(selectedTeam!.id, values),
    onSuccess: () => { message.success('更新成功'); setEditOpen(false); editForm.resetFields(); setSelectedTeam(null); queryClient.invalidateQueries({ queryKey: ['teams'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '更新失败'),
  })

  // Use adminApi for delete
  const deleteMutation = useMutation({
    mutationFn: (id: number) => adminApi.deleteTeam(id),
    onSuccess: () => { message.success('删除成功'); queryClient.invalidateQueries({ queryKey: ['teams'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '删除失败'),
  })

  const addMemberMutation = useMutation({
    mutationFn: ({ teamId, userId }: { teamId: number; userId: number }) => teamApi.addMember(teamId, { user_id: userId }),
    onSuccess: () => { message.success('成员已添加'); addMemberForm.resetFields(); queryClient.invalidateQueries({ queryKey: ['team-members'] }); queryClient.invalidateQueries({ queryKey: ['teams'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '添加失败'),
  })

  const removeMemberMutation = useMutation({
    mutationFn: ({ teamId, userId }: { teamId: number; userId: number }) => teamApi.removeMember(teamId, userId),
    onSuccess: () => { message.success('成员已移除'); queryClient.invalidateQueries({ queryKey: ['team-members'] }); queryClient.invalidateQueries({ queryKey: ['teams'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '移除失败'),
  })

  // Users not in any team for auto-complete
  const availableUsers = allUsers?.items?.filter((u) => !u.team_id) || []

  const columns: ColumnsType<Team> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    {
      title: '队伍名称', dataIndex: 'name',
      render: (name: string) => (
        <span className="flex items-center gap-2">
          <span className="w-7 h-7 rounded-full bg-indigo-500/20 flex items-center justify-center text-xs font-bold text-indigo-400">{name[0]}</span>
          {name}
        </span>
      ),
    },
    { title: '描述', dataIndex: 'description', ellipsis: true },
    { title: '成员数', dataIndex: 'member_count', width: 80 },
    { title: 'Token', dataIndex: 'token', width: 160, render: (t: string) => <Tag>{t}</Tag> },
    { title: '创建时间', dataIndex: 'created_at', width: 160, render: (t: string) => formatTime(t) },
    {
      title: '操作', width: 200,
      render: (_, r) => (
        <Space size="small">
          <Button size="small" icon={<EditOutlined />} onClick={() => { setSelectedTeam(r); editForm.setFieldsValue(r); setEditOpen(true) }}>编辑</Button>
          <Button size="small" onClick={() => { setSelectedTeam(r); setMemberOpen(true) }}>成员</Button>
          <Popconfirm title="确定删除此队伍？" onConfirm={() => deleteMutation.mutate(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Title level={3} style={{ margin: 0 }}>👥 队伍管理</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>创建队伍</Button>
      </div>

      <Card>
        <Input placeholder="搜索队伍" prefix={<SearchOutlined />} style={{ width: 200, marginBottom: 16 }}
          allowClear value={search} onChange={(e) => setSearch(e.target.value)} />
        {isLoading ? (
          <div className="flex items-center justify-center h-64"><Spin size="large" /></div>
        ) : (
          <Table columns={columns} dataSource={filteredTeams} rowKey="id" pagination={false} />
        )}
      </Card>

      {/* Create Modal */}
      <Modal title="创建队伍" open={createOpen} onCancel={() => { setCreateOpen(false); createForm.resetFields() }}
        onOk={() => createForm.submit()} confirmLoading={createMutation.isPending}>
        <Form form={createForm} layout="vertical" onFinish={createMutation.mutate}>
          <Form.Item name="name" label="队伍名称" rules={[{ required: true, message: '请输入队伍名称' }]}>
            <Input placeholder="输入队伍名称" />
          </Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
        </Form>
      </Modal>

      {/* Edit Modal */}
      <Modal title="编辑队伍" open={editOpen} onCancel={() => { setEditOpen(false); editForm.resetFields(); setSelectedTeam(null) }}
        onOk={() => editForm.submit()} confirmLoading={updateMutation.isPending}>
        <Form form={editForm} layout="vertical" onFinish={updateMutation.mutate}>
          <Form.Item name="name" label="队伍名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item name="token" label="口令">
            <Input placeholder="输入新的口令，留空则不修改" />
          </Form.Item>
        </Form>
      </Modal>

      {/* Members Modal */}
      <Modal title={`成员管理 — ${selectedTeam?.name || ''}`} open={memberOpen}
        onCancel={() => { setMemberOpen(false); setSelectedTeam(null) }}
        footer={null} width={600}>
        <Space style={{ marginBottom: 16, width: '100%' }}>
          <AutoComplete
            style={{ width: 300 }}
            placeholder="搜索并添加用户"
            options={availableUsers.map((u) => ({ value: String(u.id), label: `${u.username} (${u.email || '-'})` }))}
            onSelect={(value) => {
              if (selectedTeam) addMemberMutation.mutate({ teamId: selectedTeam.id, userId: Number(value) })
            }}
          />
        </Space>
        <Table
          columns={[
            { title: 'ID', dataIndex: 'id', width: 60 },
            { title: '用户名', dataIndex: 'username' },
            { title: '角色', dataIndex: 'role', render: (r: string) => <Tag color={r === 'admin' ? 'red' : 'blue'}>{r}</Tag> },
            {
              title: '操作', width: 80,
              render: (_, r) => selectedTeam && (
                <Popconfirm title="确定移除该成员？" onConfirm={() => removeMemberMutation.mutate({ teamId: selectedTeam.id, userId: r.id })}>
                  <Button size="small" danger>移除</Button>
                </Popconfirm>
              ),
            },
          ]}
          dataSource={members} rowKey="id" pagination={false} size="small" loading={membersLoading}
        />
      </Modal>
    </div>
  )
}
