import type { DocNote } from '../../api'

export interface BookTreeNode {
  name: string
  path: string
  chapters: DocNote[]
  children: BookTreeNode[]
}

export type BookTreeEntry =
  | { kind: 'chapter'; chapter: DocNote }
  | { kind: 'section'; section: BookTreeNode }

export const buildBookTree = (chapters: DocNote[]): BookTreeNode => {
  const root: BookTreeNode = { name: '', path: '', chapters: [], children: [] }

  for (const chapter of chapters) {
    let current = root
    for (const segment of chapter.chapter_path ?? []) {
      let child = current.children.find(({ name }) => name === segment)
      if (!child) {
        child = {
          name: segment,
          path: [...(current.path ? current.path.split('/') : []), segment].join('/'),
          chapters: [],
          children: [],
        }
        current.children.push(child)
      }
      current = child
    }
    current.chapters.push(chapter)
  }

  sortTree(root)
  return root
}

export const flattenBookTree = (tree: BookTreeNode): DocNote[] =>
  bookTreeEntries(tree).flatMap((entry) => entry.kind === 'chapter' ? [entry.chapter] : flattenBookTree(entry.section))

export const bookTreeEntries = (node: BookTreeNode): BookTreeEntry[] => [
  ...node.chapters.map((chapter): BookTreeEntry => ({ kind: 'chapter', chapter })),
  ...node.children.map((section): BookTreeEntry => ({ kind: 'section', section })),
].sort((a, b) => {
  const firstA = a.kind === 'chapter' ? a.chapter : flattenBookTree(a.section)[0]
  const firstB = b.kind === 'chapter' ? b.chapter : flattenBookTree(b.section)[0]
  if (firstA && firstB) {
    const chapterOrder = compareChapters(firstA, firstB)
    if (chapterOrder !== 0) return chapterOrder
  }
  const nameA = a.kind === 'chapter' ? a.chapter.title : a.section.name
  const nameB = b.kind === 'chapter' ? b.chapter.title : b.section.name
  return nameA.localeCompare(nameB, 'ja')
})

const sortTree = (node: BookTreeNode) => {
  node.chapters.sort(compareChapters)
  node.children.forEach(sortTree)
}

const compareChapters = (a: DocNote, b: DocNote) =>
  a.order - b.order || a.title.localeCompare(b.title, 'ja')
