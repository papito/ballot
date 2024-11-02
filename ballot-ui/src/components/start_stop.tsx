import './start_stop.css'
import axios from 'axios'
import React from 'react'
import { SessionState } from '../constants.ts'
import { Session, User } from '../views/vote.tsx'

interface StartStopProps {
    session: Session
    user: User
}

function StartStop({ session, user }: StartStopProps): React.JSX.Element {
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
    }

    if (session.status == SessionState.VOTING) {
        return (
            <div id="startStop">
                <button className="btn stop" onClick={finishVote}>
                    <i className="fas fa-stop"></i>Finish the vote
                </button>
            </div>
        )
    }

    if (session.status == SessionState.IDLE) {
        return (
            <div id="startStop">
                <button className="btn start" onClick={startVote}>
                    <i className="fas fa-play"></i>Start the vote
                </button>
            </div>
        )
    }

    return <> </>
}

export default StartStop
