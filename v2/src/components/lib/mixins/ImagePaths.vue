<script>

  const logo = require('@/components/lib/assets/slothninja_logo_fullsize.png')

  export default {
    methods: {
      imagePath: function (name) {
        switch (name) {
          case 'logo':
            return logo
        }
      }
    }
  }
</script>
