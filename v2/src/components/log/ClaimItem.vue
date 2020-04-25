<template>
  <li>
    Claimed {{value.card}} card at
    <span v-if='!value.toHand'>{{row}}{{value.from.column}}.</span>
    <span v-else>{{row}}{{value.from.column}} and placed it in hand.</span>
  </li>
</template>

<script>
  import Text from '@/components/mixins/Text'

  export default {
    mixins: [ Text ],
    name: 'sn-log-claim-item-msg',
    props: [ 'value', 'game' ],
    computed: {
      row: function () {
        var self = this
        return self.rowText(self.value.from.row)
      }
    }
  }
</script>
