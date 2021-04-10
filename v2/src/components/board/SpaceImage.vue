<template>
  <v-img
    :src="spacePath(value.kind)"
    class='board-space-image'
  >
  </v-img>
</template>

<script>

  import SpacePaths from '@/components/mixins/SpacePaths'

  export default {
    name: 'sn-space-image',
    mixins: [ SpacePaths ],
    props: ['value']
  }
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style scoped lang='scss'>
    .board-space-image {
      border-radius:10%;
      height:82px;
      width:82px;
      position:absolute;
      top:4px;
      left:4px;
    }
</style>
