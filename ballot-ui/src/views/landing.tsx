import './landing.css'
import '../components/brand.tsx'
import Brand from '../components/brand.tsx';
import Footer from '../components/footer.tsx';
import axios from 'axios';


function Landing() {
    let sessionId: string | null = null

    async function createNewSession() {
        axios
            .post("/api/session")
            .then((response) => {
                sessionId = response.data.id
            }).
            then(() => {
                console.log(sessionId)
        })
            .catch((error) => console.error(error))
    }

    return (
            <div id="Landing">
                <Brand/>

                <div className="form">
                    <form>
                        <label htmlFor=""></label>
                        <input type="text"
                               maxLength={64}
                               placeholder="Your name/alias"/>
                        <button
                            type="button"
                            onClick={createNewSession}>New Voting Space
                        </button>
                    </form>
                </div>

                <Footer/>
            </div>
    )
}

export default Landing;
