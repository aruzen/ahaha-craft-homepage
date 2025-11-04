import { useNavigate } from 'react-router-dom'

function Home() {
  const navigate = useNavigate()

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
            onClick={() => navigate('/hue-are-you')}
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
            onClick={() => navigate('/portfolio')}
          >
            ポートフォリオを見る
          </button>
        </div>
      </section>
    </main>
  )
}

export default Home
