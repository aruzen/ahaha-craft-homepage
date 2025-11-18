import React, { useState } from 'react'
import { saveHueAreYouResult } from '../../api'
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
    const normalizedName = name.trim() || '匿名'
    const hasAssignments = Object.keys(assignments).length > 0

    if (!hasAssignments) {
      throw new Error('保存できる結果がありません')
    }

    setUserName(normalizedName)

    await saveHueAreYouResult({
      user_name: normalizedName,
      record: {
        name: normalizedName,
        choice: assignments,
      },
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
