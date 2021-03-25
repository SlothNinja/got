import Vue from 'vue'
import App from './App.vue'
import vuetify from './plugins/vuetify'
import router from './router/router'
import { Plugin } from 'vue-fragment'
import visibility from 'vue-visibility-change';

Vue.use(visibility)
Vue.use(Plugin)
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
  router,
  render: h => h(App),
}).$mount('#app')
