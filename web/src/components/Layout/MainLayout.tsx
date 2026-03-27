import { useState, useEffect, type ReactNode } from 'react'
import { Outlet } from 'react-router'
import Sidebar from './Sidebar'

export default function MainLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const [isMobile, setIsMobile] = useState(false)

  // Detect mobile screen
  useEffect(() => {
    const checkMobile = () => {
      const mobile = window.innerWidth < 768
      setIsMobile(mobile)
      // Auto-collapse on mobile
      if (mobile && !collapsed) {
        setCollapsed(true)
      }
    }
    
    checkMobile()
    window.addEventListener('resize', checkMobile)
    return () => window.removeEventListener('resize', checkMobile)
  }, [collapsed])

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar collapsed={collapsed} onToggle={() => setCollapsed(!collapsed)} />
      <main
        className="flex-1 overflow-y-auto p-6 transition-all duration-300"
        style={{ 
          marginLeft: isMobile ? 0 : (collapsed ? 72 : 220),
          paddingTop: isMobile ? '80px' : '24px' // Add top padding for mobile menu button
        }}
      >
        <Outlet />
      </main>
    </div>
  )
}
