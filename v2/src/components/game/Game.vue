<template>
  <v-app id='app'>
    <div v-visibility-change="visibilityChange"></div>

    <sn-toolbar v-model='nav'>
      <v-row no-gutters>
        <v-col cols='1'>
        </v-col>
        <v-col cols='3'>
          <v-tooltip bottom color='info'>
            <template v-slot:activator='{ on }'>
              <v-btn v-on='on' icon @click.stop='toggleLog'>
                <v-icon>history</v-icon>
              </v-btn>
            </template>
            <span>Game Log</span>
          </v-tooltip>

          <v-tooltip bottom color='info'>
            <template v-slot:activator='{ on }'>
              <v-btn v-on='on' icon @click.stop='toggleChat'>
                <v-icon>chat</v-icon>
              </v-btn>
            </template>
            <span>Chat</span>
          </v-tooltip>

        </v-col>

        <v-col cols='8'>
          <sn-control-bar v-model='game' @action='action($event)' ></sn-control-bar>
        </v-col>

      </v-row>
    </sn-toolbar>

    <sn-nav-drawer v-model='nav' ></sn-nav-drawer>

    <sn-chat-drawer v-model='chatDrawer' :game='game' ></sn-chat-drawer>

    <sn-log-drawer v-model='logDrawer' :game='game' ></sn-log-drawer>

    <sn-card-bar
      v-if='selectedPlayer'
      :player='selectedPlayer'
      :game='game'
      v-model='cardbar'
      @selected-card='selected($event)'
      >
    </sn-card-bar>


      <sn-snackbar v-model='snackbar.open'>
        <div class='text-center'>
          <span v-html='snackbar.message'></span>
        </div>
      </sn-snackbar>

      <v-main>
        <v-container fluid>
          <v-row>
            <v-col cols='4'>
              <v-row>
                <v-col>
                  <sn-status-panel :game='game' :live.sync='live'></sn-status-panel>
                </v-col>
              </v-row>

              <v-row>
                <v-col>
                  <sn-player-panels
                    v-model='tab'
                    @show='cardbar = $event'
                    @pass="action({action: 'pass', data: { undo: game.undo }})"
                    :game='game'
                    >
                  </sn-player-panels>
                </v-col>
              </v-row>

            </v-col>

            <v-col cols='8'>

              <v-row>
                <v-col>
                  <sn-messagebar>{{message}}</sn-messagebar>
                </v-col>
              </v-row>

              <v-row>
                <v-col>
                  <sn-board id='board' :game='game' @selected='selected($event)' ></sn-board>
                </v-col>
              </v-row>

            </v-col>

          </v-row>

        </v-container>
      </v-main>

      <sn-footer></sn-footer>

  </v-app>
</template>

<script>
import Controlbar from '@/components/game/Controlbar'
import Toolbar from '@/components/Toolbar'
import Snackbar from '@/components/Snackbar'
import Footer from '@/components/Footer'
import NavDrawer from '@/components/NavDrawer'
import ChatDrawer from '@/components/chat/Drawer'
import LogDrawer from '@/components/log/Drawer'
import Board from '@/components/board/Board'
import Bar from '@/components/card/Bar'
import StatusPanel from '@/components/game/StatusPanel'
import Panels from '@/components/player/Panels'
import Messagebar from '@/components/game/Messagebar'
import CurrentUser from '@/components/mixins/CurrentUser'
import Messaging from '@/components/mixins/Messaging'
import Player from '@/components/mixins/Player'

const _ = require('lodash')
const axios = require('axios')

export default {
  mixins: [ CurrentUser, Player, Messaging ],
  name: 'game',
  data () {
    return {
      stop: false,
      subscribed: [],
      loaded: false,
      game: {
        title: '',
        id: 0,
        turn: 0,
        phase: 'None',
        colorMaps: [],
        options: {},
        glog: [],
        jewels: {}
      },
      tab: null,
      path: '/game',
      cardbar: false,
      nav: false,
      history: false,
      chat: false,
      chatDrawer: false,
      logDrawer: false
    }
  },
  components: {
    'sn-control-bar': Controlbar,
    'sn-toolbar': Toolbar,
    'sn-snackbar': Snackbar,
    'sn-nav-drawer': NavDrawer,
    'sn-chat-drawer': ChatDrawer,
    'sn-log-drawer': LogDrawer,
    'sn-board': Board,
    'sn-card-bar': Bar,
    'sn-status-panel': StatusPanel,
    'sn-player-panels': Panels,
    'sn-messagebar': Messagebar,
    'sn-footer': Footer
  },
  created () {
    var self = this
    self.fetchData()
    self.getToken()
    self.fbMsg().onMessage((payload) => {
      self.action(payload.data)
    })
  },
  methods: {
    visibilityChange(evt, hidden) {
      let self = this
      if(!hidden) {
        self.action({ action: 'refresh' })
      }
    },
    subscribe: function () {
      let self = this
      let obj = { token: self.token }
      axios.put(`${self.path}/subscribe/${self.$route.params.id}`, obj )
        .then(function (response) {
          let msg = _.get(response, 'data.message', false)
          let subscribed = _.get(response, 'data.subscribed', false)
          self.$nextTick()
            .then(function () {
              if (subscribed) {
                self.subscribed = subscribed
              }
              if (msg) {
                self.snackbar.message = msg
                self.snackbar.open = true
              }
            })
        })
        .catch(function (response) {
          let subscribed = self.subscribed
          if (_.size(self.subscribed) > 1) {
            self.subscribed.push(self.token)
          } else {
            self.subscribed = [ self.token ]
          }
          let msg = _.get(response, 'message', 'Server Error.  Try refreshing page.')
          self.$nextTick()
            .then(function () {
              self.subscribed = subscribed
              self.snackbar.message = msg
              self.snackbar.open = true
            })
        })
    },
    unsubscribe: function () {
      let self = this
      let obj = { token: self.token }
      axios.put(`${self.path}/unsubscribe/${self.$route.params.id}`, obj )
        .then(function (response) {
          let msg = _.get(response, 'data.message', false)
          self.$nextTick()
            .then(function () {
              if (msg) {
                self.snackbar.message = msg
                self.snackbar.open = true
              }
            })
        })
        .catch(function () {
          let subscribed = self.subscribed
          self.subscribed = []
          self.$nextTick()
            .then(function () {
              self.subscribed = subscribed
              self.snackbar.message = 'Server Error.  Try refreshing page.'
              self.snackbar.open = true
            })
        })
    },
    toggleChat: function () {
      let self = this
      self.chatDrawer = !self.chatDrawer
      if (self.chatDrawer) {
        self.logDrawer = false
      }
    },
    toggleLog: function () {
      let self = this
      self.logDrawer = !self.logDrawer
      if (self.logDrawer) {
        self.chatDrawer = false
      }
    },
    myUpdate: function (data) {
      let self = this

      if (_.has(data, 'game')) {
        self.game = data.game
        document.title = data.game.title + ' #' + data.game.id
      }

      if (_.has(data, 'message') && (data.message != '')) {
        self.snackbar.message = data.message
        self.snackbar.open = true
      }

      if (_.has(data, 'subscribed')) {
        self.subscribed = data.subscribed
      }

      if (_.has(data, 'token')) {
        self.token = data.token
      }

      if (_.has(data, 'cu')) {
        self.cu = data.cu
        self.cuLoading = false

        if (self.cu) {
          let p = self.playerByUID(data.cu.id)
          if (p) {
            self.tab = self.indexOf(p)
          }
        }
      }
    },
    fetchData: function () {
      let self = this
      axios.get(`${self.path}/show/${self.$route.params.id}`)
        .then(function (response) {
          var data = _.get(response, 'data', false)
          if (data) {
            self.myUpdate(data)
          }
          self.loaded = true
        })
        .catch(function () {
          self.snackbar.message = 'Server Error.  Try refreshing page.'
          self.snackbar.open = true
        })
    },
    selected: function (data) {
      var self = this
      switch (self.game.phase) {
        case 'Place Thieves':
          self.action({
            action: 'place-thief',
            data: {
              areaID: data.areaID,
              undo: self.game.undo
            }
          })
          break
        case 'Play Card':
          self.cardbar = false
          self.action({
            action: 'play-card',
            data: {
              kind: data,
              undo: self.game.undo
            }
          })
          break
        case 'Select Thief':
          self.action({
            action: 'select-thief',
            data: {
              areaID: data.areaID,
              undo: self.game.undo
            }
          })
          break
        case 'Move Thief':
          self.action({
            action: 'move-thief',
            data: {
              areaID: data.areaID,
              undo: self.game.undo
            }
          })
          break
      }
    },
    action: function (data) {
      var self = this
      console.log(`action data: ${JSON.stringify(data)}`)
      var action = data.action
      if (action == 'refresh') {
        console.log(`refresh fetchData`)
        self.fetchData()
        return
      }
      data.data.token = self.token
      axios.put(`${self.path}/${action}/${self.$route.params.id}`, data.data)
        .then(function (response) {
          if (_.has(response, 'data')) {
            self.myUpdate(response.data)
          }
        })
        .catch(function () {
          self.snackbar.message = 'Server Error.  Try refreshing page.'
          self.snackbar.open = true
        })
    },
  },
  computed: {
    live: {
      get: function () {
        let self = this
        return _.includes(self.subscribed, self.token)
      },
      set: function (value) {
        let self = this
        if (value) {
          self.subscribe()
        } else {
          self.unsubscribe()
        }
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
    selectedPlayer: function () {
      var self = this
      return self.playerByPID(self.cardbar)
    },
    waitMessage: function () {
      var self = this
      var name = _.get(self.cp, 'user.name', 'current player')
      return `Please wait for ${name} to take a turn.`
    },
    message: function () {
      var self = this
      if (self.game.status == 2) {
        return 'Game Over'
      }
      switch (self.game.phase) {
        case 'None':
          return self.waitMessage
        case 'Place Thieves':
          if (!self.isCP) {
            return self.waitMessage
          }

          if (self.cp.performedAction) {
            return 'Finish turn by selecting above check mark.'
          } else {
            return 'Select empty space in grid to place thief.'
          }
        case 'Play Card':
          if (!self.isCP) {
            return self.waitMessage
          }

          if (!self.cardbar) {
            return 'Select hand or pass'
          } else {
            return 'Select card from hand'
          }
        case 'Select Thief':
          if (!self.isCP) {
            return self.waitMessage
          }

          return 'Select thief in grid'
        case 'Move Thief':
          if (!self.isCP) {
            return self.waitMessage
          }

          if (self.cp.performedAction) {
            return 'Finish turn by selecting above check mark.'
          }

          return 'Select highlighted spot in grid to move thief'
        case 'Passed':
          return 'Finish turn by selecting above check mark.'
      }
      return ''
    }
  }
}
</script>
