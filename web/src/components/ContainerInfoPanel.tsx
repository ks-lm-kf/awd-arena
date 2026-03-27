
import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Button, message, Spin } from 'antd';
import { CopyOutlined, ReloadOutlined } from '@ant-design/icons';
import { getMyContainers, type Container } from '@/api/container';

// Use Container from api/container.ts
type ContainerInfo = Container;

const ContainerInfoPanel: React.FC<{ gameId: number }> = ({ gameId }) => {
  const [containers, setContainers] = useState<ContainerInfo[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadContainers();
  }, [gameId]);

  const loadContainers = async () => {
    setLoading(true);
    try {
      const res = await getMyContainers(gameId);
      setContainers(res || []);
    } catch (error) {
      message.error('加载容器信息失败');
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    message.success('已复制到剪贴板');
  };

  const columns = [
    {
      title: '题目',
      dataIndex: 'challenge_name',
      key: 'challenge_name',
    },
    {
      title: 'IP地址',
      dataIndex: 'ip_address',
      key: 'ip_address',
      render: (ip: string) => (
        <span>
          {ip}
          <Button
            type="link"
            size="small"
            icon={<CopyOutlined />}
            onClick={() => copyToClipboard(ip)}
          />
        </span>
      ),
    },
    {
      title: 'SSH信息',
      key: 'ssh',
      render: (_: any, record: ContainerInfo) => (
        <div>
          <div>
            用户: {record.ssh_user}
            <Button
              type="link"
              size="small"
              icon={<CopyOutlined />}
              onClick={() => record.ssh_user && copyToClipboard(record.ssh_user)}
            />
          </div>
          <div>
            密码: {record.ssh_password}
            <Button
              type="link"
              size="small"
              icon={<CopyOutlined />}
              onClick={() => record.ssh_password && copyToClipboard(record.ssh_password)}
            />
          </div>
          <div>端口: {record.ssh_port || 22}</div>
        </div>
      ),
    },
    {
      title: 'SSH命令',
      key: 'command',
      render: (_: any, record: ContainerInfo) => (
        <code>
          ssh {record.ssh_user}@{record.ip_address} -p {record.ssh_port || 22}
          <Button
            type="link"
            size="small"
            icon={<CopyOutlined />}
            onClick={() => 
              copyToClipboard(`ssh ${record.ssh_user}@${record.ip_address} -p ${record.ssh_port || 22}`)
            }
          />
        </code>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'running' ? 'green' : 'red'}>
          {status}
        </Tag>
      ),
    },
  ];

  return (
    <Card
      title="我的容器"
      extra={
        <Button icon={<ReloadOutlined />} onClick={loadContainers}>
          刷新
        </Button>
      }
    >
      <Spin spinning={loading}>
        <Table
          dataSource={containers}
          columns={columns}
          rowKey="id"
          pagination={false}
        />
      </Spin>
    </Card>
  );
};

export default ContainerInfoPanel;




