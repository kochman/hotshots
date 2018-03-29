import Vue from 'vue'
import VueMoment from 'vue-moment'
import navBar from './nav-bar.vue'
import dashboard from './dashboard.vue'

const $ = require('jquery');
window.jQuery = $;
window.$ = $;

window.axios = require('axios');
Vue.prototype.$http = window.axios;

Vue.use('VueMoment');

new Vue({
  el: '#hotshots-dashboard',
  components: {
    navBar,
    dashboard
  }
})