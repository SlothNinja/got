<script>

  const lampSpace = require(`@/assets/lamp-space.png`)
  const camelSpace = require(`@/assets/camel-space.png`)
  const startlampSpace = require(`@/assets/startlamp-space.png`)
  const startcamelSpace = require(`@/assets/startcamel-space.png`)
  const carpetSpace = require(`@/assets/carpet-space.png`)
  const coinsSpace = require(`@/assets/coins-space.png`)
  const guardSpace = require(`@/assets/guard-space.png`)
  const jewelsSpace = require(`@/assets/jewels-space.png`)
  const swordSpace = require(`@/assets/sword-space.png`)
  const turbanSpace = require(`@/assets/turban-space.png`)

  export default {
    methods: {
      spacePath: function (kind) {
        switch (kind) {
          case 'lamp':
            return lampSpace
          case 'camel':
            return camelSpace
          case 'start-lamp':
            return startlampSpace
          case 'start-camel':
            return startcamelSpace
          case 'carpet':
            return carpetSpace
          case 'coins':
            return coinsSpace
          case 'guard':
            return guardSpace
          case 'jewels':
            return jewelsSpace
          case 'sword':
            return swordSpace
          case 'turban':
            return turbanSpace
        }
      }
    }
  }
</script>
