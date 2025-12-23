import { useEffect, useState, type FormEvent, type MouseEvent } from 'react'
import { ApiError, login, signIn, type SessionResponce } from '../api'
import './LoginModal.css'

export type LoginModalState = 'login' | 'signup' | null

interface LoginModalProps {
  modalState: LoginModalState
  onClose: () => void
  onLogin: (payload: { username: string; session: SessionResponce }) => void
}

const LoginModal = ({ modalState, onClose, onLogin }: LoginModalProps) => {
  const [state, setState] = useState<LoginModalState>(modalState)
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    setState(modalState)
    if (modalState !== 'signup') {
      setEmail('')
      setConfirmPassword('')
    }
  }, [modalState])

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (!username.trim() || !password.trim()) {
      alert('ユーザー名とパスワードを入力してください')
      return
    }

    if (state === 'signup' && !email.trim()) {
      alert('メールアドレスを入力してください')
      return
    }

    if (state === 'signup' && password !== confirmPassword) {
      alert('パスワードが一致しません')
      return
    }

    setIsLoading(true)

    try {
      const session =
        state === 'signup'
          ? await signIn({ name: username, email, password })
          : await login({ name: username, password })

      if (!session?.token || !session.role) {
        throw new Error('サーバーから不正なレスポンスを受信しました。')
      }

      onLogin({ username, session })
      onClose()
      setUsername('')
      setEmail('')
      setPassword('')
      setConfirmPassword('')
    } catch (error) {
      if (state === 'signup' && error instanceof ApiError) {
        const fieldLabel = error.field ? `[${error.field}] ` : ''
        const duplicateHint = error.code === 'duplicate' ? '\n同じ情報のアカウントが既に存在します。' : ''
        const detail = error.message || 'サインアップに失敗しました'
        alert(`サインアップに失敗しました: ${fieldLabel}${detail}${duplicateHint}`)
      } else {
        const message = error instanceof Error ? error.message : '予期せぬエラーが発生しました'
        alert(message)
      }
    } finally {
      setIsLoading(false)
    }
  }

  const handleBackdropClick = (e: MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  if (state === null) return null

  return (
    <div className="login-modal-backdrop" onClick={handleBackdropClick}>
      <div className="login-modal">
        <button className="close-btn" onClick={onClose}>
          ×
        </button>
        
        <div className="login-header">
          <h2>{state === 'login' ? 'Login' : 'Sign Up'}</h2>
          <p>{state === 'login' ? 'AhahaCraftにログイン' : 'AhahaCraftに新規登録'}</p>
        </div>

        <form onSubmit={handleSubmit} className="login-form">
          <div className="form-group">
            <label htmlFor="username">ユーザー名</label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="ユーザー名を入力"
              disabled={isLoading}
            />
          </div>

          {state === 'signup' && (
            <div className="form-group">
              <label htmlFor="email">メールアドレス</label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="メールアドレスを入力"
                disabled={isLoading}
              />
            </div>
          )}

          <div className="form-group">
            <label htmlFor="password">パスワード</label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="パスワードを入力"
              disabled={isLoading}
            />
          </div>

          {state === 'signup' && (
            <div className="form-group">
              <label htmlFor="confirmPassword">パスワード確認</label>
              <input
                id="confirmPassword"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="パスワードを再入力"
                disabled={isLoading}
              />
            </div>
          )}

          <button 
            type="submit" 
            className="submit-btn"
            disabled={isLoading}
          >
            {isLoading
              ? '処理中...'
              : state === 'login'
                ? 'ログイン'
                : 'サインアップ'}
          </button>
        </form>

        <div className="switch-mode">
          <p>
            {state === 'login'
              ? 'アカウントをお持ちでない方は'
              : 'すでにアカウントをお持ちの方は'}
            <button
              type="button"
              className="switch-btn"
              onClick={() => (state === 'login' ? setState('signup') : setState('login'))}
              disabled={isLoading}
            >
              {state === 'login' ? 'サインアップ' : 'ログイン'}
            </button>
          </p>
        </div>
      </div>
    </div>
  )
}

export default LoginModal
