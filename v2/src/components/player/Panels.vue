<template>
  <v-card>
  <v-card
        v-for="player in game.players"
        :key="player.id"
    >
          <sn-player-panel
            :player="player"
            :game='game'
            @show="$emit('show', player.id)"
            @pass="$emit('pass')"
          >
          </sn-player-panel>
  </v-card>
  </v-card>
</template>

<script>
  import Panel from '@/components/player/Panel'
  import Player from '@/components/mixins/Player'
  //import Button from '@/components/user/Button'
  import Color from '@/components/mixins/Color'

  export default {
    mixins: [ Player, Color ],
    name: 'sn-player-panels',
    components: {
      //'sn-user-btn': Button,
      'sn-player-panel': Panel
    },
    props: [ 'value', 'game' ],
    methods: {
      icon: function (player) {
        var self = this
        if (player.passed) {
          return 'mdi-stop-circle'
        }
        if (self.cpIs(player)) {
          return 'mdi-play-circle'
        }
        return ''
      }
    },
    computed: {
      tab: {
        get: function () {
          return this.value
        },
        set: function (value) {
          this.$emit('input', value)
        }
      }
    }
  }
</script>
