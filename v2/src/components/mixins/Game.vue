<script>
  export default {
    computed: {
      game: {
        get: function () {
          return this.$root.game
        },
        set: function (value) {
          this.$root.game = value
        }
      }
    }
  }
</script>
