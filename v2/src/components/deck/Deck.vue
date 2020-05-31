<template>
  <div>
    <div class='text-center'>
      <v-badge
        color='info'
        dark
        overlap
        bottom
      >
        <span slot="badge">{{deck ? deck.length : 0}}</span>
          <sn-card
            :show='show'
            :kind='topCard(deck)'
          >
          </sn-card>
      </v-badge>
    </div>
    <div class='text-center font-weight-black'>{{label}}</div>
  </div>
</template>

<script>
  import Card from '@/components/card/Card'
  import Player from '@/components/mixins/Player'

  const _ = require('lodash')

  export default {
    mixins: [ Player ],
    name: 'sn-deck',
    components: {
      'sn-card': Card
    },
    props: [ 'deck', 'show', 'label' ],
    methods: {
      topCard: function (deck) {
        var card = _.first(deck)
        return card ? card.kind : 'card-back'
      },
    }
  }
</script>
