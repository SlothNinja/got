<template>
  <v-app id='app'>

    <sn-toolbar v-model='nav'>
      <v-row>

        <v-col cols='3'>

          <v-tooltip bottom color='info'>
            <template v-slot:activator='{ on }'>
              <v-btn v-on='on' icon @click.stop='history = !history'>
                <v-icon>history</v-icon>
              </v-btn>
            </template>
            <span>Game Log</span>
          </v-tooltip>

          <v-tooltip bottom color='info'>
            <template v-slot:activator='{ on }'>
              <v-btn v-on='on' icon @click.stop='chat = !chat'>
                <v-icon>chat</v-icon>
              </v-btn>
            </template>
            <span>Chat</span>
          </v-tooltip>
        </v-col>

        <v-col cols='7' class='d-flex justify-center'>
          <sn-control-bar
            v-model='game'
            @action='action($event)'
          >
          </sn-control-bar>
        </v-col>

      </v-row>
    </sn-toolbar>

    <sn-nav-drawer
      v-model='nav'
      app
    >
    </sn-nav-drawer>

    <sn-rdrawer
      v-model='history'
      app
    >
      <sn-game-log
        @message='sbMessage = $event; sbOpen = true'
        v-if='history'
        :stack='stack'
      >
      </sn-game-log>
    </sn-rdrawer>

    <sn-rdrawer
      v-model='chat'
      app
    >
      <sn-chat-box
        @message='sbMessage = $event; sbOpen = true'
        v-if='chat'
        :user='cu'
      >
      </sn-chat-box>
    </sn-rdrawer>

    <sn-snackbar v-model='sbOpen'>
      <div class='text-center'>
        {{sbMessage}}
      </div>
    </sn-snackbar>

    <v-content>
      <v-container fluid style='overflow:auto'>
        <v-card
          min-width='1185'
          min-height='740'
          flat
          class='theme--light v-application'
        >
          <v-row>
            <v-col cols='3'>
              <v-card
                min-width='272'
                flat
                height='740'
                class='theme--light v-application d-flex flex-column justify-space-between'
              >

                <sn-status-panel
                  :game='game'
                >
                </sn-status-panel>

                <sn-player-panels
                  v-model='tab'
                  @show='cardbar = $event'
                  @pass="action({action: 'pass', data: {} })"
                  :game='game'
                >
                </sn-player-panels>

              </v-card>
            </v-col>

            <v-col cols='9'>

              <v-card
                flat
                height='740'
                class='theme--light v-application d-flex flex-column justify-space-between'
              >

                <sn-messagebar>
                  {{message}}
                </sn-messagebar>

                <sn-board
                  id='board'
                  :game='game'
                  @selected='selected($event)'
                >
                </sn-board>

              </v-card>
            </v-col>

          </v-row>
        </v-card>

        <sn-card-bar
          v-if='selectedPlayer'
          :player='selectedPlayer'
          :game='game'
          v-model='cardbar'
          @selected-card='selectedCard($event)'
        >
        </sn-card-bar>

      </v-container>
    </v-content>

    <sn-footer
      app
    >
    </sn-footer>

  </v-app>
</template>

<script>
  import Controlbar from '@/components/game/Controlbar'
  import Toolbar from '@/components/Toolbar'
  import Snackbar from '@/components/Snackbar'
  import Footer from '@/components/Footer'
  import NavDrawer from '@/components/NavDrawer'
  import RDrawer from '@/components/rdrawer/Drawer'
  import Board from '@/components/board/Board'
  import Bar from '@/components/card/Bar'
  import StatusPanel from '@/components/game/StatusPanel'
  import Panels from '@/components/player/Panels'
  import Messagebar from '@/components/game/Messagebar'
  import ChatBox from '@/components/chat/Box'
  import GameLog from '@/components/log/Box'
  import CurrentUser from '@/components/mixins/CurrentUser'
  import Player from '@/components/mixins/Player'

  const _ = require('lodash')
  const axios = require('axios')
  const phaseNone = 0
  const phasePlaceThieves = 1
  const phasePlayCard = 2
  const phaseSelectThief = 3
  const phaseMoveThief = 4
  const phaseGameOver = 5

  export default {
    mixins: [ CurrentUser, Player ],
    name: 'game',
    data () {
      return {
        game: {
          header: { title: '', id: 0, turn: 0, phase: 0, colorMaps: [], options: {} },
          state: { glog: [], jewels: {} }
        },
        tab: 'player-1',
        path: '/game',
        cardbar: false,
        nav: false,
        history: false,
        chat: false,
        sbOpen: false,
        sbMessage: ''
      }
    },
    components: {
      'sn-control-bar': Controlbar,
      'sn-toolbar': Toolbar,
      'sn-snackbar': Snackbar,
      'sn-nav-drawer': NavDrawer,
      'sn-rdrawer': RDrawer,
      'sn-board': Board,
      'sn-card-bar': Bar,
      'sn-status-panel': StatusPanel,
      'sn-player-panels': Panels,
      'sn-chat-box': ChatBox,
      'sn-game-log': GameLog,
      'sn-messagebar': Messagebar,
      'sn-footer': Footer
    },
    created () {
      var self = this
      self.fetchData()
    },
    methods: {
      myUpdate: function (data) {
        var self = this

        if (_.has(data, 'game')) {
          self.game = data.game
          document.title = data.game.header.title + ' #' + data.game.id
        }

        if (_.has(data, 'message') && (data.message != '')) {
          self.sbMessage = data.message
          self.sbOpen = true
        }

        if (_.has(data, 'cu')) {
          self.cu = data.cu
        }

        var cuid = _.get(self.cu, 'id', false)
        if (cuid) {
          self.tab = `player-${self.pidByUID(cuid)}`
        } 
      },
      fetchData: function () {
        var self = this
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
        switch (self.game.header.phase) {
          case phasePlaceThieves:
            self.action({
              action: 'place-thief',
              data: {
                row: data.row,
                column: data.column
              }
            })
            break
          case phasePlayCard:
            self.cardbar = false
            self.action({
              action: 'play-card',
              data: {
                kind: data
              }
            })
            break
          case phaseSelectThief:
            self.action({
              action: 'select-thief',
              data: {
                row: data.row,
                column: data.column
              }
            })
            break
          case phaseMoveThief:
            self.action({
              action: 'move-thief',
              data: {
                row: data.row,
                column: data.column
              }
            })
            break
        }
      },
      selectedCard: function (data) {
        var self = this
        self.cardbar = false
        self.action({
          action: 'play-card',
          data: {
            kind: data
          }
        })
      },
      action: function (data) {
        var self = this
        var action = data.action
        if (action == 'refresh') {
          self.fetchData()
          return
        }
        self.tab = `player-${self.pidByUID(self.cu.id)}`
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
      stack: function () {
        var self = this
        return self.game.undoStack
      },
      waitMessage: function () {
        var self = this
        var name = _.get(self.cp, 'user.name', 'current player')
        return `Please wait for ${name} to take a turn.`
      },
      message: function () {
        var self = this
        switch (self.game.header.phase) {
          case phaseNone:
            return self.waitMessage
          case phasePlaceThieves:
            if (!self.isCP) {
              return self.waitMessage
            }

            if (self.cp.performedAction) {
              return 'Finish turn by selecting above check mark.'
            } else {
              return 'Select empty space in grid to place thief.'
            }
          case phasePlayCard:
            if (!self.isCP) {
              return self.waitMessage
            }

            if (!self.cardbar) {
              return 'Select hand or pass'
            } else {
              return 'Select card from hand'
            }
          case phaseSelectThief:
            if (!self.isCP) {
              return self.waitMessage
            }

            return 'Select thief in grid'
          case phaseMoveThief:
            if (!self.isCP) {
              return self.waitMessage
            }

            if (self.cp.performedAction) {
              return 'Finish turn by selecting above check mark.'
            }

            return 'Select highlighted spot in grid to move thief'
          case phaseGameOver: 
            return 'Game Over'
        }
        return ''
      }
    }
  }
</script>