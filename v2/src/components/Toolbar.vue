<template>
  <v-app-bar
    absolute
    clipped-left
    clipped-right
    color='green'
    dark
    app
  >
  <v-row no-gutters>
    <v-col cols='1'>
      <v-tooltip bottom color='info' dark>
        <template v-slot:activator='{ on }'>
          <v-app-bar-nav-icon v-on='on' @click.stop="$emit('input', !value)" ></v-app-bar-nav-icon>
        </template>
        <span>Menu</span>
      </v-tooltip>
    </v-col>
    <v-col>
        <slot></slot>
      </v-col>
      <v-col cols='1' align='right'>
        <v-card :to="{ name: 'sng-home' }" color='white' width='100' >
          <v-img width='100' :src="require('../assets/slothninja_logo_fullsize.png')" />
        </v-card>
      </v-col>
    </v-row>
  </v-app-bar>
</template>


<script>
  import CurrentUser from '@/components/mixins/CurrentUser'

  export default {
    mixins: [ CurrentUser ],
    name: 'sn-toolbar',
    props: [ 'value' ]
  }
</script>
