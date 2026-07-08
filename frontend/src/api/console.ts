import api from './index'

export interface CommandResult {
  command: string
  result: string
}

export async function execCommand(command: string): Promise<CommandResult> {
  const res = await api.post('/console/exec', { command })
  return res.data
}
