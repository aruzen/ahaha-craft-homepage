import { useState } from 'react'
import HueAreYouApp from './pages/hue-are-you/HueAreYouApp'
import Portfolio from './pages/portfolio/Portfolio'
import LoginModal, { LoginModalState } from './components/LoginModal'
import './App.css'

type CurrentPage = 'home' | 'hue-are-you' | 'toy-space' | 'contact' | 'portfolio'

function App() {
  const [currentPage, setCurrentPage] = useState<CurrentPage>('home')
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const [modalState, setLoginModalState] = useState<LoginModalState>(null)
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [user, setUser] = useState<string | null>(null)

  const renderPage = () => {
    switch (currentPage) {
      case 'hue-are-you':
        return <HueAreYouApp />
      case 'portfolio':
        return <Portfolio />
      case 'home':
      default:
        return (
          <main>
            <section id="home">
              <h2>Welcome</h2>
              <p>このサイトは現在開発中です。</p>
              <div style={{ marginTop: '2rem' }}>
                <button
                  style={{
                    padding: '1rem 2rem',
                    fontSize: '1.2rem',
                    background: 'linear-gradient(45deg, #4ecdc4, #45b7d1)',
                    border: 'none',
                    borderRadius: '8px',
                    color: 'white',
                    cursor: 'pointer',
                    marginRight: '1rem'
                  }}
                  onClick={() => setCurrentPage('hue-are-you')}
                >
                  Hue Are You? を試す
                </button>
                <button
                  style={{
                    padding: '1rem 2rem',
                    fontSize: '1.2rem',
                    background: 'linear-gradient(45deg, #e74c3c, #c0392b)',
                    border: 'none',
                    borderRadius: '8px',
                    color: 'white',
                    cursor: 'pointer'
                  }}
                  onClick={() => setCurrentPage('portfolio')}
                >
                  ポートフォリオを見る
                </button>
              </div>
            </section>
          </main>
        )
    }
  }

  return (
    <div className="App">
      <header className="App-header">
        <div className="header-content">
          <div className="header-left">
            <img
              src="/resource/ahahacraft.png"
              alt="AhahaCraft Logo"
              className="logo"
            />
            <h1>AhahaCraft</h1>
          </div>
          <div className="nav-container">
            <button
              className={`hamburger-menu ${isMobileMenuOpen ? 'active' : ''}`}
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
              aria-label="メニュー"
            >
              <span></span>
              <span></span>
              <span></span>
            </button>
            <nav className={`nav ${isMobileMenuOpen ? 'nav-open' : ''}`}>
              <ul>
                <li>
                  <a
                    href="#home"
                    onClick={(e) => { e.preventDefault(); setCurrentPage('home'); setIsMobileMenuOpen(false) }}
                    className={currentPage === 'home' ? 'active' : ''}
                  >
                    ホーム
                  </a>
                </li>
                <li>
                  <a
                    href="#hue-are-you"
                    onClick={(e) => { e.preventDefault(); setCurrentPage('hue-are-you'); setIsMobileMenuOpen(false) }}
                    className={currentPage === 'hue-are-you' ? 'active' : ''}
                  >
                    Hue Are You?
                  </a>
                </li>
                <li>
                  <a
                    href="#portfolio"
                    onClick={(e) => { e.preventDefault(); setCurrentPage('portfolio'); setIsMobileMenuOpen(false) }}
                    className={currentPage === 'portfolio' ? 'active' : ''}
                  >
                    ポートフォリオ
                  </a>
                </li>
                <li>
                  <a
                    href="#toy-space"
                    onClick={(e) => { e.preventDefault(); setCurrentPage('toy-space'); setIsMobileMenuOpen(false) }}
                    className={currentPage === 'toy-space' ? 'active' : ''}
                  >
                    Toy Space
                  </a>
                </li>
                <li>
                  <a
                    href="#contact"
                    onClick={(e) => { e.preventDefault(); setCurrentPage('contact'); setIsMobileMenuOpen(false) }}
                    className={currentPage === 'contact' ? 'active' : ''}
                  >
                    コンタクト
                  </a>
                </li>
                <li><a href="https://github.com" target="_blank" rel="noopener noreferrer">GitHub</a></li>
                <li>
                  <div className="auth-buttons">
                    {isAuthenticated ? (
                      <div className="user-info">´
                        <span className="user-name">Welcome, {user}</span>
                        <button
                          className="logout-btn"
                          onClick={() => { setIsAuthenticated(false); setUser(null) }}
                        >
                          Logout
                        </button>
                      </div>
                    ) : (
                      <>
                        <button
                          className="signin-btn"
                          onClick={() => setLoginModalState('signup')}
                        >
                          Sign In
                        </button>
                        <button
                          className="login-btn"
                          onClick={() => setLoginModalState('login')}
                        >
                          Login
                        </button>
                      </>
                    )}
                  </div>
                </li>
              </ul>
            </nav>
          </div>
        </div>
      </header>
      {renderPage()}
      <LoginModal
        modalState={modalState}
        onClose={() => setLoginModalState(null)}
        onLogin={(username) => {
          setIsAuthenticated(true)
          setUser(username)
        }}
      />
    </div>
  )
}

export default App