<template>
  <v-card>
    <v-container>
      <v-row>
        <v-col>
          <span class='font-weight-black mr-1'>Title:</span>{{game.title}}
        </v-col>
      </v-row>
      <v-divider class='my-1'></v-divider>
      <v-row no-gutters justify='space-between'>
        <v-col cols='8'>
          <div><span class='font-weight-black'>ID:</span> {{game.id}}</div>
          <div><span class='font-weight-black'>Turn:</span> {{game.turn}}</div>
          <div><span class='font-weight-black'>Current Player:</span> {{nameFor(cp)}}</div>
          <v-checkbox
            v-model='checked'
            label='Live Updates'
            ref='live'
            >
          </v-checkbox>
        </v-col>
        <v-col cols='4'>
          <v-row v-if='game.status == 2'>
            <v-col>
              <v-dialog v-model='dialog' max-width='600px'>
                <template v-slot:activator="{ on }">
                  <v-btn small class='mt-5' color='info' dark v-on='on'>Results</v-btn>
                </template>
                <v-card>
                  <sn-results-table :game='game'></sn-results-table>
                </v-card>
              </v-dialog>
            </v-col>
          </v-row>
          <v-row v-else>
            <v-col>
              <v-row no-gutters>
                <v-col align='center' class='font-weight-black'>
                  Jewels
                </v-col>
              </v-row>
              <v-row no-gutters>
                <v-col align='center'>
                  <v-card color='green' height='90' min-width='90' width='90' >
                  <v-tooltip bottom>
                    <template v-slot:activator="{ on }">
                      <space-image v-on="on" :value='game.jewels'></space-image>
                    </template>
                    <span>{{tooltip(game.jewels)}}</span>
                  </v-tooltip>
                  </v-card>
                </v-col>
              </v-row>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
    </v-container>
  </v-card>
</template>

<script>
  import SpaceImage from '@/components/board/SpaceImage'
  import Tooltip from '@/components/mixins/Tooltip'
  import ResultsTable from '@/components/game/ResultsTable'
  import Player from '@/components/mixins/Player'

  export default {
    mixins: [ Player, Tooltip ],
    name: 'sn-status-panel',
    data () {
      return {
        dialog: true
      }
    },
    components: {
      'space-image': SpaceImage,
      'sn-results-table': ResultsTable
    },
    props: [ 'game', 'live' ],
    computed: {
      checked: {
        get: function () {
          return this.live
        },
        set: function (value) {
          this.$emit('update:live', value)
        }
      }
    }
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
