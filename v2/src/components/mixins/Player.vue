<script>

  const _ = require('lodash')

  export default {
    computed: {
      cp: function () {
        var self = this
        return self.playerByPID(self.cpid)
      },
      cpid: function () {
        var self = this
        return _.get(self.game.cpids, 0, -1)
      },
      isCP: function () {
        var self = this
        return self.isPlayerFor(self.cp, self.$root.cu)
      },
      isCPorAdmin: function () {
        var self = this
        return (self.$root.cu && self.$root.cu.admin) || self.isCP
      }
    },
    methods: {
      cpIs: function (player) {
        var self = this
        var pid = _.get(player, 'id', -2)
        return self.cpid == pid
      },
      playerByPID: function (pid) {
        var self = this
        return _.find(self.game.players, ['id', pid])
      },
      playersByPIDS: function (pids) {
        var self = this
        return _.map(pids, function (pid) {
          return self.playerByPID(pid)
        })
      },
      playerByUID: function (uid) {
        var self = this
        return _.find(self.game.players, ['user.id', uid])
      },
      pidByUID: function (uid) {
        var self = this
        return _.get(self.playerByUID(uid), 'id', -1)
      },
      isPlayerFor: function (player, user) {
        var admin = _.get(user, 'admin', false)
        if (admin) {
          return true
        }
        var pid = _.get(player, 'user.id', -1)
        var uid = _.get(user, 'id', -2)
        return pid === uid
      }
    }
  }
</script>
