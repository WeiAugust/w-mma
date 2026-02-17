export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const AUTH_TOKEN_KEY = 'admin_auth_token'

let memoryToken = ''

function hasStorageApi(): boolean {
  return (
    typeof localStorage !== 'undefined' &&
    typeof localStorage.getItem === 'function' &&
    typeof localStorage.setItem === 'function' &&
    typeof localStorage.removeItem === 'function'
  )
}

function readToken(): string {
  if (memoryToken) {
    return memoryToken
  }
  if (hasStorageApi()) {
    return localStorage.getItem(AUTH_TOKEN_KEY) || ''
  }
  return ''
}

export function getAuthToken(): string {
  return readToken()
}

export function setAuthToken(token: string): void {
  memoryToken = token
  if (hasStorageApi()) {
    if (token) {
      localStorage.setItem(AUTH_TOKEN_KEY, token)
    } else {
      localStorage.removeItem(AUTH_TOKEN_KEY)
    }
  }
}

export function clearAuthToken(): void {
  setAuthToken('')
}

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const token = readToken()
  const res = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(init?.headers || {}),
    },
    ...init,
  })

  if (!res.ok) {
    let message = `request failed: ${res.status}`
    try {
      const data = (await res.json()) as { error?: string }
      if (data?.error) {
        message = data.error
      }
    } catch {
      // Keep fallback message when response is not JSON.
    }
    throw new Error(message)
  }

  return (await res.json()) as T
}
