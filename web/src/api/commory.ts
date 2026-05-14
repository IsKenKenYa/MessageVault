import request from '@/utils/http'
import axios from 'axios'
import { useUserStore } from '@/store/modules/user'

export function fetchDashboardSummary() {
  return request.get<Api.Commory.DashboardSummary>({
    url: '/api/dashboard'
  })
}

export function fetchTimeline(params: Api.Commory.SearchParams = {}) {
  return request.get<Api.Commory.TimelineItem[]>({
    url: '/api/timeline',
    params
  })
}

export function fetchSearch(params: Api.Commory.SearchParams = {}) {
  return request.get<Api.Commory.TimelineItem[]>({
    url: '/api/search',
    params
  })
}

export function fetchIdentities() {
  return request.get<Api.Commory.Identity[]>({
    url: '/api/identities'
  })
}

export function fetchIdentity(id: string) {
  return request.get<Api.Commory.Identity>({
    url: `/api/identities/${id}`
  })
}

export function fetchThread(id: string) {
  return request.get<Api.Commory.TimelineItem[]>({
    url: `/api/threads/${id}`
  })
}

export function fetchImports() {
  return request.get<Api.Commory.ImportSummary[]>({
    url: '/api/imports'
  })
}

export function uploadImport(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post<{ import_id: string; msglayer_version: string }>({
    url: '/api/imports/upload',
    data: formData
  })
}

export function validateImport(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post<{ valid: boolean }>({
    url: '/api/validate/upload',
    data: formData
  })
}

export function exportImportUrl(importId: string) {
  return `/api/imports/${importId}/export`
}

export async function downloadImport(importId: string) {
  const userStore = useUserStore()
  const response = await axios.get(exportImportUrl(importId), {
    responseType: 'blob',
    headers: {
      Authorization: `Bearer ${userStore.accessToken}`
    }
  })
  const href = URL.createObjectURL(response.data)
  const link = document.createElement('a')
  link.href = href
  link.download = `${importId}.json`
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(href)
}
