<template>
  <v-card>

    <v-container
      fluid
      ref='chatbox'
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
            v-if='loading'
            >
          </v-progress-linear>

            <div id='chat-bottom'></div>

      </v-card>
    </v-container>

    <v-divider></v-divider>

    <v-container>
      <v-card class='mb-2'>
        <v-card-text>

          <v-textarea
            auto-grow
            color='green'
            label='Message'
            placeholder="Type Message.  Press 'Enter' Key To Send."
            v-model='message'
            rows=1
            clearable
            autofocus
            v-on:keyup.enter='send'
            >
          </v-textarea>

        </v-card-text>
      </v-card>

    </v-container>
  </v-card>
</template>

<script>
import Message from '@/components/chat/Message'
import goTo from 'vuetify/es5/services/goto'

const _ = require('lodash')
const axios = require('axios')

export default {
  data: function () {
    return {
      count: 0,
      messages: [],
      message: '',
      loading: true
    }
  },
  components: {
    'sn-message': Message
  },
  name: 'sn-chat-box',
  props: [ 'user', 'game' ],
  // created () {
  //   this.fetchData()
  // },
  // activated: function () {
  //   let self = this
  //   self.scroll()
  //   self.$emit('title', self.title)
  // },
  computed: {
    msgsPath: function() {
      let self = this
      return `mlog/${self.$route.params.id}`
    },
    title: function () {
      let self = this
      return `Chat (${self.messages.length} messages)`
    }
  },
  watch: {
    loading: function (isLoading, wasLoading) {
      let self = this
      if (wasLoading && !isLoading) {
        self.scroll()
      }
    }
  },
  methods: {
    msgId: function(index) {
      return `msg-${index}`
    },
    fetchData: _.debounce(
      function () {
        let self = this
        self.loading = true
        axios.get(self.msgsPath)
          .then(function (response) {
            console.log(`fetchData: ${JSON.stringify(response)}`)
            let msgs = _.get(response, 'data.messages', false)
            if (msgs) {
              self.messages = self.messages.concat(msgs)
              self.$emit('title', self.title)
            }
            self.loading = false
          })
          .catch(function () {
            self.loading = false
            self.$emit('message', 'Server Error.  Try refreshing page.')
          })
      },
      500
    ),
    clear: function () {
      let self = this
      self.message = ''
    },
    send: function () {
      let self = this
      let obj = {
        message: self.message,
        creator: self.user
      }

      self.loading = true

      axios.put(self.msgsPath+"/add", obj)
        .then(function (response) {
          let msg = _.get(response, 'data', false)
          if (msg) {
            self.add(msg)
            self.$emit('title', self.title)
            self.clear()
          }
          self.loading = false
        })
        .catch(function () {
          self.loading = false
          self.$emit('message', 'Server Error.  Try refreshing page.')
        })
    },
    scroll: function() {
      let self = this
      self.$nextTick(function () {
        goTo('#chat-bottom', { container: self.$refs.chatbox })
      })
    },
    add: function (message) {
      console.log(`message: ${JSON.stringify(message)}`)
      let self = this
      let msg = _.get(message, 'message', false)
      if (msg) {
        self.messages.push(msg)
        self.scroll()
      }
    }
  }
}
</script>
