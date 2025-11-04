import React, { useRef, useEffect } from 'react'
import './Portfolio.css'

interface CareerItem {
  year: string
  description: string
}

interface SkillCategory {
  title: string
  items: string[]
}

interface TechIcon {
  type: 'devicon' | 'simple'
  className?: string
  name?: string
}

interface TechStack {
  name: string
  tools?: string[]
  libraries?: string[]
  usage: string
  icons?: TechIcon[]
}

const Portfolio: React.FC = () => {
  const introCanvasRef = useRef<HTMLCanvasElement>(null)
  const skillsCanvasRef = useRef<HTMLCanvasElement>(null)
  const portfolioCanvasRef = useRef<HTMLCanvasElement>(null)
  const portfolioRef = useRef<HTMLDivElement>(null)

  const careerData: CareerItem[] = [
    { year: '2019年度', description: '神奈川総合産業高等学校 入学' },
    { year: '2020年', description: 'VICTORIA(現 H.A.C.K.S 株式会社倖一)のスタートアップに参加' },
    { year: '2022年度', description: '神奈川総合産業高等学校 卒業' },
    { year: '2023年度', description: '神奈川大学 情報学部 システム数理学科 入学' },
    { year: '2023年', description: '入学後神奈川大学プログラミングサークルを立ち上げる' },
    { year: '2027年度', description: '神奈川大学 卒業予定' }
  ]

  const skillCategories: SkillCategory[] = [
    {
      title: '技術・知識理解',
      items: [
        'アセンブラレベルでの最適化経験(コンパイラ最適化を意識したリファクタリングなど)',
        'OSへの理解(Windows/Linux/Macの内部動作やシステムコールの知識, ファイルシステム, プロセス管理など)',
        'ネットワークの基礎知識(TCP/UDP, ソケット, SSL, REST APIなど)',
        'シェル・コマンド操作(grep, sed, scp, Powershellなど)',
        '各種アルゴリズム・データ構造の応用知識',
        '並列処理・スレッド・非同期処理の理解',
        'GPU・CPU間の処理パイプラインやレンダリングパフォーマンス最適化への理解',
        'メモリ管理・ポインタ操作・キャッシュ/アライメントなど低レイヤ性能知識',
        '様々な技術への一般的な知識'
      ]
    },
    /*
    {
      title: 'OS',
      items: ['Windows', 'Mac', 'Linux']
    }
      */
  ]

  const techStacks: TechStack[] = [
    {
      name: 'C/C++',
      tools: ['CMake', 'gdb/lldb', 'Valgrind', 'QEMU'],
      libraries: ['Vulkan', 'OpenGL', 'OpenCV', 'SDL3', 'UnrealEngine', 'Qt'],
      usage: 'ゲーム, CUIアプリ, 実験的なプログラム, 大規模プロジェクト',
      icons: [
        { type: 'simple', name: 'cplusplus' }
      ]
    },
    {
      name: 'Java',
      tools: ['Gradle/Maven', 'JNI'],
      libraries: ['SpringBoot', 'Java EE', 'JUnit', 'Mockito', 'Log4j'],
      usage: '案件開発での主力, MinecaftのPlugin/Mod開発',
      icons: [
        { type: 'devicon', className: 'devicon-java-plain' }
      ]
    },
    {
      name: 'Go',
      usage: '個人開発でのバックエンドサーバ(特にREST API), 小規模プロジェクト',
      icons: [
        { type: 'simple', name: 'go' }
      ]
    },
    {
      name: 'Python',
      tools: ['Conda', 'Rye', 'Poetry', 'JupyterNotebook'],
      libraries: ['TensorFlow', 'PyTorch', 'OpenCV2', 'SymPy', 'Selenium', 'PyGame'],
      usage: '数値実験, 統計処理, AI研究, スクレイピング',
      icons: [
        { type: 'devicon', className: 'devicon-python-plain' }
      ]
    },
    {
      name: 'JS/TS',
      tools: ['Vite'],
      libraries: ['React', 'Next.js', 'Excalibur.js'],
      usage: 'Webフロントエンド, ブラウザゲーム',
      icons: [
        { type: 'simple', name: 'javascript' }
      ]
    },
    {
      name: 'C#',
      libraries: ['Unity', 'WPF', 'UWP', 'Xamarin'],
      usage: 'GUIアプリ',
      icons: [
        { type: 'devicon', className: 'devicon-csharp-plain' }
      ]
    }
  ]

  useEffect(() => {
    const initCanvas = (canvas: HTMLCanvasElement | null, color: string) => {
      if (!canvas) return
      const ctx = canvas.getContext('2d')
      if (!ctx) return
      
      canvas.width = canvas.offsetWidth
      canvas.height = canvas.offsetHeight
      
      ctx.fillStyle = color
      ctx.fillRect(0, 0, canvas.width, canvas.height)
    }

    initCanvas(introCanvasRef.current, '#2c3e50')
    initCanvas(skillsCanvasRef.current, '#4f9582ff') 
    initCanvas(portfolioCanvasRef.current, '#3498db')

  }, [])

  return (
    <div className="portfolio" ref={portfolioRef}>
      <section className="intro-section">
        <canvas ref={introCanvasRef} className="section-canvas" />
        <div className="section-content">
          <div className="intro-header">
            <h2>自己紹介</h2>
            <h3>森本 響 (Hibiki Morimoto)</h3>
          </div>
          <div className="career-section">
            <h4>経歴</h4>
            <ul className="career-list">
              {careerData.map((item, index) => (
                <li key={index} className="career-item">
                  <span className="career-year">{item.year}</span>
                  <span className="career-description">{item.description}</span>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </section>

      <section className="skills-section">
        <canvas ref={skillsCanvasRef} className="section-canvas" />
        <div className="section-content">
          <h2>スキルセット</h2>
          <div className="skills-grid">
            {skillCategories.map((category, index) => (
              <div key={index} className="skill-category">
                <h3>{category.title}</h3>
                <ul>
                  {category.items.map((skill, skillIndex) => (
                    <li key={skillIndex}>{skill}</li>
                  ))}
                </ul>
              </div>
            ))}
            
            <div className="tech-stacks">
              {techStacks.map((tech, index) => (
                <div key={index} className="tech-stack">
                  <div className="tech-stack-header">
                    {tech.icons && (
                      <div className="tech-icons" aria-hidden="true">
                        {tech.icons.map((icon, iconIndex) =>
                          icon.type === 'devicon' && icon.className ? (
                            <i key={iconIndex} className={`tech-icon ${icon.className}`} />
                          ) : icon.type === 'simple' && icon.name ? (
                            <img
                              key={iconIndex}
                              src={`https://cdn.jsdelivr.net/npm/simple-icons@v11/icons/${icon.name}.svg`}
                              alt=""
                              className="tech-icon"
                              loading="lazy"
                            />
                          ) : null
                        )}
                      </div>
                    )}
                    <h3>{tech.name}</h3>
                  </div>
                  {tech.tools && (
                    <div className="tech-detail">
                      <strong>ツール:</strong> {tech.tools.join(', ')}
                    </div>
                  )}
                  {tech.libraries && (
                    <div className="tech-detail">
                      <strong>ライブラリ/フレームワーク:</strong> {tech.libraries.join(', ')}
                    </div>
                  )}
                  <div className="tech-detail">
                    <strong>主な使用先:</strong> {tech.usage}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      <section className="portfolio-section">
        <canvas ref={portfolioCanvasRef} className="section-canvas" />
        <div className="section-content">
          <h2>制作物</h2>
          <div className="portfolio-items">
            <div className="portfolio-item-template">
              <p>制作物は今後追加予定です</p>
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}

export default Portfolio
