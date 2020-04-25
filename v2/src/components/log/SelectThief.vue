<template>
  <li>
    Selected thief on {{value.card}} at {{row}}{{value.at.column}}.
  </li>
</template>

<script>
  import Text from '@/components/mixins/Text'

  export default {
    mixins: [ Text ],
    name: 'sn-log-select-thief-msg',
    props: [ 'value', 'game' ],
    computed: {
      row: function () {
        var self = this
        return self.rowText(self.value.at.row)
      }
    }
  }
</script>
