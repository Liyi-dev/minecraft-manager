import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as authApi from '../api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<{ id: number; username: string; role: string } | null>(null)

  const isLoggedIn = computed(() => !!token.value)

  async function login(username: string, password: string) {
    const resp = await authApi.login({ username, password })
    token.value = resp.token
    localStorage.setItem('token', resp.token)
    user.value = {
      id: resp.user_id,
      username: resp.username,
      role: resp.role,
    }
    return resp
  }

  async function fetchUser() {
    try {
      const info = await authApi.getMe()
      user.value = info
    } catch {
      // Token invalid, clear
      logout()
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    // Attempt server-side logout (fire and forget)
    authApi.logout().catch(() => {})
  }

  return {
    token,
    user,
    isLoggedIn,
    login,
    logout,
    fetchUser,
  }
})
