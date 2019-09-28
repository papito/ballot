<template>
  <div id="landing">
    <div class="row">
      <div class="col-4 offset-4">
        <form>
          <div class="form-group">
            <label for="name"></label>
            <input type="text" v-model="user.name" class="form-control" id="name" placeholder="Your name/alias">
          </div>
          <button type="button" class="btn btn-lg btn-warning" v-on:click="goVote">New Voting Space</button>
        </form>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
  import {User} from "../models";
  import HttpMixin from "./HttpMixin";
  import { Component, Mixins } from 'vue-mixin-decorator';

  @Component
  export default class Landing extends Mixins<HttpMixin>(HttpMixin)  {
    user = new User();

    beforeCreate() {
      document.body.className = 'bg';
    }

    goVote() {
      const sessionCreateFut: Promise<{[key:string]:string}> = this.postRequest("/api/session", {});

      sessionCreateFut.then((sessionRes) => {
        let sessionId: string= sessionRes['id'];

        const addUserFut: Promise<{[key:string]:string}> = this.postRequest("/api/user", {
          'name': this.user.name,
          'session_id': sessionId
        });

        addUserFut.then((userRes) => {
          this.user.id = userRes['id'];

          const sessionStartFut: Promise<{[key:string]:string}> = this.putRequest("/api/vote/start", {
            'session_id': sessionId
          });
          sessionStartFut.then((sessionStartRes) => {
            this.$router.push({ name: 'ballot', params: {
                sessionId: sessionId,
                userId: this.user.id
              }
            });
          });
        });
      });
    }
  }
</script>
