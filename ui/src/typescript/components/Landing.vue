<template>
  <div id="landing">
    <div class="row">
      <div class="col-4 offset-4">
        <form @submit.prevent="goVote">
          <div class="form-group">
            <label for="name"></label>
            <input type="text" v-model="user.name" class="form-control" id="name" placeholder="Your name/alias">
            <div class="invalid-feedback">
            </div>
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
    sessionId: string = "";

    beforeCreate() {
      document.body.className = 'bg';
    }

    goVote() {
      this.postRequest("/api/session", {}).then((sessionRes) => {
        return sessionRes.json();
      }).then((json) => {
        this.sessionId = json['id'];
        return Promise.resolve(this.sessionId);
      }).then(() => {
        return this.postRequest("/api/user", {
          'name': this.user.name,
          'session_id': this.sessionId
        });
      }).then((resp) => {
        return resp.json()
      }).then((json) => {
        this.user.id = json['id'];
        this.putRequest("/api/vote/start", {
          'session_id': this.sessionId
        });
      }).then(() => {
        this.$router.push({name: 'ballot', params: {
            sessionId: this.sessionId,
            userId: this.user.id
          }});
      })
    }
  }
</script>
