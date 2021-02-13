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
      <h3>Invitations</h3>
    </v-card-title>
    <v-card-text>
      <v-data-table
        :headers="headers"
        :items="items"
        show-expand
        single-expand
      >
         <template v-slot:item.players="{ item }">
           <span class="py-1" v-for="user in item.users" :key="user.id" >{{user.name}}</span>
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
    </v-card-text>
  </v-card>
      </v-container>
    </v-main>
    <sn-footer app></sn-footer>
  </v-app>
</template>

<script>
  import Expansion from '@/components/invitation/Expansion'
  import Toolbar from '@/components/Toolbar'
  import NavDrawer from '@/components/NavDrawer'
  import Snackbar from '@/components/Snackbar'
  import Footer from '@/components/Footer'

  const _ = require('lodash')
  const axios = require('axios')

  export default {
    name: 'index',
    components: {
      'sn-expanded-row': Expansion,
      'sn-toolbar': Toolbar,
      'sn-nav-drawer': NavDrawer,
      'sn-snackbar': Snackbar,
      'sn-footer': Footer
    },
    data () {
      return {
        password: '',
        show: false,
        headers: [
          { text: '', value: 'data-table-expand' },
          { text: 'ID', align: 'left', sortable: true, value: 'id' },
          { text: 'Title', value: 'title' },
          { text: 'Creator', value: 'creator.name' },
          { text: 'Num Players', value: 'numPlayers' },
          { text: 'Players', value: 'players' },
          { text: 'Last Updated', value: 'lastUpdated' },
          { text: 'Public/Private', value: 'public' }
        ],
        rules: {
          required: value => !!value || 'Required.',
          min: v => _.size(v) >= 8 || 'Min 8 characters'
        },
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
        var self = this
        axios.get('/invitations')
          .then(function (response) {
            var msg = _.get(response, 'data.message', false)
            if (msg) {
              self.snackbar.message = msg
              self.snackbar.open = true
            }
            var invitations = _.get(response, 'data.invitations', false)
            if (invitations) {
              self.items = invitations
            }
            self.loading = false
          })
          .catch(function () {
            self.loading = false
            self.snackbar.message = 'Server Error.  Try refreshing page.'
            self.snackbar.open = true
        })
      },
      action: function (obj) {
        var self = this
        var data = {}
        var action = _.get(obj, 'action', false)
        if (action == 'accept') {
          data = { password: self.password }
        }
        var id = _.get(obj, 'item.id', 0)
        axios.put(`/invitation/${action}/${id}`, data)
          .then(function (response) {
            var msg = _.get(response, 'data.message', false)
            if (msg) {
              self.snackbar.message = msg
              self.snackbar.open = true
            }
            var invitation = _.get(response, 'data.invitation', false)
            if (invitation) {
              var index = _.findIndex(self.items, [ 'id', id ])
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
        var self = this
        var item = self.getItem(id)
        return !self.joined(item) && item.status === 1 // recruiting is a status 1
      },
      canDrop: function (id) {
        var self = this
        var item = self.getItem(id)
        return self.joined(item) && item.status === 1 // recruiting is a status 1
      },
      joined: function (item) {
        var self = this
        return _.find(item.users, [ 'id', self.cu.id ])
      },
      getItem: function (id) {
        var self = this
        return _.find(self.items, [ 'id', id ])
      },
      publicPrivate: function (item) {
        return item.public ? 'Public' : 'Private'
      }
    },
    computed: {
      disabled: function () {
        var self = this
        return _.size(self.password) < 8
      },
      cu: {
        get: function () {
          return this.$root.cu
        },
        set: function (value) {
          this.$root.cu = value
        }
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
