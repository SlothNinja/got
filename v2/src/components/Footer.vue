<template>
  <v-footer app color="green" dark padless absolute >
    <v-col
      class="text-center text-caption"
      cols="12"
    >
    © 2010-{{ new Date().getFullYear() }} — <strong>SlothNinja</strong>. All logos and trademarks are property of their respective owner.
    </v-col>
  </v-footer>
</template>

<script>
  export default {
    name: 'sn-footer'
  }
</script>
