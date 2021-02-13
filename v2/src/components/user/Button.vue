<template>
  <v-btn 
    fab
    :x-small="size ? size === 'x-small' : true"
    :small="size ? size === 'small' : true"
    :medium="size ? size === 'medium' : true"
    :large="size ? size === 'large' : true"
    :color="color || 'black' "
    :href="showlink"
  >
    <v-avatar :size="avatarSize" >
      <img :src="gravatar(user.emailHash, size, user.gravType)" />
    </v-avatar>
  </v-btn>
</template>

<script>
  import Gravatar from '@/components/mixins/Gravatar'

  export default {
    mixins: [ Gravatar ],
    name: 'sn-user-btn',
    props: [ 'color', 'user', 'size' ],
    computed: {
      avatarSize: function () {
        var self = this
        switch (self.size) {
          case 'x-small':
            return '24px'
          case 'small':
            return '30px'
          case 'medium':
            return '48px'
          default:
            return '54px'
        }
      },
      showlink: function () {
        var self = this
        let userv = process.env.VUE_APP_USER_DOMAIN
        return `${userv}#/show/${self.user.id}`
      }
    }
  }
</script>
