/**
 * SSH Hosts Hooks
 * TanStack Query hooks for SSH host management
 */
import { useQuery } from '@tanstack/solid-query'
import * as sshHostService from '../services/sshHostService'
import type { SSHHostRequest } from '../../types'
import { createQueryOptions, useInvalidateMutation, useSimpleMutation } from '@api/utils'

const keys = {
  all: ['ssh-hosts'] as const,
  detail: (uid: string) => ['ssh-hosts', uid] as const,
}

// Query options for better type safety
const sshHostQueries = {
  all: () => createQueryOptions(
    keys.all,
    sshHostService.fetchSSHHosts,
    {
      staleTime: 2 * 60 * 1000, // 2 minutes
    }
  ),
  detail: (uid: string | undefined) => createQueryOptions(
    keys.detail(uid || ''),
    () => sshHostService.fetchSSHHostById(uid!),
    {
      enabled: !!uid,
      staleTime: 5 * 60 * 1000, // 5 minutes
    }
  ),
}

export const useSSHHosts = () => {
  // Get all SSH hosts
  const getAll = () => useQuery(sshHostQueries.all)

  // Get single SSH host
  const getById = (uid: () => string | undefined) => 
    useQuery(() => sshHostQueries.detail(uid()))

  // Create SSH host
  const create = useInvalidateMutation(
    (data: SSHHostRequest) => sshHostService.createSSHHost(data),
    [[...keys.all]]
  )

  // Update SSH host - invalidates both list and detail
  const update = useInvalidateMutation(
    ({ uid, data }: { uid: string; data: SSHHostRequest }) =>
      sshHostService.updateSSHHost(uid, data),
    (_, variables) => [[...keys.all], [...keys.detail(variables.uid)]]
  )

  // Delete SSH host
  const deleteMutation = useInvalidateMutation(
    (uid: string) => sshHostService.deleteSSHHost(uid),
    [[...keys.all]]
  )

  // Test SSH host - no cache invalidation needed
  const testMutation = useSimpleMutation(
    (uid: string) => sshHostService.testSSHHost(uid)
  )

  return {
    queries: { getAll, getById },
    mutations: {
      create,
      update,
      delete: deleteMutation,
      test: testMutation,
    },
  }
}
