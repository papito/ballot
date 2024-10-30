import './start_stop.css'
import axios from 'axios'
import React from 'react'
import { SessionState } from '../constants.ts'
import { ISessionState, IUserState } from '../views/vote.tsx'

interface StartStopProps {
    session: ISessionState
    user: IUserState
    funcClearState: () => void
}

function StartStop({ session, user, funcClearState }: StartStopProps): React.JSX.Element {
    if (!user.is_admin) {
        return <div> </div>
    }

    const finishVote = async (): Promise<void> => {
        await axios.put('/api/vote/finish', {
            session_id: session.id,
        })
    }

    const startVote = async (): Promise<void> => {
        await axios.put('/api/vote/start', {
            session_id: session.id,
        })

        funcClearState()
    }

    if (session.status == SessionState.VOTING) {
        return (
            <div id="startStop">
                <button className="btn stop" onClick={finishVote}>
                    Finish the vote
                </button>
            </div>
        )
    }

    if (session.status == SessionState.IDLE) {
        return (
            <div id="startStop">
                <button className="btn start" onClick={startVote}>
                    Start
                </button>
            </div>
        )
    }

    return <> </>
}

export default StartStop
