<template>
  <v-navigation-drawer
    width="500"
    v-model="drawer"
    style='max-height: 100%'
    right 
    light
    app
  >
    <v-toolbar
      color='green'
      dark
      dense
      flat
    >
      <v-toolbar-title>Game Log</v-toolbar-title>

    </v-toolbar>

    <keep-alive>
      <sn-game-log
        @message='sbMessage = $event; sbOpen = true'
        :game='game'
      >
      </sn-game-log>
    </keep-alive>

  </v-navigation-drawer>
</template>

<script>
  import GameLog from '@/components/log/Box'
  import CurrentUser from '@/components/mixins/CurrentUser'

  export default {
    name: 'sn-log-drawer',
    mixins: [ CurrentUser ],
    props: [ 'value', 'game' ],
    components: {
      'sn-game-log': GameLog
    },
    computed: {
      drawer: {
        get: function () {
          var self = this
          return self.value
        },
        set: function (value) {
          var self = this
          self.$emit('input', value)
        }
      }
    }
  }
</script>
