import { NO_ESTIMATE } from './constants.ts'

export class User {
    public static fromJson(json: { [key: string]: string }): User {
        const user = new User()
        user.id = json.id
        user.name = json.name
        user.estimate = json.estimate
        user.voted = Boolean(json.voted)
        user.joined = json.joined
        user.is_observer = Boolean(json.is_observer)
        user.is_admin = Boolean(json.is_admin)
        return user
    }
    public id: string = ''
    public name: string = ''
    public estimate: string = NO_ESTIMATE
    public voted: boolean = false
    public joined: string = ''
    public is_observer: boolean = false
    public is_admin: boolean = false
}

export class PendingVote {
    public static fromJson(json: { [key: string]: string }): PendingVote {
        const vote = new PendingVote()
        vote.user_id = json.user_id
        vote.session_id = json.session_id
        return vote
    }
    public user_id: string = ''
    public session_id: string = ''
}
