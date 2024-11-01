import './voter.css'
import React from 'react'
import { SessionState } from '../constants.ts'
import { Session, User } from '../views/vote.tsx'

interface VoterProps {
    voter: User
    session: Session
}

function Voter({ voter, session }: VoterProps): React.JSX.Element {
    const estimateJsx =
        session.status == SessionState.IDLE && voter.voted ? <span className="done">{voter.estimate}</span> : <></>

    const waitingJsx = session.status == SessionState.VOTING && !voter.voted ? <div className="waiting"></div> : <></>

    const idleJsx =
        session.status == SessionState.IDLE && !voter.voted ? <div className="idle">[not voted yet]</div> : <></>

    return (
        <div className="voter">
            <div className="name">{voter.name}</div>
            <div className="voteStatus">
                <img src={voter.voted ? '/v.png' : '/x.png'} alt={voter.voted ? 'Voted' : 'Not voted'} />
            </div>
            <div className="estimate">
                {estimateJsx}
                {waitingJsx}
                {idleJsx}
            </div>
        </div>
    )
}

export default Voter
