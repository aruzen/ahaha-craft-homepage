import type {
  LoginPayload,
  SignInPayload,
  FetchHueAreYouDataParams,
  HueAreYouDataResponse,
  SaveHueAreYouResultPayload,
  SessionResponce,
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
): Promise<void> =>
  request<void>('hue-are-you/save-result', {
    method: 'POST',
    body: payload,
    signal: options?.signal,
  })

export const fetchHueAreYouRecords = async (
  params: FetchHueAreYouDataParams,
  options?: { signal?: AbortSignal }
): Promise<HueAreYouDataResponse> =>
  request<HueAreYouDataResponse>('hue-are-you/get-data', {
    method: 'GET',
    body: {
      session: params.session,
      'data-range': params.dataRange,
    },
    signal: options?.signal,
  })

export * from './types'
