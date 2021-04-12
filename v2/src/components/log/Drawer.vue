<template>
  <v-navigation-drawer
    width="500"
    v-model="drawer"
    style='max-height: 100%'
    right 
    light
    app
    >
    <v-card class='d-flex flex-column' height='100%' >
    <v-toolbar
      color='green'
      dark
      dense
      flat
      class='flex-grow-0 flex-shrink-0'
      >
      <v-toolbar-title>Game Log</v-toolbar-title>

    </v-toolbar>


      <v-container
        ref='gamelog'
        id='gamelog'
        style='overflow-y: auto'
        >
        <sn-log-entry
          class='my-1'
          v-for="(entry, index) in game.log"
          :key="index"
          :entry='entry'
          :game='game'
          >
        </sn-log-entry>
          <div class='gamelog'></div>
      </v-container>

    </v-card>
  </v-navigation-drawer>
</template>

<script>
// import GameLog from '@/components/log/Box'
import CurrentUser from '@/components/lib/mixins/CurrentUser'
import Entry from '@/components/log/Entry'

export default {
  name: 'sn-log-drawer',
  mixins: [ CurrentUser ],
  props: [ 'value', 'game' ],
  components: {
    'sn-log-entry': Entry
  },
  watch: {
    drawer: function (oldValue, newValue) {
      if (oldValue != newValue) {
        this.scroll()
      }
    }
  },
  methods: {
    scroll: function() {
      let self = this
      self.$nextTick(function () {
        self.$vuetify.goTo('.gamelog', { container: self.$refs.gamelog } )
      })
    },
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
