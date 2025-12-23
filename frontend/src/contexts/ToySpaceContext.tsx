import { createContext, useContext, useMemo, useState, type ReactNode } from 'react'
import type { ToyEntry, ToySearchCriteria } from '../types/toy'
import { toyEntries } from '../data/toys'

interface ToySpaceContextValue {
  criteria: ToySearchCriteria
  setQuery: (value: string) => void
  toggleTag: (tagId: string) => void
  setCategory: (value: ToySearchCriteria['category']) => void
  setDifficulty: (value: ToySearchCriteria['difficulty']) => void
  setSortOrder: (value: ToySearchCriteria['sortOrder']) => void
  filteredToys: ToyEntry[]
  resetFilters: () => void
}

const defaultCriteria: ToySearchCriteria = {
  query: '',
  selectedTagIds: [],
  category: 'all',
  difficulty: 'all',
  sortOrder: 'latest',
}

const ToySpaceContext = createContext<ToySpaceContextValue | null>(null)

const sorters: Record<ToySearchCriteria['sortOrder'], (a: ToyEntry, b: ToyEntry) => number> = {
  latest: (a, b) => (a.lastUpdated > b.lastUpdated ? -1 : 1),
  popular: (a, b) => a.title.localeCompare(b.title),
}

const matchesCriteria = (toy: ToyEntry, criteria: ToySearchCriteria) => {
  if (criteria.category !== 'all' && toy.category !== criteria.category) {
    return false
  }

  if (criteria.difficulty !== 'all' && toy.difficulty !== criteria.difficulty) {
    return false
  }

  if (
    criteria.selectedTagIds.length > 0 &&
    !criteria.selectedTagIds.every((tagId) => toy.tags.includes(tagId))
  ) {
    return false
  }

  if (criteria.query.trim().length > 0) {
    const keyword = criteria.query.trim().toLowerCase()
    const haystack = `${toy.title} ${toy.summary}`.toLowerCase()
    if (!haystack.includes(keyword)) {
      return false
    }
  }

  return true
}

export const ToySpaceProvider = ({ children }: { children: ReactNode }) => {
  const [criteria, setCriteria] = useState<ToySearchCriteria>(defaultCriteria)

  const filteredToys = useMemo(() => {
    return toyEntries
      .filter((toy) => matchesCriteria(toy, criteria))
      .sort(sorters[criteria.sortOrder])
  }, [criteria])

  const updateCriteria = (updater: (prev: ToySearchCriteria) => ToySearchCriteria) => {
    setCriteria((prev) => updater(prev))
  }

  const value: ToySpaceContextValue = {
    criteria,
    setQuery: (value) => updateCriteria((prev) => ({ ...prev, query: value })),
    toggleTag: (tagId) =>
      updateCriteria((prev) => {
        const exists = prev.selectedTagIds.includes(tagId)
        return {
          ...prev,
          selectedTagIds: exists
            ? prev.selectedTagIds.filter((id) => id !== tagId)
            : [...prev.selectedTagIds, tagId],
        }
      }),
    setCategory: (value) => updateCriteria((prev) => ({ ...prev, category: value })),
    setDifficulty: (value) => updateCriteria((prev) => ({ ...prev, difficulty: value })),
    setSortOrder: (value) => updateCriteria((prev) => ({ ...prev, sortOrder: value })),
    filteredToys,
    resetFilters: () => setCriteria(defaultCriteria),
  }

  return <ToySpaceContext.Provider value={value}>{children}</ToySpaceContext.Provider>
}

export const useToySpace = () => {
  const context = useContext(ToySpaceContext)
  if (!context) {
    throw new Error('useToySpace must be used within ToySpaceProvider')
  }
  return context
}
