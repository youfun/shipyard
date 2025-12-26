export interface User {
  id: string
  username: string
  two_factor_enabled: boolean
}

export interface LoginStep1Response {
  access_token?: string
  two_factor_required?: boolean
  temp_2fa_token?: string
}

export interface LoginStep2Response {
  access_token: string
}

export interface ChangePasswordRequest {
  current_password: string
  new_password: string
}

export interface Setup2FAResponse {
  secret: string
  qr_code_url: string
  recovery_codes: string[]
}

export interface Enable2FARequest {
  secret: string
  otp: string
}

export interface Enable2FAResponse {
  message: string
  recovery_codes: string[]
}

export interface Disable2FARequest {
  password: string
  otp: string
}