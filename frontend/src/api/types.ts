export interface HueAreYouRecord {
  name: string
  choice: Record<string, string>
}

export interface HueValue {
  r: number
  g: number
  b: number
}

export interface HueAreYouResultResponse {
  hue: HueValue
  message: string
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
export type SaveHueAreYouResultResponse = HueAreYouResultResponse

export interface FetchHueAreYouDataParams {
  session: SessionData
  dataRange: [number, number]
}

export interface HueAreYouDataResponse {
  records: HueAreYouRecord[]
}

export interface DocVault {
  slug: string
  title: string
  branch?: string
  local_path?: string
  status?: string
  last_synced_at?: string
	 source_type: 'git_vault' | 'local_upload'
	default_published?: boolean
}

export interface DocTag {
  slug: string
  name: string
}

export interface DocReference {
  raw: string
  label?: string
  target_slug?: string
  asset_path?: string
}

export interface DocNoteMetadata {
  links: DocReference[]
  embeds: DocReference[]
}

export interface DocNote {
  slug: string
  title: string
  summary: string
  content_type: 'markdown' | 'html'
  published?: boolean
  order: number
  group?: string
  tags: DocTag[]
  metadata: DocNoteMetadata
  updated_at: string
  content_url: string
	asset_base_url: string
}

export interface DocVaultsResponse {
  vaults: DocVault[]
}

export interface DocNotesResponse {
  notes: DocNote[]
}

export interface DocToy { source: DocVault; note: DocNote }
export interface DocToysResponse { toys: DocToy[] }

export interface DocBranchesResponse {
  branches: string[]
}

export interface RegisterDocVaultPayload {
  session: SessionData
  branch: string
  slug?: string
  title?: string
}
