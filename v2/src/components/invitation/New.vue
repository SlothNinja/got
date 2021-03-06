<template>
  <v-app id='app'>
    <sn-toolbar v-model="nav"></sn-toolbar>
    <sn-nav-drawer v-model="nav" app></sn-nav-drawer>
    <sn-snackbar v-model="snackbar.open">
      <div class="text-xs-center">
        {{snackbar.message}}
      </div>
    </sn-snackbar>
    <v-main>  
      <v-container grid-list-md >
        <v-layout row wrap>
          <v-flex xs6>
            <v-card height="31em">
              <v-card-title primary-title>
                <h3>New Invitation</h3>
              </v-card-title>
              <v-card-text>
                <v-form action="/got/invitation" method="post">
                  <v-text-field
                    label="Title"
                    v-model="invitation.title"
                    >
                  </v-text-field>
                    <v-select
                      label="Number Players"
                      :items="npItems"
                      v-model="invitation.numPlayers"
                      >
                    </v-select> 
                      <v-select 
                      label="Two Thief Variant"
                      :items="optItems"
                      v-model="invitation.twoThief"
                      >
                      </v-select> 
                        <v-text-field
                          label='Password'
                          v-model='invitation.password'
                          :append-icon="show ? 'mdi-eye' : 'mdi-eye-off'"
                          :type="show ? 'text' : 'password'"
                          placeholder="Enter Password for Private Game"
                          clearable
                          @click:append="show = !show"
                          >
                        </v-text-field>
                          <v-btn color='green' dark @click="putData">Submit</v-btn>
                </v-form>
              </v-card-text>
            </v-card>
          </v-flex>
          <v-flex xs6>
            <v-card height="31em">
              <v-img height='200px' :src="boxPath()" />
                <v-card-text>
                  <v-layout row>
                    <v-flex xs5>Designer</v-flex>
                    <v-flex>Adam E. Daulton</v-flex>
                  </v-layout>
                  <v-layout row>
                    <v-flex xs5>Artists</v-flex> 
                    <v-flex>Jeremy Montz</v-flex> 
                  </v-layout>
                  <v-layout row> 
                    <v-flex xs5>Publisher</v-flex> 
                    <v-flex><a href="http://www.thegamecrafter.com/">The Game Crafter, LLC</a></v-flex>
                  </v-layout>
                  <v-layout row>
                    <v-flex xs5>Year Published</v-flex>
                    <v-flex>2012</v-flex>
                  </v-layout>
                  <v-layout row> 
                    <v-flex xs5>On-Line Developer</v-flex> 
                    <v-flex>Jeff Huter</v-flex> 
                  </v-layout> 
                  <v-layout row> 
                    <v-flex xs5>Permission Provided By</v-flex> 
                    <v-flex>Adam E Daulton</v-flex> 
                  </v-layout> 
                  <v-layout row> 
                    <v-flex xs5>Rules (pdf)</v-flex> 
                    <v-flex><a href="/static/rules/got.pdf">Guild Of Thieves (English)</a></v-flex> 
                  </v-layout> 
                </v-card-text>
            </v-card>
          </v-flex>
        </v-layout>
      </v-container>
    </v-main>
    <sn-footer></sn-footer>
  </v-app>
</template>

<script>

import Toolbar from '@/components/lib/Toolbar'
import NavDrawer from '@/components/lib/NavDrawer'
import Snackbar from '@/components/lib/Snackbar'
import Footer from '@/components/lib/Footer'

import CurrentUser from '@/components/lib/mixins/CurrentUser'
import BoxPath from '@/components/mixins/BoxPath'

const _ = require('lodash')
const axios = require('axios')

export default {
  name: 'newInvitation',
  mixins: [ CurrentUser, BoxPath ],
  data () {
    return {
      show: false,
      invitation: {
        title: '',
        id: 0,
        turn: 0,
        phase: 0,
        colorMaps: [],
        options: {},
        glog: [],
        jewels: {}
      },
      path: '/invitation/new',
      nav: false,
      npItems: [
        { text: '2', value: 2 },
        { text: '3', value: 3 },
        { text: '4', value: 4 }
      ],
      optItems: [
        { text: 'Yes', value: true },
        { text: 'No', value: false }
      ]
    }
  },
  components: {
    'sn-toolbar': Toolbar,
    'sn-snackbar': Snackbar,
    'sn-nav-drawer': NavDrawer,
    'sn-footer': Footer
  },
  created () {
    let self = this
    self.fetchData()
  },
  watch: {
    '$route': 'fetchData'
  },
  computed: {
    snackbar: {
      get: function () {
        return this.$root.snackbar
      },
      set: function (value) {
        this.$root.snackbar = value
      }
    },
  },
  methods: {
    updateInvitation (data) {
      let self = this

      let msg = _.get(data, 'data.message', false)
      if (msg) {
        self.snackbar.message = msg
        self.snackbar.open = true
      }
      let invitation = _.get(data, 'data.invitation', false)
      if (invitation) {
        self.invitation = invitation
      }
      let cu = _.get(data, 'data.cu', false)
      if (cu) {
        self.cu = cu
        self.cuLoading = false
      }
    },
    fetchData () {
      let self = this
      axios.get(self.path)
        .then(function (response) {
          self.updateInvitation(response)
          self.loading = false
        })
        .catch(function () {
          self.loading = false
          self.snackbar.message = 'Server Error.  Try refreshing page.'
          self.snackbar.open = true
        })
    },
    putData () {
      let self = this
      self.loading = true
      axios.put(self.path, self.invitation)
        .then(function (response) {
          self.updateInvitation(response)
          self.loading = false
        })
        .catch(function () {
          self.loading = false
          self.snackbar.message = 'Server Error.  Try again.'
          self.snackbar.open = true
          self.$router.push({ name: 'home'})
        })
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h1, h2, h3 {
  font-weight: normal;
}
</style>
