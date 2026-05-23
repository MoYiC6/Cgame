import request from '@/utils/http'

export function fetchLogin(params: Api.Auth.LoginParams) {
  return request.post<Api.Auth.LoginResponse>({
    url: '/api/v1/auth/login',
    params,
    skipAuthRefresh: true
  })
}

export function fetchRefresh() {
  return request.post<Api.Auth.LoginResponse>({
    url: '/api/v1/auth/refresh',
    skipAuthRefresh: true,
    showErrorMessage: false
  })
}

export function fetchLogout() {
  return request.post<{ success: boolean }>({
    url: '/api/v1/auth/logout',
    skipAuthRefresh: true,
    showErrorMessage: false
  })
}

export function fetchGetUserInfo() {
  return request.get<Api.Auth.MeResponse>({
    url: '/api/v1/auth/me'
  })
}
