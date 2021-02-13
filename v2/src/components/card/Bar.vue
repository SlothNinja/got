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
                  :user='userFor(player)'
                  size='large'
                  :color='color'
                >
                </sn-user-btn>

              </div>

              <div>
                {{userFor(player).name}}
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

  const _ = require('lodash')

  export default {
    mixins: [ CurrentUser, Player, Color ],
    name: 'sn-card-bar',
    components: {
      'sn-user-btn': Button,
      'sn-card-with-count': WithCount
    },
    props: [ 'value', 'player', 'game' ],
    methods: {
      selected: function (kind) {
        let self = this
        if (self.canClick) {
          self.$emit('selected-card', kind)
        }
      }
    },
    computed: {
      color: function () {
        let self = this
        return self.colorByPID(self.player.id)
      },
      btnID: function () {
        let self = this
        return `cardbar-button-${self.userFor(self.player).id}`
      },
      cards: function () {
        let self = this
        return _.countBy(self.player.hand, function (card) {
          if (card.faceUp) {
            return card.kind
          }
          return 'card-back'
        })
      },
      canClick: function () {
        let self = this
        const playCardPhase = 'Play Card'
        return (self.game.phase === playCardPhase) && (self.isPlayerFor(self.player, self.cu))
      },
      cardbar: {
        get: function () {
          let self = this
          return self.value
        },
        set: function (value) {
          let self = this
          self.$emit('input', value)
        }
      }
    }
  }
</script>
