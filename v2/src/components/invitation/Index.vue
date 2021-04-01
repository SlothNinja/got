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
            Invitations
          </v-card-title>
          <v-data-table
            @click:row="expandRow"
            :expanded.sync="expanded"
            :headers="headers"
            :items="items"
            :loading="loading"
            :options.sync="options"
            loading-text="Loading... Please wait"
            :server-items-length="totalItems"
            :items-per-page='10'
            :footer-props="{ itemsPerPageOptions: [ 10, 25, 50 ] }"
            show-expand
            single-expand
            >
            <template v-slot:item.creator="{ item }">
              <sn-user-btn :user="creator(item)" size="x-small"></sn-user-btn>&nbsp;{{creator(item).name}}
            </template>
            <template v-slot:item.players="{ item }">
              <span class="py-1" v-for="user in users(item)" :key="user.id" >
                <sn-user-btn :user="user" size="x-small"></sn-user-btn>&nbsp;{{user.name}}
              </span>
            </template>
            <template v-slot:item.public="{ item }">
              {{publicPrivate(item)}}
            </template>
            <template v-slot:expanded-item="{ headers, item }">
              <sn-expanded-row
                :span='headers.length'
                :item='item'
                @action='action($event)'
                >
              </sn-expanded-row>
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
import Expansion from '@/components/invitation/Expansion'
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
    'sn-expanded-row': Expansion,
    'sn-toolbar': Toolbar,
    'sn-nav-drawer': NavDrawer,
    'sn-snackbar': Snackbar,
    'sn-footer': Footer
  },
  data () {
    return {
      cursors: [ "" ],
      expanded: [],
      loading: 'false',
      totalItems: 0,
      options: {},
      password: '',
      show: false,
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
  mounted () {
    this.fetchData()
  },
  methods: {
    expandRow: function(item) {
      this.expanded = item === this.expanded[0] ? [] : [item]
    },
    fetchData: _.debounce(function () {
      let self = this
      self.loading = true
      axios.post('/invitations', { options: self.options, forward: self.forward })
        .then(function (response) {

          let msg = _.get(response, 'data.message', false)
          if (msg) {
            self.snackbar.message = msg
            self.snackbar.open = true
          }

          let invitations = _.get(response, 'data.invitations', false)
          if (invitations) {
            self.items = invitations
          }

          let totalItems = _.get(response, 'data.totalItems', false)
          if (totalItems) {
            self.totalItems = totalItems
          }

          let forward = _.get(response, 'data.forward', false)
          if (forward) {
            self.forward = forward
          }
 
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
    }, 500),
    action: function (obj) {
      let self = this
      let data = {}
      let action = _.get(obj, 'action', false)
      if (action == 'acceptWith') {
        let password = _.get(obj, 'password', '')
        data = { password: password }
        action = 'accept'
      }
      let id = _.get(obj, 'item.id', 0)
      axios.put(`/invitation/${action}/${id}`, data)
        .then(function (response) {
          let msg = _.get(response, 'data.message', false)
          if (msg) {
            self.snackbar.message = msg
            self.snackbar.open = true
          }
          let invitation = _.get(response, 'data.invitation', false)
          if (invitation) {
            let index = _.findIndex(self.items, [ 'id', id ])
            if (index >= 0) {
              if (invitation.status === 1) { // recruiting is a status of 1
                self.items.splice(index, 1, invitation)
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
      return _.find(self.users(item), [ 'id', self.cuid ])
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
    headers () {
      return [
        { text: '', sortable: false, value: 'data-table-expand' },
        { text: 'ID', align: 'left', sortable: false, value: 'id' },
        { text: 'Title', sortable: false, value: 'title' },
        { text: 'Creator', sortable: false, value: 'creator' },
        { text: 'Num Players', sortable: false, value: 'numPlayers' },
        { text: 'Players', sortable: false, value: 'players' },
        { text: 'Last Updated', sortable: false, value: 'lastUpdated' },
        { text: 'Public/Private', sortable: false, value: 'public' },
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

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h1, h2, h3 {
  font-weight: normal;
}
</style>
