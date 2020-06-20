<template>
  <v-simple-table>
    <template v-slot:default>
      <thead>
        <tr>
          <th></th>
          <th class='text-center' v-for="player in game.players" :key="`player-${player.id}-score`">
            <sn-user-btn :user='player.user' :color='colorByPID(player.id)' size='small' ></sn-user-btn>
            {{player.user.name}}
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
        <tr>
          <td>Thief 1 Placed On</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-thief-1`">{{thief(player, 0)}}</td>
        </tr>
        <tr>
          <td>Thief 2 Placed On</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-thief-2`">{{thief(player, 1)}}</td>
        </tr>
        <tr>
          <td>Thief 3 Placed On</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-thief-3`">{{thief(player, 2)}}</td>
        </tr>
        <tr>
          <td>Claimed Lamps</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-claimed-lamps`">{{claimed(player, 'Lamp')}}</td>
        </tr>
        <tr>
          <td>Claimed Camels</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-claimed-camels`">{{claimed(player, 'Camel')}}</td>
        </tr>
        <tr>
          <td>Claimed Cards</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-claimed-cards`">{{claimCount(player)}}</td>
        </tr>
        <tr>
          <td>Claimed Swords</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-claimed-swords`">{{claimed(player, 'Sword')}}</td>
        <tr>
          <td>Claimed Carpets</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-claimed-carpets`">{{claimed(player, 'Carpet')}}</td>
        </tr>
        <tr>
          <td>Claimed Coins</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-claimed-coins`">{{claimed(player, 'Coins')}}</td>
        </tr>
        <tr>
          <td>Claimed Turbans</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-claimed-turbans`">{{claimed(player, 'Turban')}}</td>
        </tr>
        <tr>
          <td>Claimed Jewels</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-claimed-jewels`">{{claimed(player, 'Jewels')}}</td>
        </tr>
        <tr>
          <td>Claimed Guards</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-claimed-guards`">{{claimed(player, 'Guard')}}</td>
        </tr>
        <tr>
          <td>Played Lamps</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-played-lamps`">{{played(player, 'Lamp')}}</td>
        </tr>
        <tr>
          <td>Played Camels</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-played-camels`">{{played(player, 'Camel')}}</td>
        </tr>
        <tr>
          <td>Played Swords</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-played-swords`">{{played(player, 'Sword')}}</td>
        <tr>
          <td>Played Carpets</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-played-carpets`">{{played(player, 'Carpet')}}</td>
        </tr>
        <tr>
          <td>Played Coins</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-played-coins`">{{played(player, 'Coins')}}</td>
        </tr>
        <tr>
          <td>Played Turbans</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-played-turbans`">{{played(player, 'Turban')}}</td>
        </tr>
        <tr>
          <td>Jewels Played As Lamps</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-jewelsAs-lamps`">{{jewelsAs(player, 'Lamp')}}</td>
        </tr>
        <tr>
          <td>Jewels Played As Camels</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-jewelsAs-camels`">{{jewelsAs(player, 'Camel')}}</td>
        </tr>
        <tr>
          <td>Jewels Played As Swords</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-jewelsAs-swords`">{{jewelsAs(player, 'Sword')}}</td>
        <tr>
          <td>Jewels Played As Carpets</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-jewelsAs-carpets`">{{jewelsAs(player, 'Carpet')}}</td>
        </tr>
        <tr>
          <td>Jewels Played As Coins</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-jewelsAs-coins`">{{jewelsAs(player, 'Coins')}}</td>
        </tr>
        <tr>
          <td>Jewels Played As Turbans</td>
          <td class='text-center' v-for="player in game.players" :key="`player-${player.id}-jewelsAs-turbans`">{{jewelsAs(player, 'Turban')}}</td>
        </tr>
      </tbody>
    </template>
  </v-simple-table>
</template>

<script>
  import Player from '@/components/mixins/Player'
  import Button from '@/components/user/Button'
  import Color from '@/components/mixins/Color'

  const _ = require('lodash')

  export default {
    mixins: [ Player, Color ],
    name: 'sn-results-table',
    props: [ 'game' ],
    components: {
      'sn-user-btn': Button
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
      played: function (player, kind) {
        var played = _.get(player, 'stats.played', false)
        if (!played) {
          return 0
        }
        return _.get(played, kind, 0)
      },
      jewelsAs: function (player, kind) {
        var jewelAs = _.get(player, 'stats.jewelsAs', false)
        if (!jewelAs) {
          return 0
        }
        return _.get(jewelAs, kind, 0)
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
      claimCount: function (player) {
        var claimed = _.get(player, 'stats.claimed', 0)
        if (claimed == 0) {
          return 0
        }
        return _.reduce(claimed, function(sum, n) {
          return sum + n
        }, 0)
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
