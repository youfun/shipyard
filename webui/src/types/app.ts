export interface Application {
  uid: string
  name: string
  description?: string
  status?: string
  domain?: string
  primary_domain?: string // actual domain field from backend
  target_port?: number
  targetPort?: number  // alias for target_port
  active_port?: number // actual port field from backend
  branch?: string
  created_at: string
  updated_at: string
  createdAt?: string
  updatedAt?: string
  last_deployed_at?: string
  linked_host?: string
  instances?: ApplicationInstance[]
}

export interface ApplicationLog {
  timestamp: string
  level: string
  message: string
}

export interface DeploymentHistory {
  uid: string
  version: string
  status: string
  host_name: string
  port: number
  created_at: string
  output?: string
}

export interface ApplicationInstance {
  uid: string
  host_id: string
  host_name: string
  host_addr: string
  status: string
  active_port: number
}

export interface Secret {
  key: string
  value: string
}

export interface Domain {
  uid: string
  domainName: string
  hostPort: number
  isActive: boolean
  createdAt: string
}

export interface EnvironmentVariable {
  uid: string
  key: string
  value: string
  isEncrypted: boolean
  created_at?: string
  updated_at?: string
}

export interface CreateEnvironmentVariableRequest {
  key: string
  value: string
  isEncrypted: boolean
}

export interface ApplicationToken {
  uid: string
  name: string
  expires_at?: string
  last_used_at?: string
  created_at: string
}

export interface CreateApplicationTokenRequest {
  name: string
  expires_at?: string
}

export interface CreateApplicationTokenResponse {
  uid: string
  name: string
  token: string
  created_at: string
}

export interface RecentDeployment {
  uid: string
  app_name: string
  host_name: string
  host_addr: string
  version: string
  git_commit: string
  status: string
  port: number
  deployed_at?: string
}
