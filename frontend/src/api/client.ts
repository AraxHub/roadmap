import type {
  AuthUser,
  CreatedUser,
  HomeView,
  ModuleView,
  Role,
  SubmoduleView,
  AdminTree,
  AdminUser,
  FeedbackItem,
  FeedbackSchedule,
  ContentBlock,
} from './types.ts'

type TokenListener = (token: string | null) => void

let accessToken: string | null = null
const listeners = new Set<TokenListener>()

export function getAccessToken() {
  return accessToken
}

export function setAccessToken(token: string | null) {
  accessToken = token
  listeners.forEach((l) => l(token))
}

export function onAccessTokenChange(listener: TokenListener) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

type ApiError = Error & { status?: number }

async function parseError(res: Response): Promise<ApiError> {
  let message = res.statusText || 'request failed'
  try {
    const data = (await res.json()) as { error?: string }
    if (data.error) message = data.error
  } catch {
    // ignore
  }
  const err = new Error(message) as ApiError
  err.status = res.status
  return err
}

let refreshPromise: Promise<boolean> | null = null

async function tryRefresh(): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = (async () => {
      const res = await fetch('/api/v1/auth/refresh', {
        method: 'POST',
        credentials: 'include',
      })
      if (!res.ok) {
        setAccessToken(null)
        return false
      }
      const data = (await res.json()) as { access_token: string }
      setAccessToken(data.access_token)
      return true
    })().finally(() => {
      refreshPromise = null
    })
  }
  return refreshPromise
}

export async function apiFetch<T>(
  path: string,
  init: RequestInit = {},
  retry = true,
): Promise<T> {
  const headers = new Headers(init.headers)
  if (!headers.has('Content-Type') && init.body) {
    headers.set('Content-Type', 'application/json')
  }
  if (accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`)
  }

  const res = await fetch(path, {
    ...init,
    headers,
    credentials: 'include',
  })

  if (res.status === 401 && retry && !path.includes('/auth/login') && !path.includes('/auth/refresh')) {
    const ok = await tryRefresh()
    if (ok) return apiFetch<T>(path, init, false)
  }

  if (!res.ok) {
    throw await parseError(res)
  }

  if (res.status === 204) {
    return undefined as T
  }

  return (await res.json()) as T
}

export async function login(loginName: string, password: string) {
  const data = await apiFetch<{
    access_token: string
    user: AuthUser
  }>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ login: loginName, password }),
  }, false)
  setAccessToken(data.access_token)
  return data.user
}

export async function logout() {
  try {
    await apiFetch('/api/v1/auth/logout', { method: 'POST' })
  } finally {
    setAccessToken(null)
  }
}

export async function fetchMe() {
  return apiFetch<AuthUser>('/api/v1/auth/me')
}

export async function bootstrapSession(): Promise<AuthUser | null> {
  const ok = await tryRefresh()
  if (!ok) return null
  try {
    return await fetchMe()
  } catch {
    setAccessToken(null)
    return null
  }
}

export function getHome() {
  return apiFetch<HomeView>('/api/v1/home')
}

export function getModule(slug: string) {
  return apiFetch<ModuleView>(`/api/v1/modules/${encodeURIComponent(slug)}`)
}

export function getSubmodule(moduleSlug: string, submoduleSlug: string) {
  return apiFetch<SubmoduleView>(
    `/api/v1/modules/${encodeURIComponent(moduleSlug)}/submodules/${encodeURIComponent(submoduleSlug)}`,
  )
}

export function completeSubmodule(id: string) {
  return apiFetch<{ status: string }>(`/api/v1/submodules/${encodeURIComponent(id)}/complete`, {
    method: 'POST',
  })
}

export function createUser(loginName: string, role: Role) {
  return apiFetch<CreatedUser>('/api/v1/admin/users', {
    method: 'POST',
    body: JSON.stringify({ login: loginName, role }),
  })
}

export function listAdminUsers() {
  return apiFetch<{ users: AdminUser[] }>('/api/v1/admin/users')
}

export function listUserFeedback(userId: string) {
  return apiFetch<{ feedback: FeedbackItem[] }>(
    `/api/v1/admin/users/${encodeURIComponent(userId)}/feedback`,
  )
}

export function requestFeedbackRound() {
  return apiFetch<{ status: string }>('/api/v1/admin/feedback/request-round', { method: 'POST' })
}

export function getFeedbackSchedule() {
  return apiFetch<FeedbackSchedule>('/api/v1/admin/feedback/schedule')
}

export function updateFeedbackMessage(messageText: string) {
  return apiFetch<FeedbackSchedule>('/api/v1/admin/feedback/message', {
    method: 'PUT',
    body: JSON.stringify({ message_text: messageText }),
  })
}

export function requestUserFeedback(userId: string) {
  return apiFetch<{ status: string }>(
    `/api/v1/admin/users/${encodeURIComponent(userId)}/feedback/request`,
    { method: 'POST' },
  )
}

export function getAdminTree() {
  return apiFetch<AdminTree>('/api/v1/admin/tree')
}

export function createSprint(body: {
  title: string
  slug?: string
  description?: string
  position?: number
  is_published: boolean
}) {
  return apiFetch('/api/v1/admin/sprints', { method: 'POST', body: JSON.stringify(body) })
}

export function updateSprint(
  id: string,
  body: {
    title: string
    slug?: string
    description?: string
    position: number
    is_published: boolean
  },
) {
  return apiFetch(`/api/v1/admin/sprints/${id}`, { method: 'PUT', body: JSON.stringify(body) })
}

export function deleteSprint(id: string) {
  return apiFetch(`/api/v1/admin/sprints/${id}`, { method: 'DELETE' })
}

export function createModule(body: {
  sprint_id: string
  title: string
  slug?: string
  description?: string
  position?: number
  is_published: boolean
}) {
  return apiFetch('/api/v1/admin/modules', { method: 'POST', body: JSON.stringify(body) })
}

export function updateModule(
  id: string,
  body: {
    sprint_id?: string
    title: string
    slug?: string
    description?: string
    position: number
    is_published: boolean
  },
) {
  return apiFetch(`/api/v1/admin/modules/${id}`, { method: 'PUT', body: JSON.stringify(body) })
}

export function deleteModule(id: string) {
  return apiFetch(`/api/v1/admin/modules/${id}`, { method: 'DELETE' })
}

export function createSubmodule(body: {
  module_id: string
  title: string
  slug?: string
  position?: number
  is_published: boolean
  body_md?: string
  blocks?: ContentBlock[]
}) {
  return apiFetch('/api/v1/admin/submodules', { method: 'POST', body: JSON.stringify(body) })
}

export function updateSubmodule(
  id: string,
  body: {
    module_id?: string
    title: string
    slug?: string
    position: number
    is_published: boolean
  },
) {
  return apiFetch(`/api/v1/admin/submodules/${id}`, { method: 'PUT', body: JSON.stringify(body) })
}

export function deleteSubmodule(id: string) {
  return apiFetch(`/api/v1/admin/submodules/${id}`, { method: 'DELETE' })
}

export function getSubmoduleContent(id: string) {
  return apiFetch<{ submodule_id: string; blocks: ContentBlock[] }>(
    `/api/v1/admin/submodules/${id}/content`,
  )
}

export function putSubmoduleContent(id: string, blocks: ContentBlock[]) {
  return apiFetch(`/api/v1/admin/submodules/${id}/content`, {
    method: 'PUT',
    body: JSON.stringify({ blocks }),
  })
}

export async function uploadSubmoduleMarkdown(id: string, file: File) {
  const form = new FormData()
  form.append('file', file)
  const headers = new Headers()
  const token = getAccessToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  const res = await fetch(`/api/v1/admin/submodules/${id}/content/upload`, {
    method: 'POST',
    headers,
    body: form,
    credentials: 'include',
  })
  if (!res.ok) {
    const err = new Error((await res.json().catch(() => ({}))).error || res.statusText)
    throw err
  }
  return res.json() as Promise<{ submodule_id: string; blocks: ContentBlock[] }>
}

export async function uploadSubmoduleImage(submoduleId: string, file: File) {
  const form = new FormData()
  form.append('file', file)
  const headers = new Headers()
  const token = getAccessToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  const res = await fetch(`/api/v1/admin/submodules/${submoduleId}/images`, {
    method: 'POST',
    headers,
    body: form,
    credentials: 'include',
  })
  if (!res.ok) {
    const err = new Error((await res.json().catch(() => ({}))).error || res.statusText)
    throw err
  }
  return res.json() as Promise<{ id: string; url: string }>
}

export function deleteContentImage(imageId: string) {
  return apiFetch(`/api/v1/admin/content-images/${imageId}`, { method: 'DELETE' })
}

export function contentImagePath(imageId: string) {
  return `/api/v1/content-images/${imageId}`
}

/** Загрузка картинки с Bearer (для img без Authorization). */
export async function fetchContentImageBlob(imageId: string): Promise<string> {
  const headers = new Headers()
  const token = getAccessToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  const res = await fetch(contentImagePath(imageId), {
    headers,
    credentials: 'include',
  })
  if (!res.ok) {
    throw new Error((await res.json().catch(() => ({}))).error || res.statusText)
  }
  const blob = await res.blob()
  return URL.createObjectURL(blob)
}
