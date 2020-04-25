<template>
  <li>
    Placed thief on {{value.area.card.kind}} at {{row}}{{value.area.column}}.
  </li>
</template>

<script>
  import Text from '@/components/mixins/Text'

  export default {
    mixins: [ Text ],
    name: 'sn-log-place-thief-msg',
    props: [ 'value', 'game' ],
    computed: {
      row: function () {
        var self = this
        return self.rowText(self.value.area.row)
      }
    }
  }
</script>
