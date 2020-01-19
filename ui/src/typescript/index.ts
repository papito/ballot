/*
 * The MIT License
 *
 * Copyright (c) 2020,  Andrei Taranchenko
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

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

Vue.directive('focus', {
  // When the bound element is inserted into the DOM...
  inserted: function (el) {
    // Focus the element
    el.focus()
  }
});