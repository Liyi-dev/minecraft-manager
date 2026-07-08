<template>
  <AppLayout>
    <div class="players-page">
      <div class="page-header">
        <h3>在线玩家 ({{ players.length }})</h3>
        <van-button size="small" :loading="loading" @click="fetchPlayers" icon="replay">
          刷新
        </van-button>
      </div>

      <van-skeleton :loading="loading" :row="3" animate>
        <!-- Player List -->
        <div v-if="players.length > 0">
          <van-swipe-cell v-for="player in players" :key="player.name">
            <van-cell
              :title="player.name"
              :label="player.uuid || 'UUID 未获取'"
              icon="friends-o"
              is-link
              @click="showPlayerActions(player)"
            />
            <template #right>
              <van-button
                square
                type="danger"
                text="Kick"
                @click="handleKick(player.name)"
                style="height: 100%"
              />
            </template>
          </van-swipe-cell>
        </div>

        <!-- Empty State -->
        <van-empty v-else description="暂无在线玩家" image="search" />
      </van-skeleton>

      <!-- Player Action Sheet -->
      <van-action-sheet
        v-model:show="actionSheetVisible"
        :title="selectedPlayer?.name"
        :actions="playerActions"
        @select="onActionSelect"
        cancel-text="取消"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AppLayout from '../components/AppLayout.vue'
import * as playerApi from '../api/player'
import type { PlayerInfo } from '../api/player'
import { showToast, showDialog } from 'vant'

const players = ref<PlayerInfo[]>([])
const loading = ref(false)
const actionSheetVisible = ref(false)
const selectedPlayer = ref<PlayerInfo | null>(null)

const playerActions = [
  { name: '踢出 (Kick)', value: 'kick' },
  { name: '封禁 (Ban)', value: 'ban' },
  { name: '设为 OP', value: 'op' },
  { name: '取消 OP', value: 'deop' },
]

async function fetchPlayers() {
  loading.value = true
  try {
    const resp = await playerApi.getPlayers()
    players.value = resp.players || []
  } catch (e: any) {
    showToast('获取玩家列表失败')
  } finally {
    loading.value = false
  }
}

function showPlayerActions(player: PlayerInfo) {
  selectedPlayer.value = player
  actionSheetVisible.value = true
}

async function onActionSelect(action: { value: string }) {
  if (!selectedPlayer.value) return
  actionSheetVisible.value = false

  const name = selectedPlayer.value.name

  switch (action.value) {
    case 'kick':
      await handleKick(name)
      break
    case 'ban':
      await handleBan(name)
      break
    case 'op':
      await handleOp(name)
      break
    case 'deop':
      await handleDeop(name)
      break
  }
}

async function handleKick(name: string) {
  try {
    const reason = await showDialog({
      title: `踢出 ${name}`,
      message: '请输入原因（可选）',
    }).then(
      () => (document.querySelector('.van-dialog__input') as HTMLInputElement)?.value || ''
    )
    // Actually, let's use a simpler approach
    await playerApi.kickPlayer(name)
    showToast({ message: `已踢出 ${name}`, icon: 'success' })
    fetchPlayers()
  } catch (e: any) {
    if (e !== 'cancel') {
      showToast('操作失败: ' + (e.response?.data?.error || e.message))
    }
  }
}

async function handleBan(name: string) {
  try {
    await playerApi.banPlayer(name)
    showToast({ message: `已封禁 ${name}`, icon: 'success' })
    fetchPlayers()
  } catch (e: any) {
    showToast('操作失败: ' + (e.response?.data?.error || e.message))
  }
}

async function handleOp(name: string) {
  try {
    await playerApi.opPlayer(name)
    showToast({ message: `已将 ${name} 设为 OP`, icon: 'success' })
  } catch (e: any) {
    showToast('操作失败: ' + (e.response?.data?.error || e.message))
  }
}

async function handleDeop(name: string) {
  try {
    await playerApi.deopPlayer(name)
    showToast({ message: `已取消 ${name} 的 OP`, icon: 'success' })
  } catch (e: any) {
    showToast('操作失败: ' + (e.response?.data?.error || e.message))
  }
}

onMounted(() => {
  fetchPlayers()
  // Auto-refresh player list every 15 seconds
  setInterval(fetchPlayers, 15000)
})
</script>

<style scoped>
.players-page {
  max-width: 600px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.page-header h3 {
  font-size: 18px;
  color: #323233;
}
</style>
