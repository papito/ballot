export enum SessionState {
    IDLE = 0,
    VOTING = 1,
}

export const NO_ESTIMATE: string = ''

export enum WebsocketAction {
    USER_ADDED = 'USER_ADDED',
    USER_LEFT = 'USER_LEFT',
    OBSERVER_ADDED = 'OBSERVER_ADDED',
    WATCHING = 'WATCHING',
    VOTING = 'VOTING',
    USER_VOTED = 'USER_VOTED',
    VOTE_FINISHED = 'VOTE_FINISHED',
    OBSERVER_LEFT = 'OBSERVER_LEFT',
}
