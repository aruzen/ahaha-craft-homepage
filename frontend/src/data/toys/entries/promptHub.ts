import type { ToyEntry } from '../../../types/toy'

export const promptHub: ToyEntry = {
  id: 'toy-3',
  title: 'LLMプロンプトテストハブ',
  summary: 'AI向けプロンプトをタグ管理しながら効果測定できるシンプルなWebツール。',
  category: 'blog',
  tags: ['ai', 'react', 'tooling'],
  difficulty: 'beginner',
  lastUpdated: '2025-05-04',
  heroImage: '/resource/toy-space/prompt-hub.png',
  repositoryUrl: 'https://github.com/aruzen/prompt-hub',
  slug: 'prompt-hub',
  content: `## プロダクト概要

- プロンプトと評価結果を1画面で比較
- Markdown形式で履歴を残し、GitHub Gistへエクスポート

## UI構成

- 左: Prompt編集欄 (monaco)
- 右上: 実行設定 (Model, Temp, Seed)
- 右下: 返信ログとタグ付け

## テストの進め方

1. [ ] ベースラインとなるZero-shotを作る
2. [ ] Chain-of-thought版を追加
3. [ ] ユーザー発話ログでA/Bテスト

## React側コード

\`\`\`tsx
const [result, setResult] = useState<Result | null>(null)
const runPrompt = async () => {
  setLoading(true)
  const payload = buildRequest(prompt, settings)
  const data = await invokeLLM(payload)
  setResult(data)
  setLoading(false)
}
\`\`\`

## これから

- Playground共有リンクを発行
- OpenAI Functions向けのテンプレート整備
`,
}
