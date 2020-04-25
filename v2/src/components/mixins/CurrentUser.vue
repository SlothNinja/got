<script>
  const _ = require('lodash')

  export default {
    computed: {
      cu: {
        get: function () {
          return this.$root.cu
        },
        set: function (value) {
          this.$root.cu = value
        }
      },
      cuid: function () {
        var self = this
        return _.get(self.cu, 'id', false)
      },
      cuLoading: function () {
        return this.$root.cuLoading
      }
    }
  }
</script>
