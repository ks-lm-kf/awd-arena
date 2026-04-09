import { useState } from 'react'
import { Form, Input, Button, Card, message } from 'antd'
import { LockOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router'
import { authApi } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'

export default function ChangePasswordPage() {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  const handleSubmit = async (values: any) => {
    if (values.new_password !== values.confirm_password) {
      message.error('两次输入的密码不一致')
      return
    }

    setLoading(true)
    try {
      await authApi.changePassword({
        old_password: values.old_password,
        new_password: values.new_password,
      })
      message.success('密码修改成功')
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      useAuthStore.getState().logout()
      navigate('/login')
    } catch (error: any) {
      message.error(error.response?.data?.message || '密码修改失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-950 px-4">
      <Card className="w-full max-w-md" title="修改密码">
        <p className="mb-4 text-gray-400">首次登录需要修改密码</p>
        <Form layout="vertical" onFinish={handleSubmit}>
          <Form.Item label="旧密码" name="old_password" rules={[{ required: true, message: '请输入旧密码' }]}>
            <Input.Password prefix={<LockOutlined />} size="large" />
          </Form.Item>
          <Form.Item 
            label="新密码" 
            name="new_password" 
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 8, message: '密码至少8个字符' },
              { pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[!@#$%^&*()_+\-=\[\]{}|;:',.<>?/\\`~"]).+$/, message: '必须包含大小写字母、数字和特殊字符' }
            ]}
          >
            <Input.Password prefix={<LockOutlined />} size="large" />
          </Form.Item>
          <Form.Item label="确认密码" name="confirm_password" rules={[{ required: true, message: '请确认新密码' }]}>
            <Input.Password prefix={<LockOutlined />} size="large" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} size="large" block>
              修改密码
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
