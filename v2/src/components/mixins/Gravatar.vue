<script>
  var _ = require('lodash')

  export default {
    methods: {
      gravatar: function (hash, size, t) {
        let self = this
        let sz = _.get(self.gravSizes(), size, '64')

        if (!(_.includes(self.gravTypes(), t))) {
          t = 'monsterid'
        }
        if (t == 'personal') {
          return `https://www.gravatar.com/avatar/${hash}?s=${sz}`
        }
        return `https://www.gravatar.com/avatar/${hash}?s=${sz}&d=${t}&f=y`
      },
      gravTypes: function () {
        return ['personal', 'identicon', 'monsterid', 'retro', 'robohash']
      },
      gravSizes: function () {
        return { 'x-small': '24', 'small': '30', 'medium': '48', 'large': '54' }
      }
    }
  }
</script>
