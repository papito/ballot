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
  <div id="landing">
    <div class="row">
      <div class="col-sm-4 offset-sm-4">
        <form @submit.prevent="goVote">
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
      this.postRequest("/api/session", {}).then((resp) => {
        return resp.json();
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
      }).catch((err: Error) => {
        this.showError(err);
      });
    }
  }
</script>
