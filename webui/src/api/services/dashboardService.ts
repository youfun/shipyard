/**
 * Dashboard Service
 * API service functions for dashboard stats
 */
import apiClient from '../client'
import type { ApiResponse } from '../../types'
import type { RecentDeployment } from '../../types/app'

export interface DashboardStats {
  applications_count: number
  hosts_count: number
  deployments_count: number
}

export const getDashboardStats = async (): Promise<DashboardStats> => {
  const response = await apiClient.get<ApiResponse<DashboardStats>>('/dashboard/stats')
  return response.data.data!
}

export const getRecentDeployments = async (): Promise<RecentDeployment[]> => {
  const response = await apiClient.get<ApiResponse<RecentDeployment[]>>('/dashboard/recent-deployments')
  return response.data.data || []
}
