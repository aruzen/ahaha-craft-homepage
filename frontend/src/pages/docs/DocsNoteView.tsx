import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import rehypeHighlight from 'rehype-highlight'
import remarkGfm from 'remark-gfm'
import {
  ApiError,
  fetchDocContent,
  fetchDocNote,
  getDocAssetUrl,
  type DocNote,
  type DocReference,
} from '../../api'
import './Docs.css'

const DocsNoteView = () => {
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
      .then(([noteResponse, body]) => {
        setNote(noteResponse)
        setContent(body)
        setError(null)
      })
      .catch((err) => {
        if (!controller.signal.aborted) {
          setError(err instanceof ApiError ? err.message : 'ノートの取得に失敗しました')
        }
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setIsLoading(false)
        }
      })
    return () => controller.abort()
  }, [vaultSlug, noteSlug])

  const renderedMarkdown = useMemo(() => {
    if (!note) {
      return ''
    }
    return transformObsidianMarkdown(content, vaultSlug, note.metadata.links, note.metadata.embeds)
  }, [content, note, vaultSlug])

  return (
    <main className="docs-page toy-doc-page">
      <div className="docs-shell">
        <header className="docs-header">
          <div>
            <h2>{note?.title ?? noteSlug}</h2>
            {note?.summary && <p>{note.summary}</p>}
          </div>
          <Link className="docs-back" to="/toy-space">
            Toy Spaceへ戻る
          </Link>
        </header>
        {isLoading && <p className="docs-empty">読み込み中...</p>}
        {error && <p className="docs-error">{error}</p>}
        {!isLoading && !error && note && (
          <article className="docs-viewer">
            {note.content_type === 'html' ? (
              <iframe className="docs-html-frame" title={note.title} sandbox="allow-same-origin" srcDoc={content} />
            ) : (
              <div className="docs-markdown">
                <ReactMarkdown
                  remarkPlugins={[remarkGfm]}
                  rehypePlugins={[rehypeHighlight]}
                  components={{
                    a: ({ href, children }) => {
                      const target = href?.startsWith('/toy-space/') ? undefined : '_blank'
                      return (
                        <a href={href} target={target} rel={target ? 'noreferrer' : undefined}>
                          {children}
                        </a>
                      )
                    },
                    img: ({ src, alt }) => (
                      <img src={resolveMarkdownImage(vaultSlug, src)} alt={alt ?? ''} loading="lazy" />
                    ),
                  }}
                >
                  {renderedMarkdown}
                </ReactMarkdown>
              </div>
            )}
          </article>
        )}
      </div>
    </main>
  )
}

const transformObsidianMarkdown = (
  content: string,
  vaultSlug: string,
  links: DocReference[],
  embeds: DocReference[]
) => {
  const linkMap = new Map(links.map((link) => [link.raw, link]))
  const embedMap = new Map(embeds.map((embed) => [embed.raw, embed]))

  return content.replace(/(!?)\[\[([^\]|#]+)(?:#[^\]|]+)?(?:\|([^\]]+))?\]\]/g, (_match, bang, raw, alias) => {
    const label = alias || raw
    if (bang) {
      const embed = embedMap.get(raw)
      if (!embed?.asset_path) {
        return label
      }
      return `![${label}](${getDocAssetUrl(vaultSlug, embed.asset_path)})`
    }
    const link = linkMap.get(raw)
    if (!link?.target_slug) {
      return label
    }
    return `[${label}](/toy-space/${vaultSlug}/${link.target_slug})`
  })
}

const resolveMarkdownImage = (vaultSlug: string, src?: string) => {
  if (!src || /^https?:\/\//i.test(src) || src.startsWith('data:')) {
    return src
  }
  return getDocAssetUrl(vaultSlug, src.replace(/^\.?\//, ''))
}

export default DocsNoteView
