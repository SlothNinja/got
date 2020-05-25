<template>
  <li>
    Placed thief on {{message.area.card.kind}} at {{row}}{{message.area.column}}.
  </li>
</template>

<script>
  import Text from '@/components/mixins/Text'

  export default {
    mixins: [ Text ],
    name: 'sn-log-place-thief-msg',
    props: [ 'message', 'game' ],
    computed: {
      row: function () {
        var self = this
        return self.rowText(self.message.area.row)
      }
    }
  }
</script>
