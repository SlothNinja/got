<template>
  <div>
    <sn-toolbar app v-model="nav" >
      <sn-snackbar v-model="snackbar.open">
        <div class="text-xs-center">
          {{snackbar.message}}
        </div>
      </sn-snackbar>
    </sn-toolbar>
  </div>
</template>

<script>
  import Toolbar from '@/components/Toolbar'
  import Snackbar from '@/components/Snackbar'

  export default {
    data () {
      return {
        nav: false,
        idToken: ''
      }
    },
    components: {
      'sn-snackbar': Snackbar,
      'sn-toolbar': Toolbar
    },
    computed: {
      snackbar: {
        get: function () {
          return this.$root.snackbar
        },
        set: function (value) {
          this.$root.snackbar = value
        }
      }
    }
  }
</script>
