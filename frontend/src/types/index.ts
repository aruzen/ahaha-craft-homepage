export interface HueAreYouData {
  name: string
  choice: Record<string, string>
}

export interface UserData {
  timestamp: string
  sessionId: string
}

export interface SaveResultRequest {
  user_data: UserData
  record: HueAreYouData
}

export interface AdminLoginRequest {
  name: string
  hashed_password: string
}

export interface AdminLoginResponse {
  token: string
}

export interface GetDataRequest {
  token: string
  data_range: [number, number]
}

export interface GetDataResponse {
  records: HueAreYouData[]
}

export type Color = 
  | 'ピンク' 
  | '黒' 
  | '灰色' 
  | '白' 
  | '赤' 
  | '青' 
  | '黄色' 
  | 'オレンジ' 
  | '緑' 
  | '紫' 
  | '茶'

export interface WordColorAssignment {
  word: string
  color: Color | null
}