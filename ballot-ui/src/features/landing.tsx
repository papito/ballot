import React, { useState } from 'react'
import Brand from '../components/brand.tsx'
import Byline from '../components/byline.tsx'
import Footer from '../components/footer.tsx'
import GeneralError from '../components/general_error.tsx'

import axios from 'axios'
import Tagline from '../components/tagline.tsx'
import { useErrorContext } from '../contexts/error_context.tsx'

function Landing(): React.JSX.Element {
    let sessionId: string | null = null
    let userId: string | null = null
    const [name, setName] = useState<string | null>()
    const { generalError, formError } = useErrorContext()

    async function createNewSession(event: React.FormEvent<HTMLFormElement>): Promise<void> {
        event.preventDefault()

        try {
            const createSessionResponse = await axios.post('/api/session')
            sessionId = createSessionResponse.data.id
            console.assert(sessionId, 'sessionId is required')
        } catch {
            return
        }

        try {
            const createUserResponse = await axios.post('/api/user', {
                name: name,
                session_id: sessionId,
                is_admin: 1,
            })
            userId = createUserResponse.data.id
            console.assert(userId, 'userId is required')
        } catch {
            return
        }

        window.location.assign(`/vote/s/${sessionId}/u/${userId}`)
    }

    return (
        <div id="Landing" className="view">
            <Brand session={null} />
            <GeneralError error={generalError} />
            <div className="entry-point">
                <Tagline />
                <form onSubmit={createNewSession}>
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
                    <button type="submit" className="success">
                        New Voting Space
                    </button>
                </form>
                <Byline />
            </div>
            <Footer />
        </div>
    )
}

export default Landing
