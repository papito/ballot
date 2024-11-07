import './vote.css'

import axios from 'axios'
import React from 'react'
import { useParams } from 'react-router-dom'
import Brand from '../components/brand.tsx'
import Footer from '../components/footer.tsx'
import GeneralError from '../components/general_error.tsx'
import StartStop from '../components/start_stop.tsx'
import Voter from '../components/voter.tsx'
import { SessionState } from '../constants.ts'
import { useErrorContext } from '../contexts/error_context.tsx'
import { User } from '../types/types.tsx'
import { useVoteManager } from './vote_manager.ts'

function Vote(): React.JSX.Element {
    const params = useParams()
    const sessionId = params.sessionId
    console.assert(sessionId, 'sessionId is required')
    const userId = params.userId
    console.assert(userId, 'userId is required')

    const { generalError, setGeneralError } = useErrorContext()

    const { user, setUser, session, voters, observers } = useVoteManager({ userId, sessionId })

    const observerNames = observers.map((observer: User) => observer.name).join(', ')

    const cardValues: Readonly<string[]> = ['?', '0', '1', '2', '3', '5', '8', '13', '20', '40', '100']

    const castVote = async (estimate: string): Promise<void> => {
        try {
            await axios.put('/api/vote/cast', {
                session_id: sessionId,
                user_id: userId,
                estimate: estimate,
            })
        } catch (error) {
            setGeneralError(`${error}`)
            return
        }

        setUser((draft: { estimate: string; voted: boolean }) => {
            draft.estimate = estimate
            draft.voted = true
        })
    }

    const votersJsx = voters.map((voter: User) => {
        return <Voter voter={voter} session={session} key={voter.id} />
    })

    const cardValuesJsx = cardValues.map((estimate: string) => {
        return (
            <div key={estimate}>
                <button
                    className={'btn estimate ' + (user.estimate === estimate ? 'selected' : '')}
                    onClick={() => castVote(estimate)}
                >
                    {estimate}
                </button>
            </div>
        )
    })
    const cardsJsx =
        session.status === SessionState.VOTING && !user.is_observer ? <div id="cards">{cardValuesJsx}</div> : <></>

    const observerNamesJsx: React.JSX.Element = observers.length ? (
        <div id="observerMessage">
            <span>
                <strong>Watching this session</strong>: {observerNames}
            </span>
        </div>
    ) : (
        <></>
    )

    const startMessageJsx: React.JSX.Element =
        voters.length == 1 ? (
            <div id="startMessage">
                <div className="hey">Hey...</div>
                <div>
                    Looks like you are the only one here!{' '}
                    <a href={'/vote/s/' + session.id} target="_blank" rel="noreferrer">
                        Join this session
                    </a>{' '}
                    in a different tab to test with more than one user.
                </div>
                <div>To invite others, copy the link to this session (top right corner) and share it with them.</div>
            </div>
        ) : (
            <></>
        )

    const tallyJsx: React.JSX.Element =
        session.status == SessionState.IDLE && session.tally ? (
            <div id="tally">
                <span>Estimate: {session.tally}</span>
            </div>
        ) : (
            <></>
        )

    const voterPromptJsx: React.JSX.Element =
        session.status == SessionState.VOTING && !user.voted && !user.is_observer ? (
            <div id="prompt">
                <span className="pick-a-card">Pick a card!</span>
            </div>
        ) : (
            <></>
        )

    const observerPromptJsx: React.JSX.Element =
        session.status == SessionState.VOTING && user.is_observer ? (
            <div id="prompt">
                <span className="voting">Voting in progress...</span>
            </div>
        ) : (
            <></>
        )

    const waitingPromptJsx: React.JSX.Element =
        !user.is_admin && session.status == SessionState.IDLE ? (
            <div id="prompt">
                <span className="waiting">Waiting for admin to start next vote...</span>
            </div>
        ) : (
            <></>
        )

    return (
        <div id="Vote" className="view">
            <Brand session={session} />
            <GeneralError error={generalError} />
            <div id="voteContainer">
                <div id="voteHeader">
                    <StartStop session={session} user={user} />
                    {voterPromptJsx}
                    {observerPromptJsx}
                    {waitingPromptJsx}
                </div>
                {observerNamesJsx}
                {startMessageJsx}
                {cardsJsx}
                {tallyJsx}
                <div id="voters">{votersJsx}</div>
            </div>
            <Footer />
        </div>
    )
}

export default Vote
