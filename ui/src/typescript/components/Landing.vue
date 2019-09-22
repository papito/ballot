<template>
  <div class="row">
    <div class="col align-self-center">
      <router-link class="nav-link" to="/new">New Voting Space</router-link>
    </div>
  </div>
</template>

<script lang="ts">
  import Vue from "vue";
  import { Component } from 'vue-mixin-decorator';

  @Component
  export default class Landing extends Vue {}
</script>
