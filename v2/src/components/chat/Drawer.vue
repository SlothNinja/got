<template>
  <v-navigation-drawer
    style='min-height:100%'
    width="500"
    fixed
    v-model="drawer"
    right 
    light
    app
    >

    <v-card height='100%' class='d-flex flex-column' >
      <v-toolbar
        color='green'
        dark
        dense
        flat
        class='flex-grow-0 flex-shrink-0'
        >
        <v-toolbar-title>Chat</v-toolbar-title>

      </v-toolbar>

      <v-container
        ref='chatbox'
        fluid
        style='overflow-y: auto'
        >
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

            <div class='messagebox'></div>
      </v-container>
    <v-spacer></v-spacer>

    <v-divider></v-divider>

    <v-card>
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

    </v-card>

  </v-navigation-drawer>
</template>

<script>

import CurrentUser from '@/components/lib/mixins/CurrentUser'
import Message from '@/components/chat/Message'

const axios = require('axios')
const _ = require('lodash')

export default {
  name: 'sn-chat-drawer',
  mixins: [ CurrentUser ],
  props: [ 'value', 'game', 'unread' ],
  data: function () {
    return {
      count: 0,
      messages: [],
      message: '',
      loaded: false,
      loading: true,
      sending: false
    }
  },
  components: {
    'sn-message': Message
  },
  watch: {
    fetch: function () {
      let self = this
      if (self.fetch) {
        self.fetchData()
      }
    }
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
    fetchData: _.debounce(function () {
      let self = this
      axios.get(self.msgsPath)
        .then(function (response) {
          if (_.has(response, 'data.messages')) {
            self.messages = self.messages.concat(response.data.messages)
          }

          if (_.has(response, 'data.unread')) {
            self.$emit('update:unread', response.data.unread)
          }

          self.loaded = true
          self.loading = false
          self.scroll()
        })
        .catch(function () {
          self.loading = false
          self.$emit('message', 'Server Error.  Try refreshing page.')
        })
    }, 2000),
    scroll: function() {
      let self = this
      self.$nextTick(function () {
        self.$vuetify.goTo('.messagebox', { container: self.$refs.chatbox } )
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
      let self = this
      return self.value && !self.loaded
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
