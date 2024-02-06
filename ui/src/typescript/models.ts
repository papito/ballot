/*
 * The MIT License
 *
 * Copyright (c) 2020,  Andrei Taranchenko
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

export const NO_ESTIMATE: string = '';

export class User {

  public static fromJson(json: {[key: string]: string}) {
    const user = new User();
    user.id = json.id;
    user.name = json.name;
    user.estimate = json.estimate;
    user.voted = Boolean(json.voted);
    user.joined = json.joined;
    user.is_observer = Boolean(json.is_observer);
    user.is_admin = Boolean(json.is_admin);
    return user;
  }
  public id: string = '';
  public name: string = '';
  public estimate: string = NO_ESTIMATE;
  public voted: boolean = false;
  public joined: string = '';
  // tslint:disable-next-line:variable-name
  public is_observer: boolean = false;
  // tslint:disable-next-line:variable-name
  public is_admin: boolean = false;
}

// tslint:disable-next-line:max-classes-per-file
export class PendingVote {

  public static fromJson(json: {[key: string]: string}) {
    const vote = new PendingVote();
    vote.user_id = json.user_id;
    vote.session_id = json.session_id;
    return vote;
  }
  // tslint:disable-next-line:variable-name
  public user_id: string = '';
  // tslint:disable-next-line:variable-name
  public session_id: string = '';
}

// tslint:disable-next-line:max-classes-per-file
export class Session {
  public id: string = '';
  public tally: string = '';
  public status: SessionState = SessionState.IDLE;
  public users: User[] = [];
  public observers: User[] = [];

  public url(): string {
    const domain: string | null | undefined =
        document.getElementById('paramDomain')?.textContent;
    return `${domain}/#/vote/${this.id}`;
  }
}
