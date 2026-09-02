import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import rehypeHighlight from 'rehype-highlight'
import remarkGfm from 'remark-gfm'
import { ApiError, fetchDocContent, fetchDocNote, getDocAssetUrl, type DocNote, type DocReference } from '../../api'
import './ToyDetail.css'
import 'highlight.js/styles/github-dark.css'

const ToyDetail = () => {
  const { vaultSlug = '', noteSlug = '' } = useParams()
  const [note, setNote] = useState<DocNote | null>(null)
  const [content, setContent] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const controller = new AbortController()
    setIsLoading(true)
    Promise.all([
      fetchDocNote(vaultSlug, noteSlug, { signal: controller.signal }),
      fetchDocContent(vaultSlug, noteSlug, { signal: controller.signal }),
    ])
      .then(([loadedNote, body]) => {
        setNote(loadedNote)
        setContent(body)
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

  return (
    <main className="toy-detail">
      <section className="toy-detail-hero">
        <div>
          <p className="eyebrow">Toy Space</p>
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
        <Link className="toy-detail-back" to="/toy-space">Toy Spaceへ戻る</Link>
      </section>

      {isLoading && <p className="toy-detail-state">読み込み中...</p>}
      {error && <p className="toy-detail-state toy-detail-error">{error}</p>}
      {!isLoading && !error && note && (
        <section className="toy-content">
          {note.content_type === 'html' ? (
            <iframe className="toy-html-frame" title={note.title} sandbox="allow-same-origin" srcDoc={html} />
          ) : (
            <ReactMarkdown
              remarkPlugins={[remarkGfm]}
              rehypePlugins={[rehypeHighlight]}
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
