<template>
  <div>
    <v-tooltip bottom color='info'>
      <template v-slot:activator="{ on }">
        <v-btn
          v-on="on"
          icon
          :disabled="!canReset"
          @click.native="$emit('action', { action: 'reset', data: { undo: value.undo }})"
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
        @click="$emit('action', { action: 'undo', data: { undo: value.undo }})"
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
        @click="$emit('action', { action: 'redo', data: { undo: value.undo }})"
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
              @click="$emit('action', { action : finishAction, data: { undo: value.undo }})"
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
      undoValue: function () {
        var self = this
        return self.value.undo.current - 1
      },
      canUndo: function () {
        var self = this
        return (self.isCPorAdmin) && (self.value.undo.current > self.value.undo.committed)
      },
      canRedo: function () {
        var self = this
        return (self.isCPorAdmin) && (self.value.undo.current < self.value.undo.updated)
      },
      canReset: function () {
        var self = this
        return self.isCPorAdmin
      },
      canFinish: function () {
        var self = this
        return self.isCPorAdmin ? (_.get(self.cp, 'performedAction', true)) : false
      },
      finishAction: function () {
        var self = this
        if (self.game.phase == 'Place Thieves') {
          return 'ptfinish'
        }
        return 'mtfinish'
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
