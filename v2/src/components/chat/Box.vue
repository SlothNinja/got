<template>
  <v-card class='d-flex flex-column' style='height: 100vh'>

    <v-system-bar
      color='green'
      class='white--text'
    >
      <span class='title'>Chat</span>
      <v-spacer></v-spacer>
      <span class='font-weight-black'>{{messages.length}} messages</span>
    </v-system-bar>

    <v-container
      ref='chatbox'
      style='overflow-y: auto'
    >
      <sn-message
        class='my-2'
        v-for='(message, index) in messages'
        :key='index'
        :message='message'
        :id='msgId(index)'
      >
      </sn-message>

      <v-progress-linear
        color='green'
        indeterminate
        v-if='loading'
      >
      </v-progress-linear>

      <div id='chat-bottom'></div>

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
    props: [ 'user' ],
    created () {
      this.fetchData()
    },
    activated () {
      this.scroll()
    },
    computed: {
      msgsPath: function() {
        var self = this
        return `game/message/${self.$route.params.id}`
      }
    },
    watch: {
      loading: function (isLoading, wasLoading) {
        var self = this
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
          var self = this
          self.loading = true
          axios.get(self.msgsPath)
            .then(function (response) {
              console.log(`fetchData: ${JSON.stringify(response)}`)
              var msgs = _.get(response, 'data.messages', false)
              if (msgs) {
                self.messages = self.messages.concat(msgs)
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
        var self = this
        self.message = ''
      },
      send: function () {
        var self = this
        var obj = {
          message: self.message,
          creator: self.user
        }

        self.loading = true

        axios.put(self.msgsPath+"/add", obj)
          .then(function (response) {
            var msg = _.get(response, 'data', false)
            if (msg) {
              self.add(msg)
              self.clear()
            }
            self.loading = false
          })
          .catch(function () {
            self.loading = false
            self.$emit('message', 'Server Error.  Try refreshing page.')
          })
      },
      // scrollHeight: function() {
      //   var self = this
      //   var height = self.$refs.chatbox.scrollHeight
      //   console.log(`height: ${height}`)
      //   return height
      // },
      scroll: function() {
        var self = this
        self.$nextTick(function () {
          goTo('#chat-bottom', { container: self.$refs.chatbox })
        })
      },
      add: function (message) {
        var self = this
        var msg = _.get(message, 'message', false)
        if (msg) {
          self.messages.push(msg)
          self.scroll()
        }
      }
    }
  }
</script>
