import React, { useState } from 'react'
import { colorToHex } from '../../../data/colors'
import './ResultScreen.css'

interface ResultScreenProps {
  assignments: Record<string, string>
  userName: string
  onRestart: () => void
  onSave: (name: string) => Promise<void>
}

const ResultScreen: React.FC<ResultScreenProps> = ({ 
  assignments, 
  userName, 
  onRestart, 
  onSave 
}) => {
  const [name, setName] = useState(userName)
  const [isSaving, setIsSaving] = useState(false)
  const [isSaved, setIsSaved] = useState(false)

  const handleSave = async () => {
    if (!name.trim()) {
      alert('名前を入力してください')
      return
    }
    
    setIsSaving(true)
    try {
      await onSave(name.trim())
      setIsSaved(true)
    } catch (error) {
      const message = error instanceof Error ? error.message : '保存に失敗しました'
      alert(message)
    } finally {
      setIsSaving(false)
    }
  }

  const groupedByColor = Object.entries(assignments).reduce((acc, [word, color]) => {
    if (!acc[color]) acc[color] = []
    acc[color].push(word)
    return acc
  }, {} as Record<string, string[]>)

  const totalWords = Object.keys(assignments).length

  return (
    <div className="result-screen">
      <div className="result-header">
        <h1>結果</h1>
        <p>全{totalWords}語の分類が完了しました！</p>
      </div>

      <div className="result-summary">
        {Object.entries(groupedByColor).map(([color, words]) => (
          <div key={color} className="color-group">
            <div className="color-header">
              <div 
                className="color-indicator"
                style={{ 
                  backgroundColor: colorToHex[color as keyof typeof colorToHex],
                  border: color === '白' ? '2px solid #ccc' : 'none'
                }}
              />
              <span className="color-name">{color}</span>
              <span className="word-count">({words.length}語)</span>
            </div>
            <div className="word-list">
              {words.map((word, index) => (
                <span key={index} className="word-tag">
                  {word}
                </span>
              ))}
            </div>
          </div>
        ))}
      </div>

      <div className="result-actions">
        {!isSaved && (
          <div className="save-section">
            <div className="name-input-group">
              <label htmlFor="name">名前（オプション）:</label>
              <input
                id="name"
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="匿名"
                disabled={isSaving}
              />
            </div>
            <button 
              className="save-button"
              onClick={handleSave}
              disabled={isSaving}
            >
              {isSaving ? '保存中...' : '結果を保存'}
            </button>
          </div>
        )}

        {isSaved && (
          <div className="save-success">
            ✓ 結果が保存されました
          </div>
        )}

        <div className="action-buttons">
          <button className="restart-button" onClick={onRestart}>
            もう一度やる
          </button>
          <button 
            className="home-button" 
            onClick={() => window.location.href = '/'}
          >
            ホームに戻る
          </button>
        </div>
      </div>
    </div>
  )
}

export default ResultScreen
