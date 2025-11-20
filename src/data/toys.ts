import type { ToyEntry, ToyTag } from '../types/toy'
import { colorToHex } from './colors'

export const toyTags: ToyTag[] = [
  { id: 'react', label: 'React', color: colorToHex['青'] },
  { id: 'webgl', label: 'WebGL', color: colorToHex['紫'] },
  { id: 'backend', label: 'Backend', color: colorToHex['茶'] },
  { id: 'ai', label: 'AI/ML', color: colorToHex['ピンク'] },
  { id: 'performance', label: 'Performance', color: colorToHex['赤'] },
  { id: 'tooling', label: 'Tooling', color: colorToHex['緑'] },
]

export const toyEntries: ToyEntry[] = [
  {
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
  },
  {
    id: 'toy-2',
    title: 'Rust×Go Hybrid API Gateway',
    summary: '高負荷APIをRustで、業務ロジックをGoで記述するハイブリッド構成の検証。',
    category: 'reference',
    tags: ['backend', 'performance', 'tooling'],
    difficulty: 'advanced',
    lastUpdated: '2025-07-18',
    heroImage: '/resource/toy-space/gateway.png',
    repositoryUrl: 'https://github.com/aruzen/rust-go-gateway',
    slug: 'rust-go-gateway',
  },
  {
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
  },
]
