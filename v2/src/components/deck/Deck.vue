<template>
  <div>
    <div class='text-center'>
      <v-badge
        color='info'
        dark
        overlap
        bottom
        :content='count'
        >
        <sn-card
          :show='show'
          :kind='topCard(deck)'
          :width='50'
          >
        </sn-card>
      </v-badge>
    </div>
    <div class='text-center font-weight-black'>
      <slot></slot>
    </div>
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
    props: [ 'deck', 'show', 'size' ],
    methods: {
      topCard: function (deck) {
        var card = _.first(deck)
        return card ? card.kind : 'card-back'
      },
    },
    computed: {
      count: function () {
        return _.toString(_.size(this.deck))
      }
    }
  }
</script>
