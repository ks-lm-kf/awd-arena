import { useLocation, useNavigate } from 'react-router'
import { Layout, Menu, Typography, Tag, Avatar, Badge, Button, Drawer } from 'antd'
import {
  DashboardOutlined,
  TrophyOutlined,
  ControlOutlined,
  TeamOutlined,
  ThunderboltOutlined,
  SafetyCertificateOutlined,
  SettingOutlined,
  AppstoreOutlined,
  UserOutlined,
  RocketOutlined,
  LogoutOutlined,
  MenuOutlined,
  CloseOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import { authApi } from '@/api/auth'
import { useWebSocket } from '@/hooks/useWebSocket'
import { useEffect, useState } from 'react'

const { Sider } = Layout
const { Text } = Typography

const adminMenuItems = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: '仪表盘' },
  { key: '/games', icon: <ControlOutlined />, label: '比赛管理' },
  { key: '/teams', icon: <TeamOutlined />, label: '队伍管理' },
  { key: '/admin/containers', icon: <AppstoreOutlined />, label: '容器管理' },
  { key: '/users', icon: <UserOutlined />, label: '用户管理' },
  { key: '/docker-images', icon: <AppstoreOutlined />, label: '镜像管理' },
  { key: '/settings', icon: <SettingOutlined />, label: '系统设置' },
]

const playerMenuItems = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: '比赛概况' },
  { key: '/ranking', icon: <TrophyOutlined />, label: '排行榜' },
  { key: '/attack', icon: <ThunderboltOutlined />, label: '提交 Flag' },
]

interface Props {
  collapsed: boolean
  onToggle: () => void
}

export default function Sidebar({ collapsed, onToggle }: Props) {
  const navigate = useNavigate()
  const location = useLocation()
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  const { connected } = useWebSocket()
  const [isMobile, setIsMobile] = useState(false)
  const [mobileMenuVisible, setMobileMenuVisible] = useState(false)

  // Detect mobile screen
  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768)
    }
    
    checkMobile()
    window.addEventListener('resize', checkMobile)
    return () => window.removeEventListener('resize', checkMobile)
  }, [])

  const isAdmin = user?.role === 'admin' || user?.role === 'organizer'
  const menuItems = isAdmin ? adminMenuItems : playerMenuItems
  const selectedKey = menuItems.find(
    (item) => location.pathname === item.key || location.pathname.startsWith(item.key + '/'),
  )?.key || '/dashboard'

  const handleLogout = async () => {
    try { await authApi.logout() } catch { /* ignore */ }
    useAuthStore.getState().logout()
    navigate('/login')
  }

  const handleMenuClick = (key: string) => {
    navigate(key)
    if (isMobile) {
      setMobileMenuVisible(false)
    }
  }

  const toggleMobileMenu = () => {
    setMobileMenuVisible(!mobileMenuVisible)
  }

  // Mobile hamburger menu button
  const MobileMenuButton = () => (
    <Button
      type="text"
      icon={mobileMenuVisible ? <CloseOutlined /> : <MenuOutlined />}
      onClick={toggleMobileMenu}
      className="mobile-menu-button"
      style={{
        position: 'fixed',
        top: '16px',
        left: '16px',
        zIndex: 1000,
        background: '#1f2937',
        color: '#fff',
        width: '48px',
        height: '48px',
        display: isMobile ? 'flex' : 'none',
        alignItems: 'center',
        justifyContent: 'center',
        borderRadius: '8px',
        border: '1px solid #2a2a4a',
      }}
    />
  )

  // Menu content (reusable for both mobile and desktop)
  const MenuContent = () => (
    <>
      <div className="flex items-center justify-center py-4 border-b border-[#2a2a4a] cursor-pointer" onClick={() => handleMenuClick('/dashboard')}>
        {(!collapsed || isMobile) ? (
          <div className="flex items-center gap-2">
            <span className="text-2xl">⚔️</span>
            <Text strong className="text-lg text-indigo-400 tracking-wide">AWD Arena</Text>
          </div>
        ) : <span className="text-2xl">⚔️</span>}
      </div>

      <div className="px-4 py-2 flex items-center gap-2">
        <Badge status={connected ? 'success' : 'error'} />
        <Text className="text-xs" type={connected ? 'success' : 'danger'}>
          {connected ? '实时已连接' : '连接断开'}
        </Text>
      </div>

      <Menu
        mode="inline"
        selectedKeys={[selectedKey]}
        items={menuItems}
        onClick={({ key }) => handleMenuClick(key)}
        style={{ background: 'transparent', border: 'none' }}
        theme="dark"
        inlineCollapsed={!isMobile && collapsed}
      />

      {!isAdmin && (!collapsed || isMobile) && (
        <div className="px-4 pt-2 mt-2 border-t border-[#2a2a4a]">
          <Menu
            mode="inline"
            selectedKeys={location.pathname === '/profile' ? ['/profile'] : []}
            items={[{ key: '/profile', icon: <UserOutlined />, label: '个人中心' }]}
            onClick={({ key }) => handleMenuClick(key)}
            style={{ background: 'transparent', border: 'none' }}
            theme="dark"
          />
        </div>
      )}

      <div className="absolute bottom-0 left-0 right-0 border-t border-[#2a2a4a]" style={{ background: '#0d1117' }}>
        <div className="p-4 flex items-center gap-2">
          <Avatar
            style={{ backgroundColor: isAdmin ? '#6366f1' : '#10b981', flexShrink: 0 }}
            icon={<UserOutlined />}
            size={32}
          />
          {(!collapsed || isMobile) && (
            <div className="flex-1 min-w-0">
              <div className="text-sm truncate">{user?.username || '未知用户'}</div>
              <Tag color={isAdmin ? 'purple' : 'green'} className="text-xs mt-0.5">
                {isAdmin ? '管理员' : '选手'}
              </Tag>
            </div>
          )}
        </div>
        
        {/* Logout Button */}
        {(!collapsed || isMobile) && (
          <div className="px-4 pb-4">
            <Button
              type="primary"
              danger
              icon={<LogoutOutlined />}
              onClick={handleLogout}
              block
              style={{
                background: '#dc2626',
                borderColor: '#dc2626',
                marginTop: '-8px'
              }}
            >
              退出登录
            </Button>
          </div>
        )}
        
        {collapsed && !isMobile && (
          <div className="px-2 pb-2 flex justify-center">
            <Button
              type="text"
              danger
              icon={<LogoutOutlined />}
              onClick={handleLogout}
              style={{ color: '#dc2626' }}
            />
          </div>
        )}
      </div>
    </>
  )

  // Mobile view - use Drawer
  if (isMobile) {
    return (
      <>
        <MobileMenuButton />
        <Drawer
          placement="left"
          open={mobileMenuVisible}
          onClose={() => setMobileMenuVisible(false)}
          width={280}
          className="mobile-sidebar-drawer"
          styles={{
            body: { padding: 0, background: '#111827' },
            header: { display: 'none' },
          }}
          maskClosable={true}
          maskStyle={{ background: 'rgba(0, 0, 0, 0.65)' }}
        >
          <div style={{ height: '100%', position: 'relative', background: '#111827' }}>
            <MenuContent />
          </div>
        </Drawer>
      </>
    )
  }

  // Desktop view - use Sider
  return (
    <Sider
      collapsible
      collapsed={collapsed}
      onCollapse={onToggle}
      trigger={null}
      width={220}
      collapsedWidth={72}
      style={{
        overflow: 'auto',
        overflowY: 'auto',
        height: '100vh',
        position: 'fixed',
        left: 0,
        top: 0,
        bottom: 0,
        background: '#111827',
        borderRight: '1px solid #2a2a4a',
        zIndex: 100,
        display: 'flex',
        flexDirection: 'column'
      }}
    >
      <MenuContent />
    </Sider>
  )
}

