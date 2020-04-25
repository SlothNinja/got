<template>
  <li>
    Played {{kind}} card.
  </li>
</template>

<script>
  export default {
    name: 'sn-log-play-card-msg',
    props: [ 'value', 'game' ],
    computed: {
      kind: function () {
        var self = this
        switch (self.value.card) {
          case 'start-lamp':
            return 'lamp'
          case 'start-camel':
            return 'camel'
          default:
            return self.value.card
        }
      }
    }
  }
</script>
