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
          <v-card-title primary-title>
            <h3>{{ status }} Games</h3>
          </v-card-title>
          <v-card-text>
            <v-data-table
              :headers="headers"
              :items="items"
              >
              <template v-slot:item.creator="{ item }">
                <sn-user-btn :user="creator(item)" size="x-small"></sn-user-btn>&nbsp;{{creator(item).name}}
              </template>
              <template v-slot:item.players="{ item }">
                <div class="py-1" v-for="user in users(item)" :key="user.id" >
                  <sn-user-btn :user="user" size="x-small"></sn-user-btn>&nbsp;<span :class='cpClass(item, user)'>{{user.name}}</span>
                </div>
              </template>
              <template v-slot:item.public="{ item }">
                {{publicPrivate(item)}}
              </template>
              <template v-slot:item.actions="{ item }">
                <v-btn 
                        x-small
                        rounded
                        width='62'
                        v-if="canAccept(item.id)"
                        @click="action('accept', item.id)"
                        color='info'
                        dark
                        >
                        Accept
                </v-btn>
                  <v-btn 
                        x-small
                        rounded
                        width='62'
                        v-if="canDrop(item.id)"
                        @click="action('drop', item.id)"
                        color='info'
                        dark
                        >
                        Drop
                  </v-btn>
                    <v-btn 
                        x-small
                        rounded
                        width='62'
                        v-if="status == 'Running'"
                        :to="{ name: 'game', params: { id: item.id }}"
                        color='info'
                        dark
                        >
                        Show
                    </v-btn>
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
      headers: [
        {
          text: 'ID',
          align: 'left',
          sortable: true,
          value: 'id'
        },
        { text: 'Title', value: 'title' },
        { text: 'Creator', value: 'creator' },
        { text: 'Num Players', value: 'numPlayers' },
        { text: 'Players', value: 'players' },
        { text: 'Last Updated', value: 'lastUpdated' },
        { text: 'Public/Private', value: 'public' },
        { text: 'Actions', value: 'actions' }
      ],
      items: []
    }
  },
  created () {
    this.$root.toolbar = 'sn-toolbar'
    this.fetchData()
  },
  watch: {
    '$route': 'fetchData'
  },
  methods: {
    fetchData: function () {
      let self = this
      axios.get(`/games/${self.$route.params.status}`)
        .then(function (response) {
          console.log(`response: ${JSON.stringify(response)}`)
          let msg = _.get(response, 'data.message', false)
          if (msg) {
            self.snackbar.message = msg
            self.snackbar.open = true
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
        })
        .catch(function () {
          self.loading = false
          self.snackbar.message = 'Server Error.  Try refreshing page.'
          self.snackbar.open = true
        })
    },
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
      return (pid == item.cpids[0]) ? 'font-weight-black red--text text--darken-4' : ''
    },
  },
  computed: {
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
