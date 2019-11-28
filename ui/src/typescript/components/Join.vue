<template>
  <div>
    <div id="join-session" class="row">
      <div class="col-sm-4 offset-sm-4">
        <form @submit.prevent="setUsername">
          <div class="form-group">
            <label for="this.user.name"></label>
            <input type="text"
                   maxlength="60"
                   v-model="user.name"
                   class="form-control"
                   id="this.user.name"
                   placeholder="Your name/alias">
            <div class="invalid-feedback">
            </div>
          </div>
          <button type="submit" class="btn btn-lg btn-warning">Join</button>
        </form>
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

    errors = {
      'user.name': ''
    };

    beforeCreate() {
      document.body.className = 'bg';
    }

    created () {
      this.session.id = this.$route.params["sessionId"];
      console.log(`Joining session ${this.session.id}`);
    }

    setUsername() {
      this.postRequest(
          "/api/user", {
            "name": this.user.name,
            "session_id": this.session.id
          }
      ).then((resp) => {
        return resp.json();
      }).then((json) => {
        this.user.id = json["id"];

        this.$router.push({name: 'ballot', params: {
            sessionId: this.session.id,
            userId: this.user.id
          }});
      }).catch((err: Error) => {
        this.showError(err);
      });
    }
  }
</script>
