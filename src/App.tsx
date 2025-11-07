import { useEffect, useState } from 'react'
import { Navigate, NavLink, Route, Routes, useLocation } from 'react-router-dom'
import LoginModal, { type LoginModalState } from './components/LoginModal'
import Home from './pages/home/Home'
import HueAreYouApp from './pages/hue-are-you/HueAreYouApp'
import Portfolio from './pages/portfolio/Portfolio'
import ToySpace from './pages/toy-space/ToySpace'
import Contact from './pages/contact/Contact'
import './App.css'

const navClassName = ({ isActive }: { isActive: boolean }) =>
  isActive ? 'active' : ''

function App() {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const [modalState, setLoginModalState] = useState<LoginModalState>(null)
  const [authToken, setAuthToken] = useState<string | null>(null)
  const [user, setUser] = useState<string | null>(null)
  const [isAdmin, setIsAdmin] = useState(false)
  const location = useLocation()
  const isAuthenticated = authToken !== null

  useEffect(() => {
    setIsMobileMenuOpen(false)
  }, [location])

  const handleNavClick = () => {
    setIsMobileMenuOpen(false)
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
              onClick={() => setIsMobileMenuOpen((prev) => !prev)}
              aria-label="メニュー"
            >
              <span></span>
              <span></span>
              <span></span>
            </button>
            <nav className={`nav ${isMobileMenuOpen ? 'nav-open' : ''}`}>
              <ul>
                <li>
                  <NavLink to="/" end className={navClassName} onClick={handleNavClick}>
                    ホーム
                  </NavLink>
                </li>
                <li>
                  <NavLink to="/hue-are-you" className={navClassName} onClick={handleNavClick}>
                    Hue Are You?
                  </NavLink>
                </li>
                <li>
                  <NavLink to="/portfolio" className={navClassName} onClick={handleNavClick}>
                    ポートフォリオ
                  </NavLink>
                </li>
                <li>
                  <NavLink to="/toy-space" className={navClassName} onClick={handleNavClick}>
                    Toy Space
                  </NavLink>
                </li>
                <li>
                  <NavLink to="/contact" className={navClassName} onClick={handleNavClick}>
                    コンタクト
                  </NavLink>
                </li>
                <li>
                  <a href="https://github.com" target="_blank" rel="noopener noreferrer">
                    GitHub
                  </a>
                </li>
                <li>
                  <div className="auth-buttons">
                    {isAuthenticated ? (
                      <div className="user-info">
                        <span className="user-name">
                          Welcome, {user}
                          {isAdmin ? ' (Admin)' : ''}
                        </span>
                        <button
                          className="logout-btn"
                          onClick={() => {
                            setAuthToken(null)
                            setUser(null)
                            setIsAdmin(false)
                          }}
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
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/hue-are-you" element={<HueAreYouApp />} />
        <Route path="/portfolio" element={<Portfolio />} />
        <Route path="/toy-space" element={<ToySpace />} />
        <Route path="/contact" element={<Contact />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
      <LoginModal
        modalState={modalState}
        onClose={() => setLoginModalState(null)}
        onLogin={({ username, token, isAdmin }) => {
          setAuthToken(token)
          setUser(username)
          setIsAdmin(isAdmin)
        }}
      />
    </div>
  )
}

export default App
