import { SessionState } from '../constants.ts'

export interface Session {
    id: string | undefined
    tally: string
    status: SessionState
    users: User[]
    observers: User[]
}

export interface User {
    id: string | undefined
    name: string
    estimate: string
    voted: boolean
    is_observer: boolean
    is_admin: boolean
}

export type TError = string | null

export type TValidationError = {
    error: string
}
