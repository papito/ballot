import './landing.css'
import React, { useState } from 'react'
import Brand from '../components/brand.tsx'
import Footer from '../components/footer.tsx'
import axios from 'axios'

function Landing(): React.JSX.Element {
    let sessionId: string | null = null
    const [name, setName] = useState<string | null>()
    const [formError, setFormError] = useState<string | null>()

    async function createNewSession(): Promise<void> {
        // get new session id
        const newSessionResponse = axios
            .post('/api/session')
            .then((response) => {
                sessionId = response.data.id
            })
            .catch((error) => {
                console.error(error.response)
                return
            })

        // create new admin user for the session
        newSessionResponse.then(() => {
            axios
                .post('/api/user', {
                    name: name,
                    session_id: sessionId,
                    is_admin: 1,
                })
                .catch((error) => {
                    switch (error.response.status) {
                        case 400:
                            console.error('Bad Request: ', error.response.data.error)
                            setFormError(error.response.data.error)
                            break
                        default:
                            console.error('An error occurred: ', error.response.data)
                    }
                })
        })
    }

    return (
        <div id="Landing">
            <Brand />

            <div className="form">
                <form>
                    <label htmlFor=""></label>
                    <div className={formError ? 'error' : 'hidden'}>{formError}</div>
                    <input
                        className={formError ? 'error' : ''}
                        type="text"
                        maxLength={64}
                        placeholder="Your name/alias"
                        onChange={(e) => setName(e.target.value)}
                    />
                    <button type="button" onClick={createNewSession}>
                        New Voting Space
                    </button>
                </form>
            </div>

            <Footer />
        </div>
    )
}

export default Landing
