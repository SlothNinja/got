<template>
  <v-card>

    <v-card-text>

      <v-row>
        <v-col>
          <span class='font-weight-black mr-1'>Title:</span>{{game.title}}
        </v-col>
      </v-row>
      <v-divider class='my-1'></v-divider>
      <v-row no-gutters justify='space-between'>
        <v-col cols='6'>
          <div><span class='font-weight-black'>ID:</span> {{game.id}}</div>
          <div><span class='font-weight-black'>Turn:</span> {{game.turn}}</div>
        </v-col>
        <v-col cols='4'>
            <div class='text-center font-weight-black'>Jewels</div>
            <v-card
              color='green'
              height='90'
              width='90'
            >
              <v-tooltip bottom>
                <template v-slot:activator="{ on }">
                  <space-image v-on="on" :value='game.jewels'></space-image>
                </template>
                <span>{{tooltip(game.jewels)}}</span>
              </v-tooltip>
            </v-card>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script>
  import SpaceImage from '@/components/board/SpaceImage'
  // import Button from '@/components/player/Button'
  import Tooltip from '@/components/mixins/Tooltip'

  export default {
    mixins: [ Tooltip ],
    name: 'sn-status-panel',
    components: {
      // 'sn-player-btn': Button,
      'space-image': SpaceImage
    },
    props: [ 'game' ]
  }
</script>

<style scoped lang="scss">
  .jewel-card {
    height:90px;
    width:90px;
    margin: 0 auto;
    border-radius:10px;
  }
</style>
