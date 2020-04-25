<template>
  <v-layout row>
      <div v-if='!signedIn'>
        <div id='sn-signin2'></div>
      </div>
      <div v-else>
        <v-tooltip bottom color='green'>
          <v-btn
            slot='activator'
            icon
            @click.native='signOut'
          >
            <v-icon>power_settings_new</v-icon>
          </v-btn>
          <span>Sign Out</span>
        </v-tooltip>
      </div>
    <v-flex>
      <div v-if="createAccount">
        <v-btn
          color='green'
          @click="$router.push({ name: 'newuser'})"
        >
          Create Account
        </v-btn>
      </div>
        <div v-if="haveCU" >
          <sn-user-btn size="small" :user="cu" ></sn-user-btn>
          <span class="body-2">{{cu.name}}</span>
        </div>
    </v-flex>
  </v-layout>
</template>

<script>
  import VS2 from 'vue-script2'
  import UserButton from '@/components/user/Button'

  var _ = require('lodash')

  export default {
    data () {
      return {
        gauth: {},
        profile: {}
      }
    },
    name: 'sn-signin2',
    props: [ 'clientID' ],
    components: {
      'sn-user-btn': UserButton
    },
    mounted: function () {
      console.log('mounted')
      this.renderLogin()
    },
    computed: {
      signedIn: function () {
        return this.gauth.isSignedIn ? this.gauth.isSignedIn.get() : false
      },
      createAccount: function () {
        return this.signedIn && !(this.cu) && (this.$route.name !== 'newuser')
      },
      idToken: {
        get: function () {
          return this.$root.idToken
        },
        set: function (value) {
          this.$root.idToken = value
        }
      },
      snackbar: {
        get: function () {
          return this.$root.snackbar
        },
        set: function (value) {
          this.$root.snackbar = value
        }
      },
      cu: {
        get: function () {
          return this.$root.cu
        },
        set: function (value) {
          this.$root.cu = value
        }
      },
      haveCU: function () {
        return _.get(this.cu, 'name', '') !== ''
      }
    },
    methods: {
      renderLogin () {
        var self = this
        VS2.load('https://apis.google.com/js/platform.js').then(function () {
          window.gapi.load('auth2', function () {
            window.gapi.auth2.init({client_id: self.clientID}).then(self.gauthSuccess, self.gauthError)
          })
        })
      },
      gauthSuccess (gauth) {
        this.gauth = gauth
        window.gapi.signin2.render('sn-signin2',
          {
            scope: 'email',
            width: 200,
            height: 50,
            longtitle: false,
            theme: 'dark',
            onsuccess: this.onSignInSuccess
          }
        )
        if (!this.signedIn) {
          this.idToken = ''
        }
      },
      gauthError (obj) {
        console.log('gauthError', obj)
      },
      onSignInSuccess (googleUser) {
        // `googleUser` is the GoogleUser object that represents the just-signed-in user.
        // See https://developers.google.com/identity/sign-in/web/reference#users
        this.profile = googleUser.getBasicProfile() // etc etc
        console.log('ID: ' + this.profile.getId())
        console.log('Full Name: ' + this.profile.getName())
        console.log('Given Name: ' + this.profile.getGivenName())
        console.log('Family Name: ' + this.profile.getFamilyName())
        console.log('Image URL: ' + this.profile.getImageUrl())
        console.log('Email: ' + this.profile.getEmail())
        this.idToken = googleUser.getAuthResponse().id_token
        console.log('Token: ' + this.idToken)
        this.signIn({token: this.idToken})
      },
      onSignInError (error) {
        // `error` contains any error occurred.
        console.log('onSignInError', error)
      },
      signOut () {
        var self = this
        self.gauth.signOut().then(function () {
          self.gauth = {}
          if (self.$route.name === 'login') {
            self.renderLogin()
          } else {
            self.$router.push({ name: 'login' })
          }
          console.log('User signed out.')
        })
      },
      signIn: function (data) {
        var self = this
        $.ajax({
          url: 'http://dev.slothninja.com:8080/user/signin',
          dataType: 'json',
          data: JSON.stringify(data),
          contentType: 'application/json; charset=utf-8',
          type: 'PUT',
          success: function (data) {
            console.log('data: ' + JSON.stringify(data))
            if (data.error) {
              self.snackbar = { open: true, message: data.message }
            }
            self.cu = data.cu
          }
        })
      }
    }
  }
</script>
