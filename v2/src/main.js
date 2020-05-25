import Vue from 'vue'
import App from './App.vue'
import vuetify from './plugins/vuetify'
import router from './router/router'
import axios from 'axios'
import { Plugin } from 'vue-fragment'

Vue.use(Plugin)

const _ = require('lodash')

Vue.config.productionTip = false

new Vue({
  vuetify,
  data () {
    return {
      dev: false,
      cu: null,
      cuLoading: true,
      idToken: '',
      nav: false,
      animate: true,
      extendedToolbar: 'sn-toolbar-extension-none',
      snackbar: { open: false, message: '' }
    }
  },
  created () {
      var self = this
      self.fetchCurrentUser()
  },
  methods: {
    fetchCurrentUser () {
      var self = this
      axios.get('/user/current')
        .then(function (response) {
          var cu = _.get(response, 'data.cu', false)
          if (cu) {
            self.cu = cu
          }
                
          var dev = _.get(response, 'data.dev', false)
          if (dev) {
            self.dev = dev
          }

          self.cuLoading = false
        })
        .catch(function () {
          self.snackbar.message = 'Server Error.  Try again.'
          self.snackbar.open = true
          self.$router.push({ name: 'show', params: { id: self.$route.params.id}})
        })
    },
  },
  router,
  render: h => h(App),
}).$mount('#app')
