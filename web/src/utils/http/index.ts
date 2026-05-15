import axios, {
  AxiosError,
  AxiosRequestConfig,
  AxiosResponse,
  InternalAxiosRequestConfig
} from 'axios'
import { useUserStore } from '@/store/modules/user'
import { ApiStatus } from './status'
import { HttpError, handleError, showError, showSuccess } from './error'
import { $t } from '@/locales'
import { BaseResponse } from '@/types'

const REQUEST_TIMEOUT = 15000
const MAX_RETRIES = 0
const RETRY_DELAY = 1000

interface ExtendedAxiosRequestConfig extends AxiosRequestConfig {
  showErrorMessage?: boolean
  showSuccessMessage?: boolean
  _retry?: boolean
}

const { VITE_API_URL, VITE_WITH_CREDENTIALS } = import.meta.env

const axiosInstance = axios.create({
  timeout: REQUEST_TIMEOUT,
  baseURL: VITE_API_URL,
  withCredentials: VITE_WITH_CREDENTIALS === 'true'
})

let refreshingPromise: Promise<Api.Auth.LoginResponse> | null = null

axiosInstance.interceptors.request.use(
  (request: InternalAxiosRequestConfig) => {
    const userStore = useUserStore()
    if (userStore.accessToken) {
      request.headers.set('Authorization', `Bearer ${userStore.accessToken}`)
    }

    if (request.data && !(request.data instanceof FormData) && !request.headers['Content-Type']) {
      request.headers.set('Content-Type', 'application/json')
    }

    return request
  },
  (error) => {
    showError(new HttpError($t('httpMsg.requestConfigError'), ApiStatus.error))
    return Promise.reject(error)
  }
)

axiosInstance.interceptors.response.use(
  async (response: AxiosResponse<BaseResponse>) => {
    const { code, msg } = response.data
    if (code === ApiStatus.success || code === 201) {
      return response
    }
    if (code === ApiStatus.unauthorized) {
      throw new HttpError(msg || $t('httpMsg.unauthorized'), ApiStatus.unauthorized)
    }
    throw new HttpError(msg || $t('httpMsg.requestFailed'), code || ApiStatus.error, {
      data: response.data,
      url: response.config.url,
      method: response.config.method?.toUpperCase()
    })
  },
  async (
    error: AxiosError<{
      code?: number
      msg?: string
    }>
  ) => {
    const originalConfig = (error.config || {}) as ExtendedAxiosRequestConfig
    const responseStatus = error.response?.status
    const shouldTryRefresh =
      responseStatus === ApiStatus.unauthorized &&
      !originalConfig._retry &&
      !String(originalConfig.url || '').includes('/api/auth/refresh')

    if (shouldTryRefresh) {
      originalConfig._retry = true
      try {
        const authData = await refreshAccessToken()
        originalConfig.headers = {
          ...(originalConfig.headers || {}),
          Authorization: `Bearer ${authData.token}`
        }
        return axiosInstance.request(originalConfig)
      } catch (refreshError) {
        forceLogout()
        return Promise.reject(refreshError)
      }
    }

    if (responseStatus === ApiStatus.unauthorized) {
      forceLogout()
    }
    return Promise.reject(handleError(error as AxiosError<any>))
  }
)

function createHttpError(message: string, code: number) {
  return new HttpError(message, code)
}

async function refreshAccessToken(): Promise<Api.Auth.LoginResponse> {
  if (refreshingPromise) {
    return refreshingPromise
  }

  const userStore = useUserStore()
  if (!userStore.refreshToken) {
    throw createHttpError($t('httpMsg.unauthorized'), ApiStatus.unauthorized)
  }

  refreshingPromise = axios
    .post<BaseResponse<Api.Auth.LoginResponse>>(
      '/api/auth/refresh',
      { refreshToken: userStore.refreshToken },
      {
        baseURL: VITE_API_URL,
        timeout: REQUEST_TIMEOUT,
        withCredentials: VITE_WITH_CREDENTIALS === 'true'
      }
    )
    .then((response) => {
      const payload = response.data.data
      userStore.setToken(payload.token, payload.refreshToken)
      userStore.setLoginStatus(true)
      return payload
    })
    .finally(() => {
      refreshingPromise = null
    })

  return refreshingPromise
}

function forceLogout() {
  useUserStore().logOut()
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
    if (config.showSuccessMessage && res.data.msg) {
      showSuccess(res.data.msg)
    }
    return res.data.data as T
  } catch (error) {
    if (error instanceof HttpError && error.code !== ApiStatus.unauthorized) {
      showError(error, config.showErrorMessage !== false)
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

export { axiosInstance }
export default api
