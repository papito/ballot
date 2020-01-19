/*
 * The MIT License
 *
 * Copyright (c) 2019,  Andrei Taranchenko
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

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
  is_observer: boolean = false;

  static fromJson(json: {[key:string]:string}) {
    let user = new User();
    user.id = json["id"];
    user.name = json["name"];
    user.estimate = json["estimate"];
    user.voted = Boolean(json["voted"]);
    user.joined = json["joined"];
    user.is_observer = Boolean(json["is_observer"]);
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
  observers: User[] = [];

  url(): string {
    let domain: string | null | undefined =
        document.getElementById('paramDomain')?.textContent;
    return `${domain}/#/vote/${this.id}`;
  }
}
