<template>
  <v-card class="text-center" >
    <v-row>
      <v-col cols='12'>
        <slot></slot>
      </v-col>
    </v-row>
  </v-card>
</template>

<script>
  export default {
    name: 'sn-messagebar'
  }
</script>
