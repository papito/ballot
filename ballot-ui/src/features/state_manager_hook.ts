import axios from 'axios'
import { useEffect, useRef } from 'react'
// eslint-disable-next-line import/named
import { Updater, useImmer } from 'use-immer'
import { NO_ESTIMATE, SessionState, WebsocketAction } from '../constants.ts'
import { useErrorContext } from '../contexts/error_context.tsx'
import { Session, User } from '../types/types.tsx'
import Websockets from '../websockets.ts'

export function useVoteManager({ userId, sessionId }: { userId: string | undefined; sessionId: string | undefined }): {
    user: User
    setUser: Updater<User>
    session: Session
    voters: User[]
    observers: User[]
} {
    const { setGeneralError } = useErrorContext()

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

        fetchUser().catch((error: unknown) => {
            setGeneralError(`An error occurred (${error}). See server logs.`)
        })

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
                case WebsocketAction.USER_ADDED: {
                    userAddedWsHandler(json as User)
                    break
                }
                case WebsocketAction.OBSERVER_ADDED: {
                    observerAddedWsHandler(json as User)
                    break
                }
                case WebsocketAction.WATCHING: {
                    watchingSessionWsHandler(json as Session)
                    break
                }
                case WebsocketAction.VOTING: {
                    votingStartedWsHandler()
                    break
                }
                case WebsocketAction.USER_VOTED: {
                    userVotedWsHandler(json)
                    break
                }
                case WebsocketAction.VOTE_FINISHED: {
                    votingFinishedWsHandler(json as Session)
                    break
                }
                case WebsocketAction.USER_LEFT: {
                    userLeftWsHandler(json)
                    break
                }
                case WebsocketAction.OBSERVER_LEFT: {
                    observerLeftWsHandler(json)
                    break
                }
            }
        })

        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    return { user, setUser, session, voters, observers }
}
