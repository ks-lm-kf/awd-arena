import { useState } from 'react'
import { useNavigate, Link } from 'react-router'
import { Form, Input, Button, Card, Typography, message } from 'antd'
import { UserOutlined, LockOutlined, TeamOutlined } from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import { authApi } from '@/api/auth'

const { Title, Text } = Typography

export default function RegisterPage() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)

  const onFinish = async (values: { username: string; password: string; team_token?: string }) => {
    setLoading(true)
    try {
      const res = await authApi.register(values)
      useAuthStore.getState().setToken(res.token)
      useAuthStore.getState().setUser(res.user)
      message.success('注册成功！')
      navigate('/dashboard')
    } catch (err: any) {
      message.error(err?.response?.data?.message || '注册失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-950">
      <Card className="w-[400px] shadow-2xl" style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
        <div className="text-center mb-6">
          <span className="text-5xl">&#127919;</span>
          <Title level={3} className="mt-3 mb-1" style={{ color: '#e2e8f0' }}>注册 AWD Arena</Title>
          <Text type="secondary">加入比赛，挑战极限</Text>
        </div>
        <Form layout="vertical" onFinish={onFinish} autoComplete="off" size="large">
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }, { min: 3, message: '用户名至少3个字符' }]}>
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }, { min: 6, message: '密码至少6个字符' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          <Form.Item name="team_token" extra="可选：如果有队伍口令可以加入已有队伍">
            <Input prefix={<TeamOutlined />} placeholder="队伍口令（可选）" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block className="bg-indigo-500">
              注册
            </Button>
          </Form.Item>
        </Form>
        <div className="text-center">
          <Text type="secondary">已有账号？<Link to="/login" className="text-indigo-400">去登录</Link></Text>
        </div>
      </Card>
    </div>
  )
}
