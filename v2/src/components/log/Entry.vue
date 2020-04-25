<template>
  <v-card class='mx-auto' :id='`log-${value.id}`'>

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
            :user='player.user'
            size='x-small'
            :color='colorByPID(player.id)'
          >
          </sn-user-btn>
        </div>
        <div class='text-center'>
          {{player.user.name}}
        </div>
      </v-col>
      <v-col cols='9'>
        <ul>
          <sn-log-message
            v-for='(entry, index) in value.log'
            :key='index'
            :value='entry'
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
            v-for='(entry, index) in value.log'
            :key='index'
            :value='entry'
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
  import Button from '@/components/user/Button'
  import Color from '@/components/mixins/Color'

  const _ = require('lodash')

  export default {
    mixins: [ Color ],
    name: 'sn-log-entry',
    props: [ 'value' ],
    components: {
      'sn-log-message': Message,
      'sn-user-btn': Button
    },
    computed: {
      game: function () {
        var self = this
        return {
          header: self.value.header,
          state: self.value.state
        }
      },
      turn: function () {
        var self = this
        return _.get(self.value, 'log[0].turn', 0)
      },
      player: function () {
        var self = this
        var pid = _.get(self.value, 'log[0].pid', 0)
        return _.find(self.value.state.players, ['id', pid])
      },
      updatedAt: function () {
        var self = this
        var d = _.get(self.value, 'header.updatedAt', false)
        if (d) {
          return new Date(d).toString()
        }
        return ''
      }
    }
  }
</script>
