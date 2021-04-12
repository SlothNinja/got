<script>
  import CurrentUser from '@/components/lib/mixins/CurrentUser'
  import Player from '@/components/mixins/Player'

  const _ = require('lodash')

  export default {
    mixins: [ CurrentUser, Player ],
    computed: {
      p: function () {
        let self = this
        if (self.cuid) {
          return self.playerByUID(self.cuid)
        }
        return false
      },
      colors: function () {
        let self = this
        if (self.p) {
          self.p.colors
        }
        return _.slice([ 'yellow', 'purple', 'green', 'black' ], 0, self.game.numPlayers )
      }
    },
    methods: {
      colorIndex: function (pid) {
        let s = this.game.numPlayers
        let pid2 = this.p ? this.p.id : 0
        return (pid - pid2 + s) % s
      },
      colorByPID: function (pid) {
        let self = this
        let index = self.colorIndex(pid)
        return _.nth(self.colors, index, 'white')
      },
      colorByPlayer: function (p) {
        let self = this
        return self.colorByPID(p.id)
      },
      colorByUser: function (u) {
        let self = this
        let p = self.playerByUser(u)
        if (p) {
          return self.colorByPlayer(p)
        }
        return 'white'
      },
    }
  }
</script>
