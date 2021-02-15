<template>
  <v-app id='app'>

    <sn-toolbar v-model='nav'>
      <v-row>

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

        <v-col cols='7' class='d-flex justify-center'>
          <sn-control-bar v-model='game' @action='action($event)' ></sn-control-bar>
        </v-col>

      </v-row>
    </sn-toolbar>

    <sn-nav-drawer v-model='nav' ></sn-nav-drawer>

    <sn-chat-drawer v-model='chatDrawer' :game='game' ></sn-chat-drawer>

    <sn-log-drawer v-model='logDrawer' :game='game' ></sn-log-drawer>

    <sn-snackbar v-model='sbOpen'>
      <div class='text-center'>{{sbMessage}}</div>
    </sn-snackbar>

    <v-main>
      <v-container fluid>
        <v-row>
          <v-col cols='4'>
            <v-row>
              <v-col>
                <sn-status-panel :game='game'></sn-status-panel>
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

        <sn-card-bar
          v-if='selectedPlayer'
          :player='selectedPlayer'
          :game='game'
          v-model='cardbar'
          @selected-card='selected($event)'
          >
        </sn-card-bar>

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
// import ChatBox from '@/components/chat/Box'
// import GameLog from '@/components/log/Box'
import CurrentUser from '@/components/mixins/CurrentUser'
import Player from '@/components/mixins/Player'

const _ = require('lodash')
const axios = require('axios')

export default {
  mixins: [ CurrentUser, Player ],
  name: 'game',
  data () {
    return {
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
      sbOpen: false,
      sbMessage: '',
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
    // 'sn-chat-box': ChatBox,
    // 'sn-game-log': GameLog,
    'sn-messagebar': Messagebar,
    'sn-footer': Footer
  },
  created () {
    var self = this
    self.fetchData()
  },
  methods: {
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
        self.sbMessage = data.message
        self.sbOpen = true
      }

      if (_.has(data, 'cu')) {
        self.cu = data.cu
        self.cuLoading = false

        let p = self.playerByUID(data.cu.id)
        if (p) {
          self.tab = self.indexOf(p)
        }
      }
    },
    fetchData: function () {
      let self = this
      self.loading = true
      axios.get(`${self.path}/show/${self.$route.params.id}`)
        .then(function (response) {
          var data = _.get(response, 'data', false)
          if (data) {
            self.myUpdate(data)
          }
          self.loading = false
        })
        .catch(function () {
          self.loading = false
          self.sbMessage = 'Server Error.  Try refreshing page.'
          self.sbOpen = true
        })
    },
    selected: function (data) {
      var self = this
      console.log(`selected data: ${JSON.stringify(data)}`)
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
        self.fetchData()
        return
      }
      axios.put(`${self.path}/${action}/${self.$route.params.id}`, data.data)
        .then(function (response) {
          self.loading = false

          if (_.has(response, 'data')) {
            self.myUpdate(response.data)
          }
        })
        .catch(function () {
          self.loading = false
          self.sbMessage = 'Server Error.  Try refreshing page.'
          self.sbOpen = true
        })
    },
    animateMoveTo: function (obj, to, complete) {
      var height = obj.height()
      var width = obj.width()
      var from = obj.offset()
      var midpoint = {
        top: (from.top + to.top) / 2,
        left: (from.left + to.left) / 2
      }
      obj.velocity({
        top: midpoint.top,
        left: midpoint.left,
        height: height * 2,
        width: width * 2
      }, { duration: 200 })
        .velocity({
          top: to.top,
          left: to.left,
          height: height,
          width: width
        }, {
          duration: 200,
          complete: function () {
            if (complete) {
              complete()
            }
          }
        })
    }
  },

  computed: {
    animate: {
      get: function () {
        return this.$root.animate
      },
      set: function (value) {
        this.$root.animate = value
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
