<template>
  <v-card>
    <v-tabs
      id='player-tabs'
      v-model="tab"
      background-color="green"
      grow
      dark
      slider-color="yellow"
    >
      <v-tab
        v-for="player in game.players"
        :key="player.id"
        :href="`#player-${player.id}`"
        ripple
      >
        <v-icon>{{icon(player)}}</v-icon>
        <sn-user-btn 
          :user='player.user'
          :color='colorByPID(player.id)'
          size='small'
        >
        </sn-user-btn>
      </v-tab>
      <v-tab-item
        v-for="player in game.players"
        :key="player.id"
        :value="`player-${player.id}`"
      >
        <sn-player-panel
          :player="player"
          :game='game'
          @show="$emit('show', player.id)"
          @pass="$emit('pass')"
        >
        </sn-player-panel>
      </v-tab-item>
    </v-tabs>
  </v-card>
</template>

<script>
  import Panel from '@/components/player/Panel'
  import Player from '@/components/mixins/Player'
  import Button from '@/components/user/Button'
  import Color from '@/components/mixins/Color'

  export default {
    mixins: [ Player, Color ],
    name: 'sn-player-panels',
    components: {
      'sn-user-btn': Button,
      'sn-player-panel': Panel
    },
    props: [ 'value', 'game' ],
    methods: {
      icon: function (player) {
        var self = this
        if (player.passed) {
          return 'stop_circle'
        }
        if (self.cpIs(player)) {
          return 'play_circle_filled'
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
