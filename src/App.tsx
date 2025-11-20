import { useEffect, useRef, useState } from 'react'
import { Navigate, NavLink, Route, Routes, useLocation } from 'react-router-dom'
import LoginModal, { type LoginModalState } from './components/LoginModal'
import type { SessionResponce, UserRole } from './api'
import Home from './pages/home/Home'
import HueAreYouApp from './pages/hue-are-you/HueAreYouApp'
import Portfolio from './pages/portfolio/Portfolio'
import ToySpace from './pages/toy-space/ToySpace'
import ToyDetail from './pages/toy-space/ToyDetail'
import Contact from './pages/contact/Contact'
import AdminDashboard from './pages/admin/AdminDashboard'
import { ToySpaceProvider } from './contexts/ToySpaceContext'
import { clearSessionState, loadSessionState, saveSessionState } from './utils/sessionStorage'
import './App.css'

const navClassName = ({ isActive }: { isActive: boolean }) =>
  isActive ? 'active' : ''

function App() {
  const persistedSession = useRef(loadSessionState())
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const [modalState, setLoginModalState] = useState<LoginModalState>(null)
  const [session, setSession] = useState<SessionResponce | null>(persistedSession.current?.session ?? null)
  const [user, setUser] = useState<string | null>(persistedSession.current?.user ?? null)
  const [userRole, setUserRole] = useState<UserRole | null>(persistedSession.current?.session?.role ?? null)
  const location = useLocation()
  const isAuthenticated = session !== null
  const isAdmin = userRole === 'admin'

  useEffect(() => {
    setIsMobileMenuOpen(false)
  }, [location])

  useEffect(() => {
    const headerElement = document.querySelector<HTMLElement>('.App-header')
    if (!headerElement) {
      return
    }

    const updateHeaderHeightVar = () => {
      const { height } = headerElement.getBoundingClientRect()
      document.documentElement.style.setProperty('--header-height', `${height}px`)
    }

    updateHeaderHeightVar()

    const resizeObserver =
      typeof ResizeObserver !== 'undefined'
        ? new ResizeObserver(() => updateHeaderHeightVar())
        : null

    resizeObserver?.observe(headerElement)
    window.addEventListener('resize', updateHeaderHeightVar)

    return () => {
      resizeObserver?.disconnect()
      window.removeEventListener('resize', updateHeaderHeightVar)
    }
  }, [])

  const handleNavClick = () => {
    setIsMobileMenuOpen(false)
  }

  useEffect(() => {
    if (session && user) {
      saveSessionState({ user, session })
    }
  }, [session, user])

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
                  <a href="https://github.com/aruzen" target="_blank" rel="noopener noreferrer">
                    GitHub
                  </a>
                </li>
                {isAdmin && (
                  <li>
                    <NavLink to="/admin" className={navClassName} onClick={handleNavClick}>
                      管理者画面
                    </NavLink>
                  </li>
                )}
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
                            setSession(null)
                            setUser(null)
                            setUserRole(null)
                            clearSessionState()
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
        <Route
          path="/toy-space"
          element={
            <ToySpaceProvider>
              <ToySpace />
            </ToySpaceProvider>
          }
        />
        <Route
          path="/toy-space/:slug"
          element={
            <ToySpaceProvider>
              <ToyDetail />
            </ToySpaceProvider>
          }
        />
        <Route path="/contact" element={<Contact />} />
        <Route
          path="/admin"
          element={
            isAdmin && session ? (
              <AdminDashboard username={user ?? 'Administrator'} session={session} />
            ) : (
              <Navigate to="/" replace />
            )
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
      <LoginModal
        modalState={modalState}
        onClose={() => setLoginModalState(null)}
        onLogin={({ username, session }) => {
          setSession(session)
          setUser(username)
          setUserRole(session.role)
        }}
      />
    </div>
  )
}

export default App
