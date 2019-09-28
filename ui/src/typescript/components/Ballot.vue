<template>
  <div>
    <div>
      Voting for session {{session.id}}
    </div>

    <ul id="voters">
      <li v-for="user in session.users">
        {{ user.name }}
        <span v-show="user.estimate !== ''">{{ user.estimate }}</span>
        &nbsp;
        <span v-show="user.voted">Voted</span>
      </li>
    </ul>

    <div v-show="!user.id">
      <form @submit.prevent="setUsername">
        <div>
          <span v-show="errors.name">{{errors.name}}</span>
        </div>
        <div>
          <label for="user.name">Enter voter name:</label>
          <input id="user.name" type="text" name="user.name" maxlength="100" v-model="user.name">
        </div>
      </form>
    </div>


    <div v-show="user.id">
      <button v-if="isIdle" v-on:click="startVote">
        Start
      </button>

      <button v-if="isVoting" v-on:click="finishVote">
        End Vote Now
      </button>
    </div>

    <div v-show="user.id && isVoting" v-for="estimate in possibleEstimates">
      <button :value="estimate" v-on:click="castVote">{{estimate}}</button>
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

    readonly possibleEstimates:Number[] = Array(0, 1, 2, 3, 5, 8, 13, 20, 40, 100);

    errors = {
      'user.name': ''
    };

    beforeCreate() {
      document.body.className = 'no-bg';
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
        }
      });

      // start watching the session
      let watchCmd = {
        "action": "WATCH",
        "session_id": this.session.id
      };

      let user_id = this.$route.params["userId"];
      // if we are joining as an existing user (creator), get user details and assign to current
      // user object
      if (user_id) {
        const userResp: Promise<{[key:string]:string}> = this.getRequest(`/api/user/${user_id}`);
        userResp.then((userRes) => {
          this.user = User.fromJson(userRes);
          ws.send(JSON.stringify(watchCmd));
        });
      } else {
        ws.send(JSON.stringify(watchCmd));
      }
    }

    userAddedWsHandler(json: {[key:string]:string}) {
      let user = User.fromJson(json);
      this.session.users.push(user);
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

    get isVoting() {
      return this.session.status == SessionState.VOTING;
    }

    get isIdle() {
      return this.session.status == SessionState.IDLE;
    }

    startVote() {
      const resp: Promise<{[key:string]:string}> = this.putRequest(
          "/api/vote/start", {
            "session_id": this.session.id
          }
      );

      resp.then((res) => {
        // TODO: handle errors for user who is trying to start the vote
        // if (res.ok) {
        // }
      });
    }

    finishVote() {
      const resp: Promise<{[key:string]:string}> = this.putRequest(
          "/api/vote/finish", {
            "session_id": this.session.id
          }
      );

      resp.then((res) => {
        // TODO: handle errors for user who is trying to end the vote
        // if (res.ok) {
        // }
      });
    }

    setUsername() {
      const resp: Promise<{[key:string]:string}> = this.postRequest(
          "/api/user", {
            "name": this.user.name,
            "session_id": this.session.id
          }
      );

      resp.then((res) => {
        this.user.id = res["id"];
        this.user.name = res["name"];
      });

      console.log("setting user to", this.user);
    }

    castVote(event: Event) {
      // FIXME: May not be a number
      let estimate: Number = Number((<HTMLInputElement>event.target).value);

      const resp: Promise<{[key:string]:string}> = this.putRequest(
          "/api/vote/cast", {
            "session_id": this.session.id,
            "user_id": this.user.id,
            "estimate": estimate
          }
      );
    }
  }
</script>
