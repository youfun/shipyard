/**
 * CLI Service
 * API service functions for CLI device authorization
 */
import apiClient from '../client'

export interface DeviceContext {
  session_id: string
  os: string
  device_name: string
  public_ip: string
  request_timestamp: number
}

export interface DeviceAuthRequest {
  session_id: string
  approved: boolean
}

// Response envelope type
interface ApiResponse<T> {
  success: boolean
  data: T
}

export const getDeviceContext = async (sessionId: string): Promise<DeviceContext> => {
  const response = await apiClient.get<ApiResponse<DeviceContext>>(`/cli/sessions/${sessionId}`)
  return response.data.data
}

export const confirmDeviceAuth = async (data: DeviceAuthRequest): Promise<void> => {
  await apiClient.post<ApiResponse<void>>('/cli/confirm', data)
}
