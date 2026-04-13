import type { User, Session, Model, Record as AppRecord, AuditLog, RBACPolicy, HealthCheck, SystemMetrics, LiveRequest, FeatureFlag } from '@/types'

const API_BASE = window.location.pathname.replace(/\/admin.*/, '')

async function fetchAPI(path: string, options?: RequestInit) {
  const prefix = API_BASE.endsWith('/admin') ? '' : '/admin'
  const url = `${prefix}${path}`

  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    credentials: 'same-origin',
  })

  if (!response.ok) {
    if (response.status === 401) {
      window.location.href = `${prefix}/login`
      throw new Error('Unauthorized')
    }
    throw new Error(`API Error: ${response.status} ${response.statusText}`)
  }

  return response.json()
}

export async function login(username: string, password: string): Promise<User> {
  const prefix = API_BASE.endsWith('/admin') ? '' : '/admin'
  const formData = new URLSearchParams()
  formData.append('username', username)
  formData.append('password', password)

  const response = await fetch(`${prefix}/login`, {
    method: 'POST',
    body: formData,
    credentials: 'same-origin',
  })

  if (!response.ok) {
    throw new Error('Invalid credentials')
  }

  const data = await response.json()
  return data.user
}

export async function logout(): Promise<void> {
  await fetchAPI('/api/logout', { method: 'POST' })
  window.location.href = `${API_BASE}/login`
}

export async function getCurrentUser(): Promise<User | null> {
  try {
    return await fetchAPI('/api/me')
  } catch {
    return null
  }
}

export async function getModels(): Promise<Model[]> {
  return fetchAPI('/api/models')
}

export async function getModelSchema(name: string): Promise<Model> {
  return fetchAPI(`/api/models/${name}/schema`)
}

export async function getRecords(name: string, params?: Record<string, string>): Promise<AppRecord[]> {
  const searchParams = new URLSearchParams(params)
  return fetchAPI(`/api/models/${name}?${searchParams}`)
}

export async function createRecord(name: string, data: AppRecord): Promise<AppRecord> {
  return fetchAPI(`/api/models/${name}`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function updateRecord(name: string, id: string, data: AppRecord): Promise<AppRecord> {
  return fetchAPI(`/api/models/${name}/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export async function deleteRecord(name: string, id: string): Promise<void> {
  await fetchAPI(`/api/models/${name}/${id}`, { method: 'DELETE' })
}

export async function getSessions(): Promise<Session[]> {
  return fetchAPI('/api/sessions')
}

export async function getAuditLogs(params?: Record<string, string>): Promise<AuditLog[]> {
  const searchParams = new URLSearchParams(params)
  return fetchAPI(`/api/audit?${searchParams}`)
}

export async function getRBACPolicies(): Promise<RBACPolicy[]> {
  return fetchAPI('/api/rbac/policies')
}

export async function createRBACPolicy(policy: Partial<RBACPolicy>): Promise<void> {
  await fetchAPI('/api/rbac/policies', {
    method: 'POST',
    body: JSON.stringify(policy),
  })
}

export async function deleteRBACPolicy(policy: Partial<RBACPolicy>): Promise<void> {
  await fetchAPI('/api/rbac/policies', {
    method: 'DELETE',
    body: JSON.stringify(policy),
  })
}

export async function getHealthChecks(): Promise<HealthCheck[]> {
  return fetchAPI('/api/health')
}

export async function getSystemMetrics(): Promise<SystemMetrics> {
  return fetchAPI('/api/system/snapshot')
}

export async function getLiveRequests(): Promise<LiveRequest[]> {
  return fetchAPI('/api/live/snapshot')
}

export function getLiveWebSocket(): WebSocket | null {
  try {
    const prefix = API_BASE.endsWith('/admin') ? '' : '/admin'
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return new WebSocket(`${protocol}//${window.location.host}${prefix}/api/live/ws`)
  } catch {
    return null
  }
}

export async function getFeatureFlags(): Promise<FeatureFlag[]> {
  return fetchAPI('/api/features')
}

export async function toggleFeatureFlag(name: string, enabled: boolean): Promise<void> {
  await fetchAPI(`/api/features/${name}`, {
    method: 'PUT',
    body: JSON.stringify({ enabled }),
  })
}

export async function exportData(format: 'csv' | 'json' | 'sql', modelName?: string): Promise<string> {
  const response = await fetchAPI('/api/export', {
    method: 'POST',
    body: JSON.stringify({ format, model: modelName }),
  })
  return response.url
}

export async function importData(file: File): Promise<void> {
  const formData = new FormData()
  formData.append('file', file)

  const prefix = API_BASE.endsWith('/admin') ? '' : '/admin'
  await fetch(`${prefix}/api/import/upload`, {
    method: 'POST',
    body: formData,
    credentials: 'same-origin',
  })
}
