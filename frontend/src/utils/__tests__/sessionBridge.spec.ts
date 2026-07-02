import { describe, expect, it } from 'vitest'

import {
  SESSION_BRIDGE_ATTEMPT_KEY,
  buildSessionBridgeUrl,
  markSessionBridgeAttempt
} from '@/utils/sessionBridge'

describe('session bridge utils', () => {
  it('builds an api-origin bridge url with a marked return url', () => {
    const redirectUrl = 'https://www.tokengate.to/login?redirect=%2Fkeys'

    const bridgeUrl = buildSessionBridgeUrl({
      apiBaseUrl: 'https://api.tokengate.to/api/v1',
      currentOrigin: 'https://www.tokengate.to',
      redirectUrl
    })

    const parsed = new URL(bridgeUrl!)
    expect(parsed.origin).toBe('https://api.tokengate.to')
    expect(parsed.pathname).toBe('/auth/session-bridge')

    const returnUrl = new URL(parsed.searchParams.get('redirect')!)
    expect(returnUrl.origin).toBe('https://www.tokengate.to')
    expect(returnUrl.pathname).toBe('/login')
    expect(returnUrl.searchParams.get('redirect')).toBe('/keys')
    expect(returnUrl.searchParams.get(SESSION_BRIDGE_ATTEMPT_KEY)).toBe('1')
  })

  it('does not build a bridge when api and app share an origin', () => {
    expect(buildSessionBridgeUrl({
      apiBaseUrl: '/api/v1',
      currentOrigin: 'https://api.tokengate.to',
      redirectUrl: 'https://api.tokengate.to/login'
    })).toBeNull()
  })

  it('marks an existing url as a completed bridge attempt', () => {
    const marked = markSessionBridgeAttempt('https://www.tokengate.to/login?redirect=%2Fkeys')
    expect(new URL(marked).searchParams.get(SESSION_BRIDGE_ATTEMPT_KEY)).toBe('1')
  })
})
