import './voter.css'
import React from 'react'
import { User } from '../models.ts';

interface VoterProps {
    voter: User
}

function Voter({ voter }: VoterProps): React.JSX.Element {
    return (
        <div className="voter">
            <div className="name">{voter.name}</div>
            <div className="voteStatus">{voter.voted ? 'V' : 'X'}</div>
            <div className="estimate">{voter.estimate ? voter.estimate : '[still voting...]'}</div>
        </div>
    )
}

export default Voter
