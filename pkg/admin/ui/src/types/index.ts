export interface User {
  id: number
  username: string
  email: string
  is_superuser: boolean
}

export interface Session {
  id: string
  user_id: number
  username: string
  ip: string
  user_agent: string
  created_at: string
  last_activity: string
}

export interface Model {
  name: string
  table: string
  fields: Field[]
  count?: number
}

export interface Field {
  name: string
  type: string
  primary: boolean
  nullable: boolean
}

export interface Record {
  [key: string]: any
}

export interface AuditLog {
  id: number
  timestamp: string
  user: string
  action: string
  resource: string
  details: string
}

export interface RBACPolicy {
  ptype: string
  v0: string
  v1: string
  v2: string
}

export interface HealthCheck {
  name: string
  status: 'healthy' | 'unhealthy' | 'unknown'
  latency?: number
  error?: string
}

export interface SystemMetrics {
  goroutines: number
  memory: {
    alloc: number
    total_alloc: number
    sys: number
    num_gc: number
  }
  cpu_usage: number
  db_pools: {
    name: string
    open_connections: number
    in_use: number
    idle: number
  }[]
}

export interface LiveRequest {
  id: string
  method: string
  path: string
  status: number
  duration: number
  timestamp: string
}

export interface FeatureFlag {
  name: string
  enabled: boolean
}
