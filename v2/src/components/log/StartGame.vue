<template>
  <div>
    <div>
      Good luck:
    </div>
    <v-row>
      <v-col
        v-for='(player, index) in players'
        :player='player'
        :key='index'
      >
        <sn-user-btn
          :user='player.user'
          size='x-small'
          :color='colorByPID(player.id)'
        >
        </sn-user-btn>
        <span class='ml-1'>{{player.user.name}}</span>
      </v-col>
    </v-row>
    <div>
      Have fun.
    </div>
  </div>
</template>

<script>
  import Button from '@/components/user/Button'
  import Player from '@/components/mixins/Player'
  import Color from '@/components/mixins/Color'

  export default {
    mixins: [ Player, Color ],
    name: 'sn-log-start-game-msg',
    props: [ 'message', 'game' ],
    components: {
      'sn-user-btn': Button
    },
    computed: {
      players: function () {
        var self = this
        return self.playersByPIDS(self.message.pids)
      }
    }
  }
</script>
