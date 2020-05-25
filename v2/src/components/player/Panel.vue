<template>
  <v-card>
    <div class='mx-4'>
      <v-row>
        <v-col>
          <sn-user-btn 
            :id="`panel-button-${player.user.id}`"
            :user='player.user'
            :color='color'
            size='large'
          >
          </sn-user-btn>
        </v-col>
        <v-col>
          <v-icon :color="iconColor">{{icon}}</v-icon>
          {{player.user.name}}
          <div><strong>Score:</strong> {{player.score}}</div>
        </v-col>
      </v-row>
      <v-divider></v-divider>
      <v-row>
        <v-col>
          <sn-deck :id='`hand-${player.id}`' label='Hand' :deck='player.hand' :show='false'></sn-deck>
        </v-col>
        <v-col>
          <sn-deck :id='`draw-${player.id}`' label='Draw' :deck='player.drawPile' :show='false'></sn-deck>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <v-btn
            small
            class='mb-3'
            width='64'
            :disabled='!canShow'
            color='info'
            dark
            @click.stop="$emit('show')"
          >
            Hand
          </v-btn>
          <v-btn
            small
            width='64'
            :disabled='!canPass'
            color='info'
            dark
            @click.stop="$emit('pass')"
          >
            Pass
          </v-btn>
        </v-col>
        <v-col>
          <sn-deck
            :id='`discard-${player.id}`'
            label='Discard'
            :deck='player.discardPile'
            :show='true'
          >
          </sn-deck>
        </v-col>
      </v-row>
    </div>
  </v-card>
</template>

<script>
  import Deck from '@/components/deck/Deck'
  import Button from '@/components/user/Button'
  import CurrentUser from '@/components/mixins/CurrentUser'
  import Player from '@/components/mixins/Player'
  import Color from '@/components/mixins/Color'

  const _ = require('lodash')

  export default {
    mixins: [ CurrentUser, Player, Color ],
    name: 'sn-player-panel',
    components: {
      'sn-user-btn': Button,
      'sn-deck': Deck
    },
    props: [ 'player', 'game' ],
    computed: {
      length: function () {
        var self = this
        return _.size(self.player.discard)
      },
      color: function () {
        var self = this
        return self.colorByPID(self.player.id)
      },
      playIcon: function () {
        var self = this
        return self.cpIs(self.player)
      },
      icon: function () {
        var self = this
        return self.playIcon ? 'play_arrow' : 'stop'
      },
      iconColor: function () {
        var self = this
        return self.playIcon ? 'green' : 'red'
      },
      canPass: function () {
        var self = this
        return self.isCPorAdmin && !self.cp.performedAction && self.isPlayerFor(self.player, self.cu) && self.game.phase == 'Play Card' 
      },
      canShow: function () {
        var self = this
        return self.isPlayerFor(self.player, self.cu)
      }
    }
  }
</script>
