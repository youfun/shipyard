/**
 * CLI Hooks
 * TanStack Query hooks for CLI device authorization
 */
import { useQuery } from '@tanstack/solid-query'
import * as cliService from '../services/cliService'
import type { DeviceAuthRequest } from '../services/cliService'
import { createQueryOptions, useSimpleMutation } from '@api/utils'

const keys = {
  deviceContext: (sessionId: string) => ['cli', 'device', sessionId] as const,
}

// Query options for better type safety
const cliQueries = {
  deviceContext: (sessionId: string | null) => createQueryOptions(
    keys.deviceContext(sessionId || ''),
    () => cliService.getDeviceContext(sessionId!),
    {
      enabled: !!sessionId,
      staleTime: 30 * 1000, // 30 seconds for device context
    }
  ),
}

export const useCLI = () => {
  // Get device context
  const getDeviceContext = (sessionId: () => string | null) => 
    useQuery(() => cliQueries.deviceContext(sessionId()))

  // Confirm device auth
  const confirmAuthMutation = useSimpleMutation(
    (data: DeviceAuthRequest) => cliService.confirmDeviceAuth(data)
  )

  return {
    queries: { getDeviceContext },
    mutations: { confirmAuth: confirmAuthMutation },
  }
}
