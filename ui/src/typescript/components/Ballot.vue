<template>
  <div>
    <div id="ctrl-panels" class="row">
      <div id="start-ctrl-panel" class="col-6">
        <div>
          <button v-if="isIdle" v-on:click="startVote" class="btn btn-outline-success">
            Start
          </button>

          <button v-if="isVoting" v-on:click="finishVote" class="btn btn-outline-danger">
            End Vote Now
          </button>
        </div>
      </div>


        <div id="copy-ctrl-panel" class="col-6">
          <form @submit.prevent="copyJoinUrl">
          <div class="row">
            <div class="col-8">
              <label for="sessionUrl" class="sr-only"></label>
              <input type="text" :value="session.url()" class="form-control form-control-sm" id="sessionUrl" readonly>
            </div>
            <div class="col-4">
              <button v-on:click="copyJoinUrl" class="btn btn-outline-light btn-sm mb-2">
                Copy session url
              </button>
            </div>
          </div>
          </form>
        </div>

    </div>

    <div id="choices" v-show="isVoting">
      <div v-for="estimate in possibleEstimates" class="choice">
        <button :value="estimate" v-on:click="castVote(estimate)" class="btn btn-outline-warning">{{estimate}}</button>
      </div>
      <div v-show="!isVoting" v-for="estimate in possibleEstimates" class="choice">
        <button  class="btn btn-outline-warning">&nbsp;&nbsp;</button>
      </div>
    </div>

    <div id="voters">
      <div v-for="user in session.users" class="card" v-bind:class="{ voted: user.voted }">
        {{ user.name }}
        <div>
          <img v-show="user.voted" src="/ui/img/v.png">
          <img v-show="!user.voted" src="/ui/img/x.png">
        </div>
        <div class="estimate" v-show="user.estimate !== ''">{{ user.estimate }}</div>
        <div class="estimate" v-show="user.estimate === ''">--</div>
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

    readonly possibleEstimates:string[] = Array('¯\\_(ツ)_/¯', '0', '1', '2', '3', '5', '8', '13', '20', '40', '100');

    errors = {
      'user.name': ''
    };

    beforeCreate() {
      document.body.className = 'bg';
    }

    created () {
      this.session.id = this.$route.params["sessionId"];
      console.log(`Voting for session ${this.session.id}`);

      ws.socket.onMessage((data: string) => {
        console.log("onMessage: " + data);
        let json = JSON.parse(data);
        let event = json["event"];

        switch (event) {
          case "USER_ADDED": {
            this.userAddedWsHandler(json);
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
        }
      });

      let user_id = this.$route.params["userId"];
      console.log(`Voting as user ID ${user_id}`);
      if (user_id) {
        this.user.id = user_id;
        this.getRequest(`/api/user/${user_id}`).then((resp) => {
          return resp.json();
        }).then((json) => {
          this.user = User.fromJson(json);
          let watchCmd = {
            "action": "WATCH",
            "session_id": this.session.id,
            "user_id": user_id
          };
          ws.send(JSON.stringify(watchCmd));
        }).catch((err: Error) => {
          this.showError(err);
        });
      }
    }

    userAddedWsHandler(json: {[key:string]:string}) {
      let user = User.fromJson(json);
      if (user.id != this.user.id) {
        this.session.users.push(user);
      }
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

      let users: any = json["users"] || [];
      for (let userJson of users) {
        let user = User.fromJson(userJson);
        this.session.users.push(user);
      }
    }

    votingStartedWsHandler() {
      console.log("updating session state");
      this.session.status = SessionState.VOTING;

      for (let user of this.session.users) {
        user.estimate = NO_ESTIMATE;
        user.voted = false;
      }
    }

    votingFinishedWsHandler(json: {[key:string]:any}) {
      console.log("vote finished");
      this.session.status = SessionState.IDLE;
      let usersJson: any = json["users"] || [];

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
      }
    }

    userLeftWsHandler(json: {[key:string]:any}) {
      console.log("user left");

      let userId = json['user_id'];

      this.session.users = this.session.users.filter(function (user: User) {
        return user.id != userId;
      });
    }

    get isVoting() {
      return this.session.status == SessionState.VOTING;
    }

    get isIdle() {
      return this.session.status == SessionState.IDLE;
    }

    startVote() {
      this.putRequest(
          "/api/vote/start", {
            "session_id": this.session.id
          }
      ).catch((err: Error) => {
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
      ).catch((err: Error) => {
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
  }
</script>
