import Vue from 'vue'
import Router from 'vue-router'
import Home from '@/components/home/Home'
import New from '@/components/invitation/New'
import Invitations from '@/components/invitation/Index'
import Games from '@/components/game/Index'
import Rank from '@/components/rank/Index'
import Game from '@/components/game/Game'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/invitation/new',
      name: 'new',
      component: New
    },
    {
      path: '/invitations',
      name: 'invitations',
      component: Invitations
    },
    {
      path: '/games/:status',
      name: 'games',
      component: Games
    },
    {
      path: '/game/:id',
      name: 'game',
      component: Game
    },
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/rank',
      name: 'rank',
      component: Rank
    },
    {
      path: '/logout',
      name: 'logout',
      beforeEnter() {
        window.location.href = '/logout'
      }
    },
    {
      path: '/sng-home',
      name: 'sng-home',
      beforeEnter() {
        let sngHome = process.env.VUE_APP_SNG_HOME
        window.location.href = sngHome
      }
    },
    {
      path: '/sng-games/:type/:status',
      name: 'sng-games',
      beforeEnter(to) {
        let sngHome = process.env.VUE_APP_SNG_HOME
        window.location.href = `${sngHome}${to.params.type}/games/${to.params.status}`
      }
    },
    {
      path: '/sng-new-game/:type',
      name: 'sng-new-game',
      beforeEnter(to) {
        let sngHome = process.env.VUE_APP_SNG_HOME
        window.location.href = `${sngHome}${to.params.type}/game/new`
      }
    }
  ]
})
