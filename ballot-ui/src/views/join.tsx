import React, { useState } from 'react'
import { useParams } from 'react-router-dom'
import Brand from '../components/brand.tsx'
import Footer from '../components/footer.tsx'
import GeneralError from '../components/general_error.tsx'

// https://github.com/axios/axios/discussions/5859
// eslint-disable-next-line import/named
import axios, { AxiosError, AxiosResponse, isAxiosError } from 'axios'

function Join(): React.JSX.Element {
    const params = useParams()
    const sessionId = params.sessionId
    console.debug('Session ID:', sessionId)

    const [name, setName] = useState<string | null>()
    const [formError, setFormError] = useState<string | null>()
    const [generalError, setGeneralError] = useState<string | null>(null)

    function setUnknownError(msg: string): void {
        console.error(msg)
        setGeneralError(msg)
    }

    async function join({ isObserver }: { isObserver: number }): Promise<void> {
        setFormError(null)
        setGeneralError(null)

        try {
            const response: AxiosResponse = await axios.post('/api/user', {
                name: name,
                session_id: sessionId,
                is_observer: isObserver,
            })
            console.debug(response.data)
            const userId: string | null = response.data.id
            console.assert(userId, 'userId is required')
            window.location.assign(`/p/vote/s/${sessionId}/u/${userId}`)
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
    }

    return (
        <div id="Join" className="view">
            <Brand />
            <GeneralError error={generalError} />

            <div className="form">
                <form>
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
                    <button className="success" type="button" onClick={() => join({ isObserver: 0 })}>
                        Join as a voter
                    </button>
                    <button className="warn" type="button" onClick={() => join({ isObserver: 1 })}>
                        Join as an innocent bystander
                    </button>
                </form>
            </div>

            <Footer />
        </div>
    )
}

export default Join
