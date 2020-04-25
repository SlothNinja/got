<template>
  <v-card class='d-flex flex-column' style='height: 100vh'>

    <v-system-bar
      color='green'
      class='white--text'
    >
      <span class='title'>Chat</span>
      <v-spacer></v-spacer>
      <span class='font-weight-black'>{{messages.length}} of {{count}}</span>
    </v-system-bar>

    <v-container>
      <v-card class='mb-2'>
        <v-card-text>

          <v-textarea
            auto-grow
            color='green'
            label='Message'
            placeholder="Type Message.  Press 'Send' button."
            v-model='message'
            rows=1
            clearable
            autofocus
          >
          </v-textarea>

          <div>
            <v-btn
              small
              color='info'
              class='white--text'
              :disabled="(message === '')"
              @click='send'
            >
              Send
            </v-btn>
          </div>

        </v-card-text>
      </v-card>
    </v-container>

    <v-divider></v-divider>

    <v-container
      id='chatbox'
      class='flex-grow-1'
      style='overflow-y: auto'
      v-scroll:#chatbox='onScroll'
    >
      <sn-message
        class='my-2'
        v-for='message in messages'
        :key='message.id'
        :message='message'
      >
      </sn-message>

      <v-progress-linear
        color='green'
        indeterminate
        v-if='loading'
      >
      </v-progress-linear>

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
        offset: 5,
        msgsPath: 'game/messages',
        addMsgPath: 'game/add-message',
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
    methods: {
      onScroll ({ target: { scrollTop, clientHeight, scrollHeight }}) {
        if ((scrollTop + clientHeight >= scrollHeight) && (this.messages.length < this.count)) {
          this.fetchData()
        }
      },
      fetchData: _.debounce(
        function () {
          var self = this
          var offset = self.messages.length
          self.loading = true
          axios.get(`${self.msgsPath}/${self.$route.params.id}/${offset}`)
            .then(function (response) {

              var msg = _.get(response, 'data.message', false)
              if (msg) {
                self.$emit('message', msg)
              }

              var msgs = _.get(response, 'data.messages', false)
              if (msgs) {
                self.messages = self.messages.concat(msgs)
              }

              var cnt = _.get(response, 'data.count', false)
              if (cnt) {
                  self.count = cnt
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

        axios.put(`${self.addMsgPath}/${self.$route.params.id}`, obj)
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
      add: function (message) {
        var self = this
        self.messages.unshift(message)
        self.count += 1
        self.$nextTick(function () {
          goTo(`#msg-${self.count}`, { container: '#chatbox' })
        })
      }
    }
  }
</script>
