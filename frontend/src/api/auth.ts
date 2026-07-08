import api from './index'

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user_id: number
  username: string
  role: string
}

export interface UserInfo {
  id: number
  username: string
  role: string
}

export async function login(data: LoginRequest): Promise<LoginResponse> {
  const res = await api.post('/login', data)
  return res.data
}

export async function logout(): Promise<void> {
  await api.post('/logout')
}

export async function getMe(): Promise<UserInfo> {
  const res = await api.get('/me')
  return res.data
}
