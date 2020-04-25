<template>
  <v-navigation-drawer
    clipped
    width="500"
    temporary
    hide-overlay
    fixed
    v-model="drawer"
    right 
    light
    app
  >
    <slot></slot>
  </v-navigation-drawer>
</template>

<script>
  export default {
    name: 'sn-rdrawer',
    props: [ 'value' ],
    computed: {
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
