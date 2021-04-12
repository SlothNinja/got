<template>
  <v-app id='app'>
    <sn-toolbar v-model='nav'></sn-toolbar>
    <sn-nav-drawer v-model='nav' app></sn-nav-drawer>
    <sn-snackbar v-model='snackbar.open'>
      <div class='text-center'>
        <span v-html='snackbar.message'></span>
      </div>
    </sn-snackbar>
    <v-main>
      <v-container>
        <v-card>
          <v-card-title>
            Guild of Thieves Rankings
          </v-card-title>
          <v-card-text>
            <v-data-table
              :headers="headers"
              :items="items"
              :loading="loading"
              :options.sync="options"
              loading-text="Loading... Please wait"
              :server-items-length="totalItems"
              :items-per-page='10'
              :footer-props="{ itemsPerPageOptions: [ 10, 25, 50 ] }"
              >
              <template v-slot:item.user="{ item }">
                <sn-user-btn :user="item.user" size="x-small">{{item.user.name}}</sn-user-btn>
              </template>
              <template v-slot:item.current="{ item }">
                {{display(item.current)}}
              </template>
              <template v-slot:item.projected="{ item }">
                {{display(item.projected)}}
              </template>
            </v-data-table>
          </v-card-text>
        </v-card>
      </v-container>
    </v-main>
    <sn-footer app></sn-footer>
  </v-app>
</template>

<script>
import UserButton from '@/components/lib/user/Button'
import CurrentUser from '@/components/lib/mixins/CurrentUser'
import Toolbar from '@/components/lib/Toolbar'
import NavDrawer from '@/components/lib/NavDrawer'
import Snackbar from '@/components/lib/Snackbar'
import Footer from '@/components/lib/Footer'

const _ = require('lodash')
const axios = require('axios')

export default {
  name: 'rank',
  mixins: [ CurrentUser ],
  components: {
    'sn-user-btn': UserButton,
    'sn-toolbar': Toolbar,
    'sn-nav-drawer': NavDrawer,
    'sn-snackbar': Snackbar,
    'sn-footer': Footer
  },
  data () {
    return {
      cursors: [ "" ],
      loading: 'false',
      totalItems: 0,
      options: {},
      items: []
    }
  },
  created () {
    this.fetchData()
  },
  watch: {
    '$route': 'fetchData',
    options: {
      handler (val, oldVal) {
        if (val.itemsPerPage != oldVal.itemsPerPage) {
          this.cursors = [ "" ]
        }
        this.fetchData()
      },
      deep: true,
    },
  },
  methods: {
    fetchData: _.debounce(function () {
      let self = this
      self.loading = true
      axios.post('/rankings', { options: self.options, forward: self.forward })
        .then(function (response) {
          let msg = _.get(response, 'data.message', false)
          if (msg) {
            self.snackbar.message = msg
            self.snackbar.open = true
          }
          let totalItems = _.get(response, 'data.totalItems', false)
          if (totalItems) {
            self.totalItems = totalItems
          }
          let forward = _.get(response, 'data.forward', false)
          if (forward) {
            self.forward = forward
          }
          let gheaders = _.get(response, 'data.gheaders', false)
          if (gheaders) {
            self.items = gheaders
          }
          self.loading = false
          let cu = _.get(response, 'data.cu', false)
          if (cu) {
            self.cu = cu
            self.cuLoading = false
          }
          self.loading = false
        })
        .catch(function () {
          self.loading = false
          self.snackbar.message = 'Server Error.  Try refreshing page.'
          self.snackbar.open = true
        })
    }, 1000),
    display: rating => `${_.toInteger(rating.low)} (${_.toInteger(rating.r)}:${_.toInteger(rating.rd)})`,
  },
  computed: {
    forward: {
      get: function () {
        return this.cursors[this.options.page-1]
      },
      set: function (value) {
        this.cursors.splice(this.options.page, 1, value)
      }
    },
    headers: function () {
      return [
        { text: 'Rank', align: 'left', sortable: true, value: 'rank' },
        { text: 'User', value: 'user' },
        { text: 'Current', value: 'current' },
        { text: 'Projected', value: 'projected' },
      ]
    },
    snackbar: {
      get: function () {
        return this.$root.snackbar
      },
      set: function (value) {
        this.$root.snackbar = value
      }
    },
    nav: {
      get: function () {
        return this.$root.nav
      },
      set: function (value) {
        this.$root.nav = value
      }
    },
  }
}
</script>
