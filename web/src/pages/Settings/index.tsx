import { useState, useEffect } from 'react'
import { Card, Form, Input, Button, Divider, Typography, Space, message, Tabs, Descriptions, InputNumber, Switch } from 'antd'
import { SaveOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from '@/stores/authStore'
import { useAuth } from '@/hooks/useAuth'
import { post } from '@/api/client'
import { settingsApi } from '@/api/settings'

const { Title } = Typography

export default function Settings() {
  const user = useAuthStore((s) => s.user)
  const { logout } = useAuth()
  const queryClient = useQueryClient()
  const [passwordForm] = Form.useForm()
  const [settingsForm] = Form.useForm()

  const { data: settings, isLoading } = useQuery({
    queryKey: ['settings'],
    queryFn: () => settingsApi.get(),
  })

  useEffect(() => {
    if (settings) settingsForm.setFieldsValue(settings)
  }, [settings, settingsForm])

  const settingsMutation = useMutation({
    mutationFn: (values: any) => settingsApi.update(values),
    onSuccess: () => { message.success('设置已保存'); queryClient.invalidateQueries({ queryKey: ['settings'] }) },
    onError: (err: any) => message.error(err.response?.data?.message || '保存失败'),
  })

  const handlePasswordChange = async (values: { old_password: string; new_password: string; confirm_password: string }) => {
    if (values.new_password !== values.confirm_password) { message.error('两次密码不一致'); return }
    try {
      await post('/auth/password', values)
      message.success('密码修改成功')
      passwordForm.resetFields()
    } catch (err: any) { message.error(err.response?.data?.message || '修改失败') }
  }

  const tabItems = [
    {
      key: 'profile',
      label: '个人信息',
      children: (
        <Card loading={false}>
          <Descriptions bordered column={1}>
            <Descriptions.Item label="用户名">{user?.username || '-'}</Descriptions.Item>
            <Descriptions.Item label="邮箱">{user?.email || '-'}</Descriptions.Item>
            <Descriptions.Item label="角色"><span style={{ textTransform: 'uppercase' }}>{user?.role || '-'}</span></Descriptions.Item>
            <Descriptions.Item label="队伍">{user?.team_name || (user?.team_id ? `ID: ${user.team_id}` : '未加入')}</Descriptions.Item>
            <Descriptions.Item label="注册时间">{user?.created_at || '-'}</Descriptions.Item>
          </Descriptions>
          <Divider />
          <Button danger onClick={logout}>退出登录</Button>
        </Card>
      ),
    },
    {
      key: 'password',
      label: '修改密码',
      children: (
        <Card>
          <Form form={passwordForm} layout="vertical" onFinish={handlePasswordChange} style={{ maxWidth: 400 }}>
            <Form.Item name="old_password" label="当前密码" rules={[{ required: true }]}>
              <Input.Password />
            </Form.Item>
            <Form.Item name="new_password" label="新密码" rules={[{ required: true }, { min: 6, message: '密码至少6位' }]}>
              <Input.Password />
            </Form.Item>
            <Form.Item name="confirm_password" label="确认新密码" rules={[{ required: true }]}>
              <Input.Password />
            </Form.Item>
            <Button type="primary" icon={<SaveOutlined />} htmlType="submit">修改密码</Button>
          </Form>
        </Card>
      ),
    },
    {
      key: 'system',
      label: '系统设置',
      children: (
        <Card loading={isLoading}>
          <Form form={settingsForm} layout="vertical" onFinish={settingsMutation.mutate} style={{ maxWidth: 600 }}>
            <Title level={5}>基本设置</Title>
            <Form.Item name="site_name" label="平台名称">
              <Input placeholder="AWD Arena" />
            </Form.Item>
            <Form.Item name="announcement" label="系统公告">
              <Input.TextArea rows={3} placeholder="公告内容，支持 HTML" />
            </Form.Item>

            <Title level={5} style={{ marginTop: 24 }}>比赛参数</Title>
            <Form.Item name="flag_format" label="Flag 格式">
              <Input placeholder="flag{...}" />
            </Form.Item>
            <Space size="large">
              <Form.Item name="initial_score" label="初始分数">
                <InputNumber min={0} step={100} />
              </Form.Item>
              <Form.Item name="attack_weight" label="攻击权重">
                <InputNumber min={0} step={0.1} />
              </Form.Item>
              <Form.Item name="defense_weight" label="防守权重">
                <InputNumber min={0} step={0.1} />
              </Form.Item>
            </Space>
            <Space size="large">
              <Form.Item name="round_duration" label="每轮时长(秒)">
                <InputNumber min={30} step={60} />
              </Form.Item>
              <Form.Item name="break_duration" label="休息时长(秒)">
                <InputNumber min={0} step={30} />
              </Form.Item>
              <Form.Item name="max_team_size" label="每队最大人数">
                <InputNumber min={1} max={20} />
              </Form.Item>
            </Space>

            <Button type="primary" icon={<SaveOutlined />} htmlType="submit" loading={settingsMutation.isPending}>
              保存设置
            </Button>
          </Form>
        </Card>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <Title level={3} style={{ margin: 0 }}>⚙️ 系统设置</Title>
      <Tabs items={tabItems} />
    </div>
  )
}
