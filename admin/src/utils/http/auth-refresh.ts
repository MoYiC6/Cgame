import type { AxiosRequestConfig } from 'axios'
import { fetchRefresh } from '@/api/auth'
import { useUserStore } from '@/store/modules/user'

let refreshingPromise: Promise<string> | null = null

export interface AuthRetryConfig extends AxiosRequestConfig {
  _retry?: boolean
  skipAuthRefresh?: boolean
}

export async function ensureAccessToken(): Promise<string> {
  if (!refreshingPromise) {
    refreshingPromise = refreshAccessToken().finally(() => {
      refreshingPromise = null
    })
  }
  return refreshingPromise
}

function asyncRefreshAllowed(config?: AuthRetryConfig): boolean {
  return !config?.skipAuthRefresh && !config?._retry
}

export function canAttemptAuthRefresh(config?: AuthRetryConfig): boolean {
  return asyncRefreshAllowed(config)
}

async function refreshAccessToken(): Promise<string> {
  const userStore = useUserStore()
  const response = await fetchRefresh()
  userStore.setToken(response.access_token)
  userStore.setLoginStatus(true)
  if (response.user) {
    userStore.setUserInfo({ user: response.user, session_id: userStore.info.sessionId || '' })
  }
  return response.access_token
}
