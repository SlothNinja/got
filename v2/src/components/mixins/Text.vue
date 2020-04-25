<script>
  export default {
    methods: {
      rowText: function (row) {
        return String.fromCharCode(64 + row)
      }
    }
  }
</script>
