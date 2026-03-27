import { useState } from 'react'
import { useParams, useNavigate } from 'react-router'
import {
  Card, Tabs, Table, Button, Space, Modal, Form, Input, Select, InputNumber,
  Tag, Typography, message, Popconfirm, Spin, Row, Col, Descriptions, Badge,
  Divider, Empty, AutoComplete, Tooltip
} from 'antd'
import {
  ArrowLeftOutlined, PlusOutlined, EditOutlined, DeleteOutlined,
  TeamOutlined, TrophyOutlined, AppstoreOutlined, HistoryOutlined,
  InfoCircleOutlined
} from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { ColumnsType } from 'antd/es/table'
import type { Game, Team, Challenge, GameStatus, GameMode, Difficulty } from '@/types'
import { gameApi } from '@/api/game'
import { adminApi } from '@/api/admin'
import { teamApi } from '@/api/team'
import { challengeApi } from '@/api/challenge'
import { dockerImageApi } from '@/api/dockerImage'
import { formatTime, statusLabel, statusColor, difficultyColor, difficultyLabel } from '@/utils/format'

const { Title, Text } = Typography
const { Option } = Select
const { TabPane } = Tabs

// Helper to safely parse exposed_ports (can be string or array)
const parseExposedPorts = (ports: any): { container: number; protocol: string }[] => {
  if (!ports) return []
  if (Array.isArray(ports)) return ports
  if (typeof ports === 'string') {
    try {
      const parsed = JSON.parse(ports)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
  }
  return []
}


const modeLabel: Record<GameMode, string> = { awd_score: 'AWD 经典', awd_mix: '攻防混合', koh: '山顶争夺' }

export default function GameDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const gameId = Number(id)

  // 状态管理
  const [addTeamModalOpen, setAddTeamModalOpen] = useState(false)
  const [addChallengeModalOpen, setAddChallengeModalOpen] = useState(false)
  const [editChallengeModalOpen, setEditChallengeModalOpen] = useState(false)
  const [selectedChallenge, setSelectedChallenge] = useState<Challenge | null>(null)
  const [teamSearch, setTeamSearch] = useState('')
  const [addTeamForm] = Form.useForm()
  const [addChallengeForm] = Form.useForm()
  const [editChallengeForm] = Form.useForm()

  // 获取比赛详情
  const { data: game, isLoading: gameLoading } = useQuery({
    queryKey: ['game', gameId],
    queryFn: () => gameApi.get(gameId),
    enabled: !!gameId,
  })

  // 获取比赛中的队伍
  const { data: gameTeams, isLoading: teamsLoading } = useQuery({
    queryKey: ['game-teams', gameId],
    queryFn: () => adminApi.getGameTeams(gameId),
    enabled: !!gameId,
  })

  // 获取所有队伍（用于添加）
  const { data: allTeams } = useQuery<Team[]>({
    queryKey: ['teams'],
    queryFn: () => teamApi.list(),
  })

  // 获取比赛的题目
  const { data: challenges, isLoading: challengesLoading } = useQuery({
    queryKey: ['game-challenges', gameId],
    queryFn: () => challengeApi.list(gameId),
    enabled: !!gameId,
  })

  // 获取 Docker 镜像列表
  const { data: dockerImages } = useQuery<{ items: any[]; total: number }>({
    queryKey: ['docker-images', { page: 1, page_size: 100 }],
    queryFn: () => dockerImageApi.list({ page: 1, page_size: 100 }),
  })

  // 添加队伍 Mutation
  const addTeamMutation = useMutation({
    mutationFn: (teamId: number) => adminApi.addTeamToGame(gameId, teamId),
    onSuccess: () => {
      message.success('队伍已添加')
      queryClient.invalidateQueries({ queryKey: ['game-teams', gameId] })
      setTeamSearch('')
    },
    onError: (err: any) => message.error(err.response?.data?.message || '添加失败'),
  })

  // 移除队伍 Mutation
  const removeTeamMutation = useMutation({
    mutationFn: (teamId: number) => adminApi.removeTeamFromGame(gameId, teamId),
    onSuccess: () => {
      message.success('队伍已移除')
      queryClient.invalidateQueries({ queryKey: ['game-teams', gameId] })
    },
    onError: (err: any) => message.error(err.response?.data?.message || '移除失败'),
  })

  // 添加题目 Mutation
  const addChallengeMutation = useMutation({
    mutationFn: (values: any) => challengeApi.create(gameId, values),
    onSuccess: () => {
      message.success('题目已添加')
      setAddChallengeModalOpen(false)
      addChallengeForm.resetFields()
      queryClient.invalidateQueries({ queryKey: ['game-challenges', gameId] })
    },
    onError: (err: any) => message.error(err.response?.data?.message || '添加失败'),
  })

  // 更新题目 Mutation
  const updateChallengeMutation = useMutation({
    mutationFn: (values: any) => challengeApi.update(gameId, selectedChallenge!.id, values),
    onSuccess: () => {
      message.success('题目已更新')
      setEditChallengeModalOpen(false)
      editChallengeForm.resetFields()
      setSelectedChallenge(null)
      queryClient.invalidateQueries({ queryKey: ['game-challenges', gameId] })
    },
    onError: (err: any) => message.error(err.response?.data?.message || '更新失败'),
  })

  // 删除题目 Mutation
  const deleteChallengeMutation = useMutation({
    mutationFn: (challengeId: number) => challengeApi.delete(gameId, challengeId),
    onSuccess: () => {
      message.success('题目已删除')
      queryClient.invalidateQueries({ queryKey: ['game-challenges', gameId] })
    },
    onError: (err: any) => message.error(err.response?.data?.message || '删除失败'),
  })

  // 过滤可选队伍（未添加到比赛的）
  const availableTeams = allTeams?.filter(
    (t) => !gameTeams?.some((gt: any) => gt.id === t.id)
  ).filter(
    (t) => !teamSearch || t.name.toLowerCase().includes(teamSearch.toLowerCase())
  ) || []

  // 队伍表格列定义
  const teamColumns: ColumnsType<any> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    {
      title: '队伍名称',
      dataIndex: 'name',
      render: (name: string) => (
        <span className="flex items-center gap-2">
          <span className="w-7 h-7 rounded-full bg-indigo-500/20 flex items-center justify-center text-xs font-bold text-indigo-400">
            {name?.[0] || '?'}
          </span>
          {name}
        </span>
      ),
    },
    { title: '描述', dataIndex: 'description', ellipsis: true },
    { title: '成员数', dataIndex: 'member_count', width: 80 },
    { title: '分数', dataIndex: 'score', width: 100, render: (s: number) => <Tag color="gold">{s ?? 0}</Tag> },
    {
      title: '操作',
      width: 100,
      render: (_, r) => (
        <Popconfirm
          title="确定移除此队伍？"
          onConfirm={() => removeTeamMutation.mutate(r.id)}
        >
          <Button size="small" danger icon={<DeleteOutlined />}>移除</Button>
        </Popconfirm>
      ),
    },
  ]

  // 题目表格列定义
  const challengeColumns: ColumnsType<Challenge> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '题目名称', dataIndex: 'name', ellipsis: true },
    {
      title: '难度',
      dataIndex: 'difficulty',
      width: 90,
      render: (d: Difficulty) => <Tag color={difficultyColor(d)}>{difficultyLabel(d)}</Tag>,
    },
    { title: '分数', dataIndex: 'base_score', width: 80 },
    {
      title: '镜像',
      width: 200,
      ellipsis: true,
      render: (_, r) => (
        <Tooltip title={`${r.image_name}:${r.image_tag}`}>
          <code className="text-xs text-gray-400">{r.image_name}:{r.image_tag}</code>
        </Tooltip>
      ),
    },
    { title: '端口', dataIndex: 'exposed_ports', width: 100, render: (ports: any[]) => Array.isArray(ports) ? ports.map(p => p.container).join(', ') || '-' : '-' },
    { title: 'CPU', dataIndex: 'cpu_limit', width: 80, render: (v: number) => v ? `${v}` : '-' },
    { title: '内存', dataIndex: 'mem_limit', width: 80, render: (v: number) => v ? `${v}MB` : '-' },
    {
      title: '操作',
      width: 140,
      render: (_, r) => (
        <Space size="small">
          <Button
            size="small"
            icon={<EditOutlined />}
            onClick={() => {
              setSelectedChallenge(r)
              editChallengeForm.setFieldsValue({
                ...r,
                image_name: `${r.image_name}:${r.image_tag}`,
                exposed_ports: Array.isArray(r.exposed_ports) ? r.exposed_ports.map(p => p.container).join(',') : '',
              })
              setEditChallengeModalOpen(true)
            }}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定删除此题目？"
            onConfirm={() => deleteChallengeMutation.mutate(r.id)}
          >
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  // 处理添加队伍
  const handleAddTeam = (teamId: number) => {
    addTeamMutation.mutate(teamId)
  }

  // 处理添加题目
  const handleAddChallenge = (values: any) => {
    const [imageName, imageTag] = values.image_name.split(':')
    addChallengeMutation.mutate({
      ...values,
      image_name: imageName,
      image_tag: imageTag || 'latest',
    })
  }

  // 处理更新题目
  const handleUpdateChallenge = (values: any) => {
    const [imageName, imageTag] = values.image_name.split(':')
    updateChallengeMutation.mutate({
      ...values,
      image_name: imageName,
      image_tag: imageTag || 'latest',
    })
  }

  if (gameLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Spin size="large" />
      </div>
    )
  }

  if (!game) {
    return (
      <div className="flex flex-col items-center justify-center h-64">
        <Empty description="比赛不存在" />
        <Button type="primary" onClick={() => navigate('/games')}>返回列表</Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* 头部 */}
      <div className="flex items-center gap-4">
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/games')}>
          返回
        </Button>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <Title level={3} style={{ margin: 0 }}>{game.title}</Title>
            <Tag color={statusColor(game.status)}>{statusLabel(game.status)}</Tag>
            <Tag>{modeLabel[game.mode]}</Tag>
          </div>
          <Text type="secondary">{game.description || '暂无描述'}</Text>
        </div>
      </div>

      {/* 比赛状态卡片 */}
      <Card size="small">
        <Row gutter={[24, 16]}>
          <Col span={4}>
            <StatisticItem label="当前轮次" value={`${game.current_round} / ${game.total_rounds}`} />
          </Col>
          <Col span={4}>
            <StatisticItem label="每轮时长" value={`${game.round_duration}秒`} />
          </Col>
          <Col span={4}>
            <StatisticItem label="休息时长" value={`${game.break_duration}秒`} />
          </Col>
          <Col span={4}>
            <StatisticItem label="开始时间" value={formatTime(game.start_time) || '未开始'} />
          </Col>
          <Col span={4}>
            <StatisticItem label="结束时间" value={formatTime(game.end_time) || '未结束'} />
          </Col>
          <Col span={4}>
            <StatisticItem label="Flag 格式" value={game.flag_format} />
          </Col>
        </Row>
      </Card>

      {/* Tab 内容 */}
      <Tabs defaultActiveKey="teams">
        <TabPane
          tab={<span><TeamOutlined /> 队伍管理</span>}
          key="teams"
        >
          <Card>
            <div className="flex items-center justify-between mb-4">
              <Space>
                <Input
                  placeholder="搜索队伍"
                  prefix={<InfoCircleOutlined />}
                  style={{ width: 200 }}
                  value={teamSearch}
                  onChange={(e) => setTeamSearch(e.target.value)}
                  allowClear
                />
              </Space>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => setAddTeamModalOpen(true)}
              >
                添加队伍
              </Button>
            </div>
            <Table
              columns={teamColumns}
              dataSource={gameTeams || []}
              rowKey="id"
              loading={teamsLoading}
              pagination={false}
              locale={{ emptyText: '暂无队伍' }}
            />
          </Card>
        </TabPane>

        <TabPane
          tab={<span><AppstoreOutlined /> 题目管理</span>}
          key="challenges"
        >
          <Card>
            <div className="flex items-center justify-between mb-4">
              <Text>管理比赛的题目/Challenge</Text>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => setAddChallengeModalOpen(true)}
              >
                添加题目
              </Button>
            </div>
            <Table
              columns={challengeColumns}
              dataSource={challenges || []}
              rowKey="id"
              loading={challengesLoading}
              pagination={false}
              scroll={{ x: 1000 }}
              locale={{ emptyText: '暂无题目' }}
            />
          </Card>
        </TabPane>

        <TabPane
          tab={<span><TrophyOutlined /> 比赛概览</span>}
          key="overview"
        >
          <Row gutter={16}>
            <Col span={12}>
              <Card title="比赛信息" size="small">
                <Descriptions column={1} size="small">
                  <Descriptions.Item label="比赛名称">{game.title}</Descriptions.Item>
                  <Descriptions.Item label="比赛模式">{modeLabel[game.mode]}</Descriptions.Item>
                  <Descriptions.Item label="比赛状态">
                    <Tag color={statusColor(game.status)}>{statusLabel(game.status)}</Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="当前轮次">{game.current_round} / {game.total_rounds}</Descriptions.Item>
                  <Descriptions.Item label="每轮时长">{game.round_duration} 秒</Descriptions.Item>
                  <Descriptions.Item label="休息时长">{game.break_duration} 秒</Descriptions.Item>
                  <Descriptions.Item label="Flag 格式">{game.flag_format}</Descriptions.Item>
                  
                  <Descriptions.Item label="攻击权重">{game.attack_weight}</Descriptions.Item>
                  <Descriptions.Item label="防守权重">{game.defense_weight}</Descriptions.Item>
                </Descriptions>
              </Card>
            </Col>
            <Col span={12}>
              <Card title="参赛统计" size="small">
                <Descriptions column={1} size="small">
                  <Descriptions.Item label="参赛队伍">{gameTeams?.length || 0} 支</Descriptions.Item>
                  <Descriptions.Item label="题目数量">{challenges?.length || 0} 个</Descriptions.Item>
                  <Descriptions.Item label="开始时间">{formatTime(game.start_time) || '未开始'}</Descriptions.Item>
                  <Descriptions.Item label="结束时间">{formatTime(game.end_time) || '未结束'}</Descriptions.Item>
                </Descriptions>
              </Card>
            </Col>
          </Row>
        </TabPane>

        <TabPane
          tab={<span><HistoryOutlined /> 轮次记录</span>}
          key="rounds"
        >
          <Card>
            <Empty description="轮次记录功能开发中..." />
          </Card>
        </TabPane>
      </Tabs>

      {/* 添加队伍模态框 */}
      <Modal
        title="添加队伍"
        open={addTeamModalOpen}
        onCancel={() => {
          setAddTeamModalOpen(false)
          setTeamSearch('')
        }}
        footer={null}
        width={500}
      >
        <div className="mb-4">
          <Input
            placeholder="搜索队伍名称"
            value={teamSearch}
            onChange={(e) => setTeamSearch(e.target.value)}
            allowClear
          />
        </div>
        {availableTeams.length === 0 ? (
          <Empty description="没有可添加的队伍" />
        ) : (
          <div className="max-h-64 overflow-y-auto space-y-2">
            {availableTeams.map((team) => (
              <div
                key={team.id}
                className="flex items-center justify-between p-2 rounded hover:bg-gray-800 cursor-pointer"
              >
                <div className="flex items-center gap-2">
                  <span className="w-8 h-8 rounded-full bg-indigo-500/20 flex items-center justify-center text-xs font-bold text-indigo-400">
                    {team.name?.[0] || '?'}
                  </span>
                  <div>
                    <div className="font-medium">{team.name}</div>
                    <div className="text-xs text-gray-400">{team.description || '暂无描述'}</div>
                  </div>
                </div>
                <Button
                  size="small"
                  type="primary"
                  onClick={() => handleAddTeam(team.id)}
                  loading={addTeamMutation.isPending}
                >
                  添加
                </Button>
              </div>
            ))}
          </div>
        )}
      </Modal>

      {/* 添加题目模态框 */}
      <Modal
        title="添加题目"
        open={addChallengeModalOpen}
        onCancel={() => {
          setAddChallengeModalOpen(false)
          addChallengeForm.resetFields()
        }}
        onOk={() => addChallengeForm.submit()}
        confirmLoading={addChallengeMutation.isPending}
        width={600}
      >
        <Form
          form={addChallengeForm}
          layout="vertical"
          onFinish={handleAddChallenge}
          initialValues={{ difficulty: 'easy', base_score: 100, image_tag: 'latest' }}
        >
          <Form.Item
            name="name"
            label="题目名称"
            rules={[{ required: true, message: '请输入题目名称' }]}
          >
            <Input placeholder="输入题目名称" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} placeholder="题目描述" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="image_name"
                label="Docker 镜像"
                rules={[{ required: true, message: '请选择镜像' }]}
              >
                <Select
                  placeholder="选择 Docker 镜像"
                  showSearch
                  optionFilterProp="label"
                  options={dockerImages?.items?.map((img: any) => ({
                    value: `${img.image_id}:${img.tag}`,
                    label: `${img.name} (${img.image_id}:${img.tag})`,
                  })) || []}
                />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item
                name="difficulty"
                label="难度"
                rules={[{ required: true }]}
              >
                <Select>
                  <Option value="easy">简单</Option>
                  <Option value="medium">中等</Option>
                  <Option value="hard">困难</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item
                name="base_score"
                label="基础分数"
                rules={[{ required: true }]}
              >
                <InputNumber min={1} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={8}>
              <Form.Item name="exposed_ports" label="暴露端口" help="多个端口用逗号分隔">
                <Input placeholder="80,8080" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="cpu_limit" label="CPU 限制（核）">
                <InputNumber min={0.1} step={0.1} style={{ width: '100%' }} placeholder="0.5" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="mem_limit" label="内存限制（MB）">
                <InputNumber min={64} step={64} style={{ width: '100%' }} placeholder="256" />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>

      {/* 编辑题目模态框 */}
      <Modal
        title="编辑题目"
        open={editChallengeModalOpen}
        onCancel={() => {
          setEditChallengeModalOpen(false)
          editChallengeForm.resetFields()
          setSelectedChallenge(null)
        }}
        onOk={() => editChallengeForm.submit()}
        confirmLoading={updateChallengeMutation.isPending}
        width={600}
      >
        <Form
          form={editChallengeForm}
          layout="vertical"
          onFinish={handleUpdateChallenge}
        >
          <Form.Item
            name="name"
            label="题目名称"
            rules={[{ required: true, message: '请输入题目名称' }]}
          >
            <Input placeholder="输入题目名称" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} placeholder="题目描述" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="image_name"
                label="Docker 镜像"
                rules={[{ required: true, message: '请选择镜像' }]}
              >
                <Select
                  placeholder="选择 Docker 镜像"
                  showSearch
                  optionFilterProp="label"
                  options={dockerImages?.items?.map((img: any) => ({
                    value: `${img.image_id}:${img.tag}`,
                    label: `${img.name} (${img.image_id}:${img.tag})`,
                  })) || []}
                />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item
                name="difficulty"
                label="难度"
                rules={[{ required: true }]}
              >
                <Select>
                  <Option value="easy">简单</Option>
                  <Option value="medium">中等</Option>
                  <Option value="hard">困难</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item
                name="base_score"
                label="基础分数"
                rules={[{ required: true }]}
              >
                <InputNumber min={1} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={8}>
              <Form.Item name="exposed_ports" label="暴露端口" help="多个端口用逗号分隔">
                <Input placeholder="80,8080" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="cpu_limit" label="CPU 限制（核）">
                <InputNumber min={0.1} step={0.1} style={{ width: '100%' }} placeholder="0.5" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="mem_limit" label="内存限制（MB）">
                <InputNumber min={64} step={64} style={{ width: '100%' }} placeholder="256" />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </div>
  )
}

// 统计项组件
function StatisticItem({ label, value }: { label: string; value: string | number }) {
  return (
    <div>
      <Text type="secondary" className="text-xs">{label}</Text>
      <div className="font-medium">{value}</div>
    </div>
  )
}
