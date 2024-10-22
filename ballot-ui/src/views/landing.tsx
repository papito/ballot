import './landing.css'
import React, { useState } from 'react'
import Brand from '../components/brand.tsx'
import Footer from '../components/footer.tsx'
import GeneralError from '../components/general_error.tsx'

// https://github.com/axios/axios/discussions/5859
// eslint-disable-next-line import/named
import axios, { AxiosResponse } from 'axios'

function Landing(): React.JSX.Element {
    let sessionId: string | null = null
    let userId: string | null = null
    const [name, setName] = useState<string | null>()
    const [formError, setFormError] = useState<string | null>()
    const [generalError, setGeneralError] = useState<string | null>(null)

    function setError(response: AxiosResponse): void {
        const msg = `An error occurred: <b>${response.statusText}</b>. See server logs.`
        console.error(msg)
        setGeneralError(msg)
    }

    async function createNewSession(): Promise<void> {
        setFormError(null)
        setGeneralError(null)

        /** A bit of boilerplate here but since this is the only place of direct API calls,
         * a more convoluted approach is not warranted.
         */

        // get new session id
        await axios
            .post('/api/session')
            .then((response) => {
                sessionId = response.data.id
                console.log('Session ID is: ', sessionId)
            })
            .catch((error) => {
                setError(error.response)
            })

        if (sessionId === null) {
            return
        }

        console.log('Creating new user for session')
        // create new admin user for the session
        await axios
            .post('/api/user', {
                name: name,
                session_id: sessionId,
                is_admin: 1,
            })
            .then((response) => {
                userId = response.data.id
                console.log('Session user ID is: ', userId)
            })
            .catch((error) => {
                switch (error.response.status) {
                    case 400:
                        console.error('Bad Request: ', error.response.data.error)
                        setFormError(error.response.data.error)
                        break
                    default:
                        setError(error.response)
                }
            })

        if (userId === null) {
            return
        }

        await axios.put(`/api/vote/start`, { session_id: sessionId }).then(() => {
            window.location.href = `/p/vote?session_id=${sessionId}&user_id=${userId}`
        })
    }

    return (
        <div id="Landing">
            <Brand />
            <GeneralError error={generalError} />

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
