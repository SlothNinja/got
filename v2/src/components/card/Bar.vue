<template>
  <v-bottom-sheet
    v-model='cardbar'
    inset
    hide-overlay
  >
    <v-card class='d-flex'>
      <v-container>
        <v-row>
          <v-col cols='2'>
            <v-card
              class='pa-4 text-center d-flex flex-column justify-space-between'
              height='180'
            >

              <div class='font-weight-black'>
                Hand Of:
              </div>

              <div>
                <sn-user-btn
                  :id='btnID'
                  :user='player.user'
                  size='large'
                  :color='color'
                >
                </sn-user-btn>

              </div>

              <div>
                {{player.user.name}}
              </div>

            </v-card>

          </v-col>

          <v-col cols='10'>
            <v-card
              class='d-flex justify-center pa-4'
              height='180'
            >
              <div  
                class='mr-2'
                v-for='(count, kind) in cards'
                :key='kind'
                @click='selected(kind)'
                    
              >
                <sn-card-with-count
                  :id='cardID(kind)'
                  :kind='kind'
                  :count='count'
                >
                </sn-card-with-count>
              </div>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-bottom-sheet>
</template>

<script>
  import WithCount from '@/components/card/WithCount'
  import Button from '@/components/user/Button'
  import Player from '@/components/mixins/Player'
  import Color from '@/components/mixins/Color'
  import CurrentUser from '@/components/mixins/CurrentUser'

  var _ = require('lodash')

  export default {
    mixins: [ CurrentUser, Player, Color ],
    name: 'sn-card-bar',
    components: {
      'sn-user-btn': Button,
      'sn-card-with-count': WithCount
    },
    props: [ 'value', 'player', 'game' ],
    methods: {
      cardID: function (kind) {
        return `hand-${kind}`
      },
      selected: function (kind) {
        var self = this
        if (self.canClick) {
          self.$emit('selected-card', kind)
        }
      }
    },
    computed: {
      color: function () {
        var self = this
        return self.colorByPID(self.player.id)
      },
      btnID: function () {
        var self = this
        return `cardbar-button-${self.player.user.id}`
      },
      cards: function () {
        var self = this
        return _.countBy(self.player.hand, function (card) {
          if (card.facing) {
            return card.kind
          }
          return 'card-back'
        })
      },
      canClick: function () {
        var self = this
        const playCardPhase = 2
        return (self.game.header.phase === playCardPhase) && (self.isPlayerFor(self.player, self.cu))
      },
      cardbar: {
        get: function () {
          var self = this
          return self.value
        },
        set: function (value) {
          var self = this
          self.$emit('input', value)
        }
      }
    }
  }
</script>
