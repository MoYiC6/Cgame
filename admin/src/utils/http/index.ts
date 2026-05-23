import axios, { AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { useUserStore } from '@/store/modules/user'
import { ApiStatus, isSuccessCode } from './status'
import { HttpError, handleError, showError, showSuccess } from './error'
import { $t } from '@/locales'
import { BaseResponse } from '@/types'
import { AuthRetryConfig, canAttemptAuthRefresh, ensureAccessToken } from './auth-refresh'

const REQUEST_TIMEOUT = 15000
const MAX_RETRIES = 0
const RETRY_DELAY = 1000

interface ExtendedAxiosRequestConfig extends AxiosRequestConfig {
  showErrorMessage?: boolean
  showSuccessMessage?: boolean
  skipAuthRefresh?: boolean
  _retry?: boolean
}

const { VITE_API_URL, VITE_WITH_CREDENTIALS } = import.meta.env

const axiosInstance = axios.create({
  timeout: REQUEST_TIMEOUT,
  baseURL: VITE_API_URL,
  withCredentials: VITE_WITH_CREDENTIALS === 'true',
  validateStatus: (status) => status >= 200 && status < 300,
  transformResponse: [
    (data, headers) => {
      const contentType = headers['content-type']
      if (contentType?.includes('application/json')) {
        try {
          return JSON.parse(data)
        } catch {
          return data
        }
      }
      return data
    }
  ]
})

axiosInstance.interceptors.request.use(
  (request: InternalAxiosRequestConfig) => {
    const { accessToken } = useUserStore()
    if (accessToken) {
      request.headers.set('Authorization', `Bearer ${accessToken}`)
    }

    if (request.data && !(request.data instanceof FormData) && !request.headers['Content-Type']) {
      request.headers.set('Content-Type', 'application/json')
      request.data = JSON.stringify(request.data)
    }

    return request
  },
  (error) => {
    showError(createHttpError($t('httpMsg.requestConfigError'), ApiStatus.error))
    return Promise.reject(error)
  }
)

axiosInstance.interceptors.response.use(
  async (response: AxiosResponse<BaseResponse>) => {
    const { code } = response.data
    const message = response.data.message || response.data.msg

    if (isSuccessCode(code)) return response
    if (code === ApiStatus.unauthorized && canAttemptAuthRefresh(response.config as AuthRetryConfig)) {
      return retryAfterRefresh(response.config as AuthRetryConfig)
    }
    throw createHttpError(message || $t('httpMsg.requestFailed'), normalizeErrorCode(code))
  },
  async (error) => {
    const config = error.config as AuthRetryConfig | undefined
    if (error.response?.status === ApiStatus.unauthorized && canAttemptAuthRefresh(config)) {
      try {
        return await retryAfterRefresh(config)
      } catch (refreshError) {
        await handleUnauthorizedFailure(refreshError)
        return Promise.reject(refreshError)
      }
    }
    return Promise.reject(handleError(error))
  }
)

function createHttpError(message: string, code: number) {
  return new HttpError(message, code)
}

function normalizeErrorCode(code: number | string) {
  return typeof code === 'number' ? code : ApiStatus.error
}

async function retryAfterRefresh(config?: AuthRetryConfig) {
  if (!canAttemptAuthRefresh(config)) {
    const error = createHttpError($t('httpMsg.unauthorized'), ApiStatus.unauthorized)
    await handleUnauthorizedFailure(error)
    throw error
  }

  const nextConfig: AuthRetryConfig = {
    ...(config ?? {}),
    _retry: true
  }

  const accessToken = await ensureAccessToken()
  nextConfig.headers = {
    ...(nextConfig.headers as Record<string, unknown> | undefined),
    Authorization: `Bearer ${accessToken}`
  }

  return axiosInstance.request(nextConfig)
}

async function handleUnauthorizedFailure(error?: unknown): Promise<void> {
  const userStore = useUserStore()
  await userStore.logOut({ revoke: false })
  if (error instanceof HttpError) {
    showError(error, true)
  }
}

function shouldRetry(statusCode: number) {
  return [
    ApiStatus.requestTimeout,
    ApiStatus.internalServerError,
    ApiStatus.badGateway,
    ApiStatus.serviceUnavailable,
    ApiStatus.gatewayTimeout
  ].includes(statusCode)
}

async function retryRequest<T>(
  config: ExtendedAxiosRequestConfig,
  retries: number = MAX_RETRIES
): Promise<T> {
  try {
    return await request<T>(config)
  } catch (error) {
    if (retries > 0 && error instanceof HttpError && shouldRetry(error.code)) {
      await delay(RETRY_DELAY)
      return retryRequest<T>(config, retries - 1)
    }
    throw error
  }
}

function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

async function request<T = any>(config: ExtendedAxiosRequestConfig): Promise<T> {
  if (
    ['POST', 'PUT'].includes(config.method?.toUpperCase() || '') &&
    config.params &&
    !config.data
  ) {
    config.data = config.params
    config.params = undefined
  }

  try {
    const res = await axiosInstance.request<BaseResponse<T>>(config)
    const message = res.data.message || res.data.msg

    if (config.showSuccessMessage && message) {
      showSuccess(message)
    }

    return res.data.data as T
  } catch (error) {
    if (error instanceof HttpError) {
      const showMsg = config.showErrorMessage !== false
      showError(error, showMsg)
    }
    return Promise.reject(error)
  }
}

const api = {
  get<T>(config: ExtendedAxiosRequestConfig) {
    return retryRequest<T>({ ...config, method: 'GET' })
  },
  post<T>(config: ExtendedAxiosRequestConfig) {
    return retryRequest<T>({ ...config, method: 'POST' })
  },
  put<T>(config: ExtendedAxiosRequestConfig) {
    return retryRequest<T>({ ...config, method: 'PUT' })
  },
  del<T>(config: ExtendedAxiosRequestConfig) {
    return retryRequest<T>({ ...config, method: 'DELETE' })
  },
  request<T>(config: ExtendedAxiosRequestConfig) {
    return retryRequest<T>(config)
  }
}

export default api
