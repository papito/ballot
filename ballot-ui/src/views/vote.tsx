import './vote.css'

// https://github.com/axios/axios/discussions/5859
// eslint-disable-next-line import/named
import axios, { AxiosResponse } from 'axios'
import { produce } from 'immer'
import React, { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useImmer } from 'use-immer'
import Brand from '../components/brand.tsx'
import Footer from '../components/footer.tsx'
import GeneralError from '../components/general_error.tsx'
import StartStop from '../components/start_stop.tsx'
import Voter from '../components/voter.tsx'
import { NO_ESTIMATE, SessionState } from '../constants.ts'
import Websockets from '../websockets.ts'

export interface ISessionState {
    id: string | undefined
    status: SessionState
    tally: string
}

export interface IUserState {
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
    const [user, setUser] = useImmer<IUserState>({
        id: userId,
        name: '',
        estimate: NO_ESTIMATE,
        voted: false,
        is_observer: false,
        is_admin: false,
    })

    const [session, setSession] = useState<ISessionState>({
        id: sessionId,
        status: SessionState.IDLE,
        tally: NO_ESTIMATE,
    })
    const [observerNames, setObserverNames] = useState<string>('')
    const [voters, setVoters] = useImmer<IUserState[]>([])

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
    useEffect(() => {
        console.debug('!!!!!!!!!!! Use effect')
        // see https://stackoverflow.com/questions/60152922/proper-way-of-using-react-hooks-websockets
        // https://ably.com/blog/websockets-react-tutorial
        const ws: Websockets = new Websockets()

        const fetchUser = async (): Promise<void> => {
            const response: AxiosResponse = await axios.get(`/api/user/${userId}`)
            console.debug('User:', response.data)

            const userJson = response.data

            setUser({
                id: userJson['id'],
                name: userJson['name'],
                estimate: userJson['estimate'],
                voted: userJson['voted'],
                is_observer: userJson['is_observer'],
                is_admin: userJson['is_admin'],
            })

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

            const usersJson: { [key: string]: never }[] = json['users'] || []
            const sessionVoters: IUserState[] = []

            for (const userJson of usersJson) {
                sessionVoters.push({
                    id: userJson['id'],
                    name: userJson['name'],
                    estimate: userJson['estimate'],
                    voted: userJson['voted'],
                    is_observer: userJson['is_observer'],
                    is_admin: userJson['is_admin'],
                })
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

        function userAddedWsHandler(userJson: { [key: string]: never }): void {
            const newUserId = userJson['id']

            const isExisting = voters.findIndex((voter: IUserState) => voter.id === newUserId)
            if (isExisting >= 0) {
                return
            }

            produce(voters, (draft) => {
                draft.push({
                    id: userJson['id'],
                    name: userJson['name'],
                    estimate: userJson['estimate'],
                    voted: userJson['voted'],
                    is_observer: userJson['is_observer'],
                    is_admin: userJson['is_admin'],
                })
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

            const usersJson: { [key: string]: never }[] = json['users'] || []
            const sessionVoters: IUserState[] = []

            for (const userJson of usersJson) {
                sessionVoters.push({
                    id: userJson['id'],
                    name: userJson['name'],
                    estimate: userJson['estimate'],
                    voted: userJson['voted'],
                    is_observer: userJson['is_observer'],
                    is_admin: userJson['is_admin'],
                })
            }

            setVoters(sessionVoters)
        }

        ws.socket.onMessage((data: string) => {
            const json = JSON.parse(data)
            const event: string = json['event']
            console.log(event, json)

            switch (event) {
                case 'USER_ADDED': {
                    userAddedWsHandler(json)
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

    const votersJsx = voters.map((voter: IUserState) => {
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
