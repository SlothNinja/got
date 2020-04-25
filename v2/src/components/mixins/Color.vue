<script>
  import CurrentUser from '@/components/mixins/CurrentUser'
  import Player from '@/components/mixins/Player'

  const _ = require('lodash')

  export default {
    mixins: [ CurrentUser, Player ],
    computed: {
      p : function () {
        var self = this
        if (self.cuid) {
          return self.playerByUID(self.cuid)
        }
        return self.cp
      }
    },
    methods: {
      colorIndex: function (p, pid) {
        var s = _.size(p.colors)
        return (pid - p.id + s) % s
      },
      colorByPID: function (pid) {
        var self = this
        var index = self.colorIndex(self.p, pid)
        return _.get(self.p.colors, index, 'none')
      }
    }
  }
</script>
