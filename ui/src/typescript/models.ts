export enum SessionState {
  IDLE = 0,
  VOTING = 1
}

export const NO_ESTIMATE: string= '';

export class User {
  id: string = "";
  name: string = "";
  estimate: string = NO_ESTIMATE;
  voted: boolean = false;
  joined: string = "";

  static fromJson(json: {[key:string]:string}) {
    let user = new User();
    user.id = json["id"];
    user.name = json["name"];
    user.estimate = json["estimate"];
    user.voted = Boolean(json["voted"]);
    user.joined = json["joined"];
    return user;
  }
}

export class PendingVote {
  user_id: string = "";
  session_id: string = "";

  static fromJson(json: {[key:string]:string}) {
    let vote = new PendingVote();
    vote.user_id = json["user_id"];
    vote.session_id = json["session_id"];
    return vote;
  }
}

export class Session {
  id: string = "";
  tally: string = "";
  status: SessionState = SessionState.IDLE;
  users: User[] = [];

  url() {
    const el: HTMLElement | null = document.getElementById('paramDomain');
    let domain: string | null  =  "";

    if (el) {
      domain = el.textContent;
    }

    return `${domain}/#/vote/${this.id}`;
  }
}
