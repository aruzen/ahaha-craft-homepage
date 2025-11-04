import { useEffect, useState } from 'react'
import { NavLink, Route, Routes, useLocation } from 'react-router-dom'
import LoginModal, { LoginModalState } from './components/LoginModal'
import HueAreYouApp from './pages/hue-are-you/HueAreYouApp'
import Portfolio from './pages/portfolio/Portfolio'
import Home from './pages/home/Home'
import ToySpace from './pages/toy-space/ToySpace'
import Contact from './pages/contact/Contact'
import NotFound from './pages/not-found/NotFound'
import './App.css'

function App() {
  const location = useLocation()
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const [modalState, setLoginModalState] = useState<LoginModalState>(null)
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [user, setUser] = useState<string | null>(null)

  useEffect(() => {
    setIsMobileMenuOpen(false)
  }, [location.pathname])

  return (
    <div className="App">
      <header className="App-header">
        <div className="header-content">
          <div className="header-left">
            <img src="/resource/ahahacraft.png" alt="AhahaCraft Logo" className="logo" />
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
                  <NavLink to="/" end className={({ isActive }) => (isActive ? 'active' : undefined)}>
                    ホーム
                  </NavLink>
                </li>
                <li>
                  <NavLink
                    to="/hue-are-you"
                    className={({ isActive }) => (isActive ? 'active' : undefined)}
                  >
                    Hue Are You?
                  </NavLink>
                </li>
                <li>
                  <NavLink
                    to="/portfolio"
                    className={({ isActive }) => (isActive ? 'active' : undefined)}
                  >
                    ポートフォリオ
                  </NavLink>
                </li>
                <li>
                  <NavLink
                    to="/toy-space"
                    className={({ isActive }) => (isActive ? 'active' : undefined)}
                  >
                    Toy Space
                  </NavLink>
                </li>
                <li>
                  <NavLink
                    to="/contact"
                    className={({ isActive }) => (isActive ? 'active' : undefined)}
                  >
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
                        <span className="user-name">Welcome, {user}</span>
                        <button
                          className="logout-btn"
                          onClick={() => {
                            setIsAuthenticated(false)
                            setUser(null)
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
        <Route path="*" element={<NotFound />} />
      </Routes>

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
