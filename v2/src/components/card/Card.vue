<template>
  <div>
    <v-tooltip 
      :disabled="tooltip(kind) === '' || !show"
      left
      max-width='150px'
      open-delay='400'
    >
      <template v-slot:activator="{ on }">
        <v-img
          :src="cardPath(show ? kind : 'cardback')"
          v-on='on'
          :width='width'
          style='border-radius:5%'
          >
        </v-img>
      </template>
      <span>{{tooltip(kind)}}</span>
    </v-tooltip>
  </div>
</template>

<script>
  import Tooltip from '@/components/mixins/Tooltip'
  import CardPaths from '@/components/mixins/CardPaths'

  export default {
    mixins: [ CardPaths, Tooltip ],
    name: 'sn-card',
    props: [ 'kind', 'show', 'width' ]
  }
</script>
