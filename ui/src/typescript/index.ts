import Vue from "vue";
// @ts-ignore
import Ballot from "./components/Ballot.vue";
// @ts-ignore
import Landing from "./components/Landing.vue";
import VueRouter from 'vue-router';
import Websockets from "./websockets";
import {getUrlParameter} from "./util";

const routes = [
  { path: '/', name: "landing", component: Landing },
  { path: '/vote/:sessionId', name: "ballot", component: Ballot },
];

export const ws: Websockets = new Websockets();

const router = new VueRouter({
  mode: 'history',
  routes
});

Vue.use(VueRouter);

new Vue({
  router,
  mounted: function () {
    const sessionId: string = getUrlParameter("join");
    if (sessionId) {
      router.push({ name: 'ballot', params: { sessionId: sessionId } });
    } else {
      router.replace({name: 'landing'});
    }
  }
},).$mount('#app');