<template>
 <td :colspan='span'>
 <v-container fluid>
   <v-row no-gutters >
     <v-col cols=2></v-col>
     <v-col cols=2></v-col>
     <v-col cols=2>GLO</v-col>
     <v-col cols=2>Played</v-col>
     <v-col cols=2>Won</v-col>
     <v-col cols=2>Win %</v-col>
   </v-row>
   <v-row no-gutters>
     <v-col cols=2>Invite from:</v-col>
     <v-col cols=2>
        <sn-user-btn :user="creator(item)" size="x-small"></sn-user-btn> {{creator(item).name}}
     </v-col>
     <v-col cols=2>{{gloFor(creator(item).id)}}</v-col>
     <v-col cols=2>0</v-col>
     <v-col cols=2>0</v-col>
     <v-col cols=2>0</v-col>
   </v-row>
   <v-row no-gutters v-if='cu.id != creator(item).id'>
     <v-col cols=2>Your Experience:</v-col>
     <v-col cols=2>
        <sn-user-btn :user="cu" size="x-small"></sn-user-btn> {{cu.name}}
     </v-col>
     <v-col cols=2>{{gloFor(cu.id)}}</v-col>
     <v-col cols=2>0</v-col>
     <v-col cols=2>0</v-col>
     <v-col cols=2>0</v-col>
   </v-row>
   <v-row v-if='!item.public && canAccept(item)'>
     <v-col cols='4'>
       <v-text-field
         v-model='password'
         :append-icon="show ? 'mdi-eye' : 'mdi-eye-off'"
         :rules="[rules.required, rules.min]"
         :type="show ? 'text' : 'password'"
         label='Password'
         placeholder='Enter Password'
         clearable
         autofocus
         dense
         outlined
         rounded
         hint='At least 8 characters'
         counter
         @click:append="show = !show"
       >
       </v-text-field>
     </v-col>
     <v-col cols='2'>
       <v-btn 
         x-small
         rounded
         :disabled='disabled'
         @click.native="$emit('action', { action: 'accept', item: item })"
         color='info'
         dark
       >
         Accept
       </v-btn>
     </v-col>
   </v-row>
   <v-row v-if='item.public && canAccept(item)'>
     <v-btn 
       x-small
       rounded
       width='62'
       @click.native="$emit('action', { action: 'accept', item: item })"
       color='info'
       dark
       >
       Accept
     </v-btn>
   </v-row>
   <v-row>
     <v-btn 
       x-small
       rounded
       width='62'
       v-if="canDrop(item)"
       @click.native="$emit('action', { action: 'drop', item: item })"
       color='info'
       dark
       >
       Drop
     </v-btn>
   </v-row>
 </v-container>
 </td>
</template>

<script>
  import UserButton from '@/components/user/Button'

  const _ = require('lodash')
  const axios = require('axios')

  export default {
    name: 'sn-expanded-row',
    props: [ 'span', 'item' ],
    components: {
      'sn-user-btn': UserButton
    },
    data () {
      return {
        password: '',
        show: false,
        rules: {
          required: value => !!value || 'Required.',
          min: v => _.size(v) >= 8 || 'Min 8 characters'
        },
        id: 0,
        details: {}
      }
    },
    mounted () {
      var self = this
      if (self.item.id != self.id) {
        self.id = self.item.id
        self.fetchDetails()
      }
    },
    updated () {
      var self = this
      if (self.item.id != self.id) {
        self.id = self.item.id
        self.fetchDetails()
      }
    },
    methods: {
      fetchDetails: function () {
        var self = this
          axios.get(`/invitation/details/${self.item.id}`)
          .then(function (response) {
            console.log(`response: ${JSON.stringify(response)}`)
            var details = _.get(response, 'data.details', false)
            if (details) {
              self.details = details
            }
            self.loading = false
          })
          .catch(function () {
            self.loading = false
            self.snackbar.message = 'Server Error.  Try refreshing page.'
            self.snackbar.open = true
        })
      },
      detailsFor: function (id) {
        var self = this
        return _.find(self.details, { 'id': id })
      },
      gloFor: function (id) {
        var self = this
        return _.get(self.detailsFor(id), 'glo', 0)
      },
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
      canAccept: function (item) {
        var self = this
        return !self.joined(item) && item.status === 1 // recruiting is a status 1
      },
      canDrop: function (item) {
        var self = this
        return self.joined(item) && item.status === 1 // recruiting is a status 1
      },
      joined: function (item) {
        return _.includes(item.userIds, this.cu.id)
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
