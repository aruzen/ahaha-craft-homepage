export interface HueAreYouRecord {
  name: string
  choice: Record<string, string>
}

export type UserRole = 'admin' | 'user'

export interface SessionData {
  user_id: string
  token: string
}

export interface SessionResponce {
  user_id: string
  token: string
  role: UserRole
}

export interface LoginPayload {
  name: string
  password: string
}

export interface SignInPayload {
  name: string
  email: string
  password: string
}

export type SaveHueAreYouResultPayload = HueAreYouRecord

export interface FetchHueAreYouDataParams {
  session: SessionData
  dataRange: [number, number]
}

export interface HueAreYouDataResponse {
  records: HueAreYouRecord[]
}
