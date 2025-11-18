import type {
  AdminLoginPayload,
  AdminSignInPayload,
  FetchHueAreYouDataParams,
  HueAreYouDataResponse,
  SaveHueAreYouResultPayload,
  SessionData,
} from './types'

const DEFAULT_API_BASE_URL = 'https://www.ahaha-craft.org/api/'

const ensureTrailingSlash = (value: string) => (value.endsWith('/') ? value : `${value}/`)

const rawBaseUrl = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? DEFAULT_API_BASE_URL

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

export class ApiError extends Error {
  status: number
  payload?: unknown

  constructor(status: number, message: string, payload?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.payload = payload
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
    const messageCandidates =
      data && typeof data === 'object'
        ? [
            'message' in data && typeof (data as Record<string, unknown>).message === 'string'
              ? (data as Record<string, string>).message
              : null,
            'error' in data && typeof (data as Record<string, unknown>).error === 'string'
              ? (data as Record<string, string>).error
              : null,
          ]
        : []

    const fallbackMessage = `API request failed with status ${response.status}`
    const message = messageCandidates.find((candidate): candidate is string => Boolean(candidate)) ?? fallbackMessage

    throw new ApiError(response.status, message, data ?? undefined)
  }

  return (data ?? undefined) as T
}

export const loginAdmin = async (
  credentials: AdminLoginPayload,
  options?: { signal?: AbortSignal }
): Promise<SessionData> =>
  request<SessionData>('login', {
    method: 'POST',
    searchParams: {
      name: credentials.name,
      password: credentials.password,
    },
    signal: options?.signal,
  })

export const signInAdmin = async (
  payload: AdminSignInPayload,
  options?: { signal?: AbortSignal }
): Promise<SessionData> =>
  request<SessionData>('sign-in', {
    method: 'POST',
    searchParams: {
      name: payload.name,
      email: payload.email,
      password: payload.password,
    },
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
    searchParams: {
      session: JSON.stringify(params.session),
      'data-range': JSON.stringify(params.dataRange),
    },
    signal: options?.signal,
  })

export * from './types'
