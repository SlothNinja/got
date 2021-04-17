<template>
  <v-navigation-drawer
    overflow
    clipped
    v-model='nav'
    light
    app
  >
  <v-list-item v-if='cu'>
      <v-list-item-icon>
        <sn-user-btn size='x-small' :user='cu' ></sn-user-btn>
      </v-list-item-icon>
      <v-list-item-content>
        <v-list-item-title>
          {{cu.name}}
        </v-list-item-title>
      </v-list-item-content>
    </v-list-item>

    <v-list-item v-else :to="{ name: 'login'}" >
      <v-list-item-icon>
        <v-icon>mdi-login</v-icon>
      </v-list-item-icon>
      <v-list-item-content>
        <v-list-item-title>Login</v-list-item-title>
      </v-list-item-content>
    </v-list-item>

    <v-divider></v-divider>

    <v-list
      dense
      nav
    >
      <v-list-item :to="{ name: 'sng-home' }" exact>
        <v-list-item-icon>
          <v-icon>mdi-home</v-icon>
        </v-list-item-icon>
        <v-list-item-content>
          <v-list-item-title>Home</v-list-item-title>
        </v-list-item-content>
      </v-list-item>
      <v-list-item :to="{ name: 'sng-ugames', params: { status: 'running', type: 'all', uid: cuid} }" exact>
        <v-list-item-icon>
          <v-icon>mdi-account-details</v-icon>
        </v-list-item-icon>
        <v-list-item-content>
          <v-list-item-title>Your Games</v-list-item-title>
        </v-list-item-content>
      </v-list-item>
      <template v-if='cu'>
        <v-list-group
          no-action
          prepend-icon='mdi-pencil'
          >
          <template v-slot:activator>
            <v-list-item-title>Create</v-list-item-title>
          </template>
          <v-list-item v-for="(item, index) in items" :key='index' :to="item.createlink" >
            <v-list-item-title>{{item.title}}</v-list-item-title>
          </v-list-item>
        </v-list-group>
        <v-list-group
          no-action
          prepend-icon='mdi-plus'
          >
          <template v-slot:activator>
            <v-list-item-title>Join</v-list-item-title>
          </template>
          <v-list-item v-for="(item, index) in items" :key='index' :to="item.joinlink" >
            <v-list-item-title>{{item.title}}</v-list-item-title>
          </v-list-item>
        </v-list-group>
        <v-list-group
          no-action
          prepend-icon='mdi-play'
          >
          <template v-slot:activator>
            <v-list-item-title>Play</v-list-item-title>
          </template>
          <v-list-item v-for="(item, index) in items" :key='index' :to="item.playlink" >
            <v-list-item-title>{{item.title}}</v-list-item-title>
          </v-list-item>
        </v-list-group>
        <v-list-group
          no-action
          prepend-icon='mdi-check'
          >
          <template v-slot:activator>
            <v-list-item-title>Completed</v-list-item-title>
          </template>
          <v-list-item v-for="(item, index) in items" :key='index' :to="item.completedlink" >
            <v-list-item-title>{{item.title}}</v-list-item-title>
          </v-list-item>
        </v-list-group>
        <v-list-group
          no-action
          prepend-icon='mdi-star'
          >
          <template v-slot:activator>
            <v-list-item-title>Top Players</v-list-item-title>
          </template>
          <v-list-item v-for="(item, index) in items" :key='index' :to="item.ratingslink" >
            <v-list-item-title>{{item.title}}</v-list-item-title>
          </v-list-item>
        </v-list-group>
        <v-list-item :to="{ name: 'logout'}" >
          <v-list-item-icon>
            <v-icon>mdi-logout</v-icon>
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
import UserButton from '@/components/lib/user/Button'
import CurrentUser from '@/components/lib/mixins/CurrentUser'

const _ = require('lodash')

export default {
  mixins: [ CurrentUser ],
  name: 'nav-drawer',
  props: [ 'value' ],
  components: {
    'sn-user-btn': UserButton
  },
  created: function () {
    _.forEach(this.routes, route => this.$router.addRoute(route))
  },
  computed: {
    sngGames: () => [ 'atf', 'confucius', 'indonesia', 'tammany', 'all' ],
    routes: function () {
      return [
        {
          path: '/sng-games/:type/:status',
          name: 'sng-games',
          beforeEnter: (to, from, next) => {
            if (_.includes(this.sngGames, to.params.type)) {
              let sngHome = process.env.VUE_APP_SNG_HOME
              window.location.replace(`${sngHome}#/games/${to.params.status}/${to.params.type}`)
              next()
            } else {
              let game = process.env.VUE_APP_GAME
              if (to.params.type == game) {
                next({ name: 'games', params: { status: to.params.status } })
              } else {
                let gotHome = process.env.VUE_APP_GOT_HOME
                window.location.replace(`${gotHome}#/games/${to.params.status}`)
                next()
              }
            }
          }
        },
        {
          path: '/sng-ugames/:uid/:status/:type',
          name: 'sng-ugames',
          beforeEnter: (to, from, next) => {
            let sngHome = process.env.VUE_APP_SNG_HOME
            window.location.replace(`${sngHome}#/ugames/${to.params.uid}/${to.params.status}/${to.params.type}`)
            next()
          }
        },
        {
          path: '/sng-join-game/:type',
          name: 'sng-join-game',
          beforeEnter: (to, from, next) => {
            if (_.includes(this.sngGames, to.params.type)) {
              let sngHome = process.env.VUE_APP_SNG_HOME
              window.location.replace(`${sngHome}${to.params.type}/games/recruiting`)
              next()
            } else {
              let game = process.env.VUE_APP_GAME
              if (to.params.type == game) {
                next({ name: 'invitations'})
              } else {
                let gotHome = process.env.VUE_APP_GOT_HOME
                window.location.replace(`${gotHome}#/invitations`)
                next()
              }
            }
          }
        },
        {
          path: '/sng-new-game/:type',
          name: 'sng-new-game',
          beforeEnter: (to, from, next) => {
            if (_.includes(this.sngGames, to.params.type)) {
              let sngHome = process.env.VUE_APP_SNG_HOME
              window.location.replace(`${sngHome}${to.params.type}/game/new`)
              next()
            } else {
              let game = process.env.VUE_APP_GAME
              if (to.params.type == game) {
                next({ name: 'new'})
              } else {
                let gotHome = process.env.VUE_APP_GOT_HOME
                window.location.replace(`${gotHome}#/invitation/new`)
                next()
              }
            }
          }
        },
        {
          path: '/sng-ratings/:type',
          name: 'sng-ratings',
          beforeEnter: (to, from, next) => {
            if (_.includes(this.sngGames, to.params.type)) {
              let sngHome = process.env.VUE_APP_SNG_HOME
              window.location.replace(`${sngHome}ratings/show/${to.params.type}`)
              next()
            } else {
              let game = process.env.VUE_APP_GAME
              if (to.params.type == game) {
                next({ name: 'rank'})
              } else {
                let gotHome = process.env.VUE_APP_GOT_HOME
                window.location.replace(`${gotHome}#/rank`)
                next()
              }
            }
          }
        },
        {
          path: '/sng-home',
          name: 'sng-home',
          beforeEnter: (to, from, next) => {
            let sngHome = process.env.VUE_APP_SNG_HOME
            window.location.replace(sngHome)
            next()
          }
        },
        {
          path: '/logout',
          name: 'logout',
          beforeEnter() {
            window.location.replace('/logout')
          }
        },
        {
          path: '/login',
          name: 'login',
          beforeEnter() {
            window.location.replace('/login')
          }
        },
      ]
    },
    types: function () {
      return [
        { type: 'atf', title: 'After the Flood' },
        { type: 'confucius', title: 'Confucius' },
        { type: 'got', title: 'Guild of Thieves' },
        { type: 'indonesia', title: 'Indonesia' },
        { type: 'tammany', title: 'Tammany Hall' }
      ]
    },
    items: function () {
      return _.map(this.types, game => {
        return { 
          createlink: { name: 'sng-new-game', params: { type: game.type } },
          joinlink: { name: 'sng-join-game', params: { type: game.type, status: 'recruiting' } },
          playlink: { name: 'sng-games', params: { type: game.type, status: 'running' } },
          completedlink: { name: 'sng-games', params: { type: game.type, status: 'completed' } },
          ratingslink: { name: 'sng-ratings', params: { type: game.type } },
          title: game.title
        }
      })
    },
    nav: {
      get: function () {
        return this.value
      },
      set: function (value) {
        this.$emit('input', value)
      }
    }
  }
}
</script>
