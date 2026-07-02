export const SESSION_BRIDGE_ATTEMPT_KEY = 'tg_session_bridge'

interface BuildSessionBridgeUrlOptions {
  apiBaseUrl?: string
  currentOrigin?: string
  redirectUrl?: string
}

function defaultApiBaseUrl(): string {
  return import.meta.env.VITE_API_BASE_URL || '/api/v1'
}

function defaultOrigin(): string {
  return window.location.origin
}

function defaultHref(): string {
  return window.location.href
}

export function getApiOrigin(
  apiBaseUrl: string = defaultApiBaseUrl(),
  currentOrigin: string = defaultOrigin()
): string | null {
  if (!/^https?:\/\//i.test(apiBaseUrl)) {
    return null
  }

  try {
    return new URL(apiBaseUrl, currentOrigin).origin
  } catch {
    return null
  }
}

export function hasSessionBridgeAttempt(search: string = window.location.search): boolean {
  return new URLSearchParams(search).get(SESSION_BRIDGE_ATTEMPT_KEY) === '1'
}

export function markSessionBridgeAttempt(rawUrl: string): string {
  const url = new URL(rawUrl, defaultOrigin())
  url.searchParams.set(SESSION_BRIDGE_ATTEMPT_KEY, '1')
  return url.toString()
}

export function buildSessionBridgeUrl(options: BuildSessionBridgeUrlOptions = {}): string | null {
  const currentOrigin = options.currentOrigin ?? defaultOrigin()
  const apiOrigin = getApiOrigin(options.apiBaseUrl ?? defaultApiBaseUrl(), currentOrigin)
  if (!apiOrigin || apiOrigin === currentOrigin) {
    return null
  }

  const bridgeUrl = new URL('/auth/session-bridge', apiOrigin)
  bridgeUrl.searchParams.set('redirect', markSessionBridgeAttempt(options.redirectUrl ?? defaultHref()))
  return bridgeUrl.toString()
}

function isTokengateHost(hostname: string): boolean {
  return hostname === 'tokengate.to' || hostname.endsWith('.tokengate.to')
}

function isLocalhost(hostname: string): boolean {
  return hostname === 'localhost' || hostname === '127.0.0.1' || hostname === '::1'
}

function isAllowedRedirect(url: URL): boolean {
  if (url.protocol !== 'https:' && url.protocol !== 'http:') {
    return false
  }

  const currentHost = window.location.hostname
  if (url.hostname === currentHost) {
    return true
  }
  if (isTokengateHost(currentHost)) {
    return isTokengateHost(url.hostname)
  }
  if (isLocalhost(currentHost)) {
    return isLocalhost(url.hostname)
  }
  return false
}

export function normalizeSessionBridgeRedirect(rawRedirect: unknown, fallback = '/login'): string {
  const raw = Array.isArray(rawRedirect) ? rawRedirect[0] : rawRedirect
  const value = typeof raw === 'string' && raw.trim() ? raw : fallback

  let url: URL
  try {
    url = new URL(value, defaultOrigin())
  } catch {
    url = new URL(fallback, defaultOrigin())
  }

  if (!isAllowedRedirect(url)) {
    url = new URL(fallback, defaultOrigin())
  }

  return markSessionBridgeAttempt(url.toString())
}
