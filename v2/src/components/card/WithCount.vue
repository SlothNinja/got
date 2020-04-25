<template>
  <v-badge color='info' dark overlap bottom >
    <span slot="badge">{{count}}</span>
    <sn-card :show='true' :kind='kind'></sn-card>
  </v-badge>
</template>

<script>
  import Card from '@/components/card/Card'

  export default {
    name: 'sn-card-with-count',
    components: {
      'sn-card': Card
    },
    props: [ 'kind', 'count' ]
  }
</script>
