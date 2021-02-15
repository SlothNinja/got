<template>
  <v-card>
    <v-tabs
      id='player-tabs'
      v-model="tab"
      background-color="green"
      center-active
      icons-and-text
      grow
      dark
    >

      <v-tabs-slider color='yellow' ></v-tabs-slider>

      <v-tab
        v-for="player in game.players"
        :key="player.id"
        ripple
      >
        {{nameFor(player)}}
        <v-icon>{{icon(player)}}</v-icon>
      </v-tab>
    </v-tabs>

      <v-tabs-items v-model='tab'>
        <v-tab-item
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
        </v-tab-item>
      </v-tabs-items>
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
