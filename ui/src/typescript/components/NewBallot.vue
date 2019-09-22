<template>
  <div>
    <span v-if="sessionId">Session generated.
      <router-link :to="{ name: 'ballot', params: { sessionId: sessionId }}">Go Vote</router-link>
    </span>
    <span v-else>Generating new voting session...
    </span>

  </div>
</template>

<script lang="ts">
  import HttpMixin from "./HttpMixin";
  import { Component, Mixins } from 'vue-mixin-decorator';

  @Component
  export default class NewBallot extends Mixins<HttpMixin>(HttpMixin)  {
    sessionId: string = "";

    beforeCreate() {
      document.body.className = 'no-bg';
    }

    created () {
      const resp: Promise<{[key:string]:string}> = this.postRequest("/api/session", {});

      resp.then((res) => {
        this.sessionId = res["id"];
      });
    }
  }
</script>
