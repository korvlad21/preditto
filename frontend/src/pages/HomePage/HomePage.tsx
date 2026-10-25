import {Link} from 'react-router'
import './HomePage.css'

function HomePage() {
    return (
        <section className="page home-page">
            <div> Главная страница</div>
            <Link to="/login" className="home-login-button">
                Войти
            </Link>
        </section>
    )
}

export default HomePage
