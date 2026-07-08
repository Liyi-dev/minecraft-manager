<template>
  <AppLayout>
    <div class="dashboard">
      <van-skeleton :loading="loading" :row="3" animate>
        <!-- Server Status Card -->
        <div class="status-card" :class="{ online: status.online, offline: !status.online }">
          <div class="status-indicator">
            <span class="dot"></span>
            <span class="status-text">{{ status.online ? '在线' : '离线' }}</span>
          </div>
          <p class="version" v-if="status.version">版本: {{ status.version }}</p>
        </div>

        <!-- Stats Grid -->
        <van-grid :column-num="2" :gutter="12" style="margin-top: 16px">
          <van-grid-item>
            <van-icon name="friends-o" size="24" color="#1989fa" />
            <div class="stat-label">在线玩家</div>
            <div class="stat-value">{{ status.player_count }} / {{ status.max_players }}</div>
          </van-grid-item>

          <van-grid-item>
            <van-icon name="chart-trending-o" size="24" color="#07c160" />
            <div class="stat-label">TPS</div>
            <div class="stat-value">{{ status.tps.toFixed(1) }}</div>
          </van-grid-item>
        </van-grid>

        <!-- WebSocket Status -->
        <van-cell-group inset style="margin-top: 16px">
          <van-cell title="WebSocket 连接">
            <template #right-icon>
              <van-tag :type="wsConnected ? 'success' : 'danger'">
                {{ wsConnected ? '已连接' : '未连接' }}
              </van-tag>
            </template>
          </van-cell>
          <van-cell title="RCON 状态">
            <template #right-icon>
              <van-tag :type="status.online ? 'success' : 'danger'">
                {{ status.online ? '正常' : '离线' }}
              </van-tag>
            </template>
          </van-cell>
        </van-cell-group>
      </van-skeleton>

      <!-- Refresh Button -->
      <div style="margin-top: 20px; text-align: center">
        <van-button type="primary" :loading="loading" @click="refresh" icon="replay">
          刷新状态
        </van-button>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AppLayout from '../components/AppLayout.vue'
import { useWebSocketStore } from '../stores/websocket'
import * as serverApi from '../api/server'
import { showToast } from 'vant'

const wsStore = useWebSocketStore()

const loading = ref(true)
const status = ref<serverApi.ServerStatus>({
  online: false,
  player_count: 0,
  max_players: 20,
  tps: 0,
  version: '',
})

const wsConnected = ref(false)

// Watch WS connection
const checkWsInterval = setInterval(() => {
  wsConnected.value = wsStore.connected
}, 1000)

async function refresh() {
  loading.value = true
  try {
    status.value = await serverApi.getStatus()
    wsConnected.value = wsStore.connected
  } catch (e: any) {
    showToast('获取状态失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  refresh()
  // Auto-refresh every 30 seconds
  setInterval(refresh, 30000)
})
</script>

<style scoped>
.dashboard {
  max-width: 600px;
  margin: 0 auto;
}

.status-card {
  padding: 24px;
  border-radius: 12px;
  text-align: center;
  color: #fff;
  transition: background 0.3s;
}

.status-card.online {
  background: linear-gradient(135deg, #07c160, #06ad56);
}

.status-card.offline {
  background: linear-gradient(135deg, #ee0a24, #c8081c);
}

.status-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-bottom: 8px;
}

.dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #fff;
}

.status-text {
  font-size: 24px;
  font-weight: 700;
}

.version {
  font-size: 14px;
  opacity: 0.9;
}

.stat-label {
  font-size: 12px;
  color: #969799;
  margin-top: 4px;
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  color: #323233;
  margin-top: 2px;
}
</style>
