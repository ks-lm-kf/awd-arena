import { useState } from 'react'
import { useNavigate, Link } from 'react-router'
import { Form, Input, Button, Card, Typography, message } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useAuth } from '@/hooks/useAuth'

const { Title, Text } = Typography

export default function LoginPage() {
  const navigate = useNavigate()
  const { login } = useAuth()
  const [loading, setLoading] = useState(false)

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true)
    try {
      await login(values.username, values.password)
      message.success('登录成功！')
    } catch (err: any) {
      message.error(err?.response?.data?.message || '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-950">
      <Card className="w-[400px] shadow-2xl" style={{ background: '#1a1a2e', borderColor: '#2a2a4a' }}>
        <div className="text-center mb-6">
          <span className="text-5xl">🎯</span>
          <Title level={3} className="mt-3 mb-1" style={{ color: '#e2e8f0' }}>AWD Arena</Title>
          <Text type="secondary">Attack With Defense 竞技平台</Text>
        </div>
        <Form layout="vertical" onFinish={onFinish} autoComplete="off" size="large">
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block className="bg-indigo-500">
              登录
            </Button>
          </Form.Item>
        </Form>
        <div className="text-center">
          <Text type="secondary">还没有账号？<Link to="/register" className="text-indigo-400">去注册</Link></Text>
        </div>
      </Card>
    </div>
  )
}

