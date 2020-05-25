<template>
  <li>
    Played {{kind}} card.
  </li>
</template>

<script>

  const _ = require('lodash')

  export default {
    name: 'sn-log-play-card-msg',
    props: [ 'message', 'game' ],
    computed: {
      kind: function () {
        var self = this
        var kind = _.get(self.message, 'card.kind', 'none')
        switch (kind) {
          case 'start-lamp':
            return 'lamp'
          case 'start-camel':
            return 'camel'
          default:
            return kind
        }
      }
    }
  }
</script>
