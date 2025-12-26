/**
 * SSH Hosts Service
 * API service functions for SSH host management
 */
import apiClient from '../client'
import type { SSHHost, SSHHostRequest, ApiResponse } from '../../types'

export interface SSHHostsResponse {
  data: SSHHost[]
}

// List all SSH hosts
export const fetchSSHHosts = async (): Promise<SSHHostsResponse> => {
  const response = await apiClient.get<ApiResponse<SSHHost[]>>('/ssh-hosts')
  return { data: response.data.data! }
}

// Get single SSH host
export const fetchSSHHostById = async (uid: string): Promise<SSHHost> => {
  const response = await apiClient.get<ApiResponse<SSHHost>>(`/ssh-hosts/${uid}`)
  return response.data.data!
}

// Create SSH host
export const createSSHHost = async (data: SSHHostRequest): Promise<SSHHost> => {
  const response = await apiClient.post<ApiResponse<SSHHost>>('/ssh-hosts', data)
  return response.data.data!
}

// Update SSH host
export const updateSSHHost = async (uid: string, data: SSHHostRequest): Promise<SSHHost> => {
  const response = await apiClient.put<ApiResponse<SSHHost>>(`/ssh-hosts/${uid}`, data)
  return response.data.data!
}

// Delete SSH host
export const deleteSSHHost = async (uid: string): Promise<void> => {
  await apiClient.delete(`/ssh-hosts/${uid}`)
}

// Test SSH host connection
export const testSSHHost = async (uid: string): Promise<{ success: boolean; message?: string }> => {
  const response = await apiClient.post<{ success: boolean; message?: string }>(`/ssh-hosts/${uid}/test`)
  return response.data
}
