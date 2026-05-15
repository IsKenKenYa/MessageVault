import request from '@/utils/http'

export function fetchSetupStatus() {
  return request.get<Api.Setup.SetupStatus>({
    url: '/api/setup',
    showErrorMessage: false
  })
}

export function postSetup(data: Api.Setup.SetupRequest) {
  return request.post<{ success: boolean }>({
    url: '/api/setup',
    params: data
  })
}
