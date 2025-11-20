import { Link } from 'react-router-dom'
import { toyEntries, toyTags } from '../../data/toys'
import { useToySpace } from '../../contexts/ToySpaceContext'
import type { ToyCategory, ToyDifficulty } from '../../types/toy'
import './ToySpace.css'

const ToySpace = () => {
  const {
    criteria,
    setQuery,
    toggleTag,
    setCategory,
    setDifficulty,
    setSortOrder,
    filteredToys,
    resetFilters,
  } = useToySpace()

  return (
    <main className="toyspace">
      <section className="toyspace-hero">
        <div>
          <p className="toyspace-eyebrow">TOY SPACE</p>
          <h1>技術スケッチと小さなプロジェクトの遊び場</h1>
          <p>
            {toyEntries.length}件のToyと{toyTags.length}種類のタグを横断検索。キーワード、カテゴリ、難易度、タグの組み合わせで最短ルートにたどり着けます。
          </p>
        </div>
        <div className="toyspace-stats">
          <div>
            <span>現在の検索語</span>
            <strong>{criteria.query || '（未入力）'}</strong>
          </div>
          <div>
            <span>ヒット数</span>
            <strong>{filteredToys.length}</strong>
          </div>
          <button type="button" onClick={resetFilters}>
            条件をクリア
          </button>
        </div>
      </section>

      <ToySearchPanel
        query={criteria.query}
        onQueryChange={setQuery}
        category={criteria.category}
        onCategoryChange={setCategory}
        difficulty={criteria.difficulty}
        onDifficultyChange={setDifficulty}
        sortOrder={criteria.sortOrder}
        onSortOrderChange={setSortOrder}
        selectedTags={criteria.selectedTagIds}
        onToggleTag={toggleTag}
      />

      <ToyResults toys={filteredToys} />
      <ToyCTASection />
    </main>
  )
}

interface ToySearchPanelProps {
  query: string
  onQueryChange: (value: string) => void
  category: ToyCategory | 'all'
  onCategoryChange: (value: ToyCategory | 'all') => void
  difficulty: ToyDifficulty | 'all'
  onDifficultyChange: (value: ToyDifficulty | 'all') => void
  sortOrder: 'latest' | 'popular'
  onSortOrderChange: (value: 'latest' | 'popular') => void
  selectedTags: string[]
  onToggleTag: (tagId: string) => void
}

const categoryOptions: { value: ToyCategory | 'all'; label: string }[] = [
  { value: 'all', label: 'すべて' },
  { value: 'blog', label: 'Blog' },
  { value: 'reference', label: 'Reference' },
  { value: 'tutorial', label: 'Tutorial' },
]

const difficultyOptions: { value: ToyDifficulty | 'all'; label: string }[] = [
  { value: 'all', label: '難易度：すべて' },
  { value: 'beginner', label: 'Beginner' },
  { value: 'intermediate', label: 'Intermediate' },
  { value: 'advanced', label: 'Advanced' },
]

const ToySearchPanel = ({
  query,
  onQueryChange,
  category,
  onCategoryChange,
  difficulty,
  onDifficultyChange,
  sortOrder,
  onSortOrderChange,
  selectedTags,
  onToggleTag,
}: ToySearchPanelProps) => (
  <section className="toyspace-panel">
    <div className="search-row">
      <input
        type="search"
        value={query}
        onChange={(e) => onQueryChange(e.target.value)}
        placeholder="キーワードで検索（例: WebGL, AI, Rust）"
      />
      <select value={category} onChange={(e) => onCategoryChange(e.target.value as ToyCategory | 'all')}>
        {categoryOptions.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      <select
        value={difficulty}
        onChange={(e) => onDifficultyChange(e.target.value as ToyDifficulty | 'all')}
      >
        {difficultyOptions.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      <select value={sortOrder} onChange={(e) => onSortOrderChange(e.target.value as 'latest' | 'popular')}>
        <option value="latest">最新順</option>
        <option value="popular">人気順</option>
      </select>
    </div>
    <div className="tag-grid">
      {toyTags.map((tag) => {
        const isActive = selectedTags.includes(tag.id)
        return (
          <button
            key={tag.id}
            type="button"
            className={isActive ? 'tag-chip active' : 'tag-chip'}
            style={tag.color ? { borderColor: tag.color } : undefined}
            onClick={() => onToggleTag(tag.id)}
          >
            <span>{tag.label}</span>
          </button>
        )
      })}
    </div>
  </section>
)

const ToyResults = ({ toys }: { toys: typeof toyEntries }) => (
  <section className="toyspace-results">
    {toys.length === 0 ? (
      <p className="empty">条件に一致するToyが見つかりません。タグやキーワードを見直してください。</p>
    ) : (
      <div className="toy-grid">
        {toys.map((toy) => (
          <article key={toy.id} className="toy-card">
            <Link to={`/toy-space/${toy.slug}`} className="card-link">
              {toy.heroImage && (
                <img src={toy.heroImage} alt="" loading="lazy" />
              )}
              <div className="toy-meta">
                <span className={`badge badge-${toy.category}`}>{toy.category}</span>
                <span className="date">更新: {toy.lastUpdated}</span>
              </div>
              <h3>{toy.title}</h3>
              <p>{toy.summary}</p>
              <div className="toy-tags">
                {toy.tags.map((tagId) => {
                  const tag = toyTags.find((t) => t.id === tagId)
                  return (
                    <span key={tagId} className="chip" style={tag?.color ? { borderColor: tag.color } : undefined}>
                      {tag?.label ?? tagId}
                    </span>
                  )
                })}
              </div>
              <div className="toy-footer">
                <span className="difficulty">難易度: {toy.difficulty}</span>
                <span className="detail-link">詳細を見る →</span>
              </div>
            </Link>
          </article>
        ))}
      </div>
    )}
  </section>
)

const ToyCTASection = () => (
  <section className="toyspace-cta">
    <div>
      <h2>最新のToyをメールで受け取る</h2>
      <p>月1回、実験ログと技術ノートをダイジェストでお届けします。</p>
      <form
        onSubmit={(e) => {
          e.preventDefault()
          alert('ありがとうございます！ニュースレターに登録しました。')
        }}
      >
        <input type="email" placeholder="you@example.com" required />
        <button type="submit">登録する</button>
      </form>
    </div>
    <div>
      <h3>あなたのToyも掲載しませんか？</h3>
      <p>試作コード、技術検証、ブログ草稿などを募集しています。軽いメモでも歓迎です。</p>
      <a className="cta-link" href="mailto:contact@ahaha-craft.org">
        投稿の相談をする
      </a>
    </div>
  </section>
)

export default ToySpace
