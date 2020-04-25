<template>
  <v-navigation-drawer
    clipped
    width='200'
    v-model='drawer'
    light
    app
  >
    <v-list-item>
      <v-list-item-content>
        <v-list-item-title class='title font-weight-black text-center'>
          SlothNinja Games
        </v-list-item-title>
        <v-list-item-subtitle class='subtitle-1 font-weight-bold text-center'>
          Guild of Thieves
        </v-list-item-subtitle>
      </v-list-item-content>
    </v-list-item>

    <v-divider></v-divider>

    <v-list
      dense
      nav
    >
      <v-list-item :to="{ name: 'home' }" exact>
        <v-list-item-icon>
          <v-icon>home</v-icon>
        </v-list-item-icon>
        <v-list-item-content>
          <v-list-item-title>Home</v-list-item-title>
        </v-list-item-content>
      </v-list-item>
      <template v-if='cu'>
        <v-list-item :to="{ name: 'new' }">
          <v-list-item-icon>
            <v-icon>create</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>Create</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item :to="{ name: 'index', params: {status: 'recruiting'} }">
          <v-list-item-icon>
            <v-icon>playlist_add</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>Join</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item :to="{ name: 'index', params: {status: 'running' } }">
          <v-list-item-icon>
            <v-icon>playlist_play</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>Play</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click='logout'>
          <v-list-item-icon>
            <v-icon>exit_to_app</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>Logout</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </template>
    </v-list>
  </v-navigation-drawer>
</template>

<script>
  import CurrentUser from '@/components/mixins/CurrentUser'

  export default {
    mixins: [ CurrentUser ],
    name: 'nav-drawer',
    props: [ 'value' ],
    methods: {
      logout: function () {
        var self = this
        self.delete_cookie('sngsession')
        self.cu = null
        if (self.$route.name != 'home') {
          self.$router.push({ name: 'home'})
        }
      },
      delete_cookie: function (name) {
        document.cookie = name + '= ; domain = .slothninja.com ; expires = Thu, 01 Jan 1970 00:00:00 GMT'
      },
    },
    computed: {
      drawer: {
        get: function () {
          var self = this
          return self.value
        },
        set: function (value) {
          var self = this
          self.$emit('input', value)
        }
      }
    }
  }
</script>
