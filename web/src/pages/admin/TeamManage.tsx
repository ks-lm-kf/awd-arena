import { useState } from 'react'
import { Card, Table, Button, Modal, Form, Input, Space, Typography, Tag, message, Popconfirm, Upload, InputNumber, Divider } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, UploadOutlined, ReloadOutlined, DollarOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { ColumnsType } from 'antd/es/table'
import type { Team } from '@/types'
import { teamApi } from '@/api/team'
import { adminApi } from '@/api/admin'
import { formatTime } from '@/utils/format'

const { Title, Text } = Typography

export default function AdminTeamManage() {
  const queryClient = useQueryClient()
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [importOpen, setImportOpen] = useState(false)
  const [adjustScoreOpen, setAdjustScoreOpen] = useState(false)
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null)
  const [createForm] = Form.useForm()
  const [editForm] = Form.useForm()
  const [importForm] = Form.useForm()
  const [adjustScoreForm] = Form.useForm()
  const [importPreview, setImportPreview] = useState<Array<{ name: string; description?: string }>>([])

  const { data: teams, isLoading } = useQuery({
    queryKey: ['admin-teams'],
    queryFn: () => teamApi.list(),
  })

  const createMutation = useMutation({
    mutationFn: (values: { name: string; description?: string }) => adminApi.createTeam(values),
    onSuccess: () => { message.success('创建成功'); setCreateOpen(false); createForm.resetFields(); queryClient.invalidateQueries({ queryKey: ['admin-teams'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '创建失败'),
  })

  const updateMutation = useMutation({
    mutationFn: (values: { name: string; description?: string }) => adminApi.updateTeam(selectedTeam!.id, values),
    onSuccess: () => { message.success('更新成功'); setEditOpen(false); editForm.resetFields(); queryClient.invalidateQueries({ queryKey: ['admin-teams'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '更新失败'),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => adminApi.deleteTeam(id),
    onSuccess: () => { message.success('删除成功'); queryClient.invalidateQueries({ queryKey: ['admin-teams'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '删除失败'),
  })

  const resetMutation = useMutation({
    mutationFn: (id: number) => adminApi.resetTeam(id),
    onSuccess: () => { message.success('重置成功'); queryClient.invalidateQueries({ queryKey: ['admin-teams'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '重置失败'),
  })

  const importMutation = useMutation({
    mutationFn: (values: any) => adminApi.batchImportTeams({ teams: importPreview }),
    onSuccess: (data) => {
      message.success(`成功导入 ${data.success}/${data.total} 个队伍`)
      if (data.errors && data.errors.length > 0) {
        Modal.error({
          title: '部分导入失败',
          content: (
            <div>
              {data.errors.map((err: string, idx: number) => (
                <Text key={idx} type="danger">{err}<br /></Text>
              ))}
            </div>
          ),
        })
      }
      setImportOpen(false)
      setImportPreview([])
      importForm.resetFields()
      queryClient.invalidateQueries({ queryKey: ['admin-teams'] })
    },
    onError: (err: any) => message.error(err.response?.data?.message || '导入失败'),
  })

  const adjustScoreMutation = useMutation({
    mutationFn: (values: any) => adminApi.adjustScore(values),
    onSuccess: (data) => {
      message.success(`分数已调整: ${data.old_score} → ${data.new_score}`)
      setAdjustScoreOpen(false)
      adjustScoreForm.resetFields()
      setSelectedTeam(null)
      queryClient.invalidateQueries({ queryKey: ['admin-teams'] })
    },
    onError: (err: any) => message.error(err.response?.data?.message || '调整失败'),
  })

  const handleImportFile = (file: File) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      try {
        const text = e.target?.result as string
        const lines = text.split('\n').filter(line => line.trim())
        const parsed: Array<{ name: string; description?: string }> = []

        // 支持CSV格式: name,description
        // 或纯文本格式: 每行一个队伍名
        lines.forEach((line, idx) => {
          if (idx === 0 && line.includes('name') && line.includes('description')) {
            // 跳过CSV头
            return
          }
          const parts = line.split(',').map(p => p.trim())
          if (parts[0]) {
            parsed.push({
              name: parts[0],
              description: parts[1] || undefined,
            })
          }
        })

        setImportPreview(parsed)
        message.info(`已解析 ${parsed.length} 个队伍`)
      } catch (error) {
        message.error('文件解析失败')
      }
    }
    reader.readAsText(file)
    return false // 阻止自动上传
  }

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
    { title: '分数', dataIndex: 'score', width: 100, render: (s: number) => <Tag color="blue">{s || 0}</Tag> },
    { title: '创建时间', dataIndex: 'created_at', width: 160, render: (t: string) => formatTime(t) },
    {
      title: '操作', width: 280, fixed: 'right',
      render: (_, r) => (
        <Space size="small" wrap>
          <Button size="small" icon={<EditOutlined />} onClick={() => { setSelectedTeam(r); editForm.setFieldsValue(r); setEditOpen(true) }}>编辑</Button>
          <Button size="small" icon={<DollarOutlined />} onClick={() => { setSelectedTeam(r); setAdjustScoreOpen(true) }}>调分</Button>
          <Popconfirm title="重置队伍分数？" onConfirm={() => resetMutation.mutate(r.id)}>
            <Button size="small" icon={<ReloadOutlined />}>重置</Button>
          </Popconfirm>
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
        <Title level={3} style={{ margin: 0 }}>👥 裁判 - 队伍管理</Title>
        <Space>
          <Button icon={<UploadOutlined />} onClick={() => setImportOpen(true)}>批量导入</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>创建队伍</Button>
        </Space>
      </div>

      <Card>
        <Text type="secondary" style={{ marginBottom: 16, display: 'block' }}>
          📋 仅裁判和管理员可访问。支持批量导入队伍。所有操作将被记录。
        </Text>
        {isLoading ? <div className="flex items-center justify-center h-64" /> : (
          <Table columns={columns} dataSource={teams} rowKey="id" scroll={{ x: 1000 }} />
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
        </Form>
      </Modal>

      {/* Batch Import Modal */}
      <Modal title="批量导入队伍" open={importOpen} onCancel={() => { setImportOpen(false); setImportPreview([]); importForm.resetFields() }}
        onOk={() => importMutation.mutate({})} confirmLoading={importMutation.isPending} width={700}>
        <div className="space-y-4">
          <Upload.Dragger accept=".txt,.csv" beforeUpload={handleImportFile} showUploadList={false}>
            <p className="ant-upload-drag-icon"><UploadOutlined /></p>
            <p className="ant-upload-text">点击或拖拽文件到此区域</p>
            <p className="ant-upload-hint">
              支持格式：<br />
              • CSV: name,description（第一行为表头）<br />
              • TXT: 每行一个队伍名称
            </p>
          </Upload.Dragger>

          {importPreview.length > 0 && (
            <>
              <Divider>预览 ({importPreview.length} 个队伍)</Divider>
              <div className="max-h-64 overflow-auto space-y-2">
                {importPreview.map((team, idx) => (
                  <Tag key={idx} color="blue">{team.name} {team.description && `- ${team.description}`}</Tag>
                ))}
              </div>
            </>
          )}
        </div>
      </Modal>

      {/* Adjust Score Modal */}
      <Modal title={`调整分数 - ${selectedTeam?.name || ''}`} open={adjustScoreOpen}
        onCancel={() => { setAdjustScoreOpen(false); adjustScoreForm.resetFields(); setSelectedTeam(null) }}
        onOk={() => adjustScoreForm.submit()} confirmLoading={adjustScoreMutation.isPending}>
        <Form form={adjustScoreForm} layout="vertical" onFinish={(values) => {
          adjustScoreMutation.mutate({
            ...values,
            team_id: selectedTeam?.id,
          })
        }}>
          <Form.Item name="game_id" label="比赛 ID" rules={[{ required: true, message: '请输入比赛ID' }]}>
            <InputNumber style={{ width: '100%' }} placeholder="输入比赛ID" />
          </Form.Item>
          <Form.Item name="amount" label="调整分数" rules={[{ required: true, message: '请输入调整分数' }]}
            extra="正数为加分，负数为减分">
            <InputNumber style={{ width: '100%' }} placeholder="例如: 100 或 -50" />
          </Form.Item>
          <Form.Item name="reason" label="原因" rules={[{ required: true, message: '请输入调整原因' }]}>
            <Input.TextArea rows={3} placeholder="说明调整分数的原因" />
          </Form.Item>
          {selectedTeam && (
            <Text type="secondary">当前分数: {selectedTeam.score || 0}</Text>
          )}
        </Form>
      </Modal>
    </div>
  )
}
