import React, { useState } from 'react'
import Brand from '../components/brand.tsx'
import Footer from '../components/footer.tsx'
import GeneralError from '../components/general_error.tsx'

// https://github.com/axios/axios/discussions/5859
// eslint-disable-next-line import/named
import axios, { AxiosError, AxiosResponse, isAxiosError } from 'axios'

function Landing(): React.JSX.Element {
    let sessionId: string | null = null
    let userId: string | null = null
    const [name, setName] = useState<string | null>()
    const [formError, setFormError] = useState<string | null>()
    const [generalError, setGeneralError] = useState<string | null>(null)

    function setError(error: unknown): void {
        if (isAxiosError(error)) {
            const axiosError = error as AxiosError
            const msg = `An error occurred: <b>${axiosError.response?.statusText}</b>. See server logs.`
            console.error(error)
            setGeneralError(msg)
        } else {
            console.error(error)
            setGeneralError(`${error}`)
        }
    }

    function setUnknownError(msg: string): void {
        console.error(msg)
        setGeneralError(msg)
    }

    async function createNewSession(event: React.FormEvent<HTMLFormElement>): Promise<void> {
        event.preventDefault()
        setFormError(null)
        setGeneralError(null)

        // get new session id
        try {
            const response: AxiosResponse = await axios.post('/api/session')

            // console.log(response.data)
            sessionId = response.data.id
        } catch (error) {
            setError(error)
            return
        }

        try {
            const response: AxiosResponse = await axios.post('/api/user', {
                name: name,
                session_id: sessionId,
                is_admin: 1,
            })
            // console.log(response.data)
            userId = response.data.id
        } catch (error) {
            if (isAxiosError(error)) {
                const axiosError = error as AxiosError
                switch (axiosError.response?.status) {
                    case 400:
                        setFormError(error.response?.data.error)
                        break
                    default:
                        setFormError(error.response?.statusText)
                }
            } else {
                setUnknownError(`An error occurred: ${error}`)
            }

            return
        }

        window.location.assign(`/vote/s/${sessionId}/u/${userId}`)
    }

    return (
        <div id="Landing" className="view">
            <Brand session={null} />
            <GeneralError error={generalError} />

            <div className="form">
                <div id="tagline">Distributed scrum estimation poker</div>
                <form onSubmit={createNewSession}>
                    <label htmlFor=""></label>
                    <div data-testid="formError" id="formError" className={formError ? 'error' : 'hidden'}>
                        {formError}
                    </div>
                    <input
                        className={formError ? 'error' : ''}
                        type="text"
                        maxLength={64}
                        placeholder="Your name/alias"
                        onChange={(e) => setName(e.target.value)}
                    />
                    <button type="submit" className="success">
                        New Voting Space
                    </button>
                </form>
                <div id="byline">
                    <div className="link">
                        <div>
                            By{' '}
                            <a href="https://renegadeotter.com" target="_blank" rel="noreferrer">
                                Renegade Otter
                            </a>
                        </div>
                    </div>
                    <div>
                        <a href="https://renegadeotter.com" target="_blank" rel="noreferrer">
                            <img src="/renegade.png" alt="Renegade Otters" />
                        </a>
                    </div>
                </div>
            </div>

            <Footer />
        </div>
    )
}

export default Landing
