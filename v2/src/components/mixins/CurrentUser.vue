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
        return _.get(this.cu, 'id', false)
      },
      cuLoading: {
        get: function () {
          return this.$root.cuLoading
        },
        set: function (value) {
          this.$root.cuLoading = value
        }
      }
    }
  }
</script>
