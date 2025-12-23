import type { CSSProperties } from 'react'
import { Link } from 'react-router-dom'

const buttonStyle: CSSProperties = {
  display: 'inline-block',
  padding: '1rem 2rem',
  fontSize: '1.2rem',
  border: 'none',
  borderRadius: '8px',
  color: 'white',
  cursor: 'pointer',
  textDecoration: 'none'
}

const buttonGroupStyle: CSSProperties = {
  marginTop: '2rem',
  display: 'flex',
  gap: '1rem',
  flexWrap: 'wrap'
}

const Home = () => {
  return (
    <main id='main-content'>
      <section>
        <h2>Welcome</h2>
        <p>このサイトは現在開発中です。</p>
        <div style={buttonGroupStyle}>
          <Link
            to="/hue-are-you"
            style={{
              ...buttonStyle,
              background: 'linear-gradient(45deg, #4ecdc4, #45b7d1)'
            }}
          >
            Hue Are You? を試す
          </Link>
          <Link
            to="/portfolio"
            style={{
              ...buttonStyle,
              background: 'linear-gradient(45deg, #e74c3c, #c0392b)'
            }}
          >
            ポートフォリオを見る
          </Link>
        </div>
      </section>
    </main>
  )
}

export default Home
