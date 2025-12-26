/**
 * Auth Hooks
 * TanStack Query hooks for authentication
 */
import { useQuery, useQueryClient } from '@tanstack/solid-query'
import * as authService from '../services/authService'
import type { ChangePasswordRequest, Enable2FARequest, Disable2FARequest } from '../../types'
import { createQueryOptions, useInvalidateMutation, useSimpleMutation } from '@api/utils'

const keys = {
  user: ['auth', 'user'] as const,
}

// Query options for better type safety
const authQueries = {
  currentUser: () => createQueryOptions(
    keys.user,
    authService.getCurrentUser,
    {
      retry: false,
      staleTime: 5 * 60 * 1000, // 5 minutes
    }
  ),
}

export const useAuth = () => {
  const queryClient = useQueryClient()

  // Get current user
  const getCurrentUser = () => useQuery(authQueries.currentUser)

  // Simple mutations without cache invalidation
  const loginMutation = useSimpleMutation(authService.login)
  const login2FAMutation = useSimpleMutation(authService.login2FA)
  const changePasswordMutation = useSimpleMutation(
    (data: ChangePasswordRequest) => authService.changePassword(data)
  )
  const setup2FAMutation = useSimpleMutation(authService.setup2FA)

  // Logout mutation with custom logic to clear all queries
  const logoutMutation = useInvalidateMutation(
    authService.logout,
    [], // No specific keys to invalidate
    () => {
      queryClient.clear() // Clear all queries on logout
    }
  )

  // Mutations with cache invalidation
  const enable2FAMutation = useInvalidateMutation(
    (data: Enable2FARequest) => authService.enable2FA(data),
    [[...keys.user]]
  )

  const disable2FAMutation = useInvalidateMutation(
    (data: Disable2FARequest) => authService.disable2FA(data),
    [[...keys.user]]
  )

  return {
    queries: { getCurrentUser },
    mutations: {
      login: loginMutation,
      login2FA: login2FAMutation,
      logout: logoutMutation,
      changePassword: changePasswordMutation,
      setup2FA: setup2FAMutation,
      enable2FA: enable2FAMutation,
      disable2FA: disable2FAMutation,
    },
  }
}