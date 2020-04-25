<template>
  <div>
    <v-tooltip bottom color='info'>
      <template v-slot:activator="{ on }">
        <v-btn
          v-on="on"
          icon
          :disabled="!canReset"
                @click.native="$emit('action', { action: 'reset' })"
        >
          <v-icon>clear</v-icon>
        </v-btn>
      </template>
      <span>Reset</span>
    </v-tooltip>
    <v-tooltip bottom color='info'>
      <template v-slot:activator="{ on }">
      <v-btn
          v-on='on'
        icon
        :disabled='!canUndo'
        @click="$emit('action', { action: 'undo' })"
      >
        <v-icon>undo</v-icon>
      </v-btn>
      </template>
      <span>Undo</span>
    </v-tooltip>
    <v-tooltip bottom color='info'>
      <template v-slot:activator='{ on }'>
      <v-btn
          v-on='on'
        icon
        :disabled='!canRedo'
        @click="$emit('action', { action: 'redo' })"
      >
        <v-icon>redo</v-icon>
      </v-btn>
      </template>
      <span>Redo</span>
    </v-tooltip>

    <v-tooltip bottom color='info'>
      <template v-slot:activator='{ on }'>
      <v-btn
        v-on='on'
        icon
        :disabled='!canFinish'
        @click="$emit('action', { action : 'finish' })"
      >
        <v-icon>done</v-icon>
      </v-btn>
      </template>
      <span>Finish</span>
    </v-tooltip>

    <v-tooltip bottom color='info'>
      <template v-slot:activator='{ on }'>
        <v-btn
          v-on='on'
          icon
          @click.native="$emit('action', { action : 'refresh' })"
        >
          <v-icon>refresh</v-icon>
        </v-btn>
      </template>
      <span>Refresh</span>
    </v-tooltip>

  </div>
</template>

<script>
  import Player from '@/components/mixins/Player'
  import CurrentUser from '@/components/mixins/CurrentUser'

  var _ = require('lodash')

  export default {
    name: 'sn-controlbar',
    mixins: [ Player, CurrentUser ],
    props: [ 'value' ],
    computed: {
      canUndo: function () {
        var self = this
        return (self.isCPorAdmin) && (self.value.undoStack.current > self.value.undoStack.committed)
      },
      canRedo: function () {
        var self = this
        return (self.isCPorAdmin) && (self.value.undoStack.current < self.value.undoStack.updated)
      },
      canReset: function () {
        var self = this
        return self.isCPorAdmin
      },
      canFinish: function () {
        var self = this
        return self.isCPorAdmin ? (_.get(self.cp, 'performedAction', true)) : false
      },
      game: {
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
