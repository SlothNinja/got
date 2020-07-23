<template>
  <v-card
    class='d-flex flex-column'
    style='height: 100vh'
  >

    <v-container
      id='gamelog'
      class='flex-grow-1'
      style='overflow-y: scroll'
    >
      <sn-log-entry
        class='my-1'
        v-for="(entry, index) in game.log"
        :key="index"
        :entry='entry'
        :game='game'
      >
      </sn-log-entry>
    </v-container>

  </v-card>
</template>

<script>
  import Entry from '@/components/log/Entry'

  export default {
    data: function () {
      return {
        path: 'game/glog',
        offset: 0,
        loading: true,
        log: []
      }
    },
    props: [ 'game' ],
    components: {
      'sn-log-entry': Entry
    },
    name: 'sn-game-log',
    activated: function () {
      var self = this
      self.$emit('title', 'Game Log')
    }
  }
</script>
