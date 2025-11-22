import { NavLink } from 'react-router-dom'

const navClassName = ({ isActive }: { isActive: boolean }) =>
  isActive ? 'active' : ''

const Contact = () => {
  return (
    <main id="home">
      <section id="contact">
        <h2>Contact</h2>
        <p>ご意見・ご質問がありましたらお気軽にご連絡ください。</p>
        <ul>
          <li>ポートフォリオ: <NavLink to="/portfolio" className={navClassName}>
                    ポートフォリオ
                  </NavLink></li>
          <li>メール: <a href="mailto:mmhbk.byd@gmail.com">mmhbk.byd@gmail.com</a></li>
          <li>GitHub: <a href="https://github.com/aruzen" target="_blank" rel="noopener noreferrer">AhahaCraft</a></li>
        </ul>
      </section>
    </main>
  )
}

export default Contact
