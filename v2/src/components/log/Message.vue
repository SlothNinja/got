<template>
  <component
    :is='component'
    :message='message'
    :game='game'
  >
  </component>
</template>

<script>
  import StartGame from '@/components/log/StartGame'
  import PlaceThief from '@/components/log/PlaceThief'
  import SelectThief from '@/components/log/SelectThief'
  import MoveThief from '@/components/log/MoveThief'
  import BumpedThief from '@/components/log/BumpedThief'
  import ClaimItem from '@/components/log/ClaimItem'
  import PlayCard from '@/components/log/PlayCard'
  import DrawCard from '@/components/log/DrawCard'
  import Pass from '@/components/log/Pass'

  export default {
    name: 'sn-log-message',
    props: [ 'message', 'game' ],
    components: {
      'sn-log-start-game-msg': StartGame,
      'sn-log-place-thief-msg': PlaceThief,
      'sn-log-select-thief-msg': SelectThief,
      'sn-log-move-thief-msg': MoveThief,
      'sn-log-bumped-thief-msg': BumpedThief,
      'sn-log-claim-item-msg': ClaimItem,
      'sn-log-play-card-msg': PlayCard,
      'sn-log-draw-card-msg': DrawCard,
      'sn-log-pass-msg': Pass
    },
    computed: {
      component: function () {
        var self = this
        return `sn-log-${self.message.template}-msg`
      }
    }
  }
</script>
