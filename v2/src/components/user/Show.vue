<template>
  <div>
        <v-layout row wrap>
          <v-flex xs6>
            <v-card height="31em">
              <v-card-text v-if="loading" class="text-xs-center">
                <v-flex class="pt-5">
                  <v-progress-circular
                    indeterminate
                    color="green"
                    size="128"
                    width="10"
                  >Loading...</v-progress-circular>
                </v-flex>
              </v-card-text>
              <template v-else>
                <v-card-title primary-title>
                  <div class="font-weight-bold title">
                    <sn-user-btn size="medium" :user="u" ></sn-user-btn>
                    {{u.name}}
                  </div>
                </v-card-title>
                <v-card-text>
                  <v-text-field
                    name="name"
                    label="Screen Name"
                    v-model="u.name"
                    id="name"
                    readonly
                  >
                  </v-text-field>
                  <v-text-field
                    name="email"
                    label="Email"
                    v-model="u.email"
                    id="email"
                    readonly
                  >
                  </v-text-field>
                  <v-layout row>
                    <v-flex>
                      <v-checkbox
                        v-model="u.emailReminders"
                        :label="`Email Reminders: ${u.emailReminders.toString()}`"
                        readonly
                        color="green"
                      ></v-checkbox>
                    </v-flex>
                    <v-flex>
                      <v-checkbox
                        v-model="u.emailNotifications"
                        :label="`Email Notifications: ${u.emailNotifications.toString()}`"
                        color="green"
                        readonly
                      ></v-checkbox>
                    </v-flex>
                  </v-layout>
                  <v-btn color="green" dark :to="{ name: 'edit', params: { id: $route.params.id }}">Edit</v-btn>
                </v-card-text>
              </template>
            </v-card>
          </v-flex>
          <v-flex xs6>
            <v-card height="31em">
              <v-card-text>
                <div>
                  <h1>Welcome to SlothNinja Games</h1>
                  <p>
                    SlothNinja Games is a play-by-web (PBW) site that permits registered members to play boardgames
                    with other members via the Internet, in a turn-based manner. Registration is required in order
                    to play. Registration currently requires a Google Account, but is free.
                  </p>
                  <p>
                    Please specify the screen name for you account.  If you have multiple Google Accounts, please
                    verify that the displayed email address is for the Google Account that you wish to register.  If the
                    wrong Google Account, select the login button in the toolbar and login in with the correct
                    Google Account.
                  </p>
                  <p>
                    Select whether the system should send an email notification when it's your turn.  Regardless
                    of selection, the site will send a daily email reminder if any games are waiting for you to
                    make a move.
                  </p>
                  <p>
                    Finally, you can also select your avatar from default avatars.  Avatars are provided by
                    gravatar.com.  See, gravatar.com to personalize your avatar.
                  </p>
                </div>
              </v-card-text>
            </v-card>
          </v-flex>
        </v-layout>
  </div>
</template>

<script>
  import UserButton from '@/components/user/Button'

  const axios = require('axios')

  export default {
    data () {
      return {
        u: { name: '', emailReminders: 'true', emailNotifications: 'true', gravType: 'monsterid' },
        auto: true,
        nav: false,
        idToken: '',
        loading: true
      }
    },
    components: {
      'sn-user-btn': UserButton
    },
    computed: {
      snackbar: {
        get: function () {
          return this.$root.snackbar
        },
        set: function (value) {
          this.$root.snackbar = value
        }
      }
    },
    created () {
      var self = this
      self.fetchData()
    },
    methods: {
      fetchData () {
        var self = this
        var path = `/json/${self.$route.params.id}`
        axios.get(path)
          .then(function (response) {
            self.loading = false
            self.u = response.data.u
          })
          .catch(function () {
            self.loading = false
            self.snackbar.message = 'Server Error.  Try again.'
            self.snackbar.open = true
            self.$router.push({ name: 'show', params: { id: self.$route.params.id}})
          })
      }
    }
  }
</script>
