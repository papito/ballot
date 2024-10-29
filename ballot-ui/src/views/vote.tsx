import './vote.css'
import React, { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import Brand from '../components/brand.tsx'
import Footer from '../components/footer.tsx'
import GeneralError from '../components/general_error.tsx'
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
    const userId = params.userId
    console.debug('Session ID:', sessionId)
    console.debug('User ID:', userId)

    console.assert(sessionId, 'sessionId is required')
    console.assert(userId, 'userId is required')

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

    // const possibleEstimates: Readonly<string[]> = ['?', '0', '1', '2', '3', '5', '8', '13', '20', '40', '100']

    // const connection: MutableRefObject<null> = useRef(null)

    useEffect(() => {
        console.log('!!!!!!!!!!! Use effect')
        // TODO: see https://stackoverflow.com/questions/60152922/proper-way-of-using-react-hooks-websockets
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

        ws.socket.onMessage((data: string) => {
            console.log('onMessage: ' + data)
            const json = JSON.parse(data)
            const event: string = json['event']

            switch (event) {
                case 'USER_ADDED': {
                    console.log('User added:', json)
                    // this.userAddedWsHandler(json)
                    break
                }
                case 'OBSERVER_ADDED': {
                    console.log('Observer added:', json)
                    // this.observerAddedWsHandler(json)
                    break
                }
                case 'WATCHING': {
                    console.log('Watching session:', json)
                    watchingSessionWsHandler(json)
                    break
                }
                case 'VOTING': {
                    console.log('Voting started:', json)
                    // this.votingStartedWsHandler()
                    break
                }
                case 'USER_VOTED': {
                    console.log('User voted:', json)
                    // this.userVotedHandler(json)
                    break
                }
                case 'VOTE_FINISHED': {
                    console.log('Voting finished:', json)
                    // this.votingFinishedWsHandler(json)
                    break
                }
                case 'USER_LEFT': {
                    console.log('User left:', json)
                    // this.userLeftWsHandler(json)
                    break
                }
                case 'OBSERVER_LEFT': {
                    console.log('Observer left:', json)
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
        return (
            <div key={voter.id} className="voter">
                {voter.name}
            </div>
        )
    })

    console.log('Users JSX', session.users.length)

    return (
        <div id="Vote" className="view">
            <Brand />
            <GeneralError error={generalError} />
            <div id="voteContainer">
                <div id="voteHeader">
                    <StartStop session={session} user={user} />
                    <div id="copySessionUrl">
                        <button>Copy session URL</button>
                    </div>
                </div>
                <div id="voters">{votersJsx}</div>
            </div>
            <Footer />
        </div>
    )
}

export default Vote
