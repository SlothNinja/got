<template>
  <td :colspan='span'>
    <v-container fluid>
      <v-card>
        <v-card-text>
          <v-simple-table dense>
            <thead>
              <tr>
                <th class='text-left'>
                </th>
                <th class='text-left'>
                  Player
                </th>
                <th class='text-center'>
                  GLO
                </th>
                <th class='text-center'>
                  Projected
                </th>
                <th class='text-center'>
                  Played
                </th>
                <th class='text-center'>
                  Won
                </th>
                <th class='text-center'>
                  Win%
                </th>
              </tr>
            </thead>
            <tbody>

              <sn-expanded-table-row
                :details='details'
                :user='creator(item)'
                >
                Invite from:
              </sn-expanded-table-row>

              <template
                v-for="(uid, index) in item.userIds"
                >
                  <sn-expanded-table-row
                    v-if="(uid != item.creatorId) && (uid != cu)"
                    :details='details'
                    :user='user(item, index)'
                    :key="uid"
                    >
                  </sn-expanded-table-row>
                  <sn-expanded-table-row
                    v-if="(uid != item.creatorId) && (uid == cu)"
                    :details='details'
                    :user='user(item, index)'
                    :key="uid"
                    >
                    Your Experience:
                  </sn-expanded-table-row>
              </template>
            </tbody>
          </v-simple-table>
        </v-card-text>
        <v-divider></v-divider>
        <v-card-actions>
          <v-container>
            <v-row>
              <v-col cols='4' v-if='!item.public && canAccept(item)'>
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
              <v-col cols='2' v-if='!item.public && canAccept(item)'>
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
              <v-col cols='4' v-if='item.public && canAccept(item)'>
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
              </v-col>
              <v-col cols='4' v-if='item.public && canDrop(item)'>
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
              </v-col>
            </v-row>
          </v-container>
        </v-card-actions>
      </v-card>
    </v-container>
  </td>
</template>

<script>
import CurrentUser from '@/components/mixins/CurrentUser'
import ExpansionRow from '@/components/invitation/ExpansionRow'

const _ = require('lodash')
const axios = require('axios')

export default {
  name: 'sn-expanded-row',
  mixins: [ CurrentUser ],
  props: [ 'span', 'item' ],
  components: {
    'sn-expanded-table-row' : ExpansionRow
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
    let self = this
    if (self.item.id != self.id) {
      self.id = self.item.id
      self.fetchDetails()
    }
  },
  updated () {
    let self = this
    if (self.item.id != self.id) {
      self.id = self.item.id
      self.fetchDetails()
    }
  },
  methods: {
    fetchDetails: function () {
      let self = this
      axios.get(`/invitation/details/${self.item.id}`)
        .then(function (response) {
          let details = _.get(response, 'data.details', false)
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
    detailsFor: function (id, stat) {
      let found = _.find(this.details, { 'id': id })
      if (found) {
        return _.get(found, stat, 0)
      }
      return 0
    },
    fetchData: function () {
      let self = this
      axios.get('/invitations')
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
          self.loading = false
        })
        .catch(function () {
          self.loading = false
          self.snackbar.message = 'Server Error.  Try refreshing page.'
          self.snackbar.open = true
        })
    },
    canAccept: function (item) {
      let self = this
      return !self.joined(item) && item.status === 1 // recruiting is a status 1
    },
    canDrop: function (item) {
      let self = this
      return self.joined(item) && item.status === 1 // recruiting is a status 1
    },
    joined: function (item) {
      return _.includes(item.userIds, this.cuid)
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
    user: function (item, index) {
      return {
        id: item.userIds[index],
        name: item.userNames[index],
        emailHash: item.userEmailHashes[index],
        gravType: item.userGravTypes[index]
      }
    }
  },
  computed: {
    disabled: function () {
      return _.size(this.password) < 8
    },
    snackbar: {
      get: function () {
        return this.$root.snackbar
      },
      set: function (value) {
        this.$root.snackbar = value
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
