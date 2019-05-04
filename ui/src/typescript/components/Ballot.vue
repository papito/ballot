<template>
  <div>
    <div>
      Voting for session {{session.id}}
    </div>

    <ul id="voters">
      <li v-for="user in session.users">
        {{ user.name }}&nbsp;&nbsp;<span v-show="user.estimate >= 0">{{ user.estimate }}</span>
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
      <button v-on:click="startVote">
        <span v-if="isVoting">
          Restart
        </span>
        <span v-else-if="isIdle">
          Start
        </span>
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
  import {Session, SessionState, User, Vote} from "../models";

  @Component
  export default class Ballot extends Mixins<HttpMixin>(HttpMixin)  {
    session: Session = new Session();
    user: User = new User();

    readonly possibleEstimates:Number[] = Array(0, 1, 2, 3, 5, 8, 13, 20, 40, 100);

    errors = {
      'user.name': ''
    };

    created () {
      this.session.id = this.$route.params["sessionId"];
      console.log(`Voting for session ${this.session.id}`);

      // start watching the session
      let data = {
        "action": "WATCH",
        "session_id": this.session.id}
      ;
      ws.send(JSON.stringify(data));

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
        }
      });
    }

    userAddedWsHandler(json: {[key:string]:string}) {
      let user = User.fromJson(json);
      this.session.users.push(user);
    }

    userVotedHandler(json: {[key:string]:string}) {
      let voteArg: Vote = Vote.fromJson(json);

      // find and update the user
      let user = this.session.users.find(function(u) {
        return u.id == voteArg.user_id;
      });

      if (user == undefined) {
        throw new Error(`Could not find user by ID ${voteArg.user_id}`);
      }

      user.estimate = voteArg.estimate;
      console.log("USER VOTED " + user.estimate)
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
        if (res.ok) {
        }
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
