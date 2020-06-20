<template>
  <v-snackbar
    bottom
    color='info'
    vertical
    v-model='open'
  >
    <slot></slot>

    <template v-slot:action='{ attrs }'>
      <v-btn
        v-bind='attrs'
        text
        @click.native='open = false'
      >
        Close
      </v-btn>
    </template>
  </v-snackbar>
</template>

<script>
  export default {
    name: 'sn-snackbar',
    props: [ 'value' ],
    computed: {
      open: {
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
