import { useState, useEffect } from 'react'
import { Card, Form, Input, Button, Typography, Descriptions, Tag, message } from 'antd'
import { UserOutlined, LockOutlined, TeamOutlined } from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import { authApi } from '@/api/auth'
import { teamApi } from '@/api/team'
import type { Team } from '@/types'

const { Title } = Typography

export default function ProfilePage() {
  const user = useAuthStore((s) => s.user)
  const [loading, setLoading] = useState(false)
  const [teamInfo, setTeamInfo] = useState<Team | null>(null)
  const [passwordForm] = Form.useForm()

  useEffect(() => {
    if (user?.team_id) {
      teamApi.get(user.team_id).then(setTeamInfo).catch(() => {})
    }
  }, [user?.team_id])

  const changePassword = async (values: { old_password: string; new_password: string }) => {
    setLoading(true)
    try {
      await authApi.changePassword(values)
      message.success('密码修改成功')
      passwordForm.resetFields()
    } catch (err: any) {
      message.error(err?.response?.data?.message || '修改失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <Title level={3}>个人中心</Title>
      <Card title="基本信息" style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
        <Descriptions column={1}>
          <Descriptions.Item label="用户名">{user?.username}</Descriptions.Item>
          <Descriptions.Item label="角色">
            <Tag color={user?.role === 'admin' ? 'purple' : 'green'}>
              {user?.role === 'admin' ? '管理员' : '选手'}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="邮箱">{user?.email || '-'}</Descriptions.Item>
          <Descriptions.Item label="注册时间">{user?.created_at ? new Date(user.created_at).toLocaleString() : '-'}</Descriptions.Item>
        </Descriptions>
      </Card>
      {teamInfo && (
        <Card title="队伍信息" style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
          <Descriptions column={1}>
            <Descriptions.Item label="队伍名称">{teamInfo.name}</Descriptions.Item>
            <Descriptions.Item label="描述">{teamInfo.description || '-'}</Descriptions.Item>
            <Descriptions.Item label="队伍口令">
              <code className="bg-gray-800 px-2 py-1 rounded text-indigo-400">{teamInfo.token}</code>
            </Descriptions.Item>
            <Descriptions.Item label="成员数">{teamInfo.member_count || '-'}</Descriptions.Item>
          </Descriptions>
        </Card>
      )}
      {!teamInfo && user?.role === 'player' && (
        <Card style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
          <div className="text-center py-4 text-gray-400">
            <TeamOutlined style={{ fontSize: 32 }} />
            <p className="mt-2">你还未加入任何队伍</p>
          </div>
        </Card>
      )}
      <Card title="修改密码" style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
        <Form form={passwordForm} layout="vertical" onFinish={changePassword} style={{ maxWidth: 400 }}>
          <Form.Item name="old_password" rules={[{ required: true, message: '请输入当前密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="当前密码" />
          </Form.Item>
          <Form.Item name="new_password" rules={[{ required: true, message: '请输入新密码' }, { min: 6, message: '密码至少6个字符' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="新密码" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading}>修改密码</Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
