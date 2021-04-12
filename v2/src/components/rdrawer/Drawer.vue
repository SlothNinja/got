<template>
  <v-navigation-drawer
    width="500"
    fixed
    v-model="drawer"
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
      <v-toolbar-title>{{title}}</v-toolbar-title>

      <v-spacer></v-spacer>

      <v-tooltip bottom color='info'>
        <template v-slot:activator='{ on }'>
          <v-btn v-on='on' icon @click.stop='mode = "glog"'>
            <v-icon>history</v-icon>
          </v-btn>
        </template>
        <span>Game Log</span>
      </v-tooltip>

      <v-tooltip bottom color='info'>
        <template v-slot:activator='{ on }'>
          <v-btn v-on='on' icon @click.stop='mode = "chat"'>
            <v-icon>chat</v-icon>
          </v-btn>
        </template>
        <span>Chat</span>
      </v-tooltip>

    </v-toolbar>

    <keep-alive>
      <sn-game-log
        @message='sbMessage = $event; sbOpen = true'
        @title='title = $event' v-if='mode == "glog"'
        :game='game'
      >
      </sn-game-log>
    </keep-alive>

    <keep-alive>
      <sn-chat-box
        @message='sbMessage = $event; sbOpen = true'
        @title='title = $event'
        v-if='mode == "chat"'
        :user='cu'
      >
      </sn-chat-box>
    </keep-alive>

  </v-navigation-drawer>
</template>

<script>
  import GameLog from '@/components/log/Box'
  import ChatBox from '@/components/chat/Box'
  import CurrentUser from '@/components/lib/mixins/CurrentUser'

  export default {
    name: 'sn-rdrawer',
    mixins: [ CurrentUser ],
    props: [ 'value', 'game' ],
    components: {
      'sn-game-log': GameLog,
      'sn-chat-box': ChatBox
    },
    data () {
      return {
        mode: 'glog',
        title: ''
      }
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
