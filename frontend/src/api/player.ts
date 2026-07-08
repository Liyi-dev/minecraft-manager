import api from './index'

export interface PlayerInfo {
  name: string
  uuid: string
  online: boolean
}

export interface PlayersResponse {
  players: PlayerInfo[]
  count: number
}

export async function getPlayers(): Promise<PlayersResponse> {
  const res = await api.get('/players')
  return res.data
}

export async function kickPlayer(name: string, reason: string = ''): Promise<any> {
  const res = await api.post('/players/kick', { name, reason })
  return res.data
}

export async function banPlayer(name: string, reason: string = ''): Promise<any> {
  const res = await api.post('/players/ban', { name, reason })
  return res.data
}

export async function opPlayer(name: string): Promise<any> {
  const res = await api.post('/players/op', { name })
  return res.data
}

export async function deopPlayer(name: string): Promise<any> {
  const res = await api.post('/players/deop', { name })
  return res.data
}
