<template>
  <div>
    <v-tooltip 
      :disabled="tooltip(kind) === '' || !show"
      left
      max-width='150px'
      open-delay='400'
    >
      <template v-slot:activator="{ on }">
        <sn-card-image v-on="on" :kind='kind' :show='show' ></sn-card-image>
      </template>
      <span>{{tooltip(kind)}}</span>
    </v-tooltip>
  </div>
</template>

<script>
  import Image from '@/components/card/Image'
  import Tooltip from '@/components/mixins/Tooltip'

  export default {
    mixins: [ Tooltip ],
    name: 'sn-card',
    props: [ 'kind', 'show' ],
    components: {
      'sn-card-image': Image
    }
  }
</script>
