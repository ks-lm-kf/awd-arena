import { useState, useEffect } from 'react'
import { Card, Table, Button, Space, Tag, Modal, Form, Input, InputNumber, Select, message, Spin, Alert, Popconfirm } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons'
import { templateApi, type ChallengeTemplate, type CreateTemplateRequest } from '@/api/template'

export default function TemplatesPage() {
  const [data, setData] = useState<ChallengeTemplate[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<ChallengeTemplate | null>(null)
  const [form] = Form.useForm()
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 })

  const fetchData = async (page: number = 1, pageSize: number = 20) => {
    setLoading(true)
    setError(null)
    try {
      const response = await templateApi.list(page, pageSize)
      setData(response.templates)
      setPagination(prev => ({ ...prev, current: page, pageSize, total: response.total }))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch templates')
      message.error('获取题目模板失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [])

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: '题目名称', dataIndex: 'name', key: 'name' },
    { title: '分类', dataIndex: 'category', key: 'category' },
    { 
      title: '难度', 
      dataIndex: 'difficulty', 
      key: 'difficulty',
      render: (diff: string) => {
        const colors: Record<string, string> = { easy: 'green', medium: 'orange', hard: 'red' }
        const labels: Record<string, string> = { easy: '简单', medium: '中等', hard: '困难' }
        return <Tag color={colors[diff] || 'default'}>{labels[diff] || diff}</Tag>
      }
    },
    { title: '分值', dataIndex: 'points', key: 'points', width: 100 },
    { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
    { 
      title: '状态', 
      dataIndex: 'is_active', 
      key: 'is_active',
      render: (active: boolean) => (
        <Tag color={active ? 'green' : 'default'}>{active ? '启用' : '禁用'}</Tag>
      )
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: ChallengeTemplate) => (
        <Space>
          <Button 
            type="link" 
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除"
            description="确定要删除这个题目模板吗?"
            onConfirm={() => handleDelete(record.id)}
            okText="是"
            cancelText="否"
          >
            <Button type="link" danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      )
    },
  ]

  const handleEdit = (template: ChallengeTemplate) => {
    setEditingTemplate(template)
    form.setFieldsValue(template)
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await templateApi.delete(id)
      message.success('删除成功')
      fetchData(pagination.current, pagination.pageSize)
    } catch (err) {
      message.error('删除失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const requestData: CreateTemplateRequest = {
        name: values.name,
        description: values.description || '',
        category: values.category,
        difficulty: values.difficulty,
        points: values.points,
        docker_image: values.docker_image || 'ubuntu:latest',
        internal_port: values.internal_port || 8080,
        flag: values.flag || `flag{${Date.now()}}`,
        is_active: values.is_active ?? true
      }

      if (editingTemplate) {
        await templateApi.update(editingTemplate.id, requestData)
        message.success('更新成功')
      } else {
        await templateApi.create(requestData)
        message.success('创建成功')
      }
      
      setModalVisible(false)
      form.resetFields()
      setEditingTemplate(null)
      fetchData(pagination.current, pagination.pageSize)
    } catch (err) {
      message.error('操作失败，请检查表单')
    }
  }

  const handleTableChange = (pag: any) => {
    fetchData(pag.current, pag.pageSize)
  }

  return (
    <div className="p-6">
      <Card 
        title="题目模板" 
        extra={
          <Space>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={() => fetchData(pagination.current, pagination.pageSize)}
            >
              刷新
            </Button>
            <Button 
              type="primary" 
              icon={<PlusOutlined />} 
              onClick={() => {
                setEditingTemplate(null)
                form.resetFields()
                setModalVisible(true)
              }}
            >
              新建题目
            </Button>
          </Space>
        }
      >
        {error && (
          <Alert 
            message="错误" 
            description={error} 
            type="error" 
            closable 
            style={{ marginBottom: 16 }}
            onClose={() => setError(null)}
          />
        )}
        
        <Spin spinning={loading}>
          <Table 
            columns={columns} 
            dataSource={data} 
            rowKey="id" 
            pagination={{
              current: pagination.current,
              pageSize: pagination.pageSize,
              total: pagination.total,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 条记录`
            }}
            onChange={handleTableChange}
          />
        </Spin>
      </Card>
      
      <Modal
        title={editingTemplate ? '编辑题目' : '新建题目'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => {
          setModalVisible(false)
          form.resetFields()
          setEditingTemplate(null)
        }}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="题目名称" rules={[{ required: true, message: '请输入题目名称' }]}>
            <Input placeholder="例如: Web登录注入" />
          </Form.Item>
          
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} placeholder="题目描述" />
          </Form.Item>
          
          <Space size="large" style={{ width: '100%' }}>
            <Form.Item name="category" label="分类" rules={[{ required: true }]} style={{ width: 200 }}>
              <Select placeholder="选择分类">
                <Select.Option value="Web">Web</Select.Option>
                <Select.Option value="Pwn">Pwn</Select.Option>
                <Select.Option value="Crypto">Crypto</Select.Option>
                <Select.Option value="Reverse">Reverse</Select.Option>
                <Select.Option value="Misc">Misc</Select.Option>
              </Select>
            </Form.Item>
            
            <Form.Item name="difficulty" label="难度" rules={[{ required: true }]} style={{ width: 200 }}>
              <Select placeholder="选择难度">
                <Select.Option value="easy">简单</Select.Option>
                <Select.Option value="medium">中等</Select.Option>
                <Select.Option value="hard">困难</Select.Option>
              </Select>
            </Form.Item>
          </Space>
          
          <Space size="large" style={{ width: '100%' }}>
            <Form.Item name="points" label="分值" rules={[{ required: true }]} style={{ width: 200 }}>
              <InputNumber min={0} placeholder="100" style={{ width: '100%' }} />
            </Form.Item>
            
            <Form.Item name="is_active" label="状态" initialValue={true} style={{ width: 200 }}>
              <Select>
                <Select.Option value={true}>启用</Select.Option>
                <Select.Option value={false}>禁用</Select.Option>
              </Select>
            </Form.Item>
          </Space>
          
          <Form.Item name="docker_image" label="Docker镜像">
            <Input placeholder="例如: ubuntu:latest (可选)" />
          </Form.Item>
          
          <Form.Item name="internal_port" label="内部端口">
            <InputNumber min={1} max={65535} placeholder="8080" style={{ width: '100%' }} />
          </Form.Item>
          
          <Form.Item name="flag" label="Flag">
            <Input placeholder="例如: flag{xxx} (可选，留空自动生成)" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
