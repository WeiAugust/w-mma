import { clearAuthToken, request, setAuthToken } from './request'

export type LoginResponse = {
  token: string
}

export async function login(username: string, password: string): Promise<string> {
  const data = await request<LoginResponse>('/admin/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
  setAuthToken(data.token || '')
  return data.token || ''
}

export function logout(): void {
  clearAuthToken()
}
