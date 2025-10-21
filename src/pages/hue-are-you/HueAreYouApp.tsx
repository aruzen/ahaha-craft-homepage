import React, { useState } from 'react'
import StartScreen from './user/StartScreen'
import SelectionScreen from './user/SelectionScreen'
import ResultScreen from './user/ResultScreen'
import './HueAreYouApp.css'

type Screen = 'start' | 'selection' | 'result'

const HueAreYouApp: React.FC = () => {
  const [currentScreen, setCurrentScreen] = useState<Screen>('start')
  const [assignments, setAssignments] = useState<Record<string, string>>({})
  const [userName, setUserName] = useState('')

  const handleStart = () => {
    setCurrentScreen('selection')
  }

  const handleComplete = (results: Record<string, string>) => {
    setAssignments(results)
    setCurrentScreen('result')
  }

  const handleSave = async (name: string) => {
    setUserName(name)
    // TODO: API呼び出しを実装
    console.log('Saving results:', { name, assignments })
    
    // 模擬API呼び出し
    return new Promise<void>((resolve) => {
      setTimeout(() => {
        resolve()
      }, 1000)
    })
  }

  const handleRestart = () => {
    setAssignments({})
    setCurrentScreen('start')
  }

  const handleBack = () => {
    setCurrentScreen('start')
  }

  return (
    <div className="hue-are-you-container">
      <div className="animated-background">
      </div>
      <div className="hue-are-you-content">
        {currentScreen === 'start' && (
          <StartScreen onStart={handleStart} />
        )}
        {currentScreen === 'selection' && (
          <SelectionScreen 
            onComplete={handleComplete}
            onBack={handleBack}
          />
        )}
        {currentScreen === 'result' && (
          <ResultScreen
            assignments={assignments}
            userName={userName}
            onRestart={handleRestart}
            onSave={handleSave}
          />
        )}
      </div>
    </div>
  )
}

export default HueAreYouApp