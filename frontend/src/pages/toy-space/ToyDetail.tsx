import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import rehypeHighlight from 'rehype-highlight'
import rehypeKatex from 'rehype-katex'
import remarkGfm from 'remark-gfm'
import remarkMath from 'remark-math'
import { ApiError, fetchDocContent, fetchDocNote, fetchDocNotes, getDocAssetUrl, type DocNote, type DocReference } from '../../api'
import BookChapterList from './BookChapterList'
import './ToyDetail.css'
import 'highlight.js/styles/github-dark.css'
import 'katex/dist/katex.min.css'

const ToyDetail = () => {
  const { vaultSlug = '', noteSlug = '' } = useParams()
  const [note, setNote] = useState<DocNote | null>(null)
  const [content, setContent] = useState('')
  const [vaultNotes, setVaultNotes] = useState<DocNote[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const controller = new AbortController()
    setIsLoading(true)
    Promise.all([
      fetchDocNote(vaultSlug, noteSlug, { signal: controller.signal }),
      fetchDocContent(vaultSlug, noteSlug, { signal: controller.signal }),
      fetchDocNotes(vaultSlug, undefined, { signal: controller.signal }).catch(() => ({ notes: [] })),
    ])
      .then(([loadedNote, body, notesResponse]) => {
        setNote(loadedNote)
        setContent(body)
        setVaultNotes(notesResponse.notes ?? [])
        setError(null)
      })
      .catch((err) => {
        if (!controller.signal.aborted) setError(err instanceof ApiError ? err.message : 'Toyの取得に失敗しました')
      })
      .finally(() => {
        if (!controller.signal.aborted) setIsLoading(false)
      })
    return () => controller.abort()
  }, [noteSlug, vaultSlug])

  const markdown = useMemo(
    () => note ? transformObsidianMarkdown(content, vaultSlug, note.metadata.links, note.metadata.embeds) : '',
    [content, note, vaultSlug]
  )
  const html = useMemo(() => note ? injectHtmlBase(content, note.asset_base_url) : content, [content, note])
  const bookChapters = useMemo(
    () => note?.group ? vaultNotes.filter(({ group }) => group === note.group) : [],
    [note, vaultNotes]
  )
  const relatedNotes = useMemo(() => {
    if (!note) return []
    const currentTags = new Set(note.tags.map(({ slug }) => slug))
    return vaultNotes
      .filter((candidate) => candidate.slug !== note.slug && (!note.group || candidate.group !== note.group))
      .map((candidate) => ({
        note: candidate,
        sharedTags: candidate.tags.filter(({ slug }) => currentTags.has(slug)).length,
      }))
      .sort((a, b) => b.sharedTags - a.sharedTags || b.note.updated_at.localeCompare(a.note.updated_at))
      .slice(0, 5)
      .map(({ note: candidate }) => candidate)
  }, [note, vaultNotes])

  return (
    <main className="toy-detail">
      <section className="toy-detail-hero">
        <div>
          <Link className="eyebrow" to="/toy-space">Toy Space</Link>
          <h1>{note?.title ?? noteSlug}</h1>
          {note?.summary && <p>{note.summary}</p>}
          {note && (
            <div className="detail-meta">
              <span className="badge badge-reference">{vaultSlug}</span>
              <span>更新: {new Date(note.updated_at).toLocaleDateString('ja-JP')}</span>
              {note.group && <span>{note.group}</span>}
            </div>
          )}
          {note && <div className="detail-tags">{note.tags.map((tag) => <span key={tag.slug} className="chip">{tag.name}</span>)}</div>}
        </div>
      </section>

      {isLoading && <p className="toy-detail-state">読み込み中...</p>}
      {error && <p className="toy-detail-state toy-detail-error">{error}</p>}
      {!isLoading && !error && note && (
        <div className="toy-detail-layout">
          <section className="toy-content">
            {note.content_type === 'html' ? (
              <iframe className="toy-html-frame" title={note.title} sandbox="allow-same-origin" srcDoc={html} />
            ) : (
              <ReactMarkdown
                remarkPlugins={[remarkGfm, remarkMath]}
                rehypePlugins={[rehypeHighlight, rehypeKatex]}
                components={{
                  a: ({ href, children }) => {
                    const internal = href?.startsWith('/toy-space/')
                    return <a href={href} target={internal ? undefined : '_blank'} rel={internal ? undefined : 'noreferrer'}>{children}</a>
                  },
                  img: ({ src, alt }) => <img src={resolveImage(note.asset_base_url, src)} alt={alt ?? ''} loading="lazy" />,
                }}
              >
                {markdown}
              </ReactMarkdown>
            )}
          </section>
          {(bookChapters.length > 0 || relatedNotes.length > 0) && (
            <aside className="toy-detail-sidebar" aria-label="関連ナビゲーション">
              {bookChapters.length > 0 && (
                <section>
                  <h2>{note.group}</h2>
                  <BookChapterList vaultSlug={vaultSlug} chapters={bookChapters} currentSlug={note.slug} />
                </section>
              )}
              {relatedNotes.length > 0 && (
                <section>
                  <h2>関連ページ</h2>
                  <nav className="toy-related-links">
                    {relatedNotes.map((related) => (
                      <Link key={related.slug} to={`/toy-space/${vaultSlug}/${related.slug}`}>
                        <strong>{related.title}</strong>
                        {related.summary && <span>{related.summary}</span>}
                      </Link>
                    ))}
                  </nav>
                </section>
              )}
            </aside>
          )}
        </div>
      )}
    </main>
  )
}

const transformObsidianMarkdown = (content: string, vaultSlug: string, links: DocReference[], embeds: DocReference[]) => {
  const linkMap = new Map(links.map((link) => [link.raw, link]))
  const embedMap = new Map(embeds.map((embed) => [embed.raw, embed]))
  return content.replace(/(!?)\[\[([^\]|#]+)(?:#[^\]|]+)?(?:\|([^\]]+))?\]\]/g, (_match, bang, raw, alias) => {
    const label = alias || raw
    if (bang) {
      const assetPath = embedMap.get(raw)?.asset_path
      return assetPath ? `![${label}](${getDocAssetUrl(vaultSlug, assetPath)})` : label
    }
    const targetSlug = linkMap.get(raw)?.target_slug
    return targetSlug ? `[${label}](/toy-space/${vaultSlug}/${targetSlug})` : label
  })
}

const resolveImage = (assetBaseUrl: string, src?: string) => {
  if (!src || /^https?:\/\//i.test(src) || src.startsWith('data:')) return src
	if (src.startsWith('/api/docs/assets/')) return src
  return new URL(src.replace(/^\.\//, ''), `${window.location.origin}${assetBaseUrl}`).toString()
}

const injectHtmlBase = (content: string, assetBaseUrl: string) => {
	const base = `<base href="${assetBaseUrl}">`
	return /<head[\s>]/i.test(content) ? content.replace(/<head([^>]*)>/i, `<head$1>${base}`) : `${base}${content}`
}

export default ToyDetail
