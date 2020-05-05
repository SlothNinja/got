<template>
  <v-card
    :color='cardColor'
    ripple
    raised
    class='ma-1'
    hover
    >
    <v-card-text class='pa-0'>
      <v-tooltip
        left
        max-width='200'
        open-delay='800'
        :disabled='!showCard'
        color='info'
      >
        <template v-slot:activator='{ on }'>
          <div
            class='board-space'
            :class='clickable'
            @click="selected"
            v-on='on'
          >
            <sn-space-image
              v-if='showCard'
              :value='value.card'
            ></sn-space-image>
            <sn-thief-image
              v-if='showThief'
              :value='thiefColor'
            >
            </sn-thief-image>
          </div>
        </template>
        <span>{{tooltip(value.card.kind)}}</span>
      </v-tooltip>
    </v-card-text>
  </v-card>
</template>

<script>
  import SpaceImage from '@/components/board/SpaceImage'
  import Tooltip from '@/components/mixins/Tooltip'
  import Player from '@/components/mixins/Player'
  import Thief from '@/components/thief/Image'
  import Color from '@/components/mixins/Color'

  export default {
    mixins: [ Tooltip, Player, Color ],
    name: 'sn-space',
    components: {
      'sn-space-image': SpaceImage,
      'sn-thief-image': Thief
    },
    props: [ 'value', 'game' ],
    methods: {
      selected: function () {
        var self = this
        if (self.value.clickable) {
          self.$emit('selected')
        }
      },
    },
    computed: {
      cardColor: function () {
        var self = this
        return self.value.clickable ? 'yellow' : 'green darken-4'
      },
      clickable: function () {
        var self = this
        return self.value.clickable ? 'clickable' : null
      },
      showCard: function () {
        var self = this
        return self.value.card.kind != 'none'
      },
      showThief: function () {
        var self = this
        return self.value.thief != 0
      },
      thiefColor: function () {
        var self = this
        return self.colorByPID(self.value.thief)
      }
    }
  }
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang='scss'>

  .board-space {
    height:90px;
    width:90px;
  }

</style>
