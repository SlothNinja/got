<template>
  <li>
    Claimed {{message.area.card.kind}} card at
    <span v-if='!message.hand'>{{row}}{{message.area.column}}.</span>
    <span v-else>{{row}}{{message.area.column}} and placed it in hand.</span>
  </li>
</template>

<script>
  import Text from '@/components/mixins/Text'

  export default {
    mixins: [ Text ],
    name: 'sn-log-claim-item-msg',
    props: [ 'message', 'game' ],
    computed: {
      row: function () {
        var self = this
        return self.rowText(self.message.area.row)
      }
    }
  }
</script>
