import Vue from "vue";
// @ts-ignore
import Ballot from "./components/Ballot.vue";
// @ts-ignore
import Join from "./components/Join.vue";
// @ts-ignore
import Landing from "./components/Landing.vue";
import VueRouter from 'vue-router';
import Websockets from "./websockets";
import {getUrlParameter} from "./util";

const routes = [
  { path: '/', name: "landing", component: Landing },
  { path: '/vote/:sessionId/u/:userId', name: "ballot", component: Ballot },
  { path: '/vote/:sessionId', name: "join", component: Join },
];

export const ws: Websockets = new Websockets();

const router = new VueRouter({
  mode: 'hash',
  routes
});

Vue.use(VueRouter);

new Vue({
  router
},).$mount('#app');