<template>
  <v-navigation-drawer
    app
    temporary
    hide-overlay
    v-model='cardbar'
    >
    <v-container>
      <v-row
        v-for='(count, kind) in cards'
        :key='kind'
        >
        <v-col class='text-center' >
          <div @click='selected(kind)' >
            <sn-card-with-count
              :kind='kind'
              :count='count'
              >
            </sn-card-with-count>
          </div>
        </v-col>
      </v-row>
    </v-container>
  </v-navigation-drawer>
</template>

<script>
import WithCount from '@/components/card/WithCount'
import Player from '@/components/mixins/Player'
import CurrentUser from '@/components/mixins/CurrentUser'

const _ = require('lodash')

export default {
  name: 'sn-card-bar',
  mixins: [ Player, CurrentUser ],
  components: {
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
