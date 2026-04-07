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
      message.error('Two passwords do not match')
      return
    }

    setLoading(true)
    try {
      await authApi.changePassword({
        old_password: values.old_password,
        new_password: values.new_password,
      })
      message.success('Password changed successfully')
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      useAuthStore.getState().logout()
      navigate('/login')
    } catch (error: any) {
      message.error(error.response?.data?.message || 'Failed to change password')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-950 px-4">
      <Card className="w-full max-w-md" title="Change Password">
        <p className="mb-4 text-gray-400">First time login requires password change</p>
        <Form layout="vertical" onFinish={handleSubmit}>
          <Form.Item label="Old Password" name="old_password" rules={[{ required: true }]}>
            <Input.Password prefix={<LockOutlined />} size="large" />
          </Form.Item>
          <Form.Item 
            label="New Password" 
            name="new_password" 
            rules={[
              { required: true },
              { min: 8, message: 'At least 8 characters' },
              { pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).+$/, message: 'Must contain uppercase, lowercase and number' }
            ]}
          >
            <Input.Password prefix={<LockOutlined />} size="large" />
          </Form.Item>
          <Form.Item label="Confirm Password" name="confirm_password" rules={[{ required: true }]}>
            <Input.Password prefix={<LockOutlined />} size="large" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} size="large" block>
              Change Password
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}

