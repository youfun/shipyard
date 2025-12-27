/**
 * Setup Service
 * API service functions for initial setup
 */
import apiClient from '../client'
import type { ApiResponse } from '../../types'

export interface SetupRequest {
  username: string
  password: string
}

export interface SetupResponse {
  success: boolean
  message?: string
}

export const setup = async (data: SetupRequest): Promise<SetupResponse> => {
  const response = await apiClient.post<ApiResponse<void>>('/setup', data)
  return { success: response.data.success, message: response.data.message }
}

export const checkSetupStatus = async (): Promise<{ setup_required: boolean }> => {
  const response = await apiClient.get<ApiResponse<{ setup_required: boolean }>>('/setup/status')
  return response.data.data!
}
