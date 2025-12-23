import React, { useMemo, useState } from 'react'
import { ApiError, type SaveHueAreYouResultResponse } from '../../../api'
import ErrorNotice, { type ErrorDescriptor } from '../../../components/ErrorNotice'
import { colorToHex } from '../../../data/colors'
import './ResultScreen.css'

interface ResultScreenProps {
  assignments: Record<string, string>
  userName: string
  onRestart: () => void
  onSave: (name: string) => Promise<SaveHueAreYouResultResponse>
}

const ResultScreen: React.FC<ResultScreenProps> = ({ 
  assignments, 
  userName, 
  onRestart, 
  onSave 
}) => {
  const [name, setName] = useState(userName)
  const [isSaving, setIsSaving] = useState(false)
  const [resultData, setResultData] = useState<SaveHueAreYouResultResponse | null>(null)
  const [errorNotice, setErrorNotice] = useState<ErrorDescriptor | null>(null)

  const handleSave = async () => {
    setIsSaving(true)
    try {
      const result = await onSave(name.trim())
      setResultData(result)
      setErrorNotice(null)
    } catch (error) {
      if (error instanceof ApiError) {
        setErrorNotice({ message: error.message, field: error.field, code: error.code })
      } else if (error instanceof Error) {
        setErrorNotice({ message: error.message })
      } else {
        setErrorNotice({ message: '保存に失敗しました' })
      }
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
  const hueColor = useMemo(() => {
    if (!resultData) return ''
    const { r, g, b } = resultData.hue
    return `rgb(${r}, ${g}, ${b})`
  }, [resultData])

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
            {isSaving ? '取得中...' : '結果を見る'}
          </button>
          {errorNotice && (
            <div className="save-error">
              <ErrorNotice {...errorNotice} onDismiss={() => setErrorNotice(null)} />
            </div>
          )}
        </div>

        {resultData && (
          <div className="result-detail">
            <div className="result-color" style={{ backgroundColor: hueColor }} />
            <div className="result-detail-body">
              <div className="result-detail-label">あなたの今日の色は</div>
              <div className="result-detail-code">{hueColor}</div>
              <p className="result-detail-message">{resultData.message}</p>
            </div>
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
