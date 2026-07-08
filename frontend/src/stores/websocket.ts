import { defineStore } from 'pinia'
import { ref, shallowRef } from 'vue'
import { WSClient } from '../utils/websocket'
import { useAuthStore } from './auth'

export const useWebSocketStore = defineStore('websocket', () => {
  const client = shallowRef<WSClient | null>(null)
  const connected = ref(false)
  const logs = ref<string[]>([])
  const maxLogLines = 500

  function connect() {
    const authStore = useAuthStore()
    if (!authStore.token) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const url = `${protocol}//${host}/ws`

    const ws = new WSClient(url, authStore.token)

    ws.on('auth_ok', () => {
      connected.value = true
      ws.subscribeLogs()
    })

    ws.on('connected', () => {
      // Connection established, waiting for auth
    })

    ws.on('disconnected', () => {
      connected.value = false
    })

    ws.on('log', (data: any) => {
      logs.value.push(data.line)
      if (logs.value.length > maxLogLines) {
        logs.value = logs.value.slice(-maxLogLines)
      }
    })

    ws.on('player_join', (data: any) => {
      console.log('Player joined:', data.name)
    })

    ws.on('player_leave', (data: any) => {
      console.log('Player left:', data.name)
    })

    ws.on('command_result', (_data: any) => {
      // Command results are handled by the console view directly
    })

    ws.on('error', (data: any) => {
      console.error('WebSocket error:', data)
    })

    ws.connect()
    client.value = ws
  }

  function disconnect() {
    if (client.value) {
      client.value.disconnect()
      client.value = null
    }
    connected.value = false
  }

  function getClient(): WSClient | null {
    return client.value
  }

  return {
    client,
    connected,
    logs,
    connect,
    disconnect,
    getClient,
  }
})
