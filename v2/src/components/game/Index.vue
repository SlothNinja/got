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
            {{ status }} Games
          </v-card-title>
          <v-data-table
            @click:row="showGame"
            :headers="headers"
            :items="items"
            :loading="loading"
            :options.sync="options"
            loading-text="Loading... Please wait"
            :server-items-length="totalItems"
            :footer-props="{ itemsPerPageOptions: [ 25, 50, 100 ] }"
            >
            <template v-slot:item.creator="{ item }">
              <sn-user-btn :user="creator(item)" size="x-small"></sn-user-btn>&nbsp;{{creator(item).name}}
            </template>
            <template v-slot:item.players="{ item }">
              <span class="px-1" v-for="user in users(item)" :key="user.id" >
                <sn-user-btn :user="user" size="x-small"></sn-user-btn>&nbsp;<span :class='cpClass(item, user)'>{{user.name}}</span>
              </span>
            </template>
          </v-data-table>
        </v-card>
      </v-container>
    </v-main>
    <sn-footer app></sn-footer>
  </v-app>
</template>

<script>
import UserButton from '@/components/user/Button'
import Toolbar from '@/components/Toolbar'
import NavDrawer from '@/components/NavDrawer'
import Snackbar from '@/components/Snackbar'
import Footer from '@/components/Footer'
import CurrentUser from '@/components/mixins/CurrentUser'

const _ = require('lodash')
const axios = require('axios')

export default {
  name: 'index',
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
      cursorMap: new Map,
      loading: 'false',
      totalItems: 0,
      forward: "",
      options: {},
      items: []
    }
  },
  created () {
    this.$root.toolbar = 'sn-toolbar'
    this.fetchData()
  },
  watch: {
    '$route': 'fetchData',
    options: {
      handler (val) {
        this.fetchData(val)
      },
      deep: true,
    },
  },
  mounted () {
    this.fetchData()
  },
  methods: {
    fetchOptions: function(val) {
      let self = this
      let key = JSON.stringify(val)
      let options = self.cursorMap.get(key)
      if (options) {
        return options
      }
      options = {
        page: val.page,
        itemsPerPage: val.itemsPerPage,
        forward: val.page == 1 ? "" : self.forward,
      }
      self.cursorMap.set(key, options)
      return options
    },
    showGame: function (item) {
      this.$router.push({name: 'game', params: { id: item.id } })
    },
    fetchData: _.debounce(function (val) {
      let self = this
      self.loading = true
      let options = self.fetchOptions(val)

      axios.post(`/games/${self.$route.params.status}`, options)
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
    action: function (action, id) {
      let self = this
      axios.put(`/game/${action}/${id}`)
        .then(function (response) {
          let msg = _.get(response, 'data.message', false)
          if (msg) {
            self.snackbar.message = msg
            self.snackbar.open = true
          }
          let header = _.get(response, 'data.header', false)
          if (header) {
            let index = _.findIndex(self.items, [ 'id', id ])
            if (index >= 0) {
              if (header.status === 1) { // recruiting is a status of 1
                self.items.splice(index, 1, header)
              } else {
                self.items.splice(index, 1)
              }
            }
          }
          self.loading = false
        })
        .catch(function () {
          self.loading = false
          self.snackbar.message = 'Server Error.  Try refreshing page.'
          self.snackbar.open = true
        })
    },
    canAccept: function (id) {
      let self = this
      let item = self.getItem(id)
      return !self.joined(item) && item.status === 1 // recruiting is a status 1
    },
    canDrop: function (id) {
      let self = this
      let item = self.getItem(id)
      return self.joined(item) && item.status === 1 // recruiting is a status 1
    },
    joined: function (item) {
      let self = this
      return _.find(item.users, [ 'id', self.cuid ])
    },
    getItem: function (id) {
      let self = this
      return _.find(self.items, [ 'id', id ])
    },
    publicPrivate: function (item) {
      return item.public ? 'Public' : 'Private'
    },
    creator: function (item) {
      return {
        id: item.creatorId,
        name: item.creatorName,
        emailHash: item.creatorEmailHash,
        gravType: item.creatorGravType
      }
    },
    users: function (item) {
      return _.map(item.userIds, function (id, i) {
        return {
          id: id,
          name: item.userNames[i],
          emailHash: item.userEmailHashes[i],
          gravType: item.userGravTypes[i],
        }
      })
    },
    cpClass: function (item, user) {
      let pid = _.indexOf(item.userIds, user.id) + 1
      let cpid = _.first(item.cpids)
      if (pid == cpid) {
        if (this.cuid == user.id) {
          return 'font-weight-black red--text text--darken-4'
        }
        return 'font-weight-black'
      }
      return ''
    },
  },
  computed: {
    headers () {
      return [
        { text: 'ID', align: 'left', sortable: false, value: 'id' },
        { text: 'Title', sortable: false, value: 'title' },
        { text: 'Creator', sortable: false, value: 'creator' },
        { text: 'Players', sortable: false, value: 'players' },
        { text: 'Last Updated', sortable: false, value: 'lastUpdated' },
      ]
    },
    status: function () {
      return _.capitalize(this.$route.params.status)
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

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h1, h2, h3 {
  font-weight: normal;
}
</style>
