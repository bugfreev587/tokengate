/**
 * User API endpoints
 * Handles user profile management and password changes
 */

import { apiClient } from './client'
import {
  resolveWeChatOAuthStartStrict,
  prepareOAuthBindAccessTokenCookie,
  type WeChatOAuthPublicSettings,
} from './auth'
import type {
  AccountPlatform,
  AccountType,
  User,
  ChangePasswordRequest,
  NotifyEmailEntry,
  UserAuthProvider,
  UserAffiliateDetail,
  AffiliateTransferResponse,
  PaginatedResponse,
} from '@/types'

export interface ConnectedAccountSummary {
  id: number
  name: string
  platform: AccountPlatform
  type: AccountType
  status: 'active' | 'inactive' | 'error' | string
  email?: string
  plan_type?: string
  group_id?: number
  group_name?: string
  capacity_source: 'connected_account' | 'tokengate' | string
  last_used_at?: string | null
  created_at: string
  updated_at: string
}

export interface ConnectedAccountAuthUrlRequest {
  proxy_id?: number | null
  redirect_uri?: string
  project_id?: string
  oauth_type?: string
  tier_id?: string
}

export interface ConnectedAccountAuthUrlResponse {
  auth_url: string
  session_id: string
  state?: string
}

export interface ConnectedAccountExchangeRequest {
  session_id: string
  code: string
  state?: string
  redirect_uri?: string
  proxy_id?: number | null
  name?: string
  concurrency?: number
  priority?: number
  oauth_type?: string
  tier_id?: string
}

export type OpenAIConnectedAccountAuthUrlRequest = ConnectedAccountAuthUrlRequest
export type OpenAIConnectedAccountAuthUrlResponse = ConnectedAccountAuthUrlResponse
export type OpenAIConnectedAccountExchangeRequest = ConnectedAccountExchangeRequest
export type ConnectedAccountOAuthProvider = Extract<AccountPlatform, 'openai' | 'anthropic' | 'gemini'>

/**
 * Get current user profile
 * @returns User profile data
 */
export async function getProfile(): Promise<User> {
  const { data } = await apiClient.get<User>('/user/profile')
  return data
}

/**
 * Update current user profile
 * @param profile - Profile data to update
 * @returns Updated user profile data
 */
export async function updateProfile(profile: {
  username?: string
  avatar_url?: string | null
  balance_notify_enabled?: boolean
  balance_notify_threshold?: number | null
  balance_notify_extra_emails?: NotifyEmailEntry[]
}): Promise<User> {
  const { data } = await apiClient.put<User>('/user', profile)
  return data
}

/**
 * Change current user password
 * @param passwords - Old and new password
 * @returns Success message
 */
export async function changePassword(
  oldPassword: string,
  newPassword: string
): Promise<{ message: string }> {
  const payload: ChangePasswordRequest = {
    old_password: oldPassword,
    new_password: newPassword
  }

  const { data } = await apiClient.put<{ message: string }>('/user/password', payload)
  return data
}

/**
 * Send verification code for adding a notify email
 * @param email - Email address to verify
 */
export async function sendNotifyEmailCode(email: string): Promise<void> {
  await apiClient.post('/user/notify-email/send-code', { email })
}

/**
 * Verify and add a notify email
 * @param email - Email address to add
 * @param code - Verification code
 */
export async function verifyNotifyEmail(email: string, code: string): Promise<void> {
  await apiClient.post('/user/notify-email/verify', { email, code })
}

/**
 * Remove a notify email
 * @param email - Email address to remove
 */
export async function removeNotifyEmail(email: string): Promise<void> {
  await apiClient.delete('/user/notify-email', { data: { email } })
}

/**
 * Toggle a notify email's disabled state
 * @param email - Email address (empty string for primary email placeholder)
 * @param disabled - Whether to disable the email
 */
export async function toggleNotifyEmail(email: string, disabled: boolean): Promise<User> {
  const { data } = await apiClient.put<User>('/user/notify-email/toggle', { email, disabled })
  return data
}

export async function sendEmailBindingCode(email: string): Promise<void> {
  await apiClient.post('/user/account-bindings/email/send-code', { email })
}

export async function bindEmailIdentity(payload: {
  email: string
  verify_code: string
  password: string
}): Promise<User> {
  const { data } = await apiClient.post<User>('/user/account-bindings/email', payload)
  return data
}

export async function unbindAuthIdentity(provider: BindableOAuthProvider): Promise<User> {
  const { data } = await apiClient.delete<User>(`/user/account-bindings/${provider}`)
  return data
}

export type BindableOAuthProvider = Exclude<UserAuthProvider, 'email'>

interface BuildOAuthBindingStartURLOptions {
  redirectTo?: string
  wechatOAuthSettings?: WeChatOAuthPublicSettings | null
}

export function resolveWeChatOAuthMode(): 'open' | 'mp' {
  if (typeof navigator === 'undefined') {
    return 'open'
  }
  return /MicroMessenger/i.test(navigator.userAgent) ? 'mp' : 'open'
}

function resolveWeChatOAuthBindingMode(
  settings?: WeChatOAuthPublicSettings | null
): 'open' | 'mp' | null {
  if (settings) {
    return resolveWeChatOAuthStartStrict(settings).mode
  }
  return resolveWeChatOAuthMode()
}

export function buildOAuthBindingStartURL(
  provider: BindableOAuthProvider,
  options: BuildOAuthBindingStartURLOptions = {}
): string | null {
  const redirectTo = options.redirectTo?.trim() || '/profile'
  const apiBase = (import.meta.env.VITE_API_BASE_URL as string | undefined) || '/api/v1'
  const normalized = apiBase.replace(/\/$/, '')
  const params = new URLSearchParams({
    redirect: redirectTo,
    intent: 'bind_current_user'
  })

  if (provider === 'wechat') {
    const mode = resolveWeChatOAuthBindingMode(options.wechatOAuthSettings)
    if (!mode) {
      return null
    }
    params.set('mode', mode)
  }

  return `${normalized}/auth/oauth/${provider}/bind/start?${params.toString()}`
}

export async function startOAuthBinding(
  provider: BindableOAuthProvider,
  options: BuildOAuthBindingStartURLOptions = {}
): Promise<void> {
  if (typeof window === 'undefined') {
    return
  }
  const startURL = buildOAuthBindingStartURL(provider, options)
  if (!startURL) {
    return
  }
  await prepareOAuthBindAccessTokenCookie()
  window.location.href = startURL
}

export async function getAffiliateDetail(): Promise<UserAffiliateDetail> {
  const { data } = await apiClient.get<UserAffiliateDetail>('/user/aff')
  return data
}

export async function transferAffiliateQuota(): Promise<AffiliateTransferResponse> {
  const { data } = await apiClient.post<AffiliateTransferResponse>('/user/aff/transfer')
  return data
}

export async function listConnectedAccounts(
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<ConnectedAccountSummary>> {
  const { data } = await apiClient.get<PaginatedResponse<ConnectedAccountSummary>>('/user/accounts', {
    params: {
      page,
      page_size: pageSize,
    },
  })
  return data
}

export async function generateOpenAIConnectedAccountAuthUrl(
  payload: OpenAIConnectedAccountAuthUrlRequest = {}
): Promise<ConnectedAccountAuthUrlResponse> {
  return generateConnectedAccountAuthUrl('openai', payload)
}

export async function exchangeOpenAIConnectedAccountCode(
  payload: OpenAIConnectedAccountExchangeRequest
): Promise<ConnectedAccountSummary> {
  return exchangeConnectedAccountCode('openai', payload)
}

export async function generateConnectedAccountAuthUrl(
  provider: ConnectedAccountOAuthProvider,
  payload: ConnectedAccountAuthUrlRequest = {}
): Promise<ConnectedAccountAuthUrlResponse> {
  const requestPayload = sanitizeConnectedAccountOAuthPayload(provider, payload)
  const { data } = await apiClient.post<ConnectedAccountAuthUrlResponse>(
    `/user/accounts/${provider}/auth-url`,
    requestPayload
  )
  return data
}

export async function exchangeConnectedAccountCode(
  provider: ConnectedAccountOAuthProvider,
  payload: ConnectedAccountExchangeRequest
): Promise<ConnectedAccountSummary> {
  const requestPayload = sanitizeConnectedAccountOAuthPayload(provider, payload)
  const { data } = await apiClient.post<ConnectedAccountSummary>(
    `/user/accounts/${provider}/exchange-code`,
    requestPayload
  )
  return data
}

function sanitizeConnectedAccountOAuthPayload<T extends ConnectedAccountAuthUrlRequest | ConnectedAccountExchangeRequest>(
  provider: ConnectedAccountOAuthProvider,
  payload: T
): T {
  if (provider !== 'openai' || !('redirect_uri' in payload)) {
    return payload
  }
  const requestPayload = { ...payload }
  delete requestPayload.redirect_uri
  return requestPayload
}

export async function refreshOpenAIConnectedAccount(id: number): Promise<ConnectedAccountSummary> {
  return refreshConnectedAccount(id)
}

export async function refreshConnectedAccount(id: number): Promise<ConnectedAccountSummary> {
  const { data } = await apiClient.post<ConnectedAccountSummary>(`/user/accounts/${id}/refresh`)
  return data
}

export async function deleteConnectedAccount(id: number): Promise<{ deleted: boolean }> {
  const { data } = await apiClient.delete<{ deleted: boolean }>(`/user/accounts/${id}`)
  return data
}

export const userAPI = {
  getProfile,
  updateProfile,
  changePassword,
  sendNotifyEmailCode,
  verifyNotifyEmail,
  removeNotifyEmail,
  toggleNotifyEmail,
  sendEmailBindingCode,
  bindEmailIdentity,
  unbindAuthIdentity,
  buildOAuthBindingStartURL,
  startOAuthBinding,
  getAffiliateDetail,
  transferAffiliateQuota,
  listConnectedAccounts,
  generateConnectedAccountAuthUrl,
  generateOpenAIConnectedAccountAuthUrl,
  exchangeConnectedAccountCode,
  exchangeOpenAIConnectedAccountCode,
  refreshConnectedAccount,
  refreshOpenAIConnectedAccount,
  deleteConnectedAccount,
}

export default userAPI
