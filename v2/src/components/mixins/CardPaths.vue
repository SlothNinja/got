<script>

  const lamp = require(`@/assets/lamp.png`)
  const camel = require(`@/assets/camel.png`)
  const startlamp = require(`@/assets/startlamp.png`)
  const startcamel = require(`@/assets/startcamel.png`)
  const carpet = require(`@/assets/carpet.png`)
  const coins = require(`@/assets/coins.png`)
  const guard = require(`@/assets/guard.png`)
  const jewels = require(`@/assets/jewels.png`)
  const sword = require(`@/assets/sword.png`)
  const turban = require(`@/assets/turban.png`)
  const cardback = require(`@/assets/cardback.png`)

  export default {
    methods: {
      cardPath: function (kind) {
        switch (kind) {
          case 'lamp':
            return lamp
          case 'camel':
            return camel
          case 'start-lamp':
            return startlamp
          case 'start-camel':
            return startcamel
          case 'carpet':
            return carpet
          case 'coins':
            return coins
          case 'guard':
            return guard
          case 'jewels':
            return jewels
          case 'sword':
            return sword
          case 'turban':
            return turban
          case 'cardback':
            return cardback
        }
        return cardback
      }
    }
  }
</script>
