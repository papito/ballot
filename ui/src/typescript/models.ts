export enum SessionState {
  IDLE = 0,
  VOTING = 1
}

const NO_ESTIMATE: number= -1;

export class User {
  id: string = "";
  name: string = "";
  estimate: number = NO_ESTIMATE;

  static fromJson(json: {[key:string]:string}) {
    let user = new User();
    user.id = json["id"];
    user.name = json["name"];
    user.estimate = Number(json["estimate"]);
    return user;
  }
}

export class Vote {
  user_id: string = "";
  session_id: string = "";
  estimate: number = NO_ESTIMATE;

  static fromJson(json: {[key:string]:string}) {
    let vote = new Vote();
    vote.user_id = json["user_id"];
    vote.session_id = json["session_id"];
    vote.estimate = Number(json["estimate"]);
    return vote;
  }
}

export class Session {
  id: string = "";
  status: SessionState = SessionState.IDLE;
  users: User[] = [];
}
