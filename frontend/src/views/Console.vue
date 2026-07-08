<template>
  <AppLayout>
    <div class="console-page">
      <!-- Output Window -->
      <div class="output-window" ref="outputEl">
        <div v-if="entries.length === 0" class="empty-hint">
          输入命令开始，或订阅日志查看实时输出
        </div>
        <div
          v-for="(entry, idx) in entries"
          :key="idx"
          class="output-entry"
          :class="entry.type"
        >
          <span class="timestamp">{{ formatTime(entry.time) }}</span>
          <span class="content">{{ entry.content }}</span>
        </div>
      </div>

      <!-- Command Input -->
      <div class="input-bar">
        <van-field
          v-model="command"
          placeholder="输入 Minecraft 命令..."
          :disabled="sending"
          @keypress="onKeyPress"
          clearable
        >
          <template #button>
            <van-button
              size="small"
              type="primary"
              :loading="sending"
              @click="sendCommand"
            >
              发送
            </van-button>
          </template>
        </van-field>
      </div>

      <!-- Quick Commands -->
      <div class="quick-commands">
        <van-tag
          v-for="cmd in quickCommands"
          :key="cmd"
          type="primary"
          plain
          size="medium"
          @click="command = cmd"
          style="cursor: pointer; margin: 4px"
        >
          /{{ cmd }}
        </van-tag>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted, onUnmounted } from 'vue'
import AppLayout from '../components/AppLayout.vue'
import { useWebSocketStore } from '../stores/websocket'
import * as consoleApi from '../api/console'
import { showToast } from 'vant'

interface OutputEntry {
  type: 'command' | 'result' | 'log' | 'error' | 'info'
  content: string
  time: Date
}

const wsStore = useWebSocketStore()
const command = ref('')
const sending = ref(false)
const entries = ref<OutputEntry[]>([])
const outputEl = ref<HTMLElement | null>(null)
const history = ref<string[]>([])
const historyIdx = ref(-1)

const quickCommands = ['list', 'tps', 'say Hello', 'time query daytime', 'weather clear']

function addEntry(type: OutputEntry['type'], content: string) {
  entries.value.push({ type, content, time: new Date() })
  // Scroll to bottom
  nextTick(() => {
    if (outputEl.value) {
      outputEl.value.scrollTop = outputEl.value.scrollHeight
    }
  })
}

async function sendCommand() {
  const cmd = command.value.trim()
  if (!cmd) return

  sending.value = true
  const cmdToSend = cmd.startsWith('/') ? cmd : cmd

  addEntry('command', `> ${cmdToSend}`)

  try {
    const result = await consoleApi.execCommand(cmdToSend)
    addEntry('result', result.result || '(无输出)')

    // Add to history
    history.value.push(cmdToSend)
    historyIdx.value = history.value.length
  } catch (e: any) {
    addEntry('error', `错误: ${e.response?.data?.error || e.message}`)
  } finally {
    sending.value = false
    command.value = ''
  }
}

function onKeyPress(e: KeyboardEvent) {
  // Arrow up for history
  if (e.key === 'ArrowUp') {
    e.preventDefault()
    if (historyIdx.value > 0) {
      historyIdx.value--
      command.value = history.value[historyIdx.value]
    }
  } else if (e.key === 'ArrowDown') {
    e.preventDefault()
    if (historyIdx.value < history.value.length - 1) {
      historyIdx.value++
      command.value = history.value[historyIdx.value]
    } else if (historyIdx.value === history.value.length - 1) {
      historyIdx.value = history.value.length
      command.value = ''
    }
  }
}

function formatTime(d: Date): string {
  return d.toLocaleTimeString('zh-CN', { hour12: false })
}

onMounted(() => {
  // Subscribe to WS logs
  const ws = wsStore.getClient()
  if (ws) {
    ws.on('log', (data: any) => {
      addEntry('log', data.line)
    })
    ws.on('command_result', (data: any) => {
      addEntry('info', `[${data.username}] ${data.command}: ${data.result}`)
    })
  }
})
</script>

<style scoped>
.console-page {
  max-width: 800px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  height: calc(100vh - 120px);
}

.output-window {
  flex: 1;
  overflow-y: auto;
  background: #1a1a2e;
  color: #e0e0e0;
  border-radius: 8px;
  padding: 12px;
  font-family: 'Cascadia Code', 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.6;
  margin-bottom: 12px;
  min-height: 200px;
}

.empty-hint {
  color: #666;
  text-align: center;
  padding: 40px 0;
}

.output-entry {
  padding: 2px 0;
  word-break: break-all;
}

.output-entry .timestamp {
  color: #666;
  margin-right: 8px;
  font-size: 11px;
}

.output-entry.command .content {
  color: #ffd700;
}

.output-entry.result .content {
  color: #a0a0a0;
}

.output-entry.log .content {
  color: #8be9fd;
}

.output-entry.error .content {
  color: #ff5555;
}

.output-entry.info .content {
  color: #50fa7b;
}

.input-bar {
  margin-bottom: 8px;
}

.quick-commands {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding-bottom: 8px;
}
</style>
