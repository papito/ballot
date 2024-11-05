import React, { useState } from 'react'
import { useParams } from 'react-router-dom'
import Brand from '../components/brand.tsx'
import Byline from '../components/byline.tsx'
import Footer from '../components/footer.tsx'
import GeneralError from '../components/general_error.tsx'

import axios from 'axios'
import Tagline from '../components/tagline.tsx'
import { useErrorContext } from '../contexts/error_context.tsx'

function Join(): React.JSX.Element {
    const params = useParams()
    const sessionId = params.sessionId
    console.debug('Session ID:', sessionId)

    const [name, setName] = useState<string | null>()
    const { generalError, formError } = useErrorContext()

    async function join({ isObserver }: { isObserver: number }): Promise<void> {
        try {
            const response = await axios.post('/api/user', {
                name: name,
                session_id: sessionId,
                is_observer: isObserver,
            })
            const userId: string | null = response.data.id
            console.assert(userId, 'userId is required')
            window.location.assign(`/vote/s/${sessionId}/u/${userId}`)
        } catch {
            return
        }
    }

    return (
        <div id="Join" className="view">
            <Brand session={null} />
            <GeneralError error={generalError} />

            <div className="entry-point">
                <Tagline />
                <form onSubmit={(e) => e.preventDefault()}>
                    <label htmlFor=""></label>
                    <div data-testid="formError" id="formError" className={formError ? 'error' : 'hidden'}>
                        {formError}
                    </div>
                    <input
                        autoFocus={true}
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
                        Join as an observer
                    </button>
                </form>
                <Byline />
            </div>

            <Footer />
        </div>
    )
}

export default Join
