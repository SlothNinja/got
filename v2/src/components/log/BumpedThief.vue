<template>
  <li>Bumped thief of {{name}} to {{message.area.card.kind}} at {{row}}{{message.area.column}}.</li>
</template>

<script>
  import Text from '@/components/mixins/Text'
  import Player from '@/components/mixins/Player'
        
  const _ = require('lodash')

  export default {
    mixins: [ Text, Player ],
    name: 'sn-log-bumped-thief-msg',
    props: [ 'message', 'game' ],
    computed: {
      row: function () {
        var self = this
        return self.rowText(self.message.area.row)
      },
      name: function () {
        var self = this
        var pid = _.get(self.message, 'area.thief', 0)
        if (pid == 0) {
          return ''
        }
        return self.nameFor(self.playerByPID(pid))
      }
    }
  }
</script>
