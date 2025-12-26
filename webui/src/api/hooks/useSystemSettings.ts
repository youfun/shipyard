import { createQuery, createMutation, useQueryClient } from '@tanstack/solid-query'
import * as systemService from '../services/systemService'

export const useSystemSettings = () => {
  const queryClient = useQueryClient()

  const settingsQuery = createQuery(() => ({
    queryKey: ['system-settings'],
    queryFn: systemService.fetchSystemSettings,
  }))

  const updateSettingsMutation = createMutation(() => ({
    mutationFn: systemService.updateSystemSettings,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['system-settings'] })
    },
  }))

  return {
    settingsQuery,
    updateSettingsMutation,
  }
}
