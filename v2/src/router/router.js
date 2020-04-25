import Vue from 'vue'
import Router from 'vue-router'
import Home from '@/components/home/Home'
import New from '@/components/game/New'
import Index from '@/components/game/Index'
import Game from '@/components/game/Game'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/game/new',
      name: 'new',
      component: New
    },
    {
      path: '/games/:status',
      name: 'index',
      component: Index
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
    }
  ]
})
