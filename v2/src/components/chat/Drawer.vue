<template>
  <v-navigation-drawer
    style='max-height: 100%'
    width="500"
    fixed
    v-model="drawer"
    right 
    light
    app
    ref='chatbox'
    id='chatbox'
    >
    <v-toolbar
      color='green'
      dark
      dense
      flat
      >
      <v-toolbar-title>Chat</v-toolbar-title>

    </v-toolbar>
    <v-card>

      <v-container
        fluid
        style='overflow-y: auto'
        >
        <v-card>
          <sn-message
            class='my-2'
            v-for='(message, index) in messages'
            :key='index'
            :message='message'
            :id='msgId(index)'
            :game='game'
            >
          </sn-message>

            <v-progress-linear
              color='green'
              indeterminate
              v-if='loading || sending'
              >
            </v-progress-linear>


        </v-card>

        <v-divider></v-divider>

        <v-card class='messagebox mb-2'>
          <v-card-text>

            <v-textarea
              auto-grow
              color='green'
              label='Message'
              placeholder="Type Message.  Press 'Enter' Key To Send."
              v-model='message'
              rows=1
              clearable
              v-on:keyup.enter='send'
              >
            </v-textarea>

          </v-card-text>
        </v-card>

      </v-container>
    </v-card>


  </v-navigation-drawer>
</template>

<script>

import CurrentUser from '@/components/mixins/CurrentUser'
import Message from '@/components/chat/Message'

const axios = require('axios')
const _ = require('lodash')

export default {
  name: 'sn-chat-drawer',
  mixins: [ CurrentUser ],
  props: [ 'value', 'game' ],
  data: function () {
    return {
      count: 0,
      messages: [],
      message: '',
      loading: true,
      sending: false
    }
  },
  components: {
    'sn-message': Message
  },
  watch: {
    fetch: function (oldValue, newValue) {
      if (oldValue != newValue) {
        this.fetchData()
      }
    }
  },
  mounted: function () {
    this.chatboxContent = this.$refs['chatbox'].$el.querySelector('div.v-navigation-drawer__content')
  },
  methods: {
    msgId: function(index) {
      return `msg-${index}`
    },
    clear: function () {
      let self = this
      self.message = ''
    },
    send: function () {
      let self = this
      let obj = {
        message: self.message,
        creator: self.cu
      }

      self.sending = true

      axios.put(self.msgsPath+"/add", obj)
        .then(function (response) {
          let msg = _.get(response, 'data', false)
          if (msg) {
            self.add(msg)
            self.clear()
          }
          self.sending = false
        })
        .catch(function () {
          self.sending = false
          self.$emit('message', 'Server Error.  Try refreshing page.')
        })
    },
    fetchData: _.once(function () {
      let self = this
      axios.get(self.msgsPath)
        .then(function (response) {
          let msgs = _.get(response, 'data.messages', false)
          if (msgs) {
            self.messages = self.messages.concat(msgs)
          }
          self.loading = false
          self.scroll()
        })
        .catch(function () {
          self.loading = false
          self.$emit('message', 'Server Error.  Try refreshing page.')
        })
    }),
    scroll: function() {
      let self = this
      self.$nextTick(function () {
        self.$vuetify.goTo('.messagebox', { container: self.chatboxContent } )
      })
    },
    add: function (message) {
      let self = this
      let msg = _.get(message, 'message', false)
      if (msg) {
        self.messages.push(msg)
        self.scroll()
      }
    }
  },
  computed: {
    fetch: function () {
      return this.value && this.loading
    },
    msgsPath: function() {
      let self = this
      return `mlog/${self.$route.params.id}`
    },
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
