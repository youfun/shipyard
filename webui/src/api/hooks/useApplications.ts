/**
 * Applications Hooks
 * TanStack Query hooks for applications
 */
import { useQuery } from '@tanstack/solid-query'
import * as applicationService from '../services/applicationService'
import type { Application, CreateEnvironmentVariableRequest, EnvironmentVariable, Domain, CreateApplicationTokenRequest } from '../../types'
import { createQueryOptions, useInvalidateMutation } from '@api/utils'

const keys = {
  all: ['applications'] as const,
  detail: (uid: string) => ['applications', uid] as const,
  deployments: (uid: string) => ['applications', uid, 'deployments'] as const,
  runningDeployments: (identifier: string) => ['applications', identifier, 'deployments', 'running'] as const,
  envVars: (uid: string) => ['applications', uid, 'environment-variables'] as const,
  domains: (uid: string) => ['applications', uid, 'domains'] as const,
  releases: (uid: string) => ['applications', uid, 'releases'] as const,
  tokens: (uid: string) => ['applications', uid, 'tokens'] as const,
}

// Query options for better type safety and reusability
const applicationQueries = {
  all: () => createQueryOptions(
    keys.all,
    applicationService.fetchApplications,
    { staleTime: 2 * 60 * 1000 } // 2 minutes
  ),
  detail: (uid: string | undefined) => createQueryOptions(
    keys.detail(uid || ''),
    () => applicationService.fetchApplicationById(uid!),
    { 
      enabled: !!uid,
      staleTime: 5 * 60 * 1000, // 5 minutes
    }
  ),
  deployments: (uid: string | undefined) => createQueryOptions(
    keys.deployments(uid || ''),
    () => applicationService.fetchApplicationDeployments(uid!),
    { 
      enabled: !!uid,
      staleTime: 1 * 60 * 1000, // 1 minute - deployments change frequently
    }
  ),
  runningDeployments: (identifier: string | undefined) => createQueryOptions(
    keys.runningDeployments(identifier || ''),
    () => applicationService.fetchRunningDeployments(identifier!),
    { 
      enabled: !!identifier,
      staleTime: 30 * 1000, // 30 seconds - very dynamic data
    }
  ),
  envVars: (uid: string | undefined) => createQueryOptions(
    keys.envVars(uid || ''),
    () => applicationService.fetchEnvironmentVariables(uid!),
    { 
      enabled: !!uid,
      staleTime: 10 * 60 * 1000, // 10 minutes
    }
  ),
  domains: (uid: string | undefined) => createQueryOptions(
    keys.domains(uid || ''),
    () => applicationService.fetchDomains(uid!),
    { 
      enabled: !!uid,
      staleTime: 10 * 60 * 1000, // 10 minutes
    }
  ),
  releases: (uid: string | undefined) => createQueryOptions(
    keys.releases(uid || ''),
    () => applicationService.fetchReleases(uid!),
    { 
      enabled: !!uid,
      staleTime: 5 * 60 * 1000, // 5 minutes
    }
  ),
  tokens: (uid: string | undefined) => createQueryOptions(
    keys.tokens(uid || ''),
    () => applicationService.fetchApplicationTokens(uid!),
    { 
      enabled: !!uid,
      staleTime: 10 * 60 * 1000, // 10 minutes
    }
  ),
}

export const useApplications = () => {
  // Query functions using modern approach
  const getAll = () => useQuery(applicationQueries.all)
  
  const getById = (uid: () => string | undefined) => 
    useQuery(() => applicationQueries.detail(uid()))
  
  const getDeployments = (uid: () => string | undefined) => 
    useQuery(() => applicationQueries.deployments(uid()))
  
  const getRunningDeployments = (identifier: () => string | undefined) => 
    useQuery(() => applicationQueries.runningDeployments(identifier()))
  
  const getEnvVars = (uid: () => string | undefined) => 
    useQuery(() => applicationQueries.envVars(uid()))
  
  const getDomains = (uid: () => string | undefined) => 
    useQuery(() => applicationQueries.domains(uid()))
  
  const getReleases = (uid: () => string | undefined) => 
    useQuery(() => applicationQueries.releases(uid()))
  
  const getTokens = (uid: () => string | undefined) => 
    useQuery(() => applicationQueries.tokens(uid()))

  // Deployment mutations
  const createDeploymentMutation = useInvalidateMutation(
    ({ uid, data }: { uid: string; data: { release_id?: string; rebuild?: boolean } }) =>
      applicationService.createDeployment(uid, data),
    (_, variables) => [[...keys.deployments(variables.uid)]]
  )

  // Environment variable mutations
  const createEnvVarMutation = useInvalidateMutation(
    ({ uid, data }: { uid: string; data: CreateEnvironmentVariableRequest }) =>
      applicationService.createEnvironmentVariable(uid, data),
    (_, variables) => [[...keys.envVars(variables.uid)]]
  )

  const updateEnvVarMutation = useInvalidateMutation(
    ({ envVarId, data, appUid: _appUid }: { envVarId: string; data: Partial<EnvironmentVariable>; appUid: string }) =>
      applicationService.updateEnvironmentVariable(envVarId, data),
    (_, variables) => [[...keys.envVars(variables.appUid)]]
  )

  const deleteEnvVarMutation = useInvalidateMutation(
    ({ envVarId, appUid: _appUid }: { envVarId: string; appUid: string }) =>
      applicationService.deleteEnvironmentVariable(envVarId),
    (_, variables) => [[...keys.envVars(variables.appUid)]]
  )

  // Domain mutations
  const createDomainMutation = useInvalidateMutation(
    ({ uid, data }: { uid: string; data: { domainName: string; hostPort: number; isActive: boolean } }) =>
      applicationService.createDomain(uid, data),
    (_, variables) => [[...keys.domains(variables.uid)]]
  )

  const updateDomainMutation = useInvalidateMutation(
    ({ domainId, data, appUid: _appUid }: { domainId: string; data: Partial<Domain>; appUid: string }) =>
      applicationService.updateDomain(domainId, data),
    (_, variables) => [[...keys.domains(variables.appUid)]]
  )

  const deleteDomainMutation = useInvalidateMutation(
    ({ domainId, appUid: _appUid }: { domainId: string; appUid: string }) =>
      applicationService.deleteDomain(domainId),
    (_, variables) => [[...keys.domains(variables.appUid)]]
  )

  // Application mutations
  const deleteApplicationMutation = useInvalidateMutation(
    (uid: string) => applicationService.deleteApplication(uid),
    [[...keys.all]] // Static invalidation
  )

  const updateApplicationMutation = useInvalidateMutation(
    ({ uid, data }: { uid: string; data: Partial<Application> }) =>
      applicationService.updateApplication(uid, data),
    (_, variables) => [[...keys.detail(variables.uid)], [...keys.all]] // Invalidate both detail and list
  )

  // Application control mutations
  const startApplicationMutation = useInvalidateMutation(
    (uid: string) => applicationService.startApplication(uid),
    (_, uid) => [[...keys.detail(uid)]]
  )

  const stopApplicationMutation = useInvalidateMutation(
    (uid: string) => applicationService.stopApplication(uid),
    (_, uid) => [[...keys.detail(uid)]]
  )

  const restartApplicationMutation = useInvalidateMutation(
    (uid: string) => applicationService.restartApplication(uid),
    (_, uid) => [[...keys.detail(uid)]]
  )

  // Instance control mutations
  const startInstanceMutation = useInvalidateMutation(
    ({ uid, appUid: _appUid }: { uid: string; appUid: string }) => applicationService.startInstance(uid),
    (_, variables) => [[...keys.detail(variables.appUid)]]
  )

  const stopInstanceMutation = useInvalidateMutation(
    ({ uid, appUid: _appUid }: { uid: string; appUid: string }) => applicationService.stopInstance(uid),
    (_, variables) => [[...keys.detail(variables.appUid)]]
  )

  const restartInstanceMutation = useInvalidateMutation(
    ({ uid, appUid: _appUid }: { uid: string; appUid: string }) => applicationService.restartInstance(uid),
    (_, variables) => [[...keys.detail(variables.appUid)]]
  )

  // Token mutations
  const createTokenMutation = useInvalidateMutation(
    ({ uid, data }: { uid: string; data: CreateApplicationTokenRequest }) =>
      applicationService.createApplicationToken(uid, data),
    (_, variables) => [[...keys.tokens(variables.uid)]]
  )

  const deleteTokenMutation = useInvalidateMutation(
    ({ uid, tokenId }: { uid: string; tokenId: string }) =>
      applicationService.deleteApplicationToken(uid, tokenId),
    (_, variables) => [[...keys.tokens(variables.uid)]]
  )

  return {
    queries: {
      getAll,
      getById,
      getDeployments,
      getRunningDeployments,
      getEnvVars,
      getDomains,
      getReleases,
      getTokens,
    },
    mutations: {
      createDeployment: createDeploymentMutation,
      createEnvVar: createEnvVarMutation,
      updateEnvVar: updateEnvVarMutation,
      deleteEnvVar: deleteEnvVarMutation,
      createDomain: createDomainMutation,
      updateDomain: updateDomainMutation,
      deleteDomain: deleteDomainMutation,
      deleteApplication: deleteApplicationMutation,
      updateApplication: updateApplicationMutation,
      startApplication: startApplicationMutation,
      stopApplication: stopApplicationMutation,
      restartApplication: restartApplicationMutation,
      startInstance: startInstanceMutation,
      stopInstance: stopInstanceMutation,
      restartInstance: restartInstanceMutation,
      createToken: createTokenMutation,
      deleteToken: deleteTokenMutation,
    },
  }
}
