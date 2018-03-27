import Vue from 'vue'
import VueResource from 'vue-resource'
import jQuery from 'jquery'
import navBar from './nav-bar.vue'
import dashboard from './dashboard.vue'


Vue.use(require('vue-moment'));

new Vue({
  el: '#hotshots-dashboard',
  components: {
    navBar,
    dashboard
  }
})