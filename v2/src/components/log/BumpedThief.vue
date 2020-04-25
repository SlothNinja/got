<template>
  <li>Bumped thief of {{name}} to {{value.card}} at {{row}}{{value.to.column}}.</li>
</template>

<script>
  import Text from '@/components/mixins/Text'
  import Player from '@/components/mixins/Player'

  export default {
    mixins: [ Text, Player ],
    name: 'sn-log-bumped-thief-msg',
    props: [ 'value', 'game' ],
    computed: {
      row: function () {
        var self = this
        return self.rowText(self.value.to.row)
      },
      name: function () {
        var self = this
        return self.playerByPID(self.value.pid).user.name
      }
    }
  }
</script>
