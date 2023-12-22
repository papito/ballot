<!--
  - The MIT License
  -
  - Copyright (c) 2020,  Andrei Taranchenko
  -
  - Permission is hereby granted, free of charge, to any person obtaining a copy
  - of this software and associated documentation files (the "Software"), to deal
  - in the Software without restriction, including without limitation the rights
  - to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
  - copies of the Software, and to permit persons to whom the Software is
  - furnished to do so, subject to the following conditions:
  -
  - The above copyright notice and this permission notice shall be included in
  - all copies or substantial portions of the Software.
  -
  - THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
  - IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
  - FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
  - AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
  - LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
  - OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
  - THE SOFTWARE.
  -->

<template>
  <div>
    <div id="ctrl-panels" class="row no-gutters">
      <div class="col-12">
        <div class="row">
          <div class="col-7 text-center">
            <div v-show="isAdmin" id="start-ctrl-panel">
              <button v-if="isIdle"
                      v-on:click="startVote"
                      class="btn btn-outline-success">
                <span class="oi oi-media-play icon" aria-hidden="true"></span>Start
              </button>

              <button v-if="isVoting"
                      v-on:click="finishVote"
                      class="btn btn-outline-danger">
                <span class="oi oi-media-stop icon" aria-hidden="true"></span>End Vote Now
              </button>
            </div>
          </div>

          <div class="col-5">
            <div id="copy-ctrl-panel">
              <form @submit.prevent="copyJoinUrl">
                <div class="row no-gutters">
                  <div class="d-none d-lg-block col-md-8 text-right">
                    <label for="sessionUrl" class="sr-only"></label>
                    <input type="text" :value="session.url()" class="form-control form-control-sm" id="sessionUrl" readonly>
                  </div>
                  <div class="d-none d-lg-block col-md-4 text-left">
                    <button v-on:click="copyJoinUrl" class="btn btn-outline-light btn-sm">
                      <span class="oi oi-clipboard icon" aria-hidden="true"></span>Copy URL
                    </button>
                  </div>
                  <div class="col-12 d-lg-none text-center">
                    <button v-on:click="copyJoinUrl" class="btn btn-outline-light">
                      Copy URL
                    </button>
                  </div>
                </div>
              </form>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div id="choices" v-show="isVoting && !user.is_observer">
      <div v-for="estimate in possibleEstimates" class="choice">
        <button :value="estimate"
                v-on:click="castVote(estimate)"
                class="btn btn-outline-warning btn-sm"
                v-bind:class="{ active: estimate === user.estimate }">
          {{estimate}}
        </button>
      </div>
      <div v-show="!isVoting" v-for="estimate in possibleEstimates" class="choice">
        <button class="btn btn-outline-warning btn-sm">&nbsp;&nbsp;</button>
      </div>
    </div>

    <div id="tally" v-show="isIdle" v-bind:class="{ observer: user.is_observer, observed: !user.is_observer }">
      {{session.tally}}
    </div>

    <div id="observers" v-show="session.observers.length > 0" v-bind:class="{ observer: user.is_observer, observed: !user.is_observer }">
      <strong v-show="!user.is_observer">Watching you</strong>
      <strong v-show="user.is_observer">Watchers</strong>: <span v-text="observerNames"/>
    </div>

    <div id="voters">
      <div v-for="u in session.users" class="card" v-bind:class="{ voted: u.voted }">
        <div class="name">
          <span class="name-text" v-show="u.id !== user.id">
          {{ u.name }}
          </span>
          <span class="name-text" v-show="u.id === user.id">
            [&nbsp;{{ u.name }}&nbsp;]
          </span>
        </div>
        <div>
          <img v-show="u.voted" src="/ui/img/v.png">
          <img v-show="!u.voted" src="/ui/img/x.png">
        </div>
        <div class="estimate" v-show="u.estimate !== ''">{{ u.estimate }}</div>
        <div class="estimate" v-show="u.estimate === ''">--</div>
      </div>
    </div>

  </div>
</template>

<script lang="ts">
  import HttpMixin from "./HttpMixin";
  import { Component, Mixins } from 'vue-mixin-decorator';
  import {ws} from "../index";
  import {Session, SessionState, User, PendingVote, NO_ESTIMATE} from "../models";

  @Component
  export default class Ballot extends Mixins<HttpMixin>(HttpMixin)  {
    session: Session = new Session();
    user: User = new User();
    observerNames: string = "";

    readonly possibleEstimates:string[] = Array('?', '0', '1', '2', '3', '5', '8', '13', '20', '40', '100');

    errors = {
      'user.name': ''
    };

    beforeCreate() {
      document.body.className = 'bg';
    }

    created () {
      this.session.id = this.$route.params["sessionId"];
      this.user.id = this.$route.params["userId"];
      console.log(`Voting for session [${this.session.id}] as user [${this.user.id}]`);

      ws.socket.onMessage((data: string) => {
        console.log("onMessage: " + data);
        let json = JSON.parse(data);
        let event = json["event"];

        switch (event) {
          case "USER_ADDED": {
            this.userAddedWsHandler(json);
            break;
          }
          case "OBSERVER_ADDED": {
            this.observerAddedWsHandler(json);
            break;
          }
          case "WATCHING": {
            this.watchingSessionWsHandler(json);
            break;
          }
          case "VOTING": {
            this.votingStartedWsHandler();
            break;
          }
          case "USER_VOTED": {
            this.userVotedHandler(json);
            break;
          }
          case "VOTE_FINISHED": {
            this.votingFinishedWsHandler(json);
            break
          }
          case "USER_LEFT": {
            this.userLeftWsHandler(json);
            break
          }
          case "OBSERVER_LEFT": {
            this.observerLeftWsHandler(json);
            break
          }
        }
      });

      this.getRequest(`/api/user/${this.user.id}`).then((resp) => {
        return resp.json();
      }).then((json) => {
        this.user = User.fromJson(json);
        let watchCmd = {
          "action": "WATCH",
          "session_id": this.session.id,
          "user_id": this.user.id,
          "is_observer": this.user.is_observer,
          "is_admin": this.user.is_admin
        };
        ws.send(JSON.stringify(watchCmd));
      }).catch((err: Error) => {
        this.showError(err);
      });
    }

    clearState() {
      this.user.estimate = NO_ESTIMATE;
      this.session.tally = NO_ESTIMATE;
    }

    userAddedWsHandler(json: {[key:string]:string}) {
      let user = User.fromJson(json);
      if (user.id == this.user.id) { // do not re-add yourself
        return;
      }

      /* To address a race condition that might happen when a user
         refreshes - check that this is not a duplicate.
       */
      let existingUser = this.session.users.find(function(u) {
        return u.id == user.id;
      });
      if (existingUser) {
        return;
      }

      this.session.users.push(user);
      this.session.users.sort((a, b) => a.joined.localeCompare(b.joined))
    }

    observerAddedWsHandler(json: {[key:string]:string}) {
      let observer = User.fromJson(json);
      if (observer.id == this.user.id) { // do not re-add yourself
        return;
      }

      /* To address a race condition that might happen when a user
         refreshes - check that this is not a duplicate.
       */
      let existingUser = this.session.users.find(function(u) {
        return u.id == observer.id;
      });
      if (existingUser) {
        return;
      }

      this.session.observers.push(observer);
      this.session.observers.sort((a, b) => a.joined.localeCompare(b.joined));
      this.observerNames = this.getObserverNames();
    }

    userVotedHandler(json: {[key:string]:string}) {
      let voteArg: PendingVote = PendingVote.fromJson(json);

      // find and update the user
      let user = this.session.users.find(function(u) {
        return u.id == voteArg.user_id;
      });

      if (user == undefined) {
        throw new Error(`Could not find user by ID ${voteArg.user_id}`);
      }
      user.voted = true;
      console.log("USER VOTED " + user.id)
    }

    watchingSessionWsHandler(json: {[key:string]:any}) {
      this.session.status = json["session_state"];
      this.session.tally = json["tally"];

      let users: any = json["users"] || [];
      for (let userJson of users) {
        let user = User.fromJson(userJson);
        this.session.users.push(user);
      }

      let observers: any = json["observers"] || [];
      for (let observerJson of observers) {
        let observer = User.fromJson(observerJson);
        this.session.observers.push(observer);
      }

      this.observerNames = this.getObserverNames();
    }

    votingStartedWsHandler() {
      this.session.status = SessionState.VOTING;
      this.clearState();

      for (let user of this.session.users) {
        user.estimate = NO_ESTIMATE;
        user.voted = false;
      }
    }

    votingFinishedWsHandler(json: {[key:string]:any}) {
      console.log("vote finished");
      this.session.status = SessionState.IDLE;
      let usersJson: any = json["users"] || [];
      let tally: string = json["tally"];

      for (let userJson of usersJson) {
        let user = User.fromJson(userJson);

        // find and update the user
        let sessionUser = this.session.users.find(function(u) {
          return u.id == user.id;
        });
        if (sessionUser == undefined) {
          throw new Error(`Could not find user by ID ${user.id}`);
        }
        sessionUser.voted = user.voted;
        sessionUser.estimate = user.estimate;
        this.session.tally = tally;
      }
    }

    userLeftWsHandler(json: {[key:string]:any}) {
      let userId = json['user_id'];
      this.session.users = this.session.users.filter(function (user: User) {
        return user.id != userId;
      });
    }

    observerLeftWsHandler(json: {[key:string]:any}) {
      let userId = json['user_id'];
      this.session.observers = this.session.observers.filter(function (user: User) {
        return user.id != userId;
      });
      this.observerNames = this.getObserverNames();
    }

    get isVoting() {
      return this.session.status == SessionState.VOTING;
    }

    get isIdle() {
      return this.session.status == SessionState.IDLE;
    }

    get isAdmin() {
      console.log("!!!!");
      console.log(this.user.is_admin);
      return this.user.is_admin;
    }

    startVote() {
      this.putRequest(
          "/api/vote/start", {
            "session_id": this.session.id
          }
      ).then(() => {
        this.clearState();
      }).catch((err: Error) => {
        this.showError(err);
      });
    }

    finishVote() {
      this.putRequest(
          "/api/vote/finish", {
            "session_id": this.session.id
          }
      ).catch((err: Error) => {
        this.showError(err);
      });
    }

    castVote(estimate: string) {
      this.putRequest(
          "/api/vote/cast", {
            "session_id": this.session.id,
            "user_id": this.user.id,
            "estimate": estimate
          }
      ).then(() => {
        // the server will not return the estimate but we need to store it for our own state
        this.user.estimate = estimate;
      }).catch((err: Error) => {
        this.showError(err);
      });
    }

    copyJoinUrl() {
      function copyToClipboard(text: string) {
        let dummy = document.createElement("textarea");
        document.body.appendChild(dummy);
        dummy.value = text;
        dummy.select();
        document.execCommand("copy");
        document.body.removeChild(dummy);
      }

      copyToClipboard(this.session.url());
    }

    getObserverNames(): string {
      let userNames = new Array<String>();

      for (let obs of this.session.observers) {
        userNames.push(obs.name);
      }
      return userNames.join(", ");
    }
  }
</script>
