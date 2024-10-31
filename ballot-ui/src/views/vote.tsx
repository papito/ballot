import './vote.css'

import axios from 'axios'
import { produce } from 'immer'
import React, { useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useImmer } from 'use-immer'
import Brand from '../components/brand.tsx'
import Footer from '../components/footer.tsx'
import GeneralError from '../components/general_error.tsx'
import StartStop from '../components/start_stop.tsx'
import Voter from '../components/voter.tsx'
import { NO_ESTIMATE, SessionState } from '../constants.ts'
import Websockets from '../websockets.ts'

export interface Session {
    id: string | undefined
    status: SessionState
    tally: string
}

export interface User {
    id: string | undefined
    name: string
    estimate: string
    voted: boolean
    is_observer: boolean
    is_admin: boolean
}

function Vote(): React.JSX.Element {
    const params = useParams()
    const sessionId = params.sessionId
    console.assert(sessionId, 'sessionId is required')
    const userId = params.userId
    console.assert(userId, 'userId is required')

    console.debug('Session ID:', sessionId)
    console.debug('User ID:', userId)

    const [generalError, setGeneralError] = useState<string | null>(null)
    const [user, setUser] = useImmer<User>({
        id: userId,
        name: '',
        estimate: NO_ESTIMATE,
        voted: false,
        is_observer: false,
        is_admin: false,
    })

    const [session, setSession] = useState<Session>({
        id: sessionId,
        status: SessionState.IDLE,
        tally: NO_ESTIMATE,
    })
    const [observerNames, setObserverNames] = useState<string>('')
    const [voters, setVoters] = useImmer<User[]>([])

    const possibleEstimates: Readonly<string[]> = ['?', '0', '1', '2', '3', '5', '8', '13', '20', '40', '100']

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

        const ws: Websockets = new Websockets()

        const fetchUser = async (): Promise<void> => {
            const { data } = await axios.get<User>(`/api/user/${userId}`)
            setUser(data)

            const watchCmd = {
                action: 'WATCH',
                session_id: sessionId,
                user_id: userId,
                is_observer: false,
                is_admin: true,
            }
            ws.send(JSON.stringify(watchCmd))
        }

        function watchingSessionWsHandler(json: { [key: string]: never }): void {
            session.status = json['session_state']
            session.tally = json['tally']

            const usersJson: never[] = json['users'] || []
            const sessionVoters: User[] = []

            for (const userJson of usersJson) {
                sessionVoters.push(userJson)
            }

            setVoters(sessionVoters)

            const observersJson: { [key: string]: never }[] = json['observers'] || []
            const names: string = observersJson
                .map((observerJson: { [key: string]: never }) => observerJson['name'])
                .join(', ')

            setObserverNames(names)
        }

        function userLeftWsHandler(json: { [key: string]: never }): void {
            const voterId = json['user_id']

            produce(voters, (draft) => {
                const index = draft.findIndex((voter) => voter.id === voterId)
                if (index !== -1) draft.splice(index, 1)
            })
        }

        function userAddedWsHandler(userJson: never): void {
            const newUserId = userJson['id']

            const isExisting = voters.findIndex((voter: User) => voter.id === newUserId)
            if (isExisting >= 0) {
                return
            }

            produce(voters, (draft) => {
                draft.push(userJson)
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
            setSession({ ...session, status: SessionState.VOTING })

            setVoters((draft) => {
                draft.forEach((voter) => {
                    voter.voted = false
                    voter.estimate = NO_ESTIMATE
                })
            })

            setUser((draft) => {
                draft.estimate = NO_ESTIMATE
            })
        }

        function votingFinishedWsHandler(json: { [key: string]: never }): void {
            const tally: string = json['tally']
            setSession({ ...session, status: SessionState.IDLE, tally: tally })

            const usersJson: never[] = json['users'] || []
            const sessionVoters: User[] = []

            for (const userJson of usersJson) {
                sessionVoters.push(userJson)
            }

            setVoters(sessionVoters)
        }

        ws.socket.onMessage((data: string) => {
            const json = JSON.parse(data)
            const event: string = json['event']
            console.log(event, json)

            setGeneralError('')

            switch (event) {
                case 'USER_ADDED': {
                    userAddedWsHandler(json as never)
                    break
                }
                case 'OBSERVER_ADDED': {
                    // this.observerAddedWsHandler(json)
                    break
                }
                case 'WATCHING': {
                    watchingSessionWsHandler(json)
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
                    votingFinishedWsHandler(json)
                    break
                }
                case 'USER_LEFT': {
                    userLeftWsHandler(json)
                    break
                }
                case 'OBSERVER_LEFT': {
                    // this.observerLeftWsHandler(json)
                    break
                }
            }
        })

        fetchUser().catch((error: unknown) => {
            setGeneralError(`An error occurred (${error}). See server logs.`)
        })
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [sessionId, userId])

    const votersJsx = voters.map((voter: User) => {
        return <Voter {...voter} key={voter.id} />
    })

    const possibleEstimatesJsx = possibleEstimates.map((estimate: string) => {
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

    const observerNamesJsx: React.JSX.Element = observerNames ? (
        <div id="observerNames">
            <span>Observers: {observerNames}</span>
        </div>
    ) : (
        <></>
    )

    const startMessageJsx: React.JSX.Element =
        voters.length == 1 ? (
            <div id="startMessage">
                Looks like you are the only one here!{' '}
                <a href="" target="_blank">
                    Join this session
                </a>{' '}
                in a different tab to test with more than one user.
            </div>
        ) : (
            <></>
        )

    return (
        <div id="Vote" className="view">
            <Brand />
            <GeneralError error={generalError} />
            <div id="voteContainer">
                <div id="voteHeader">
                    <StartStop session={session} user={user} />
                    <div id="copySessionUrl">
                        <button className="btn copy-url">Copy session URL</button>
                    </div>
                </div>
                {observerNamesJsx}
                {startMessageJsx}
                <div id="estimates">{possibleEstimatesJsx}</div>
                <div id="voters">{votersJsx}</div>
            </div>
            <Footer />
        </div>
    )
}

export default Vote
