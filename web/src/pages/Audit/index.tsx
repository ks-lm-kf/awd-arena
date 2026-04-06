import { useState, useEffect } from 'react'
import { Card, Table, Tag, Space, DatePicker, Button, Input, Select, message, Spin, Alert } from 'antd'
import { SearchOutlined, DownloadOutlined, ReloadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { auditApi, type AuditLog } from '@/api/audit'

export default function AuditPage() {
  const [data, setData] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 })

  const fetchData = async (page: number = 1, pageSize: number = 20) => {
    setLoading(true)
    setError(null)
    try {
      const response = await auditApi.getLogs(page, pageSize)
      setData(response.items)
      setPagination(prev => ({ ...prev, current: page, pageSize, total: response.total }))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch audit logs')
      message.error('获取审计日志失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [])

  const columns: ColumnsType<AuditLog> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: '时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
    { title: '用户', dataIndex: 'username', key: 'username', width: 120 },
    { title: '操作', dataIndex: 'action', key: 'action', width: 150 },
    { title: '目标类型', dataIndex: 'resource_type', key: 'resource_type', width: 120 },
    { title: '目标ID', dataIndex: 'resource_id', key: 'resource_id', width: 100 },
    { title: '详情', dataIndex: 'details', key: 'details', ellipsis: true },
    { title: 'IP地址', dataIndex: 'ip_address', key: 'ip_address', width: 150 },
  ]

  const handleTableChange = (pag: any) => {
    fetchData(pag.current, pag.pageSize)
  }

  return (
    <div className="p-6">
      <Card 
        title="审计日志" 
        extra={
          <Space>
            <DatePicker.RangePicker />
            <Select defaultValue="all" style={{ width: 120 }}>
              <Select.Option value="all">全部操作</Select.Option>
              <Select.Option value="login">登录</Select.Option>
              <Select.Option value="submit">提交</Select.Option>
            </Select>
            <Input placeholder="搜索..." prefix={<SearchOutlined />} style={{ width: 200 }} />
            <Button 
              icon={<ReloadOutlined />} 
              onClick={() => fetchData(pagination.current, pagination.pageSize)}
            >
              刷新
            </Button>
            <Button icon={<DownloadOutlined />}>导出</Button>
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
    </div>
  )
}
