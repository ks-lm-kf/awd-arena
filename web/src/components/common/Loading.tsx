import { Spin } from 'antd'
import { LoadingOutlined } from '@ant-design/icons'

export default function Loading({ tip = '加载中...' }: { tip?: string }) {
  return (
    <div className="flex items-center justify-center h-64">
      <Spin size="large" indicator={<LoadingOutlined style={{ fontSize: 32 }} />} tip={tip}>
        <div className="p-12" />
      </Spin>
    </div>
  )
}
