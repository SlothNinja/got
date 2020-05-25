<template>
  <fragment>
    <li v-if='message.shuffled'>Shuffled deck.</li>
    <li>Drew card.</li>
  </fragment>
</template>

<script>
  export default {
    name: 'sn-log-draw-card-msg',
    props: [ 'message', 'game' ]
  }
</script>
