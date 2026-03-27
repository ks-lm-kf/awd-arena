import { useEffect, useRef, useCallback, useState } from 'react'
import { useAuthStore } from '@/stores/authStore'
import type { WSEvent, WSEventType } from '@/types'

type EventHandler = (data: any) => void

const MAX_RECONNECT_DELAY = 30000
const PING_INTERVAL = 30000

export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null)
  const handlersRef = useRef<Map<WSEventType, Set<EventHandler>>>(new Map())
  const [connected, setConnected] = useState(false)
  const token = useAuthStore((s) => s.token)
  const reconnectDelay = useRef(1000)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout>>(undefined)
  const pingTimer = useRef<ReturnType<typeof setInterval>>(undefined)
  const mountedRef = useRef(true)

  const stopPing = useCallback(() => {
    if (pingTimer.current) {
      clearInterval(pingTimer.current)
      pingTimer.current = undefined
    }
  }, [])

  const connect = useCallback(() => {
    if (!token) return
    if (wsRef.current?.readyState === WebSocket.OPEN || wsRef.current?.readyState === WebSocket.CONNECTING) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws?token=${token}`

    try {
      const ws = new WebSocket(wsUrl)
      wsRef.current = ws

      ws.onopen = () => {
        if (!mountedRef.current) return
        setConnected(true)
        reconnectDelay.current = 1000
        console.log('[WS] connected')
        // Start heartbeat
        stopPing()
        pingTimer.current = setInterval(() => {
          if (ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ type: 'ping' }))
          }
        }, PING_INTERVAL)
      }

      ws.onmessage = (e) => {
        try {
          const event: WSEvent = JSON.parse(e.data)
          if ((event as any).type === 'pong') return // heartbeat response
          const eventHandlers = handlersRef.current.get(event.type)
          if (eventHandlers) {
            eventHandlers.forEach((handler) => handler(event.data))
          }
        } catch {
          // ignore parse errors
        }
      }

      ws.onclose = () => {
        if (!mountedRef.current) return
        setConnected(false)
        stopPing()
        console.log(`[WS] disconnected, reconnecting in ${reconnectDelay.current}ms...`)
        reconnectTimer.current = setTimeout(connect, reconnectDelay.current)
        // Exponential backoff
        reconnectDelay.current = Math.min(reconnectDelay.current * 2, MAX_RECONNECT_DELAY)
      }

      ws.onerror = () => {
        ws.close()
      }
    } catch {
      // Connection failed, schedule retry
      reconnectTimer.current = setTimeout(connect, reconnectDelay.current)
      reconnectDelay.current = Math.min(reconnectDelay.current * 2, MAX_RECONNECT_DELAY)
    }
  }, [token, stopPing])

  const subscribe = useCallback((eventType: WSEventType, handler: EventHandler) => {
    if (!handlersRef.current.has(eventType)) {
      handlersRef.current.set(eventType, new Set())
    }
    handlersRef.current.get(eventType)!.add(handler)
    return () => {
      handlersRef.current.get(eventType)?.delete(handler)
    }
  }, [])

  const unsubscribe = useCallback((eventType: WSEventType, handler: EventHandler) => {
    handlersRef.current.get(eventType)?.delete(handler)
  }, [])

  const send = useCallback((data: object) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data))
    }
  }, [])

  useEffect(() => {
    mountedRef.current = true
    connect()
    return () => {
      mountedRef.current = false
      clearTimeout(reconnectTimer.current)
      stopPing()
      wsRef.current?.close()
      wsRef.current = null
    }
  }, [connect, stopPing])

  return { connected, subscribe, unsubscribe, send }
}
