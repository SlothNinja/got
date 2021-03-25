<template>
  <v-row no-gutters>
    <v-col align='center'>
      <v-tooltip :disabled='!canReset' bottom color='info'>
        <template v-slot:activator="{ on }">
          <v-btn v-on="on" icon :disabled="!canReset" @click="$emit('action', { action: 'reset', data: { undo: game.undo }})" >
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </template>
        <span>Reset</span>
      </v-tooltip>
      <v-tooltip :disabled='!canUndo' bottom color='info'>
        <template v-slot:activator="{ on }">
          <v-btn v-on='on' :disabled='!canUndo' icon @click="$emit('action', { action: 'undo', data: { undo: game.undo }})" >
            <v-icon>mdi-undo</v-icon>
          </v-btn>
        </template>
        <span>Undo</span>
      </v-tooltip>
      <v-tooltip :disabled='!canRedo' bottom color='info'>
        <template v-slot:activator='{ on }'>
          <v-btn v-on='on' icon :disabled='!canRedo' @click="$emit('action', { action: 'redo', data: { undo: game.undo }})" >
            <v-icon>mdi-redo</v-icon>
          </v-btn>
        </template>
        <span>Redo</span>
      </v-tooltip>
      <v-tooltip :disabled='!canFinish' bottom color='info'>
        <template v-slot:activator='{ on }'>
          <v-btn v-on='on' icon :disabled='!canFinish' @click="$emit('action', { action : finishAction, data: { undo: game.undo }})" >
            <v-icon>mdi-check</v-icon>
          </v-btn>
        </template>
        <span>Finish</span>
      </v-tooltip>
      <v-tooltip bottom color='info'>
        <template v-slot:activator='{ on }'>
          <v-btn v-on='on' icon @click.native="$emit('action', { action : 'refresh' })" >
            <v-icon>mdi-refresh</v-icon>
          </v-btn>
        </template>
        <span>Refresh</span>
      </v-tooltip>
    </v-col>
  </v-row>
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
      return (self.game.status == 3) && (self.isCPorAdmin) && (self.game.undo.current > self.game.undo.committed)
    },
    canRedo: function () {
      var self = this
      return (self.game.status == 3) && (self.isCPorAdmin) && (self.game.undo.current < self.game.undo.updated)
    },
    canReset: function () {
      var self = this
      return (self.game.status == 3) && self.isCPorAdmin
    },
    canFinish: function () {
      var self = this
      return (self.game.status == 3) && self.isCPorAdmin ? (_.get(self.cp, 'performedAction', true)) : false
    },
    finishAction: function () {
      var self = this
      switch (self.game.phase) {
        case 'Place Thieves':
          return 'ptfinish'
        case 'Move Thief':
          return 'mtfinish'
        case 'Passed':
          return 'pfinish'
        default:
          return ''
      }
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
