<template>
  <div class="app-layout">
    <header class="app-header">
      <h2>Minecraft Manager</h2>
      <div class="user-info">
        <span>{{ authStore.user?.username }}</span>
        <van-button size="small" type="danger" plain @click="handleLogout">退出</van-button>
      </div>
    </header>

    <main class="app-main">
      <slot />
    </main>

    <van-tabbar v-model="activeTab" @change="onTabChange" route>
      <van-tabbar-item to="/" icon="home-o">仪表盘</van-tabbar-item>
      <van-tabbar-item to="/players" icon="friends-o">玩家</van-tabbar-item>
      <van-tabbar-item to="/console" icon="terminal-o">控制台</van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useWebSocketStore } from '../stores/websocket'
import { showConfirmDialog } from 'vant'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const wsStore = useWebSocketStore()

const activeTab = ref(0)

const tabMap: Record<string, number> = {
  '/': 0,
  '/players': 1,
  '/console': 2,
}

watch(
  () => route.path,
  (path) => {
    if (path in tabMap) {
      activeTab.value = tabMap[path]
    }
  },
  { immediate: true }
)

function onTabChange(index: number) {
  activeTab.value = index
}

async function handleLogout() {
  await showConfirmDialog({
    title: '退出登录',
    message: '确定要退出吗？',
  })
  wsStore.disconnect()
  authStore.logout()
  router.push('/login')
}

onMounted(() => {
  // Connect WebSocket
  if (!wsStore.connected) {
    wsStore.connect()
  }
})
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  padding-bottom: 50px;
}

.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #1989fa;
  color: #fff;
}

.app-header h2 {
  font-size: 18px;
  font-weight: 600;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.app-main {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}
</style>
