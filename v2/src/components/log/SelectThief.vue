<template>
  <li>
    Selected thief on {{message.area.card.kind}} at {{row}}{{message.area.column}}.
  </li>
</template>

<script>
  import Text from '@/components/mixins/Text'

  const _ = require('lodash')

  export default {
    mixins: [ Text ],
    name: 'sn-log-select-thief-msg',
    props: [ 'message', 'game' ],
    computed: {
      kind: function () {
        var self = this
        return _.get(self.message, 'area.card.kind', 'none')
      },
      column: function () {
        var self = this
        return _.get(self.message, 'area.column', 'none')
      },
      row: function () {
        var self = this
        return self.rowText(self.message.area.row)
      }
    }
  }
</script>
