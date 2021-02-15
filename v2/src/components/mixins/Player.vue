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
      indexOf: function (player) {
        let self = this
        return _.findIndex(self.game.players, ['id', player.id])
      },
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
      playerByUser: function (u) {
        let self = this
        let uid = _.get(u, 'id', false)
        if (uid) {
          return self.playerByUID(uid)
        }
        return {}
      },
      playerByUID: function (uid) {
        let self = this
        let index = self.uidIndex(uid)
        if (index == -1) {
          return {}
        }
        let pid = index + 1
        return _.find(self.game.players, [ 'id', pid ])
      },
      pidByUID: function (uid) {
        let self = this
        let player = self.playerByUID(uid)
        if (player == {}) {
          return -1
        }
        return player.id
      },
      isPlayerFor: function (player, user) {
        let self = this
        let pid = self.uidFor(player)
        let uid = _.get(user, 'id', -2)
        return pid === uid
      },
      uidIndex: function (uid) {
        let self = this
        return _.indexOf(self.game.userIds, uid)
      },
      nameFor: function (p) {
        let self = this
        if (p) {
          let index = p.id - 1
          return self.game.userNames[index]
        }
        return ""
      },
      uidFor: function (p) {
        let self = this
        if (p) {
          let index = p.id - 1
          return self.game.userIds[index]
        }
        return -1
      },
      emailHashFor: function (p) {
        let self = this
        if (p) {
          let index = p.id - 1
          return self.game.userEmailHashes[index]
        }
        return -1
      },
      gravTypeFor: function (p) {
        let self = this
        if (p) {
          let index = p.id - 1
          return self.game.userGravTypes[index]
        }
        return -1
      },
      userFor: function (p) {
        let self = this
        return {
          id: self.uidFor(p),
          name: self.nameFor(p),
          emailHash: self.emailHashFor(p),
          gravType: self.gravTypeFor(p)
        }
      },
    }
  }
</script>
