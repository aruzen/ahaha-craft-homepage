import type { ToyEntry } from '../../../types/toy'

export const hueVisualizer: ToyEntry = {
  id: 'toy-1',
  title: '色感覚分析ビジュアライザ',
  summary: 'Hue Are You の集計結果をWebGLで可視化する実験的ダッシュボード。',
  category: 'tutorial',
  tags: ['react', 'webgl', 'performance'],
  difficulty: 'intermediate',
  lastUpdated: '2025-10-02',
  heroImage: '/resource/toy-space/hue-viz.png',
  repositoryUrl: 'https://github.com/aruzen/hue-visualizer',
  slug: 'hue-visualizer',
  content: `## ゴール

- 色分類結果を 3D プロットで俯瞰
- 時系列比較を GPU シェーダで高速描画

## 技術スタック

| Layer | Choice |
| --- | --- |
| Frontend | React + Vite |
| Visualization | regl + GLSL |
| Data Source | Hue API (REST) |

## コード例

\`\`\`ts
const regl = createREGL({ canvas })
const draw = regl({
  frag: shader.fragment,
  vert: shader.vertex,
  attributes: {
    position: positions,
    color: colors,
  },
  count: positions.length / 3,
})

regl.frame(() => {
  draw()
})
\`\`\`

## 次の一手

1. VR空間で色分布を比較
2. 参加者属性とのクロス集計
`,
}
