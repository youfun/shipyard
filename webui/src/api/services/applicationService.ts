/**
 * Applications Service
 * API service functions for applications
 */
import apiClient from '../client'
import type { Application, DeploymentHistory, EnvironmentVariable, Domain, CreateEnvironmentVariableRequest, ApplicationToken, CreateApplicationTokenRequest, CreateApplicationTokenResponse, ApiResponse } from '../../types'

export interface ApplicationsResponse {
  data: Application[]
}

export interface ApplicationResponse {
  data: Application
}

// List all applications
export const fetchApplications = async (): Promise<ApplicationsResponse> => {
  const response = await apiClient.get<ApiResponse<Application[]>>('/applications')
  return { data: response.data.data! }
}

// Get single application
export const fetchApplicationById = async (uid: string): Promise<Application> => {
  const response = await apiClient.get<ApiResponse<Application>>(`/applications/${uid}`)
  return response.data.data!
}

// Get application deployments
export const fetchApplicationDeployments = async (uid: string): Promise<{ data: DeploymentHistory[] }> => {
  const response = await apiClient.get<ApiResponse<DeploymentHistory[]>>(`/applications/${uid}/deployments`)
  return { data: response.data.data! }
}

// Get running deployments
export const fetchRunningDeployments = async (identifier: string): Promise<{ data: any[] }> => {
  const response = await apiClient.get<ApiResponse<any[]>>(`/applications/${identifier}/deployments/running`)
  return { data: response.data.data! }
}

// Create deployment
export const createDeployment = async (uid: string, data: { release_id?: string; rebuild?: boolean }): Promise<void> => {
  await apiClient.post(`/applications/${uid}/deployments`, data)
}

// Get environment variables
export const fetchEnvironmentVariables = async (uid: string): Promise<{ data: EnvironmentVariable[] }> => {
  const response = await apiClient.get<ApiResponse<EnvironmentVariable[]>>(`/applications/${uid}/environment-variables`)
  return { data: response.data.data! }
}

// Create environment variable
export const createEnvironmentVariable = async (uid: string, data: CreateEnvironmentVariableRequest): Promise<EnvironmentVariable> => {
  const response = await apiClient.post<ApiResponse<EnvironmentVariable>>(`/applications/${uid}/environment-variables`, data)
  return response.data.data!
}

// Update environment variable
export const updateEnvironmentVariable = async (envVarId: string, data: Partial<EnvironmentVariable>): Promise<EnvironmentVariable> => {
  const response = await apiClient.put<ApiResponse<EnvironmentVariable>>(`/environment-variables/${envVarId}`, data)
  return response.data.data!
}

// Delete environment variable
export const deleteEnvironmentVariable = async (envVarId: string): Promise<void> => {
  await apiClient.delete(`/environment-variables/${envVarId}`)
}

// Get domains
export const fetchDomains = async (uid: string): Promise<{ data: Domain[] }> => {
  const response = await apiClient.get<ApiResponse<Domain[]>>(`/apps/${uid}/routings`)
  return { data: response.data.data! }
}

// Create domain
export const createDomain = async (uid: string, data: { domainName: string; hostPort: number; isActive: boolean }): Promise<Domain> => {
  const response = await apiClient.post<ApiResponse<Domain>>(`/apps/${uid}/routings`, data)
  return response.data.data!
}

// Update domain
export const updateDomain = async (domainId: string, data: Partial<Domain>): Promise<Domain> => {
  const response = await apiClient.put<Domain>(`/routings/${domainId}`, data)
  return response.data
}

// Delete domain
export const deleteDomain = async (domainId: string): Promise<void> => {
  await apiClient.delete(`/routings/${domainId}`)
}

// Get releases
export const fetchReleases = async (uid: string): Promise<{ data: any[] }> => {
  const response = await apiClient.get<{ data: any[] }>(`/applications/${uid}/releases`)
  return response.data
}

// Delete application
export const deleteApplication = async (uid: string): Promise<void> => {
  await apiClient.delete(`/applications/${uid}`)
}

// Update application
export const updateApplication = async (uid: string, data: Partial<Application>): Promise<Application> => {
  const response = await apiClient.put<Application>(`/applications/${uid}`, data)
  return response.data
}

// Start application
export const startApplication = async (uid: string): Promise<void> => {
  await apiClient.post(`/applications/${uid}/start`)
}

// Stop application
export const stopApplication = async (uid: string): Promise<void> => {
  await apiClient.post(`/applications/${uid}/stop`)
}

// Restart application
export const restartApplication = async (uid: string): Promise<void> => {
  await apiClient.post(`/applications/${uid}/restart`)
}

// Get application tokens
export const fetchApplicationTokens = async (uid: string): Promise<ApplicationToken[]> => {
  const response = await apiClient.get<ApplicationToken[]>(`/applications/${uid}/tokens`)
  return response.data
}

// Create application token
export const createApplicationToken = async (uid: string, data: CreateApplicationTokenRequest): Promise<CreateApplicationTokenResponse> => {
  const response = await apiClient.post<CreateApplicationTokenResponse>(`/applications/${uid}/tokens`, data)
  return response.data
}

// Delete application token
export const deleteApplicationToken = async (uid: string, tokenId: string): Promise<void> => {
  await apiClient.delete(`/applications/${uid}/tokens/${tokenId}`)
}

// Instance Operations
export const startInstance = async (uid: string): Promise<void> => {
  await apiClient.post(`/instances/${uid}/start`)
}

export const stopInstance = async (uid: string): Promise<void> => {
  await apiClient.post(`/instances/${uid}/stop`)
}

export const restartInstance = async (uid: string): Promise<void> => {
  await apiClient.post(`/instances/${uid}/restart`)
}

// Get instance logs
export const fetchInstanceLogs = async (uid: string, lines: number = 500): Promise<{ logs: string }> => {
  const response = await apiClient.get<{ logs: string }>(`/instances/${uid}/logs?lines=${lines}`)
  return response.data
}

// export const fetchInstanceLogs = async (uid: string, lines: number = 500): Promise<{ logs: string }> => {
//   const response = await apiClient.get<{ logs: string }>(`/instances/${uid}/logs?lines=${lines}`)
//   return response.data
// }

export const fetchDeploymentLogs = async (uid: string): Promise<{ logs: string }> => {
  const response = await apiClient.get<{ logs: string }>(`/deployments/${uid}/logs`)
  return response.data
}
