<!--
  - The MIT License
  -
  - Copyright (c) 2019,  Andrei Taranchenko
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
    <div id="join-session" class="row">
      <div class="col-sm-4 offset-sm-4">
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
        <button v-on:click="joinAsVoter" class="btn btn-lg btn-success">Join as a voter</button>
        <button v-on:click="joinAsObserver" class="btn btn-lg btn-warning">Join as an innocent observer</button>
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
    is_observer = 0;

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

    joinAsVoter() {
      this.join();
    }

    joinAsObserver() {
      this.is_observer = 1;
      this.join();
    }

    join() {
      this.postRequest(
          "/api/user", {
            "name": this.user.name,
            "session_id": this.session.id,
            "is_observer": this.is_observer
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
