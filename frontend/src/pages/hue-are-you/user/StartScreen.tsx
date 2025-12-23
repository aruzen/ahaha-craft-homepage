import React from 'react'
import './StartScreen.css'

interface StartScreenProps {
  onStart: () => void
}

const StartScreen: React.FC<StartScreenProps> = ({ onStart }) => {
  return (
    <div className="start-screen">
      <div className="start-container">
        <h1> Hue Are You?</h1>
        <p>言葉を色で分別するツールです</p>
        <div className="instructions">
          <h3>使い方</h3>
          <ul>
            <li>画面に表示される言葉を、あなたが感じる色にドラッグ&amp;ドロップしてください</li>
            <li>全部で{import.meta.env.DEV ? '5' : '103'}個の言葉があります</li>
            <li>直感的に、思った色に振り分けてください</li>
            <li>正解・不正解はありません</li>
          </ul>
        </div>
        <button className="start-button" onClick={onStart}>
          スタート
        </button>
      </div>
    </div>
  )
}

export default StartScreen