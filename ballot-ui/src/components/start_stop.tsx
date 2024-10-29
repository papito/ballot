import './start_stop.css'
import React from 'react'
import { SessionState } from '../constants.ts';
import { ISessionState, IUserState } from '../views/vote.tsx';


interface StartStopProps {
    session: ISessionState
    user: IUserState
}

function StartStop({ session, user }: StartStopProps): React.JSX.Element {
    if (!user.is_admin) {
        return <> </>
    }

    if (session.status == SessionState.VOTING) {
        return (
            <div id="startStop">
                <button>See vote results</button>
            </div>
        )
    }

    if (session.status == SessionState.IDLE) {
        return (
            <div id="startStop">
                <button>Start</button>
            </div>
        )
    }

    return <> </>
}

export default StartStop
