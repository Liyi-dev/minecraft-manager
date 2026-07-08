import api from './index'

export interface ServerStatus {
  online: boolean
  player_count: number
  max_players: number
  tps: number
  version: string
}

export async function getStatus(): Promise<ServerStatus> {
  const res = await api.get('/server/status')
  return res.data
}
