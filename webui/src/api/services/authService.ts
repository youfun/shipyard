/**
 * Auth Service
 * API service functions for authentication
 */
import apiClient from '../client'
import type { 
  User, 
  LoginStep1Response, 
  LoginStep2Response, 
  ChangePasswordRequest,
  Setup2FAResponse,
  Enable2FARequest,
  Enable2FAResponse,
  Disable2FARequest,
  ApiResponse
} from '../../types'

export interface LoginRequest {
  username: string
  password: string
}

export interface Login2FARequest {
  temp_2fa_token: string
  otp: string
}

export const login = async (data: LoginRequest): Promise<LoginStep1Response> => {
  const response = await apiClient.post<ApiResponse<LoginStep1Response>>('/auth/login', data)
  return response.data.data!
}

export const login2FA = async (data: Login2FARequest): Promise<LoginStep2Response> => {
  const response = await apiClient.post<ApiResponse<LoginStep2Response>>('/auth/login/2fa', data)
  return response.data.data!
}

export const logout = async (): Promise<void> => {
  await apiClient.post('/auth/logout')
}

export const getCurrentUser = async (): Promise<User> => {
  const response = await apiClient.get<ApiResponse<User>>('/auth/me')
  return response.data.data!
}

export const changePassword = async (data: ChangePasswordRequest): Promise<void> => {
  await apiClient.post('/auth/change-password', data)
}

export const setup2FA = async (): Promise<Setup2FAResponse> => {
  const response = await apiClient.post<ApiResponse<Setup2FAResponse>>('/auth/2fa/setup')
  return response.data.data!
}

export const enable2FA = async (data: Enable2FARequest): Promise<Enable2FAResponse> => {
  const response = await apiClient.post<ApiResponse<Enable2FAResponse>>('/auth/2fa/enable', data)
  return response.data.data!
}

export const disable2FA = async (data: Disable2FARequest): Promise<void> => {
  await apiClient.post('/auth/2fa/disable', data)
}