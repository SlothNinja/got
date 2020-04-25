<template>
  <li>Moved thief to {{value.card}} at {{row}}{{value.to.column}}.</li>
</template>

<script>
  import Text from '@/components/mixins/Text'

  export default {
    mixins: [ Text ],
    name: 'sn-log-move-thief-msg',
    props: [ 'value', 'game' ],
    computed: {
      row: function () {
        var self = this
        return self.rowText(self.value.to.row)
      }
    }
  }
</script>
