<template>
  <v-card class='mx-auto' >

    <v-system-bar
      color='green'
      class='white--text'
    >
      <span>Turn: {{turn}}</span>
    </v-system-bar>

    <v-row v-if='player'>
      <v-col cols='3'>
        <div class='text-center'>
          <sn-user-btn
            :user='userFor(player)'
            size='x-small'
            :color='colorByPID(player.id)'
          >
          </sn-user-btn>
        </div>
        <div class='text-center'>
          {{userFor(player).name}}
        </div>
      </v-col>
      <v-col cols='9'>
        <ul>
          <sn-log-message
            v-for='(message, index) in entry.messages'
            :key='index'
            :message='message'
            :game='game'
          >
          </sn-log-message>
        </ul>
      </v-col>
    </v-row>
    <v-row v-else>
      <v-col cols='12'>
        <ul>
          <sn-log-message
            v-for='(message, index) in entry.messages'
            :key='index'
            :message='message'
            :game='game'
          >
          </sn-log-message>
        </ul>
      </v-col>
    </v-row>
    <v-divider></v-divider>
    <div class='text-center caption'>
      <span>{{updatedAt}}</span>
    </div>
  </v-card>
</template>

<script>
  import Message from '@/components/log/Message'
  import Button from '@/components/lib/user/Button'
  import Color from '@/components/mixins/Color'
  import Player from '@/components/mixins/Player'

  const _ = require('lodash')

  export default {
    mixins: [ Color, Player ],
    name: 'sn-log-entry',
    props: [ 'entry', 'game' ],
    components: {
      'sn-log-message': Message,
      'sn-user-btn': Button
    },
    computed: {
      turn: function () {
        var self = this
        return _.get(self.entry, 'turn', 0)
      },
      player: function () {
        var self = this
        var pid = _.get(self.entry, 'pid', 0)
        return _.find(self.game.players, ['id', pid])
      },
      updatedAt: function () {
        var self = this
        var d = _.get(self.entry, 'updatedAt', false)
        if (d) {
          return new Date(d).toString()
        }
        return ''
      }
    }
  }
</script>
