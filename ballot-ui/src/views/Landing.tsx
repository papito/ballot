import './Landing.css'
import '../components/Brand.tsx'
import {useState} from 'react';
import Brand from '../components/Brand.tsx';
import Footer from '../components/Footer.tsx';
import axios from 'axios';


function Landing() {
    let [sessionId, setSessionId] = useState('')

    function createNewSession() {
        axios
            .post("/api/session")
            .then((response) => {
                setSessionId(response.data.id)
                console.log("session id: " + sessionId)
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
