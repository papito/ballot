import './Landing.css'
import '../components/Brand.tsx'
import Brand from '../components/Brand.tsx';
import Footer from '../components/Footer.tsx';

function MainContent() {
    return (
        <div id="Landing">
            <Brand/>

            <div className="form">
                <form>
                    <label htmlFor=""></label>
                    <input type="text"
                           maxLength={64}
                           placeholder="Your name/alias"/>
                    <button type="button">New Voting Space
                    </button>
                </form>
            </div>

            <Footer/>
        </div>
    )
}

function Landing() {
    return (
        <div>
            <MainContent/>
        </div>
    )
}

export default Landing;
