import './vote.css'
import React, { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import Brand from '../components/brand.tsx'
import Footer from '../components/footer.tsx'
import GeneralError from '../components/general_error.tsx'
import Voter from '../components/voter.tsx'
import { NO_ESTIMATE, SessionState } from '../constants.ts'

// https://github.com/axios/axios/discussions/5859
// eslint-disable-next-line import/named
import axios, { AxiosResponse } from 'axios'
import StartStop from '../components/start_stop.tsx'
import { User } from '../models.ts'
import Websockets from '../websockets.ts'

export interface ISessionState {
    id: string | undefined
    status: SessionState
    tally: { [key: string]: number }
    users: User[]
    observers: User[]
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
    const [user, setUser] = useState<IUserState>({
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
        tally: {},
        users: [],
        observers: [],
    })
    const [observerNames, setObserverNames] = useState<string>('')

    const possibleEstimates: Readonly<string[]> = ['?', '0', '1', '2', '3', '5', '8', '13', '20', '40', '100']

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

            const thisUser = User.fromJson(response.data)

            setUser({
                ...user,
                name: thisUser.name,
                is_admin: thisUser.is_admin,
                is_observer: thisUser.is_observer,
                voted: thisUser.voted,
                estimate: thisUser.estimate,
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

            const sessionUsers: User[] = []
            const usersJson: { [key: string]: never }[] = json['users'] || []
            for (const userJson of usersJson) {
                const aUser = User.fromJson(userJson)
                sessionUsers.push(aUser)
            }

            session.users = sessionUsers
            setSession({ ...session, users: sessionUsers })

            const sessionObservers: User[] = []
            const observersJson: { [key: string]: never }[] = json['observers'] || []
            for (const observerJson of observersJson) {
                const observer = User.fromJson(observerJson)
                sessionObservers.push(observer)
            }

            setObserverNames(sessionObservers.map((observer: User) => observer.name).join(', '))
        }

        function userLeftWsHandler(json: { [key: string]: never }): void {
            const voterId = json['user_id']
            const userIndex = session.users.findIndex((voter: User) => voter.id === voterId)
            if (userIndex !== -1) {
                session.users.splice(userIndex, 1)
                setSession({ ...session, users: session.users })
            }
        }

        function userAddedWsHandler(json: { [key: string]: never }): void {
            const newUser = User.fromJson(json)

            const isExisting = session.users.findIndex((voter: User) => voter.id === newUser.id)
            if (isExisting >= 0) {
                return
            }

            session.users.push(newUser)
            setSession({ ...session, users: session.users })
        }

        ws.socket.onMessage((data: string) => {
            console.log('Received:', data)
            const json = JSON.parse(data)
            const event: string = json['event']
            console.log(event, json)

            switch (event) {
                case 'USER_ADDED': {
                    console.log(json)
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
                    // this.votingStartedWsHandler()
                    break
                }
                case 'USER_VOTED': {
                    // this.userVotedHandler(json)
                    break
                }
                case 'VOTE_FINISHED': {
                    // this.votingFinishedWsHandler(json)
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
    }, [sessionId, userId])

    const votersJsx = session.users.map((voter: User) => {
        return <Voter voter={voter} key={voter.id} />
    })

    const possibleEstimatesJsx = possibleEstimates.map((estimate: string) => {
        return (
            <div key={estimate}>
                <button className="btn estimate">{estimate}</button>
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
        session.users.length == 1 ? (
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
