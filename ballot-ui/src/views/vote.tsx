import './vote.css'
import React, { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import Brand from '../components/brand.tsx'
import Footer from '../components/footer.tsx'
import GeneralError from '../components/general_error.tsx'

// https://github.com/axios/axios/discussions/5859
// eslint-disable-next-line import/named
import axios, { AxiosResponse } from 'axios'
import { User } from '../models.ts'
import Websockets from '../websockets.ts'

function Vote(): React.JSX.Element {
    const [generalError, setGeneralError] = useState<string | null>(null)
    const [user, setUser] = useState<User | null>(null)

    const params = useParams()
    // const possibleEstimates: Readonly<string[]> = ['?', '0', '1', '2', '3', '5', '8', '13', '20', '40', '100']

    const sessionId = params.sessionId
    console.assert(sessionId, 'sessionId is required')
    const userId = params.userId
    console.assert(userId, 'userId is required')

    console.log('Session ID:', sessionId)
    console.log('User ID:', userId)

    // const connection: MutableRefObject<null> = useRef(null)

    const ws: Websockets = new Websockets()

    useEffect(() => {
        const fetchUser = async (): Promise<void> => {
            const response: AxiosResponse = await axios.get(`/api/user/${userId}`)

            setUser(User.fromJson(response.data))
            console.log('User:', response.data)

            const watchCmd = {
                action: 'WATCH',
                session_id: sessionId,
                user_id: userId,
                is_observer: false,
                is_admin: true,
            }
            ws.send(JSON.stringify(watchCmd))
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
                    // this.watchingSessionWsHandler(json)
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
    }, [userId])

    return (
        <div id="Vote">
            <Brand />
            <GeneralError error={generalError} />
            <div id="voteContainer"></div>
            <Footer />
        </div>
    )
}

export default Vote
