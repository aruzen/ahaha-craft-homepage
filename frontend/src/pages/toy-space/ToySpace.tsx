import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { ApiError, fetchDocToys, type DocNote, type DocVault } from '../../api'
import BookChapterList from './BookChapterList'
import './ToySpace.css'

interface PublishedToy {
  vault: DocVault
  note: DocNote
}

const ToySpace = () => {
  const [toys, setToys] = useState<PublishedToy[]>([])
  const [query, setQuery] = useState('')
  const [vaultSlug, setVaultSlug] = useState('all')
  const [selectedTags, setSelectedTags] = useState<string[]>([])
  const [contentType, setContentType] = useState<'all' | 'markdown' | 'html' | 'book'>('all')
  const [sortOrder, setSortOrder] = useState<'latest' | 'title'>('latest')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [expandedBooks, setExpandedBooks] = useState<Set<string>>(() => new Set())

  useEffect(() => {
    const controller = new AbortController()
    setIsLoading(true)
    fetchDocToys({ signal: controller.signal })
      .then(({ toys: loadedToys }) => {
        setToys((loadedToys ?? []).map(({ source, note }) => ({ vault: source, note })))
        setError(null)
      })
      .catch((err) => {
        if (!controller.signal.aborted) {
          setError(err instanceof ApiError ? err.message : 'Toyの取得に失敗しました')
        }
      })
      .finally(() => {
        if (!controller.signal.aborted) setIsLoading(false)
      })
    return () => controller.abort()
  }, [])

  const vaults = useMemo(
    () => Array.from(new Map(toys.map(({ vault }) => [vault.slug, vault])).values()),
    [toys]
  )
  const tags = useMemo(
    () => Array.from(new Map(toys.flatMap(({ note }) => note.tags.map((tag) => [tag.slug, tag]))).values()),
    [toys]
  )
  const filteredToys = useMemo(() => {
    const keyword = query.trim().toLowerCase()
    return toys
      .filter(({ vault, note }) => {
        if (vaultSlug !== 'all' && vault.slug !== vaultSlug) return false
        if (selectedTags.length > 0 && !selectedTags.every((tag) => note.tags.some(({ slug }) => slug === tag))) return false
        if (!keyword) return true
        return `${note.title} ${note.summary} ${note.group ?? ''} ${note.tags.map((tag) => tag.name).join(' ')}`
          .toLowerCase()
          .includes(keyword)
      })
      .sort((a, b) => sortOrder === 'latest'
        ? b.note.updated_at.localeCompare(a.note.updated_at)
        : a.note.title.localeCompare(b.note.title, 'ja'))
  }, [query, selectedTags, sortOrder, toys, vaultSlug])

  const standaloneToys = useMemo(
    () => filteredToys.filter(({ note }) => !note.group && (contentType === 'all' || contentType === note.content_type)),
    [contentType, filteredToys]
  )
  const books = useMemo(() => {
    if (contentType !== 'all' && contentType !== 'book') return []
    const grouped = new Map<string, { vault: DocVault; title: string; chapters: DocNote[] }>()
    for (const { vault, note } of filteredToys) {
      if (!note.group) continue
      const key = `${vault.slug}/${note.group}`
      const book = grouped.get(key) ?? { vault, title: note.group, chapters: [] }
      book.chapters.push(note)
      grouped.set(key, book)
    }
    return Array.from(grouped.values()).map((book) => ({
      ...book,
      chapters: book.chapters.sort((a, b) => a.order - b.order || a.title.localeCompare(b.title, 'ja')),
    }))
  }, [contentType, filteredToys])

  const resetFilters = () => {
    setQuery('')
    setVaultSlug('all')
    setSelectedTags([])
    setContentType('all')
    setSortOrder('latest')
  }

  return (
    <main className="toyspace">
      <section className="toyspace-hero">
        <div>
          <p className="toyspace-eyebrow">TOY SPACE</p>
          <h1>技術スケッチとソースコードの遊び場</h1>
        </div>
      </section>

      <section className="toyspace-panel">
        <div className="search-input-row">
          <input type="search" value={query} onChange={(e) => setQuery(e.target.value)} placeholder="キーワードで検索" />
        </div>
        <div className="filter-row">
          <select value={vaultSlug} onChange={(e) => setVaultSlug(e.target.value)}>
            <option value="all">すべてのソース</option>
            {vaults.map((vault) => <option key={vault.slug} value={vault.slug}>{vault.title}</option>)}
          </select>
          <select value={sortOrder} onChange={(e) => setSortOrder(e.target.value as 'latest' | 'title')}>
            <option value="latest">最新順</option>
            <option value="title">タイトル順</option>
          </select>
          <select value={contentType} onChange={(e) => setContentType(e.target.value as typeof contentType)}>
            <option value="all">すべての形式</option>
            <option value="markdown">Markdown</option>
            <option value="html">HTML</option>
            <option value="book">Book</option>
          </select>
        </div>
        {tags.length > 0 && (
          <div className="tag-grid">
            {tags.map((tag) => {
              const active = selectedTags.includes(tag.slug)
              return (
                <button
                  key={tag.slug}
                  type="button"
                  className={active ? 'tag-chip active' : 'tag-chip'}
                  onClick={() => setSelectedTags((current) => active ? current.filter((slug) => slug !== tag.slug) : [...current, tag.slug])}
                >
                  {tag.name}
                </button>
              )
            })}
          </div>
        )}
      </section>

      <div className="toyspace-stats">
        <div><span>現在の検索語</span><strong>{query || '（未入力）'}</strong></div>
        <div><span>ヒット数</span><strong>{standaloneToys.length + books.length}</strong></div>
        <button type="button" onClick={resetFilters}>条件をクリア</button>
      </div>

      <section className="toyspace-results">
        {isLoading && <p className="empty">読み込み中...</p>}
        {error && <p className="empty toyspace-error">{error}</p>}
        {!isLoading && !error && standaloneToys.length === 0 && books.length === 0 && <p className="empty">公開中のToyがありません。</p>}
        <div className="toy-grid">
          {books.map((book) => {
            const bookKey = `${book.vault.slug}/${book.title}`
            const expanded = expandedBooks.has(bookKey)
            return (
              <article key={bookKey} className="toy-card toy-book">
                <div className="toy-meta">
                  <span className="badge badge-tutorial">Book</span>
                  <span>{book.vault.title}</span>
                </div>
                <h3>{book.title}</h3>
                <div className="toy-book-chapters">
                  <BookChapterList vaultSlug={book.vault.slug} chapters={book.chapters} limit={expanded ? undefined : 3} />
                  {book.chapters.length >= 4 && (
                    <button
                      type="button"
                      className="book-expand-button"
                      aria-expanded={expanded}
                      onClick={() => setExpandedBooks((current) => {
                        const next = new Set(current)
                        if (expanded) next.delete(bookKey)
                        else next.add(bookKey)
                        return next
                      })}
                    >
                      {expanded ? '折り畳む' : `残り${book.chapters.length - 3}章を表示`}
                    </button>
                  )}
                </div>
              </article>
            )
          })}
          {standaloneToys.map(({ vault, note }) => (
            <article key={`${vault.slug}/${note.slug}`} className="toy-card">
              <Link to={`/toy-space/${vault.slug}/${note.slug}`} className="card-link">
                <div className="toy-meta">
                  <span className="badge badge-reference">{vault.title}</span>
                  <span className="date">更新: {new Date(note.updated_at).toLocaleDateString('ja-JP')}</span>
                </div>
                <h3>{note.title}</h3>
                {note.summary && <p>{note.summary}</p>}
                <div className="toy-tags">
                  {note.group && <span className="chip">{note.group}</span>}
                  {note.tags.map((tag) => <span key={tag.slug} className="chip">{tag.name}</span>)}
                </div>
                <div className="toy-footer"><span>{note.content_type}</span><span className="detail-link">詳細を見る →</span></div>
              </Link>
            </article>
          ))}
        </div>
      </section>
    </main>
  )
}

export default ToySpace
