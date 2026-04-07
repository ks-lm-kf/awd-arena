import { useState } from 'react'
import {
  Card, Table, Button, Space, Modal, Form, Input, Select, InputNumber, Tag,
  Typography, message, Popconfirm, Tooltip, Drawer, Spin, Empty, Row, Col,
  Tabs, Alert
} from 'antd'
import { 
  PlusOutlined, EditOutlined, DeleteOutlined, CloudDownloadOutlined, 
  CloudServerOutlined, SearchOutlined, CloudUploadOutlined, 
  BuildOutlined, InfoCircleOutlined, ExclamationCircleOutlined 
} from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { ColumnsType } from 'antd/es/table'
import { 
  dockerImageApi, type DockerImage, type DockerImageParams, type HostImage,
  type PullImageParams, type BuildImageParams
} from '@/api/dockerImage'
import { formatBytes, difficultyColor, difficultyLabel } from '@/utils/format'

const { Title, Text } = Typography
const { Option } = Select

export default function DockerImages() {
  const queryClient = useQueryClient()
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<DockerImage | null>(null)
  const [form] = Form.useForm()
  const [filter, setFilter] = useState<{ category?: string; difficulty?: string; search?: string }>({})
  const [page, setPage] = useState(1)
  const [hostDrawer, setHostDrawer] = useState(false)
  const [pullModalOpen, setPullModalOpen] = useState(false)
  const [buildModalOpen, setBuildModalOpen] = useState(false)
  const [pullForm] = Form.useForm()
  const [buildForm] = Form.useForm()

  const { data, isLoading } = useQuery({
    queryKey: ['docker-images', filter, page],
    queryFn: () => dockerImageApi.list({ ...filter, page, page_size: 10 }),
  })

  const { data: hostImages, isLoading: hostLoading, refetch: refetchHostImages } = useQuery({
    queryKey: ['host-images'],
    queryFn: () => dockerImageApi.hostList(),
    enabled: hostDrawer,
  })

  const createMutation = useMutation({
    mutationFn: (values: DockerImageParams) => dockerImageApi.create(values),
    onSuccess: () => { message.success('创建成功'); closeModal(); queryClient.invalidateQueries({ queryKey: ['docker-images'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '创建失败'),
  })

  const updateMutation = useMutation({
    mutationFn: (values: DockerImageParams) => dockerImageApi.update(editing!.id, values),
    onSuccess: () => { message.success('更新成功'); closeModal(); queryClient.invalidateQueries({ queryKey: ['docker-images'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '更新失败'),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => dockerImageApi.delete(id),
    onSuccess: () => { message.success('删除成功'); queryClient.invalidateQueries({ queryKey: ['docker-images'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '删除失败'),
  })

  const pullMutation = useMutation({
    mutationFn: (id: number) => dockerImageApi.pull(id),
    onSuccess: () => { message.success('镜像拉取任务已提交') },
    onError: (err: any) => message.error(err.response?.data?.message || '拉取失败'),
  })

  const pullImageMutation = useMutation({
    mutationFn: (params: PullImageParams) => dockerImageApi.pullImage(params),
    onSuccess: () => { 
      message.success('镜像拉取成功'); 
      setPullModalOpen(false); 
      pullForm.resetFields();
      refetchHostImages();
    },
    onError: (err: any) => message.error(err.response?.data?.message || '拉取失败'),
  })

  const buildImageMutation = useMutation({
    mutationFn: (params: BuildImageParams) => dockerImageApi.buildImage(params),
    onSuccess: () => { 
      message.success('镜像构建成功'); 
      setBuildModalOpen(false); 
      buildForm.resetFields();
      refetchHostImages();
    },
    onError: (err: any) => message.error(err.response?.data?.message || '构建失败'),
  })

  const removeFromHostMutation = useMutation({
    mutationFn: ({ idOrName, force }: { idOrName: string; force?: boolean }) => 
      dockerImageApi.removeFromHost(idOrName, force),
    onSuccess: () => { 
      message.success('镜像已从主机删除'); 
      refetchHostImages();
    },
    onError: (err: any) => {
      console.error('Remove from host error:', err);
      const errMsg = err.response?.data?.message || err.message || '删除失败';
      message.error(`删除失败: ${errMsg}`);
    },
  })

  const removeCompletelyMutation = useMutation({
    mutationFn: ({ id, force }: { id: number; force?: boolean }) => 
      dockerImageApi.removeFromDBAndHost(id, force),
    onSuccess: () => { 
      message.success('镜像已完全删除'); 
      queryClient.invalidateQueries({ queryKey: ['docker-images'] });
    },
    onError: (err: any) => {
      console.error('Remove completely error:', err);
      const errMsg = err.response?.data?.message || err.message || '删除失败';
      message.error(`完全删除失败: ${errMsg}`);
    },
  })

  const closeModal = () => { setModalOpen(false); setEditing(null); form.resetFields() }

  const openEdit = (record: DockerImage) => {
    setEditing(record)
    form.setFieldsValue(record)
    setModalOpen(true)
  }

  const columns: ColumnsType<DockerImage> = [
    { title: 'ID', dataIndex: 'id', width: 60, fixed: 'left' },
    { title: '名称', dataIndex: 'name', width: 180, ellipsis: true, fixed: 'left' },
    { title: '分类', dataIndex: 'category', width: 90, render: (v: string) => <Tag>{v}</Tag> },
    { title: '难度', dataIndex: 'difficulty', width: 90, render: (v: string) => <Tag color={difficultyColor(v)}>{difficultyLabel(v)}</Tag> },
    { title: '镜像', width: 220, ellipsis: true, render: (_, r) => <code className="text-xs text-gray-400">{r.image_id}:{r.tag}</code> },
    { title: '端口', dataIndex: 'ports', width: 100 },
    { title: '创建时间', dataIndex: 'created_at', width: 160, render: (t: string) => t ? new Date(t).toLocaleString() : '-' },
    {
      title: '操作', width: 280, fixed: 'right',
      render: (_, r) => (
        <Space size="small">
          <Tooltip title="编辑">
            <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)} />
          </Tooltip>
          <Tooltip title="拉取">
            <Button size="small" icon={<CloudDownloadOutlined />} loading={pullMutation.isPending} onClick={() => pullMutation.mutate(r.id)} />
          </Tooltip>
          <Popconfirm 
            title="完全删除" 
            description="此操作将从数据库和主机同时删除镜像，确定继续？" 
            onConfirm={() => removeCompletelyMutation.mutate({ id: r.id, force: true })}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="完全删除">
              <Button size="small" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
          <Popconfirm 
            title="仅删除记录" 
            description="仅从数据库删除记录，保留主机镜像" 
            onConfirm={() => deleteMutation.mutate(r.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="仅删除数据库记录">
              <Button size="small" icon={<DeleteOutlined />} style={{ color: '#faad14' }} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const hostImageColumns: ColumnsType<HostImage> = [
    { title: '仓库', dataIndex: 'repository', width: 200, ellipsis: true },
    { title: '标签', dataIndex: 'tag', width: 100 },
    { title: '镜像ID', dataIndex: 'image_id', width: 150, render: (id: string) => <code>{id?.slice(0, 12)}</code> },
    { title: '大小', dataIndex: 'size', width: 100 },
    { title: '创建时间', dataIndex: 'created_at', width: 150 },
    {
      title: '操作', width: 150, fixed: 'right',
      render: (_, r) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button size="small" icon={<InfoCircleOutlined />} onClick={() => {
              Modal.info({
                title: '镜像详情',
                content: (
                  <div>
                    <p><strong>仓库:</strong> {r.repository}</p>
                    <p><strong>标签:</strong> {r.tag}</p>
                    <p><strong>镜像ID:</strong> <code>{r.image_id}</code></p>
                    <p><strong>大小:</strong> {r.size}</p>
                    <p><strong>创建时间:</strong> {r.created_at}</p>
                  </div>
                ),
              });
            }} />
          </Tooltip>
          <Popconfirm 
            title="删除镜像" 
            description={`确定要删除 ${r.repository}:${r.tag} 吗？`}
            onConfirm={() => removeFromHostMutation.mutate({ idOrName: r.image_id, force: false })}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button size="small" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const handleSubmit = (values: DockerImageParams) => {
    if (editing) {
      updateMutation.mutate(values)
    } else {
      createMutation.mutate(values)
    }
  }

  const handlePullImage = (values: PullImageParams) => {
    pullImageMutation.mutate(values)
  }

  const handleBuildImage = (values: BuildImageParams) => {
    buildImageMutation.mutate(values)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Title level={3} style={{ margin: 0 }}>Docker 镜像管理</Title>
        <Space>
          <Button icon={<CloudServerOutlined />} onClick={() => setHostDrawer(true)}>主机镜像</Button>
          <Button icon={<BuildOutlined />} onClick={() => setBuildModalOpen(true)}>构建镜像</Button>
          <Button icon={<CloudDownloadOutlined />} onClick={() => setPullModalOpen(true)}>拉取镜像</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>添加镜像</Button>
        </Space>
      </div>

      <Card>
        <Space style={{ marginBottom: 16 }} wrap>
          <Input placeholder="搜索镜像名称" prefix={<SearchOutlined />} style={{ width: 200 }} allowClear
            onChange={(e) => { setFilter((f) => ({ ...f, search: e.target.value || undefined })); setPage(1) }} />
          <Select placeholder="分类" allowClear style={{ width: 120 }}
            onChange={(v) => { setFilter((f) => ({ ...f, category: v })); setPage(1) }}
            options={[{ value: 'web', label: 'Web' }, { value: 'pwn', label: 'Pwn' }, { value: 're', label: 'Re' }, { value: 'crypto', label: 'Crypto' }, { value: 'general', label: 'General' }]} />
          <Select placeholder="难度" allowClear style={{ width: 120 }}
            onChange={(v) => { setFilter((f) => ({ ...f, difficulty: v })); setPage(1) }}
            options={[{ value: 'easy', label: '简单' }, { value: 'medium', label: '中等' }, { value: 'hard', label: '困难' }]} />
        </Space>
        <Table
          columns={columns} 
          dataSource={data?.items || []} 
          rowKey="id"
          loading={isLoading} 
          pagination={{ current: page, pageSize: 10, total: data?.total || 0, onChange: setPage }}
          scroll={{ x: 1400 }}
        />
      </Card>

      {/* 编辑/添加镜像模态框 */}
      <Modal
        title={editing ? '编辑镜像' : '添加镜像'} 
        open={modalOpen} 
        onCancel={closeModal}
        onOk={() => form.submit()} 
        confirmLoading={createMutation.isPending || updateMutation.isPending} 
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input placeholder="镜像名称" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="category" label="分类" rules={[{ required: true }]}>
                <Select options={[{ value: 'web', label: 'Web' }, { value: 'pwn', label: 'Pwn' }, { value: 're', label: 'Re' }, { value: 'crypto', label: 'Crypto' }, { value: 'general', label: 'General' }]} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="difficulty" label="难度" rules={[{ required: true }]}>
                <Select options={[{ value: 'easy', label: '简单' }, { value: 'medium', label: '中等' }, { value: 'hard', label: '困难' }]} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="image_id" label="镜像名" rules={[{ required: true }]}>
            <Input placeholder="registry.example.com/image-name" />
          </Form.Item>
          <Form.Item name="tag" label="标签" rules={[{ required: true }]}>
            <Input placeholder="latest" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Row gutter={16}>
            <Col span={8}>
              <Form.Item name="ports" label="暴露端口">
                <Input placeholder="例: 80,8080" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="cpu_limit" label="CPU 限制">
                <InputNumber min={0.1} step={0.1} style={{ width: '100%' }} placeholder="0.5" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="memory_limit" label="内存限制(MB)">
                <InputNumber min={64} step={64} style={{ width: '100%' }} placeholder="256" />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>

      {/* 拉取镜像模态框 */}
      <Modal
        title="拉取镜像"
        open={pullModalOpen}
        onCancel={() => { setPullModalOpen(false); pullForm.resetFields(); }}
        onOk={() => pullForm.submit()}
        confirmLoading={pullImageMutation.isPending}
      >
        <Alert 
          message="从 Registry 拉取镜像" 
          description="输入镜像名称从 Docker Registry 拉取镜像到主机"
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <Form form={pullForm} layout="vertical" onFinish={handlePullImage}>
          <Form.Item name="name" label="镜像名称" rules={[{ required: true }]} help="例如: nginx, ubuntu:20.04">
            <Input placeholder="nginx 或 ubuntu:20.04" />
          </Form.Item>
          <Form.Item name="tag" label="标签（可选）" help="不填写则默认为 latest">
            <Input placeholder="latest" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 构建镜像模态框 */}
      <Modal
        title="构建镜像"
        open={buildModalOpen}
        onCancel={() => { setBuildModalOpen(false); buildForm.resetFields(); }}
        onOk={() => buildForm.submit()}
        confirmLoading={buildImageMutation.isPending}
        width={700}
      >
        <Alert 
          message="从 Dockerfile 构建镜像" 
          description="使用 Dockerfile 构建新的镜像"
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <Form form={buildForm} layout="vertical" onFinish={handleBuildImage}>
          <Form.Item 
            name="tags" 
            label="镜像标签" 
            rules={[{ required: true }]}
            help="例如: myapp:latest, registry.example.com/myapp:v1.0"
          >
            <Select mode="tags" placeholder="输入镜像标签，按回车添加" />
          </Form.Item>
          <Form.Item 
            name="context_path" 
            label="构建上下文路径"
            help="Dockerfile 所在目录路径，默认为当前目录"
          >
            <Input placeholder="/path/to/context" />
          </Form.Item>
          <Form.Item 
            name="dockerfile" 
            label="Dockerfile 文件名"
            help="默认为 Dockerfile"
          >
            <Input placeholder="Dockerfile" />
          </Form.Item>
          <Form.Item 
            name="build_args" 
            label="构建参数"
            help="格式: KEY=value，每行一个"
          >
            <Input.TextArea 
              rows={3} 
              placeholder="VERSION=1.0&#10;ENV=production" 
            />
          </Form.Item>
          <Form.Item 
            name="no_cache" 
            label="禁用缓存"
          >
            <Select options={[
              { value: false, label: '使用缓存（推荐）' },
              { value: true, label: '禁用缓存' }
            ]} />
          </Form.Item>
        </Form>
      </Modal>

      {/* 主机镜像抽屉 */}
      <Drawer 
        title="主机上的镜像" 
        open={hostDrawer} 
        onClose={() => setHostDrawer(false)} 
        width={900}
        extra={
          <Button icon={<CloudDownloadOutlined />} onClick={() => refetchHostImages()}>
            刷新
          </Button>
        }
      >
        {hostLoading ? (
          <div className="flex justify-center py-8"><Spin /></div>
        ) : !hostImages?.length ? (
          <Empty description="主机上没有镜像" />
        ) : (
          <Table
            columns={hostImageColumns}
            dataSource={hostImages}
            rowKey={(r) => r.image_id}
            pagination={{ pageSize: 10 }}
            scroll={{ x: 900 }}
          />
        )}
      </Drawer>
    </div>
  )
}

