import type { SessionResponce } from '../api'

const STORAGE_KEY = 'ahahacraft/session'

export interface PersistedSessionState {
  user: string
  session: SessionResponce
}

const isBrowser = () => typeof window !== 'undefined' && typeof window.localStorage !== 'undefined'

export const loadSessionState = (): PersistedSessionState | null => {
  if (!isBrowser()) {
    return null
  }
  const raw = window.localStorage.getItem(STORAGE_KEY)
  if (!raw) {
    return null
  }
  try {
    const parsed = JSON.parse(raw) as Partial<PersistedSessionState>
    if (
      parsed &&
      typeof parsed.user === 'string' &&
      parsed.session &&
      typeof parsed.session.user_id === 'string' &&
      typeof parsed.session.token === 'string' &&
      typeof parsed.session.role === 'string'
    ) {
      return parsed as PersistedSessionState
    }
  } catch (error) {
    console.warn('Failed to parse session from storage', error)
  }
  return null
}

export const saveSessionState = (state: PersistedSessionState) => {
  if (!isBrowser()) {
    return
  }
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
}

export const clearSessionState = () => {
  if (!isBrowser()) {
    return
  }
  window.localStorage.removeItem(STORAGE_KEY)
}
