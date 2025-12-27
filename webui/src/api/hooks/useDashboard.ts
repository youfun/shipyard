/**
 * Dashboard Hooks
 * TanStack Query hooks for dashboard
 */
import { useQuery } from '@tanstack/solid-query'
import * as dashboardService from '@api/services/dashboardService'
import { createQueryOptions } from '@api/utils'

const keys = {
  stats: ['dashboard', 'stats'] as const,
  recentDeployments: ['dashboard', 'recent-deployments'] as const,
}

// Query options for better type safety and reusability
const dashboardQueries = {
  stats: () => createQueryOptions(
    keys.stats,
    dashboardService.getDashboardStats,
    {
      staleTime: 5 * 60 * 1000, // 5 minutes
      refetchOnWindowFocus: true,
    }
  ),
  recentDeployments: () => createQueryOptions(
    keys.recentDeployments,
    dashboardService.getRecentDeployments,
    {
      staleTime: 30 * 1000, // 30 seconds
      refetchOnWindowFocus: true,
    }
  ),
}

export const useDashboard = () => {
  // Get dashboard stats
  const getStats = () => useQuery(dashboardQueries.stats)

  // Get recent deployments
  const getRecentDeployments = () => useQuery(dashboardQueries.recentDeployments)

  return {
    queries: { getStats, getRecentDeployments },
  }
}
