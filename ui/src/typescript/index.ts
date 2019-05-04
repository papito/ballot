import Vue from "vue";
import NewBallot from './components/NewBallot.vue';
import Ballot from "./components/Ballot.vue";
import VueRouter from 'vue-router';
import Websockets from "./websockets";
import {getUrlParameter} from "./util";

const routes = [
  { path: '/new', component: NewBallot },
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
    }
  }
},).$mount('#app');
