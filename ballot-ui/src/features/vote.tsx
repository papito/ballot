import './vote.css'

import axios from 'axios'
import React, { useEffect, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { useImmer } from 'use-immer'
import Brand from '../components/brand.tsx'
import Footer from '../components/footer.tsx'
import GeneralError from '../components/general_error.tsx'
import StartStop from '../components/start_stop.tsx'
import Voter from '../components/voter.tsx'
import { NO_ESTIMATE, SessionState } from '../constants.ts'
import { useErrorContext } from '../contexts/error_context.tsx'
import { Session, User } from '../types/types.tsx'
import Websockets from '../websockets.ts'

function Vote(): React.JSX.Element {
    const params = useParams()
    const sessionId = params.sessionId
    console.assert(sessionId, 'sessionId is required')
    const userId = params.userId
    console.assert(userId, 'userId is required')

    const { generalError, setGeneralError } = useErrorContext()

    const [user, setUser] = useImmer<User>({
        id: userId,
        name: '',
        estimate: NO_ESTIMATE,
        voted: false,
        is_observer: false,
        is_admin: false,
    })

    const [session, setSession] = useImmer<Session>({
        id: sessionId,
        status: SessionState.IDLE,
        tally: NO_ESTIMATE,
        users: [],
        observers: [],
    })
    const [voters, setVoters] = useImmer<User[]>([])
    const [observers, setObservers] = useImmer<User[]>([])

    const observerNames = observers.map((observer) => observer.name).join(', ')

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

        setUser((draft) => {
            draft.estimate = estimate
            draft.voted = true
        })
    }

    /**
     * This runs once when the component is mounted.
     */
    const mounted = useRef(false)

    useEffect(() => {
        if (mounted.current) {
            return
        }
        mounted.current = true

        console.debug('Session ID:', sessionId)
        console.debug('User ID:', userId)

        const ws: Websockets = new Websockets()

        /**
         * A mobile device may have lost connection on sleep or locked screen.
         * Calling "reconnect" should be harmless, and not always effective.
         */
        document.addEventListener('visibilitychange', () => {
            if (!document.hidden) {
                ws.reconnect()
            }
        })
        window.addEventListener('blur', () => {
            ws.reconnect()
        })

        const fetchUser = async (): Promise<void> => {
            try {
                const { data } = await axios.get<User>(`/api/user/${userId}`)
                setUser(data)

                const watchCmd = {
                    action: 'WATCH',
                    session_id: sessionId,
                    user_id: data.id,
                    is_observer: data.is_observer,
                    is_admin: data.is_admin,
                }
                ws.send(JSON.stringify(watchCmd))
            } catch {
                return
            }
        }

        function userLeftWsHandler(json: { [key: string]: never }): void {
            const voterId = json['user_id']

            // this can happen if the user opens a second tab by mistake, as the same user
            if (voterId === user.id) {
                console.warn('Will not remove current user from voters list')
                return
            }

            setVoters((v) => v.filter((voter) => voter.id !== voterId))
        }

        function observerLeftWsHandler(observerJson: { [key: string]: never }): void {
            const observerId = observerJson['user_id']

            setObservers((v) => v.filter((u) => u.id !== observerId))
        }

        function userAddedWsHandler(userJson: User): void {
            const newUserId = userJson['id']

            setVoters((v) => {
                const isExisting = v.findIndex((voter: User) => voter.id === newUserId)
                if (isExisting >= 0) {
                    return
                }
                return [...v, userJson]
            })
        }

        function observerAddedWsHandler(observerJson: User): void {
            const newObserverId = observerJson['id']

            setObservers((v) => {
                const isExisting = v.findIndex((u: User) => u.id === newObserverId)
                if (isExisting >= 0) {
                    return
                }
                return [...v, observerJson]
            })
        }

        function userVotedWsHandler(json: { [key: string]: string }): void {
            const voterId = json['user_id']

            setVoters((draft) => {
                const voter = draft.find((v) => v.id === voterId)
                if (voter) {
                    voter.voted = true
                }
            })
        }

        function votingStartedWsHandler(): void {
            setSession((draft) => {
                draft.status = SessionState.VOTING
            })

            setVoters((draft) => {
                draft.forEach((voter) => {
                    voter.voted = false
                    voter.estimate = NO_ESTIMATE
                })
            })

            setUser((draft) => {
                draft.estimate = NO_ESTIMATE
                draft.voted = false
            })
        }

        function watchingSessionWsHandler(ses: Session): void {
            setSession((draft) => {
                draft.status = ses.status
                draft.tally = ses.tally
            })

            setVoters(ses.users)
            setObservers(ses.observers)
        }

        function votingFinishedWsHandler(ses: Session): void {
            setSession((draft) => {
                draft.status = SessionState.IDLE
                draft.tally = ses.tally
            })

            setVoters(ses.users)
        }

        ws.socket.onMessage((data: string) => {
            const json = JSON.parse(data)
            const event: string = json['event']
            console.log(event, json)

            setGeneralError('')

            switch (event) {
                case 'USER_ADDED': {
                    userAddedWsHandler(json as User)
                    break
                }
                case 'OBSERVER_ADDED': {
                    observerAddedWsHandler(json as User)
                    break
                }
                case 'WATCHING': {
                    watchingSessionWsHandler(json as Session)
                    break
                }
                case 'VOTING': {
                    votingStartedWsHandler()
                    break
                }
                case 'USER_VOTED': {
                    userVotedWsHandler(json)
                    break
                }
                case 'VOTE_FINISHED': {
                    votingFinishedWsHandler(json as Session)
                    break
                }
                case 'USER_LEFT': {
                    userLeftWsHandler(json)
                    break
                }
                case 'OBSERVER_LEFT': {
                    observerLeftWsHandler(json)
                    break
                }
            }
        })

        fetchUser().catch((error: unknown) => {
            setGeneralError(`An error occurred (${error}). See server logs.`)
        })

        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

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
