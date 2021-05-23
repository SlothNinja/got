<template>
  <v-simple-table dense fixed-header height='300px'>
    <thead>
      <tr>
        <th></th>
        <th class='text-center' v-for="player in game.players" :key="`player-${player.id}-score`">
          <sn-user-btn :user='userFor(player)' :color='colorByPID(player.id)' size='small' >{{nameFor(player)}}</sn-user-btn>
        </th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td>Score</td>
        <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-score`">{{player.score}}</td>
      </tr>
      <tr>
        <td>Finish</td>
        <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-finish`">{{player.stats.finish}}</td>
      </tr>
      <tr>
        <td>Moves</td>
        <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-moves`">{{moves(player)}}</td>
      </tr>
      <tr>
        <td>Time Spent</td>
        <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-think`">{{think(player)}}</td>
      </tr>
      <tr>
        <td>Time Per Move</td>
        <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-tpm`">{{perTurn(player)}}</td>
      </tr>
      <tr v-for="(num, index) in thieves" :key="`place-${index}`">
        <td>Thief {{num}} Placed On</td>
        <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-thief-{{index}}`">{{thief(player, index)}}</td>
      </tr>
      <template v-for="(kind, index) in kinds">
        <tr v-if="anyClaimed(kind)" :key="`claimed-${index}`">
          <td>Claimed {{kind}}</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-claimed-lamps`">{{claimed(player, kind)}}</td>
        </tr>
      </template>
      <template v-for="(kind, index) in kinds">
        <tr v-if="anyPlayed(kind)" :key="`played-${index}`">
          <td>Played {{kind}}</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-claimed-lamps`">{{played(player, kind)}}</td>
        </tr>
      </template>
      <template v-for="(kind, index) in kinds">
        <tr v-if="anyJewelsAs(kind)" :key="`jewels-as-${index}`">
          <td>Jewels Played As {{kind}}</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-jewelsAs-${index}`">{{jewelsAs(player, kind)}}</td>
        </tr>
      </template>
    </tbody>
  </v-simple-table>
</template>

<script>
import Player from '@/components/mixins/Player'
import Button from '@/components/lib/user/Button'
import Color from '@/components/mixins/Color'

const _ = require('lodash')

export default {
  mixins: [ Player, Color ],
  name: 'sn-results-table',
  props: [ 'game' ],
  components: {
    'sn-user-btn': Button
  },
  computed: {
    kinds: function () {
      return [ "Lamp", "Camel", "Sword", "Carpet", "Coins", "Turban", "Jewels", "Guard" ]
    },
    thieves: function () {
      return this.twoThief ? [ 1, 2 ] : [ 1, 2, 3 ]
    }
  },
  methods: {
    thief: function (player, index) {
      return _.get(_.invert(player.stats.placed[index]), 1, 'none')
    },
    claimed: function (player, kind) {
      var claimed = _.get(player, 'stats.claimed', false)
      if (!claimed) {
        return 0
      }
      return _.get(claimed, kind, 0)
    },
    anyClaimed: function (kind) {
      let self = this
      let players = self.game.players
      let played = _.reduce(players, function(sum, player) {
        return sum + self.claimed(player, kind)
      }, 0)
      return played > 0
    },
    played: function (player, kind) {
      var played = _.get(player, 'stats.cardsPlayed', false)
      if (!played) {
        return 0
      }
      return _.get(played, kind, 0)
    },
    anyPlayed: function (kind) {
      let self = this
      let players = self.game.players
      let played = _.reduce(players, function(sum, player) {
        return sum + self.played(player, kind)
      }, 0)
      return played > 0
    },
    jewelsAs: function (player, kind) {
      var jewelAs = _.get(player, 'stats.jewelsAs', false)
      if (!jewelAs) {
        return 0
      }
      return _.get(jewelAs, kind, 0)
    },
    anyJewelsAs: function (kind) {
      let self = this
      let players = self.game.players
      let played = _.reduce(players, function(sum, player) {
        return sum + self.jewelsAs(player, kind)
      }, 0)
      return played > 0
    },
    duration: function (nano) {
      return nano/1000000000
    },
    think: function (player) {
      var self = this
      var sec = _.get(player, 'stats.think', 0)/1000000000
      return self.humanize(sec)
    },
    moves: function (player) {
      return _.get(player, 'stats.moves', 1)
    },
    perTurn: function (player) {
      var self = this
      var sec = _.get(player, 'stats.think', 0)/1000000000
      var moves = self.moves(player)
      var human = self.humanize(sec/moves)
      return `${human} / move`
    },
    humanize: function (sec) {
      if (sec < 60) {
        return `${_.floor(sec, 2)} s`
      }
      if (sec < 3600) {
        return `${_.floor(sec / 60, 2)} m`
      }
      if (sec < 86400) {
        return `${_.floor(sec / 3600, 2)} h`
      }
      return `${_.floor(sec / 86400, 2)} d`
    }
  }
}
</script>
