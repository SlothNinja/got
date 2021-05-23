<template>
  <v-card class='mb-3'>
    <v-card-text>
      <v-row no-gutters align='center'>
        <v-col cols='12' md='3'>
          <div>
            <sn-user-btn 
             :user='userFor(player)'
             :color='colorByPlayer(player)'
             size='small'
             >
              {{nameFor(player)}}
            </sn-user-btn>
          </div>
          <div><strong>Score:</strong> {{player.score}}</div>
          <div v-if='player.passed'><strong>Passed</strong></div>
        </v-col>
        <v-col cols='4' md='3'>
          <sn-deck :deck='player.hand' :show='false' size='small'>Hand</sn-deck>
        </v-col>
        <v-col cols='4' md='3'>
          <sn-deck :deck='player.drawPile' :show='false' size='small'>Draw</sn-deck>
        </v-col>
        <v-col cols='4' md='3'>
          <sn-deck :deck='player.discardPile' :show='true' size='small'>Discard</sn-deck>
        </v-col>
      </v-row>
    </v-card-text>
    <v-divider v-if='showBtns'></v-divider>
    <v-card-actions v-if='showBtns'>
      <v-row
        no-gutters
        justify='space-around'
        >
        <v-col align='center' v-if='canShow'>
          <v-btn
            small
            width='64'
            color='info'
            dark
            @click.stop="$emit('show')"
            >
            Hand
          </v-btn>
        </v-col>
        <v-col align='center' v-if='canPass'>
          <v-btn
            small
            width='64'
            color='info'
            dark
            @click.stop="$emit('pass')"
            >
            Pass
          </v-btn>
        </v-col>
      </v-row>
    </v-card-actions>
  </v-card>
</template>

<script>
  import Deck from '@/components/deck/Deck'
  import CurrentUser from '@/components/lib/mixins/CurrentUser'
  import Player from '@/components/mixins/Player'
  import Color from '@/components/mixins/Color'
  import Button from '@/components/lib/user/Button'

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
      canPass: function () {
        var self = this
        return self.isCPorAdmin && !self.cp.performedAction && self.isPlayerFor(self.player, self.cu) && self.game.phase == 'Play Card' 
      },
      canShow: function () {
        var self = this
        return self.isPlayerFor(self.player, self.cu)
      },
      showBtns: function () {
        return this.canPass || this.canShow
      }
    }
  }
</script>
