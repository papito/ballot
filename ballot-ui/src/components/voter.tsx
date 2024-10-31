import './voter.css'
import React from 'react'
import { User } from '../views/vote.tsx'

function Voter({ name, estimate, voted }: User): React.JSX.Element {
    return (
        <div className="voter">
            <div className="name">{name}</div>
            <div className="voteStatus">{voted ? 'V' : 'X'}</div>
            <div className="estimate">{estimate ? estimate : '[still voting...]'}</div>
        </div>
    )
}

export default Voter
