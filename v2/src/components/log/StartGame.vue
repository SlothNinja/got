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

  export default {
    mixins: [ Player ],
    name: 'sn-log-start-game-msg',
    props: [ 'value', 'game' ],
    components: {
      'sn-user-btn': Button
    },
    computed: {
      players: function () {
        var self = this
        return self.playersByPIDS(self.value.pids)
      }
    }
  }
</script>
