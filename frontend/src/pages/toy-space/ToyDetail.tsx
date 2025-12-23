import { Link, Navigate, useParams } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeHighlight from 'rehype-highlight'
import { toyEntries, toyTags } from '../../data/toys'
import './ToyDetail.css'
import 'highlight.js/styles/github-dark.css'

const ToyDetail = () => {
  const { slug } = useParams<{ slug: string }>()
  const toy = toyEntries.find((entry) => entry.slug === slug)

  if (!toy) {
    return <Navigate to="/toy-space" replace />
  }

  const related = toyEntries.filter((entry) => entry.id !== toy.id && entry.tags.some((tag) => toy.tags.includes(tag))).slice(0, 3)

  return (
    <main className="toy-detail">
      <section className="toy-detail-hero">
        <div>
          <p className="eyebrow">Toy Space</p>
          <h1>{toy.title}</h1>
          <p>{toy.summary}</p>
          <div className="detail-meta">
            <span className={`badge badge-${toy.category}`}>{toy.category}</span>
            <span>更新: {toy.lastUpdated}</span>
            <span>難易度: {toy.difficulty}</span>
            {toy.repositoryUrl && (
              <a href={toy.repositoryUrl} target="_blank" rel="noreferrer">
                Repository ↗
              </a>
            )}
          </div>
          <div className="detail-tags">
            {toy.tags.map((tagId) => {
              const tag = toyTags.find((t) => t.id === tagId)
              return (
                <span key={tagId} className="chip" style={tag?.color ? { borderColor: tag.color } : undefined}>
                  {tag?.label ?? tagId}
                </span>
              )
            })}
          </div>
        </div>
        {toy.heroImage && <img src={toy.heroImage} alt="" loading="lazy" />}
      </section>

      <section className="toy-content">
        <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeHighlight]}>
          {toy.content}
        </ReactMarkdown>
      </section>

      {related.length > 0 && (
        <section className="toy-related">
          <div className="related-header">
            <h2>関連Toy</h2>
            <Link to="/toy-space">一覧へ戻る</Link>
          </div>
          <div className="related-grid">
            {related.map((item) => (
              <Link key={item.id} to={`/toy-space/${item.slug}`} className="related-card">
                <h3>{item.title}</h3>
                <p>{item.summary}</p>
                <span>更新: {item.lastUpdated}</span>
              </Link>
            ))}
          </div>
        </section>
      )}
    </main>
  )
}

export default ToyDetail
