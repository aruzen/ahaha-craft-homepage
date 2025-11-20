import { Link, Navigate, useParams } from 'react-router-dom'
import { toyEntries, toyTags } from '../../data/toys'
import type { ToyCategory } from '../../types/toy'
import './ToyDetail.css'

const templateCopy: Record<ToyCategory, { headline: string; sections: string[]; tips: string[] }> = {
  blog: {
    headline: '実装に至る背景と学びを共有',
    sections: ['イントロダクション', '課題と仮説', '実装のポイント', '得られた知見'],
    tips: ['コード片には要点コメントを添える', '感情の動きや失敗談も積極的に書く'],
  },
  reference: {
    headline: 'アーキテクチャの設計根拠を整理',
    sections: ['概要', 'コンポーネント構成', 'ベストプラクティス', '落とし穴 / FAQ'],
    tips: ['図表でフローを可視化', 'API I/Fは例とスキーマをセットで載せる'],
  },
  tutorial: {
    headline: 'ハンズオンで試せる手順書',
    sections: ['準備物', 'ステップ1', 'ステップ2', '発展'],
    tips: ['各ステップの所要時間を明記', '完成イメージを先に提示する'],
  },
}

const ToyDetail = () => {
  const { slug } = useParams<{ slug: string }>()
  const toy = toyEntries.find((entry) => entry.slug === slug)

  if (!toy) {
    return <Navigate to="/toy-space" replace />
  }

  const template = templateCopy[toy.category]
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

      <section className="toy-template">
        <header>
          <h2>{template.headline}</h2>
          <p>このテンプレートをベースに本文を構成すると、読者に伝わりやすくなります。</p>
        </header>
        <div className="template-grid">
          <div>
            <h3>推奨セクション</h3>
            <ol>
              {template.sections.map((section) => (
                <li key={section}>{section}</li>
              ))}
            </ol>
          </div>
          <div>
            <h3>執筆Tips</h3>
            <ul>
              {template.tips.map((tip) => (
                <li key={tip}>{tip}</li>
              ))}
            </ul>
          </div>
        </div>
      </section>

      <section className="toy-outline">
        <h2>本文のアウトライン案</h2>
        <article>
          <h3>1. 概要</h3>
          <p>プロジェクトの目的、背景、使用技術を200文字程度でまとめます。</p>
        </article>
        <article>
          <h3>2. 実装のポイント</h3>
          <p>重要なコードや設計判断を段階ごとに示し、読者が追体験できるようにします。</p>
        </article>
        <article>
          <h3>3. 成果と今後</h3>
          <p>得られた知見や改善予定をリストアップ。GitHub Issueやタスクへのリンクも歓迎。</p>
        </article>
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
