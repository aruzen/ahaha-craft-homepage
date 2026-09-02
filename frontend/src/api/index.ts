import type {
  LoginPayload,
  SignInPayload,
  FetchHueAreYouDataParams,
  HueAreYouDataResponse,
  SaveHueAreYouResultPayload,
  SaveHueAreYouResultResponse,
  SessionResponce,
  DocBranchesResponse,
  DocNote,
  DocNotesResponse,
  DocVault,
  DocVaultsResponse,
	DocToysResponse,
  RegisterDocVaultPayload,
  SessionData,
} from './types'

const DEFAULT_DEV_API_BASE_URL = 'http://localhost:8080/api/'
const DEFAULT_PROD_API_BASE_URL = 'https://www.ahaha-craft.org/api/'

const ensureTrailingSlash = (value: string) => (value.endsWith('/') ? value : `${value}/`)

const fallbackBaseUrl = import.meta.env.DEV ? DEFAULT_DEV_API_BASE_URL : DEFAULT_PROD_API_BASE_URL
const rawBaseUrl = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? fallbackBaseUrl

const resolvedBaseUrl = (() => {
  if (/^https?:\/\//i.test(rawBaseUrl)) {
    return ensureTrailingSlash(rawBaseUrl)
  }

  if (typeof window !== 'undefined') {
    const prefix = rawBaseUrl.startsWith('/') ? '' : '/'
    return ensureTrailingSlash(`${window.location.origin}${prefix}${rawBaseUrl}`)
  }

  return ensureTrailingSlash(rawBaseUrl)
})()

const buildUrl = (path: string, searchParams?: Record<string, string | number | undefined>) => {
  const sanitizedPath = path.replace(/^\/+/, '')
  const url = new URL(sanitizedPath, resolvedBaseUrl)

  if (searchParams) {
    Object.entries(searchParams).forEach(([key, value]) => {
      if (value === undefined || value === null) {
        return
      }
      url.searchParams.set(key, String(value))
    })
  }

  return url.toString()
}

interface ApiErrorResponseBody {
  error?: string
  field?: string
  message?: string
}

interface ApiErrorInit {
  status: number
  message: string
  payload?: unknown
  code?: string
  field?: string
}

export class ApiError extends Error {
  status: number
  payload?: unknown
  code?: string
  field?: string

  constructor({ status, message, payload, code, field }: ApiErrorInit) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.payload = payload
    this.code = code
    this.field = field
  }
}

type RequestMethod = 'GET' | 'POST'

interface RequestOptions {
  method?: RequestMethod
  body?: unknown
  searchParams?: Record<string, string | number | undefined>
  signal?: AbortSignal
}

const safeJsonParse = (raw: string) => {
  try {
    return JSON.parse(raw) as unknown
  } catch {
    return null
  }
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, searchParams, signal } = options
  const url = buildUrl(path, searchParams)
  const headers: Record<string, string> = {
    Accept: 'application/json',
  }

  const init: RequestInit = {
    method,
    headers,
    signal,
  }

  if (body !== undefined) {
    init.body = JSON.stringify(body)
    init.headers = {
      ...headers,
      'Content-Type': 'application/json',
    }
  }

  const response = await fetch(url, init)
  const text = await response.text()
  const data = text ? safeJsonParse(text) : null

  if (!response.ok) {
    const fallbackMessage = `API request failed with status ${response.status}`
    let message = fallbackMessage
    let code: string | undefined
    let field: string | undefined

    if (data && typeof data === 'object') {
      const errorBody = data as Partial<ApiErrorResponseBody>
      if (typeof errorBody.message === 'string' && errorBody.message.trim().length > 0) {
        message = errorBody.message
      } else if (typeof errorBody.error === 'string' && errorBody.error.trim().length > 0) {
        message = errorBody.error
      }

      code = typeof errorBody.error === 'string' ? errorBody.error : undefined
      field = typeof errorBody.field === 'string' ? errorBody.field : undefined
    }

    throw new ApiError({
      status: response.status,
      message,
      payload: data ?? undefined,
      code,
      field,
    })
  }

  return (data ?? undefined) as T
}

export const login = async (
  credentials: LoginPayload,
  options?: { signal?: AbortSignal }
): Promise<SessionResponce> =>
  request<SessionResponce>('login', {
    method: 'POST',
    body: credentials,
    signal: options?.signal,
  })

export const signIn = async (
  payload: SignInPayload,
  options?: { signal?: AbortSignal }
): Promise<SessionResponce> =>
  request<SessionResponce>('sign-in', {
    method: 'POST',
    body: payload,
    signal: options?.signal,
  })

export const saveHueAreYouResult = async (
  payload: SaveHueAreYouResultPayload,
  options?: { signal?: AbortSignal }
): Promise<SaveHueAreYouResultResponse> =>
  request<SaveHueAreYouResultResponse>('hue-are-you/save-result', {
    method: 'POST',
    body: payload,
    signal: options?.signal,
  })

export const fetchHueAreYouRecords = async (
  params: FetchHueAreYouDataParams,
  options?: { signal?: AbortSignal }
): Promise<HueAreYouDataResponse> =>
  request<HueAreYouDataResponse>('hue-are-you/get-data', {
    method: 'POST',
    body: {
      session: params.session,
      'data-range': params.dataRange,
    },
    signal: options?.signal,
  })

export const fetchDocVaults = async (
  options?: { signal?: AbortSignal }
): Promise<DocVaultsResponse> =>
  request<DocVaultsResponse>('docs/vaults', {
    signal: options?.signal,
  })

export const fetchDocToys = async (options?: { signal?: AbortSignal }): Promise<DocToysResponse> =>
  request<DocToysResponse>('docs/toys', { signal: options?.signal })

export const fetchDocNotes = async (
  vaultSlug: string,
  params?: { tag?: string; group?: string },
  options?: { signal?: AbortSignal }
): Promise<DocNotesResponse> =>
  request<DocNotesResponse>(`docs/vaults/${vaultSlug}/notes`, {
    searchParams: {
      tag: params?.tag,
      group: params?.group,
    },
    signal: options?.signal,
  })

export const fetchDocNote = async (
  vaultSlug: string,
  noteSlug: string,
  options?: { signal?: AbortSignal }
): Promise<DocNote> =>
  request<DocNote>(`docs/vaults/${vaultSlug}/notes/${noteSlug}`, {
    signal: options?.signal,
  })

export const fetchDocContent = async (
  vaultSlug: string,
  noteSlug: string,
  options?: { signal?: AbortSignal }
): Promise<string> => {
  const response = await fetch(buildUrl(`docs/content/${vaultSlug}/${noteSlug}`), {
    method: 'GET',
    headers: { Accept: 'text/markdown, text/html, text/plain' },
    signal: options?.signal,
  })

  if (!response.ok) {
    throw new ApiError({
      status: response.status,
      message: `API request failed with status ${response.status}`,
    })
  }

  return response.text()
}

export const getDocAssetUrl = (vaultSlug: string, assetPath: string): string =>
  buildUrl(`docs/assets/${vaultSlug}/${assetPath}`)

export const fetchAdminDocBranches = async (
  session: SessionData,
  options?: { signal?: AbortSignal }
): Promise<DocBranchesResponse> =>
  request<DocBranchesResponse>('docs/admin/branches', {
    method: 'POST',
    body: { session },
    signal: options?.signal,
  })

export const fetchAdminDocVaults = async (
  session: SessionData,
  options?: { signal?: AbortSignal }
): Promise<DocVaultsResponse> =>
  request<DocVaultsResponse>('docs/admin/vaults', {
    method: 'POST',
    body: { session },
    signal: options?.signal,
  })

export const registerDocVault = async (
  payload: RegisterDocVaultPayload,
  options?: { signal?: AbortSignal }
): Promise<DocVault> =>
  request<DocVault>('docs/admin/vaults/register', {
    method: 'POST',
    body: payload,
    signal: options?.signal,
  })

export const syncDocVault = async (
  session: SessionData,
  vaultSlug: string,
  options?: { signal?: AbortSignal }
): Promise<void> =>
  request<void>(`docs/admin/vaults/${vaultSlug}/sync`, {
    method: 'POST',
    body: { session },
    signal: options?.signal,
  })

export const rescanDocVault = async (
  session: SessionData,
  vaultSlug: string,
  options?: { signal?: AbortSignal }
): Promise<void> =>
  request<void>(`docs/admin/vaults/${vaultSlug}/rescan`, {
    method: 'POST',
    body: { session },
    signal: options?.signal,
  })

export const setDocDefaultPublished = async (session: SessionData, vaultSlug: string, defaultPublished: boolean): Promise<void> =>
  request<void>(`docs/admin/vaults/${vaultSlug}/settings`, { method: 'POST', body: { session, default_published: defaultPublished } })

export const disableDocVault = async (
  session: SessionData,
  vaultSlug: string,
  options?: { signal?: AbortSignal }
): Promise<void> =>
  request<void>(`docs/admin/vaults/${vaultSlug}/disable`, {
    method: 'POST',
    body: { session },
    signal: options?.signal,
  })

export const overrideDocNotePublished = async (
  session: SessionData,
  vaultSlug: string,
  noteSlug: string,
  published: boolean,
  options?: { signal?: AbortSignal }
): Promise<void> =>
  request<void>(`docs/admin/vaults/${vaultSlug}/notes/${noteSlug}/published`, {
    method: 'POST',
    body: { session, published },
    signal: options?.signal,
  })

export const fetchAdminDocNotes = async (session: SessionData, sourceSlug: string): Promise<DocNotesResponse> =>
  request<DocNotesResponse>(`docs/admin/vaults/${sourceSlug}/notes`, { method: 'POST', body: { session } })

export const updateAdminDocMetadata = async (
  session: SessionData,
  sourceSlug: string,
  noteSlug: string,
  metadata: { title?: string; summary?: string; published?: boolean; order?: number; group?: string; tags?: string[] }
): Promise<void> => request<void>(`docs/admin/vaults/${sourceSlug}/notes/${noteSlug}/metadata`, {
  method: 'POST', body: { session, ...metadata },
})

export const uploadDocContent = async (
  session: SessionData,
  input: { kind: 'note' | 'book'; slug: string; file: File; overwrite: boolean }
): Promise<void> => {
  const form = new FormData()
  form.set('session', JSON.stringify(session)); form.set('kind', input.kind); form.set('slug', input.slug)
  form.set('overwrite', String(input.overwrite)); form.set('file', input.file)
  const response = await fetch(buildUrl('docs/admin/uploads'), { method: 'POST', body: form, headers: { Accept: 'application/json' } })
  const raw = await response.text(); const data = raw ? safeJsonParse(raw) : null
  if (!response.ok) {
    const payload = data as Partial<ApiErrorResponseBody> | null
    throw new ApiError({ status: response.status, message: payload?.message ?? `API request failed with status ${response.status}`, code: payload?.error, field: payload?.field, payload: data })
  }
}

export const trashUploadedDoc = async (session: SessionData, noteSlug: string): Promise<void> =>
  request<void>('docs/admin/uploads/trash', { method: 'POST', body: { session, note_slug: noteSlug } })

export const fetchUploadedDocTrash = async (session: SessionData): Promise<{ items: string[] }> =>
  request<{ items: string[] }>('docs/admin/uploads/trash/list', { method: 'POST', body: { session } })

export const restoreUploadedDoc = async (session: SessionData, path: string): Promise<void> =>
  request<void>('docs/admin/uploads/restore', { method: 'POST', body: { session, path } })

export * from './types'
