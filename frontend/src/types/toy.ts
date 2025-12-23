export type ToyCategory = 'blog' | 'reference' | 'tutorial'

export type ToyDifficulty = 'beginner' | 'intermediate' | 'advanced'

export interface ToyTag {
  id: string
  label: string
  description?: string
  color?: string
}

export interface ToyEntry {
  id: string
  title: string
  summary: string
  category: ToyCategory
  tags: string[]
  difficulty: ToyDifficulty
  lastUpdated: string
  heroImage?: string
  repositoryUrl?: string
  slug: string
  content: string
}

export interface ToySearchCriteria {
  query: string
  selectedTagIds: string[]
  category: ToyCategory | 'all'
  difficulty: ToyDifficulty | 'all'
  sortOrder: 'latest' | 'popular'
}
