<template>
  <v-app id='app'>
    <sn-toolbar v-model='nav'></sn-toolbar>
    <sn-nav-drawer v-model='nav' app></sn-nav-drawer>
    <sn-snackbar v-model='snackbar.open'>
      <div class='text-center'>
        <span v-html='snackbar.message'></span>
      </div>
    </sn-snackbar>
    <v-content>
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
           <sn-user-btn :user="item.creator" size="x-small"></sn-user-btn>&nbsp;{{item.creator.name}}
         </template>
         <template v-slot:item.players="{ item }">
           <div class="py-1" v-for="user in item.users" :key="user.id" >
             <sn-user-btn :user="user" size="x-small"></sn-user-btn>&nbsp;{{user.name}}
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
    </v-content>
    <sn-footer app></sn-footer>
  </v-app>
</template>

<script>
  import UserButton from '@/components/user/Button'
  import Toolbar from '@/components/Toolbar'
  import NavDrawer from '@/components/NavDrawer'
  import Snackbar from '@/components/Snackbar'
  import Footer from '@/components/Footer'

  const _ = require('lodash')
  const axios = require('axios')

  export default {
    name: 'index',
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
        var self = this
        axios.get(`/games/${self.$route.params.status}`)
          .then(function (response) {
            console.log(`response: ${JSON.stringify(response)}`)
            var msg = _.get(response, 'data.message', false)
            if (msg) {
              self.snackbar.message = msg
              self.snackbar.open = true
            }
            var gheaders = _.get(response, 'data.gheaders', false)
            if (gheaders) {
              self.items = gheaders
            }
            self.loading = false
          })
          .catch(function () {
            self.loading = false
            self.snackbar.message = 'Server Error.  Try refreshing page.'
            self.snackbar.open = true
        })
      },
      action: function (action, id) {
        var self = this
        axios.put(`/game/${action}/${id}`)
          .then(function (response) {
            var msg = _.get(response, 'data.message', false)
            if (msg) {
              self.snackbar.message = msg
              self.snackbar.open = true
            }
            var header = _.get(response, 'data.header', false)
            if (header) {
              var index = _.findIndex(self.items, [ 'id', id ])
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
      cu: {
        get: function () {
          return this.$root.cu
        },
        set: function (value) {
          this.$root.cu = value
        }
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
