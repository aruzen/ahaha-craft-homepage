import type { ToyEntry, ToyTag } from '../types/toy'
import { colorToHex } from './colors'
import { hueVisualizer } from './toys/entries/hueVisualizer'
import { rustGoGateway } from './toys/entries/rustGoGateway'
import { promptHub } from './toys/entries/promptHub'

export const toyTags: ToyTag[] = [
  { id: 'react', label: 'React', color: colorToHex['青'] },
  { id: 'webgl', label: 'WebGL', color: colorToHex['紫'] },
  { id: 'backend', label: 'Backend', color: colorToHex['茶'] },
  { id: 'ai', label: 'AI/ML', color: colorToHex['ピンク'] },
  { id: 'performance', label: 'Performance', color: colorToHex['赤'] },
  { id: 'tooling', label: 'Tooling', color: colorToHex['緑'] },
]

export const toyEntries: ToyEntry[] = [hueVisualizer, rustGoGateway, promptHub]
