<template>
  <div id="landing">
    <div class="row">
      <div class="col">
        <router-link class="nav-link" to="/new">
          <button type="button" class="btn btn-lg btn-warning">New Voting Space</button>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
  import Vue from "vue";
  import { Component } from 'vue-mixin-decorator';

  @Component
  export default class Landing extends Vue {
    beforeCreate() {
      document.body.className = 'bg';
    }
  }
</script>
