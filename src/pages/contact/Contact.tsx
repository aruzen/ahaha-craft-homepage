import { NavLink } from 'react-router-dom'

function Contact() {
  return (
    <main>
      <section id="contact">
        <h2>Contact</h2>
        <p>ご質問やご相談はお気軽にメールでお送りください。</p>
        <p>
          <a href="mailto:contact@ahahacraft.dev">contact@ahahacraft.dev</a>
        </p>
        <p>
          <NavLink to="/portfolio">制作実績はこちら</NavLink>
        </p>
      </section>
    </main>
  )
}

export default Contact
