<template>
  <v-card
    class='d-flex flex-column'
    style='height: 100vh'
  >

    <v-system-bar
      color='green'
      class='white--text'
    >
      <span class='title'>Game Log</span>
      <v-spacer></v-spacer>
      <span class='font-weight-black'>{{log.length}} of {{total}}</span>
    </v-system-bar>

    <v-container
      id='gamelog'
      class='flex-grow-1'
      style='overflow-y: scroll'
      v-scroll:#gamelog='onScroll'
    >
      <sn-log-entry
        class='my-1'
        v-for="(entry, index) in log"
        :key="index"
        :value='entry'
      >
      </sn-log-entry>

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
  import Entry from '@/components/log/Entry'

  const _ = require('lodash')
  const axios = require('axios')

  export default {
    data: function () {
      return {
        path: 'game/glog',
        offset: 0,
        loading: true,
        log: []
      }
    },
    props: [ 'stack' ],
    components: {
      'sn-log-entry': Entry
    },
    name: 'sn-game-log',
    created () {
      var self = this
      self.fetchData()
    },
    computed: {
      total: function () {
        var self = this
        var skip = self.stack.current - self.stack.committed
        if (skip == 0) {
          return self.stack.current
        }
        return self.stack.current - skip + 1
      }
    },
    methods: {
      onScroll ({ target: { scrollTop, clientHeight, scrollHeight }}) {
        var self = this
        if ((scrollTop + clientHeight >= scrollHeight) && (self.log.length < self.total)) {
          console.log(`length: ${self.log.length} total: ${self.total}`)
          self.fetchData()
        }
      },
      fetchData: _.debounce(
        function () {
          var self = this
          var obj = {
            stack: self.stack,
            offset: self.offset
          }
          console.log(`obj: ${JSON.stringify(obj)}`)
          self.loading = true
          axios.put(`${self.path}/${self.$route.params.id}`, obj)
            .then(function (response) {
              var msg = _.get(response, 'data.message', false)
              if (msg) {
                self.$emit('message', msg)
              }

              var offset = _.get(response, 'data.offset', false)
              if (offset) {
                self.offset = offset
              }

              var logs = _.get(response, 'data.logs', false)
              if (logs) {
                var flogs = _.filter(logs, function(log) { return !(_.isNull(_.get(log, 'entries'))) })
                self.log = self.log.concat(flogs)
              }
              self.loading = false
            })
            .catch(function () {
              self.loading = false
              self.$emit('message', 'Server Error.  Try refreshing page.')
          })
        },
        500
      )
    }
  }
</script>
