<template>
  <v-app-bar
    absolute
    clipped-left
    clipped-right
    color='green'
    dark
    app
  >
    <div style='height:3em' v-if='cuLoading'>
      &nbsp;
    </div>

    <div style='height:3em' v-else>
      <div v-if='cu' class='font-weight-bold title'>
        <sn-user-btn size='small' :user='cu' ></sn-user-btn>
        {{cu.name}}
      </div>
      <div v-else>
              <v-btn :href='loginPath' color='info'>Login</v-btn>
      </div>
    </div>

    <v-spacer></v-spacer>

    <v-card class='mt-11' :to="{ name: 'home' }" color='white' height='100' width='220'>
      <v-img height='100' contain :src="require('../assets/slothninja_logo_fullsize.png')" />
    </v-card>

    <template v-slot:extension>
      <v-flex xs1>
        <v-tooltip bottom color='info' dark>
          <template v-slot:activator='{ on }'>
            <v-app-bar-nav-icon v-on='on' @click.stop="$emit('input', !value)" ></v-app-bar-nav-icon>
          </template>
          <span>Menu</span>
        </v-tooltip>
      </v-flex>
      <v-flex>
        <slot></slot>
      </v-flex>
    </template>

  </v-app-bar>
</template>


<script>
  import UserButton from '@/components/user/Button'
  import CurrentUser from '@/components/mixins/CurrentUser'

  export default {
    mixins: [ CurrentUser ],
    name: 'sn-toolbar',
    components: {
      'sn-user-btn': UserButton
    },
    props: [ 'value' ],
    computed: {
      loginPath: function () {
        return "/login"
        // return `http://lwww.slothninja.com:8083/user/login/?redirect=${window.btoa(window.location.href)}`
      }
    }
  }
</script>
