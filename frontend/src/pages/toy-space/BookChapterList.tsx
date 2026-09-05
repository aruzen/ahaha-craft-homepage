import { Link } from 'react-router-dom'
import type { DocNote } from '../../api'
import { bookTreeEntries, buildBookTree, flattenBookTree, type BookTreeNode } from './bookTree'
import './BookChapterList.css'

interface BookChapterListProps {
  vaultSlug: string
  chapters: DocNote[]
  currentSlug?: string
  limit?: number
}

const BookChapterList = ({ vaultSlug, chapters, currentSlug, limit }: BookChapterListProps) => {
  const tree = buildBookTree(chapters)
  const ordered = flattenBookTree(tree)
  const visible = new Set((limit === undefined ? ordered : ordered.slice(0, limit)).map(({ slug }) => slug))
  const chapterNumbers = new Map(ordered.map((chapter, index) => [chapter.slug, chapter.order || index + 1]))

  return (
    <ol className="book-chapter-tree">
      <TreeContents
        node={tree}
        vaultSlug={vaultSlug}
        currentSlug={currentSlug}
        visible={visible}
        chapterNumbers={chapterNumbers}
      />
    </ol>
  )
}

interface TreeContentsProps {
  node: BookTreeNode
  vaultSlug: string
  currentSlug?: string
  visible: Set<string>
  chapterNumbers: Map<string, number>
}

const TreeContents = ({ node, vaultSlug, currentSlug, visible, chapterNumbers }: TreeContentsProps) => (
  <>
    {bookTreeEntries(node).map((entry) => {
      if (entry.kind === 'chapter') {
        const chapter = entry.chapter
        if (!visible.has(chapter.slug)) return null
        return (
          <li key={chapter.slug} className="book-chapter-item">
            <Link
              to={`/toy-space/${vaultSlug}/${chapter.slug}`}
              aria-current={chapter.slug === currentSlug ? 'page' : undefined}
            >
              <span>{chapterNumbers.get(chapter.slug)}</span>
              <strong>{chapter.title}</strong>
            </Link>
          </li>
        )
      }
      const child = entry.section
      const childChapters = flattenBookTree(child).filter(({ slug }) => visible.has(slug))
      if (childChapters.length === 0) return null
      return (
        <li key={child.path} className="book-chapter-section">
          <div>{child.name}</div>
          <ol>
            <TreeContents
              node={child}
              vaultSlug={vaultSlug}
              currentSlug={currentSlug}
              visible={visible}
              chapterNumbers={chapterNumbers}
            />
          </ol>
        </li>
      )
    })}
  </>
)

export default BookChapterList
