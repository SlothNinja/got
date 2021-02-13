<template>
  <v-card>
    <v-container>
      <v-row>
        <v-col align='center' cols='12'>
          <slot></slot>
        </v-col>
      </v-row>
    </v-container>
  </v-card>
</template>

<script>
  export default {
    name: 'sn-messagebar'
  }
</script>
