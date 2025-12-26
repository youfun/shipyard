export interface SSHHost {
  uid: string
  name: string
  addr: string
  port: number
  user: string
  status: string
  arch: string
  has_password?: boolean
  has_private_key?: boolean
  initialized_at?: string
  created_at?: string
  updated_at?: string
}

export interface SSHHostRequest {
  name: string
  addr: string
  port?: number
  user: string
  password?: string
  private_key?: string
}
