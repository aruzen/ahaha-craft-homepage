import { useEffect, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { ApiError, fetchDocNotes, type DocNote } from '../../api'
import './Docs.css'

const DocsNotes = () => {
  const { vaultSlug = '' } = useParams()
  const [searchParams] = useSearchParams()
  const [notes, setNotes] = useState<DocNote[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const tag = searchParams.get('tag') ?? undefined
  const group = searchParams.get('group') ?? undefined

  useEffect(() => {
    const controller = new AbortController()
    setIsLoading(true)
    fetchDocNotes(vaultSlug, { tag, group }, { signal: controller.signal })
      .then((response) => {
        setNotes(response.notes ?? [])
        setError(null)
      })
      .catch((err) => {
        if (!controller.signal.aborted) {
          setError(err instanceof ApiError ? err.message : 'ノート一覧の取得に失敗しました')
        }
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setIsLoading(false)
        }
      })
    return () => controller.abort()
  }, [vaultSlug, tag, group])

  return (
    <main className="docs-page">
      <div className="docs-shell">
        <header className="docs-header">
          <div>
            <h2>{vaultSlug}</h2>
            <p>公開ノート一覧</p>
          </div>
          <Link className="docs-back" to="/docs">
            Vault一覧へ
          </Link>
        </header>
        {isLoading && <p className="docs-empty">読み込み中...</p>}
        {error && <p className="docs-error">{error}</p>}
        {!isLoading && !error && notes.length === 0 && (
          <p className="docs-empty">公開中のノートはありません。</p>
        )}
        <div className="docs-note-list">
          {notes.map((note) => (
            <Link key={note.slug} className="docs-note-row" to={`/docs/${vaultSlug}/${note.slug}`}>
              <h3>{note.title}</h3>
              {note.summary && <p>{note.summary}</p>}
              <div className="docs-note-meta">
                {note.group && <span>{note.group}</span>}
                {note.tags.map((tag) => (
                  <span key={tag.slug} className="docs-tag">
                    {tag.name}
                  </span>
                ))}
              </div>
            </Link>
          ))}
        </div>
      </div>
    </main>
  )
}

export default DocsNotes
