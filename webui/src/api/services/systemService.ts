import apiClient from '../client'
import type { ApiResponse } from '../../types'

export interface SystemSettings {
  domain: string
}

export interface SystemStatus {
  status: string
  version: string
  go_version: string
  os: string
  arch: string
  num_cpu: number
  num_goroutine: number
  memory?: {
    alloc_mb: number
    total_alloc_mb: number
    sys_mb: number
  }
}

export const fetchSystemSettings = async (): Promise<SystemSettings> => {
  const response = await apiClient.get<ApiResponse<SystemSettings>>('/system/settings')
  return response.data.data!
}

export const updateSystemSettings = async (settings: SystemSettings): Promise<SystemSettings> => {
  const response = await apiClient.post<ApiResponse<SystemSettings>>('/system/settings', settings)
  return response.data.data!
}

export const fetchSystemStatus = async (): Promise<SystemStatus> => {
  const response = await apiClient.get<ApiResponse<SystemStatus>>('/status')
  return response.data.data!
}
